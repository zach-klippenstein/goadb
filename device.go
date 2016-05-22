package adb

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/zach-klippenstein/goadb/internal/errors"
	"github.com/zach-klippenstein/goadb/wire"
)

// MtimeOfClose should be passed to OpenWrite to set the file modification time to the time the Close
// method is called.
var MtimeOfClose = time.Time{}

// Device communicates with a specific Android device.
// To get an instance, call Device() on an Adb.
type Device struct {
	server     server
	descriptor DeviceDescriptor

	// Used to get device info.
	deviceListFunc func() ([]*DeviceInfo, error)
}

func (c *Device) String() string {
	return c.descriptor.String()
}

// get-product is documented, but not implemented, in the server.
// TODO(z): Make product exported if get-product is ever implemented in adb.
func (c *Device) product() (string, error) {
	attr, err := c.getAttribute("get-product")
	return attr, wrapClientError(err, c, "Product")
}

func (c *Device) Serial() (string, error) {
	attr, err := c.getAttribute("get-serialno")
	return attr, wrapClientError(err, c, "Serial")
}

func (c *Device) DevicePath() (string, error) {
	attr, err := c.getAttribute("get-devpath")
	return attr, wrapClientError(err, c, "DevicePath")
}

func (c *Device) State() (DeviceState, error) {
	attr, err := c.getAttribute("get-state")
	state, err := parseDeviceState(attr)
	return state, wrapClientError(err, c, "State")
}

func (c *Device) DeviceInfo() (*DeviceInfo, error) {
	// Adb doesn't actually provide a way to get this for an individual device,
	// so we have to just list devices and find ourselves.

	serial, err := c.Serial()
	if err != nil {
		return nil, wrapClientError(err, c, "GetDeviceInfo(GetSerial)")
	}

	devices, err := c.deviceListFunc()
	if err != nil {
		return nil, wrapClientError(err, c, "DeviceInfo(ListDevices)")
	}

	for _, deviceInfo := range devices {
		if deviceInfo.Serial == serial {
			return deviceInfo, nil
		}
	}

	err = errors.Errorf(errors.DeviceNotFound, "device list doesn't contain serial %s", serial)
	return nil, wrapClientError(err, c, "DeviceInfo")
}

/*
RunCommand runs the specified commands on a shell on the device.

From the Android docs:
	Run 'command arg1 arg2 ...' in a shell on the device, and return
	its output and error streams. Note that arguments must be separated
	by spaces. If an argument contains a space, it must be quoted with
	double-quotes. Arguments cannot contain double quotes or things
	will go very wrong.

	Note that this is the non-interactive version of "adb shell"
Source: https://android.googlesource.com/platform/system/core/+/master/adb/SERVICES.TXT

This method quotes the arguments for you, and will return an error if any of them
contain double quotes.
*/
func (c *Device) RunCommand(cmd string, args ...string) (string, error) {
	cmd, err := prepareCommandLine(cmd, args...)
	if err != nil {
		return "", wrapClientError(err, c, "RunCommand")
	}

	conn, err := c.dialDevice()
	if err != nil {
		return "", wrapClientError(err, c, "RunCommand")
	}
	defer conn.Close()

	req := fmt.Sprintf("shell:%s", cmd)

	// Shell responses are special, they don't include a length header.
	// We read until the stream is closed.
	// So, we can't use conn.RoundTripSingleResponse.
	if err = conn.SendMessage([]byte(req)); err != nil {
		return "", wrapClientError(err, c, "RunCommand")
	}
	if _, err = conn.ReadStatus(req); err != nil {
		return "", wrapClientError(err, c, "RunCommand")
	}

	resp, err := conn.ReadUntilEof()
	return string(resp), wrapClientError(err, c, "RunCommand")
}

/*
Remount, from the official adb command’s docs:
	Ask adbd to remount the device's filesystem in read-write mode,
	instead of read-only. This is usually necessary before performing
	an "adb sync" or "adb push" request.
	This request may not succeed on certain builds which do not allow
	that.
Source: https://android.googlesource.com/platform/system/core/+/master/adb/SERVICES.TXT
*/
func (c *Device) Remount() (string, error) {
	conn, err := c.dialDevice()
	if err != nil {
		return "", wrapClientError(err, c, "Remount")
	}
	defer conn.Close()

	resp, err := conn.RoundTripSingleResponse([]byte("remount"))
	return string(resp), wrapClientError(err, c, "Remount")
}

