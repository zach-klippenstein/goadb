package goadb

import (
	"fmt"
	"io"
	"strings"

	"github.com/zach-klippenstein/goadb/wire"
)

// DeviceClient communicates with a specific Android device.
type DeviceClient struct {
	dialer     wire.Dialer
	descriptor *DeviceDescriptor
}

func (c *DeviceClient) String() string {
	return c.descriptor.String()
}

// get-product is documented, but not implemented in the server.
// TODO(z): Make getProduct exported if get-product is ever implemented in adb.
func (c *DeviceClient) getProduct() (string, error) {
	return c.getAttribute("get-product")
}

func (c *DeviceClient) GetSerial() (string, error) {
	return c.getAttribute("get-serialno")
}

func (c *DeviceClient) GetDevicePath() (string, error) {
	return c.getAttribute("get-devpath")
}

func (c *DeviceClient) GetState() (string, error) {
	return c.getAttribute("get-state")
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
func (c *DeviceClient) RunCommand(cmd string, args ...string) (string, error) {
	cmd, err := prepareCommandLine(cmd, args...)
	if err != nil {
		return "", err
	}

	conn, err := c.dialDevice()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	req := fmt.Sprintf("shell:%s", cmd)

	// Shell responses are special, they don't include a length header.
	// We read until the stream is closed.
	// So, we can't use conn.RoundTripSingleResponse.
	if err = conn.SendMessage([]byte(req)); err != nil {
		return "", err
	}
	if err = wire.ReadStatusFailureAsError(conn, req); err != nil {
		return "", err
	}

	resp, err := conn.ReadUntilEof()
	if err != nil {
		return "", err
	}

	return string(resp), nil
}

/*
Remount, from the docs,
	Ask adbd to remount the device's filesystem in read-write mode,
	instead of read-only. This is usually necessary before performing
	an "adb sync" or "adb push" request.
	This request may not succeed on certain builds which do not allow
	that.
Source: https://android.googlesource.com/platform/system/core/+/master/adb/SERVICES.TXT
*/
func (c *DeviceClient) Remount() (string, error) {
	conn, err := c.dialDevice()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	resp, err := conn.RoundTripSingleResponse([]byte("remount"))
	return string(resp), err
}

func (c *DeviceClient) ListDirEntries(path string) (*DirEntries, error) {
	conn, err := c.getSyncConn()
	if err != nil {
		return nil, err
	}

	return listDirEntries(conn, path)
}

func (c *DeviceClient) Stat(path string) (*DirEntry, error) {
	conn, err := c.getSyncConn()
	if err != nil {
		return nil, err
	}

	return stat(conn, path)
}

func (c *DeviceClient) OpenRead(path string) (io.ReadCloser, error) {
	conn, err := c.getSyncConn()
	if err != nil {
		return nil, err
	}

	return receiveFile(conn, path)
}

// getAttribute returns the first message returned by the server by running
// <host-prefix>:<attr>, where host-prefix is determined from the DeviceDescriptor.
func (c *DeviceClient) getAttribute(attr string) (string, error) {
	resp, err := wire.RoundTripSingleResponse(c.dialer,
		fmt.Sprintf("%s:%s", c.descriptor.getHostPrefix(), attr))
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

// dialDevice switches the connection to communicate directly with the device
// by requesting the transport defined by the DeviceDescriptor.
func (c *DeviceClient) dialDevice() (*wire.Conn, error) {
	conn, err := c.dialer.Dial()
	if err != nil {
		return nil, fmt.Errorf("error dialing adb server: %+v", err)
	}

	req := fmt.Sprintf("host:%s", c.descriptor.getTransportDescriptor())
	if err = wire.SendMessageString(conn, req); err != nil {
		conn.Close()
		return nil, fmt.Errorf("error connecting to device '%s': %+v", c.descriptor, err)
	}

	if err = wire.ReadStatusFailureAsError(conn, req); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

func (c *DeviceClient) getSyncConn() (*wire.SyncConn, error) {
	conn, err := c.dialDevice()
	if err != nil {
		return nil, fmt.Errorf("error connecting to device for sync: %+v", err)
	}

	// Switch the connection to sync mode.
	if err := wire.SendMessageString(conn, "sync:"); err != nil {
		return nil, fmt.Errorf("error requesting sync mode: %+v", err)
	}
	if err := wire.ReadStatusFailureAsError(conn, "sync"); err != nil {
		return nil, err
	}

	return conn.NewSyncConn(), nil
}

// prepareCommandLine validates the command and argument strings, quotes
// arguments if required, and joins them into a valid adb command string.
func prepareCommandLine(cmd string, args ...string) (string, error) {
	if isBlank(cmd) {
		return "", fmt.Errorf("command cannot be empty")
	}

	for i, arg := range args {
		if strings.ContainsRune(arg, '"') {
			return "", fmt.Errorf("arg at index %d contains an invalid double quote: %s", i, arg)
		}
		if containsWhitespace(arg) {
			args[i] = fmt.Sprintf("\"%s\"", arg)
		}
	}

	// Prepend the comand to the args array.
	if len(args) > 0 {
		cmd = fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
	}

	return cmd, nil
}
