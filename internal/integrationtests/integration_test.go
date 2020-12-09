package integrationtests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zach-klippenstein/goadb"
)

func TestListDevices(t *testing.T) {
	server, err := adb.NewServer(adb.ServerConfig{})
	require.NoError(t, err)
	client := adb.NewHostClient(server)

	devices, err := client.ListDevices()
	require.NoError(t, err)
	assert.Len(t, devices, 1)
	t.Logf("Found %d devices:", len(devices))
	for _, device := range devices {
		t.Logf("%s\tusb:%s product:%s model:%s device:%s\n",
			device.Serial, device.Usb, device.Product, device.Model, device.DeviceInfo)
	}
}

func TestShell(t *testing.T) {
	server, err := adb.NewServer(adb.ServerConfig{})
	require.NoError(t, err)
	client := adb.NewDeviceClient(server, adb.AnyDevice())

	output, err := client.RunCommand("echo", "hello world")
	require.NoError(t, err)
	require.Equal(t, "hello world", output)
}
