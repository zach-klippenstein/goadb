/*
Package goadb is a Go interface to the Android Debug Bridge (adb).

The client/server spec is defined at https://android.googlesource.com/platform/system/core/+/master/adb/OVERVIEW.TXT.

WARNING This library is under heavy development, and its API is likely to change without notice.
*/
package goadb

import (
	"fmt"
	"os/exec"
	"strconv"

	"github.com/zach-klippenstein/goadb/wire"
)

// Dialer is a function that knows how to create a connection to an adb server.
type Dialer func() (*wire.Conn, error)

/*
HostClient interacts with host services on the adb server.

Eg.
	dialer := &HostClient{wire.Dial}
	dialer.GetServerVersion()

TODO make this a real example.

TODO Finish implementing services.

See list of services at https://android.googlesource.com/platform/system/core/+/master/adb/SERVICES.TXT.
*/
type HostClient struct {
	Dialer
}

// GetServerVersion asks the ADB server for its internal version number.
func (c *HostClient) GetServerVersion() (int, error) {
	resp, err := c.roundTripSingleResponse([]byte("host:version"))
	if err != nil {
		return 0, err
	}

	version, err := strconv.ParseInt(string(resp), 16, 32)
	return int(version), err
}

/*
KillServer tells the server to quit immediately.

Corresponds to the command:
	adb kill-server
*/
func (c *HostClient) KillServer() error {
	conn, err := c.Dialer()
	if err != nil {
		return err
	}
	defer conn.Close()

	if err = conn.SendMessage([]byte("host:kill")); err != nil {
		return err
	}

	return nil
}

/*
StartServer ensures there is a server running.

Currently implemented by just running
	adb start-server
*/
func (c *HostClient) StartServer() error {
	cmd := exec.Command("adb", "start-server")
	return cmd.Run()
}

/*
ListDeviceSerials returns the serial numbers of all attached devices.

Corresponds to the command:
	adb devices
*/
func (c *HostClient) ListDeviceSerials() ([]string, error) {
	resp, err := c.roundTripSingleResponse([]byte("host:devices"))
	if err != nil {
		return nil, err
	}

	devices, err := parseDeviceList(string(resp), parseDeviceShort)
	if err != nil {
		return nil, err
	}

	serials := make([]string, len(devices))
	for i, dev := range devices {
		serials[i] = dev.Serial
	}
	return serials, nil
}

/*
ListDevices returns the list of connected devices.

Corresponds to the command:
	adb devices -l
*/
func (c *HostClient) ListDevices() ([]*Device, error) {
	resp, err := c.roundTripSingleResponse([]byte("host:devices-l"))
	if err != nil {
		return nil, err
	}

	return parseDeviceList(string(resp), parseDeviceLong)
}

func (c *HostClient) roundTripSingleResponse(req []byte) (resp []byte, err error) {
	conn, err := c.Dialer()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err = conn.SendMessage(req); err != nil {
		return nil, err
	}

	err = c.readStatusFailureAsError(conn)
	if err != nil {
		return nil, err
	}

	return conn.ReadMessage()
}

// Reads the status, and if failure, reads the message and returns it as an error.
// If the status is success, doesn't read the message.
func (c *HostClient) readStatusFailureAsError(conn *wire.Conn) error {
	status, err := conn.ReadStatus()
	if err != nil {
		return err
	}

	if !status.IsSuccess() {
		msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		return fmt.Errorf("server error: %s", msg)
	}

	return nil
}
