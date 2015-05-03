// TODO(z): Implement TrackDevices.
package goadb

import (
	"errors"
	"os/exec"
	"strconv"

	"github.com/zach-klippenstein/goadb/wire"
)

const (
	// Default port the adb server listens on.
	AdbPort = 5037
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
	dialer wire.Dialer
}

func NewHostClient() (*HostClient, error) {
	return NewHostClientPort(AdbPort)
}

func NewHostClientPort(port int) (*HostClient, error) {
	return NewHostClientDialer(wire.NewDialer("localhost", port))
}

func NewHostClientDialer(d wire.Dialer) (*HostClient, error) {
	if d == nil {
		return nil, errors.New("dialer cannot be nil.")
	}
	return &HostClient{d}, nil
}

// GetServerVersion asks the ADB server for its internal version number.
func (c *HostClient) GetServerVersion() (int, error) {
	resp, err := wire.RoundTripSingleResponse(c.dialer, "host:version")
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
	conn, err := c.dialer.Dial()
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
	resp, err := wire.RoundTripSingleResponse(c.dialer, "host:devices")
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
	resp, err := wire.RoundTripSingleResponse(c.dialer, "host:devices-l")
	if err != nil {
		return nil, err
	}

	return parseDeviceList(string(resp), parseDeviceLong)
}

func (c *HostClient) GetDevice(d *DeviceInfo) *DeviceClient {
	return c.GetDeviceWithSerial(d.Serial)
}

// GetDeviceWithSerial returns a client for the device with the specified serial number.
// Will return a client even if there is no matching device connected.
func (c *HostClient) GetDeviceWithSerial(serial string) *DeviceClient {
	return c.getDevice(deviceWithSerial(serial))
}

// GetAnyDevice returns a client for any one connected device.
func (c *HostClient) GetAnyDevice() *DeviceClient {
	return c.getDevice(anyDevice())
}

// GetUsbDevice returns a client for the USB device.
// Will return a client even if there is no device connected.
func (c *HostClient) GetUsbDevice() *DeviceClient {
	return c.getDevice(anyUsbDevice())
}

// GetLocalDevice returns a client for the local device.
// Will return a client even if there is no device connected.
func (c *HostClient) GetLocalDevice() *DeviceClient {
	return c.getDevice(anyLocalDevice())
}

func (c *HostClient) getDevice(descriptor *DeviceDescriptor) *DeviceClient {
	return &DeviceClient{c.dialer, descriptor}
}
