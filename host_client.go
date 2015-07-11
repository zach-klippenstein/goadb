// TODO(z): Implement TrackDevices.
package goadb

import (
	"strconv"

	"github.com/zach-klippenstein/goadb/wire"
)

/*
HostClient communicates with host services on the adb server.

Eg.
	client := NewHostClient()
	client.StartServer()
	client.ListDevices()
	client.GetAnyDevice() // see DeviceClient

See list of services at https://android.googlesource.com/platform/system/core/+/master/adb/SERVICES.TXT.
*/
// TODO(z): Finish implementing host services.
type HostClient struct {
	config ClientConfig
}

// func NewHostClient() (*HostClient, error) {
// 	return NewHostClientPort(AdbPort)
// }

// func NewHostClientPort(port int) (*HostClient, error) {
// 	return NewHostClientDialer(wire.NewDialer("localhost", port))
// }

// func NewHostClientDialer(d wire.Dialer) (*HostClient, error) {
// 	if d == nil {
// 		return nil, errors.New("dialer cannot be nil.")
// 	}
// 	return &HostClient{d}, nil
// }

func NewHostClient(config ClientConfig) *HostClient {
	return &HostClient{config.sanitized()}
}

// GetServerVersion asks the ADB server for its internal version number.
func (c *HostClient) GetServerVersion() (int, error) {
	resp, err := roundTripSingleResponse(c.config.Dialer, "host:version")
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
	conn, err := c.config.Dialer.Dial()
	if err != nil {
		return err
	}
	defer conn.Close()

	if err = wire.SendMessageString(conn, "host:kill"); err != nil {
		return err
	}

	return nil
}

/*
ListDeviceSerials returns the serial numbers of all attached devices.

Corresponds to the command:
	adb devices
*/
func (c *HostClient) ListDeviceSerials() ([]string, error) {
	resp, err := roundTripSingleResponse(c.config.Dialer, "host:devices")
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
func (c *HostClient) ListDevices() ([]*DeviceInfo, error) {
	resp, err := roundTripSingleResponse(c.config.Dialer, "host:devices-l")
	if err != nil {
		return nil, err
	}

	return parseDeviceList(string(resp), parseDeviceLong)
}
