package adb

import (
	"strconv"

	"github.com/zach-klippenstein/goadb/util"
	"github.com/zach-klippenstein/goadb/wire"
)

/*
HostClient communicates with host services on the adb server.

Eg.
	StartServer()
	client := NewHostClient()
	client.ListDevices()

See list of services at https://android.googlesource.com/platform/system/core/+/master/adb/SERVICES.TXT.
*/
// TODO(z): Finish implementing host services.
type HostClient struct {
	server Server
}

func NewHostClient(server Server) *HostClient {
	return &HostClient{server}
}

// GetServerVersion asks the ADB server for its internal version number.
func (c *HostClient) GetServerVersion() (int, error) {
	resp, err := roundTripSingleResponse(c.server, "host:version")
	if err != nil {
		return 0, wrapClientError(err, c, "GetServerVersion")
	}

	version, err := c.parseServerVersion(resp)
	if err != nil {
		return 0, wrapClientError(err, c, "GetServerVersion")
	}
	return version, nil
}

/*
KillServer tells the server to quit immediately.

Corresponds to the command:
	adb kill-server
*/
func (c *HostClient) KillServer() error {
	conn, err := c.server.Dial()
	if err != nil {
		return wrapClientError(err, c, "KillServer")
	}
	defer conn.Close()

	if err = wire.SendMessageString(conn, "host:kill"); err != nil {
		return wrapClientError(err, c, "KillServer")
	}

	return nil
}

/*
ListDeviceSerials returns the serial numbers of all attached devices.

Corresponds to the command:
	adb devices
*/
func (c *HostClient) ListDeviceSerials() ([]string, error) {
	resp, err := roundTripSingleResponse(c.server, "host:devices")
	if err != nil {
		return nil, wrapClientError(err, c, "ListDeviceSerials")
	}

	devices, err := parseDeviceList(string(resp), parseDeviceShort)
	if err != nil {
		return nil, wrapClientError(err, c, "ListDeviceSerials")
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
	resp, err := roundTripSingleResponse(c.server, "host:devices-l")
	if err != nil {
		return nil, wrapClientError(err, c, "ListDevices")
	}

	devices, err := parseDeviceList(string(resp), parseDeviceLong)
	if err != nil {
		return nil, wrapClientError(err, c, "ListDevices")
	}
	return devices, nil
}

func (c *HostClient) parseServerVersion(versionRaw []byte) (int, error) {
	versionStr := string(versionRaw)
	version, err := strconv.ParseInt(versionStr, 16, 32)
	if err != nil {
		return 0, util.WrapErrorf(err, util.ParseError,
			"error parsing server version: %s", versionStr)
	}
	return int(version), nil
}