func (c *Device) ListDirEntries(path string) (*DirEntries, error) {
	conn, err := c.getSyncConn()
	if err != nil {
		return nil, wrapClientError(err, c, "ListDirEntries(%s)", path)
	}

	entries, err := listDirEntries(conn, path)
	return entries, wrapClientError(err, c, "ListDirEntries(%s)", path)
}

func (c *Device) Stat(path string) (*DirEntry, error) {
	conn, err := c.getSyncConn()
	if err != nil {
		return nil, wrapClientError(err, c, "Stat(%s)", path)
	}
	defer conn.Close()

	entry, err := stat(conn, path)
	return entry, wrapClientError(err, c, "Stat(%s)", path)
}

func (c *Device) OpenRead(path string) (io.ReadCloser, error) {
	conn, err := c.getSyncConn()
	if err != nil {
		return nil, wrapClientError(err, c, "OpenRead(%s)", path)
	}

	reader, err := receiveFile(conn, path)
	return reader, wrapClientError(err, c, "OpenRead(%s)", path)
}

// OpenWrite opens the file at path on the device, creating it with the permissions specified
// by perms if necessary, and returns a writer that writes to the file.
// The files modification time will be set to mtime when the WriterCloser is closed. The zero value
// is TimeOfClose, which will use the time the Close method is called as the modification time.
func (c *Device) OpenWrite(path string, perms os.FileMode, mtime time.Time) (io.WriteCloser, error) {
	conn, err := c.getSyncConn()
	if err != nil {
		return nil, wrapClientError(err, c, "OpenWrite(%s)", path)
	}

	writer, err := sendFile(conn, path, perms, mtime)
	return writer, wrapClientError(err, c, "OpenWrite(%s)", path)
}

// getAttribute returns the first message returned by the server by running
// <host-prefix>:<attr>, where host-prefix is determined from the DeviceDescriptor.
func (c *Device) getAttribute(attr string) (string, error) {
	resp, err := roundTripSingleResponse(c.server,
		fmt.Sprintf("%s:%s", c.descriptor.getHostPrefix(), attr))
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

func (c *Device) getSyncConn() (*wire.SyncConn, error) {
	conn, err := c.dialDevice()
	if err != nil {
		return nil, err
	}

	// Switch the connection to sync mode.
	if err := wire.SendMessageString(conn, "sync:"); err != nil {
		return nil, err
	}
	if _, err := conn.ReadStatus("sync"); err != nil {
		return nil, err
	}

	return conn.NewSyncConn(), nil
}

// dialDevice switches the connection to communicate directly with the device
// by requesting the transport defined by the DeviceDescriptor.
func (c *Device) dialDevice() (*wire.Conn, error) {
	conn, err := c.server.Dial()
	if err != nil {
		return nil, err
	}

	req := fmt.Sprintf("host:%s", c.descriptor.getTransportDescriptor())
	if err = wire.SendMessageString(conn, req); err != nil {
		conn.Close()
		return nil, errors.WrapErrf(err, "error connecting to device '%s'", c.descriptor)
	}

	if _, err = conn.ReadStatus(req); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

// prepareCommandLine validates the command and argument strings, quotes
// arguments if required, and joins them into a valid adb command string.
func prepareCommandLine(cmd string, args ...string) (string, error) {
	if isBlank(cmd) {
		return "", errors.AssertionErrorf("command cannot be empty")
	}

	for i, arg := range args {
		if strings.ContainsRune(arg, '"') {
			return "", errors.Errorf(errors.ParseError, "arg at index %d contains an invalid double quote: %s", i, arg)
		}
		if containsWhitespace(arg) {
			args[i] = fmt.Sprintf("\"%s\"", arg)
		}
	}

	// Prepend the command to the args array.
	if len(args) > 0 {
		cmd = fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
	}

	return cmd, nil
}
