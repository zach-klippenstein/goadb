package adb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zach-klippenstein/goadb/util"
	"github.com/zach-klippenstein/goadb/wire"
)

func TestGetAttribute(t *testing.T) {
	s := &MockServer{
		Status:   wire.StatusSuccess,
		Messages: []string{"value"},
	}
	client := NewDeviceClient(s, DeviceWithSerial("serial"))

	v, err := client.getAttribute("attr")
	assert.Equal(t, "host-serial:serial:attr", s.Requests[0])
	assert.NoError(t, err)
	assert.Equal(t, "value", v)
}

func TestGetDeviceInfo(t *testing.T) {
	deviceLister := func() ([]*DeviceInfo, error) {
		return []*DeviceInfo{
			&DeviceInfo{
				Serial:  "abc",
				Product: "Foo",
			},
			&DeviceInfo{
				Serial:  "def",
				Product: "Bar",
			},
		}, nil
	}

	client := newDeviceClientWithDeviceLister("abc", deviceLister)
	device, err := client.GetDeviceInfo()
	assert.NoError(t, err)
	assert.Equal(t, "Foo", device.Product)

	client = newDeviceClientWithDeviceLister("def", deviceLister)
	device, err = client.GetDeviceInfo()
	assert.NoError(t, err)
	assert.Equal(t, "Bar", device.Product)

	client = newDeviceClientWithDeviceLister("serial", deviceLister)
	device, err = client.GetDeviceInfo()
	assert.True(t, util.HasErrCode(err, util.DeviceNotFound))
	assert.EqualError(t, err.(*util.Err).Cause,
		"DeviceNotFound: device list doesn't contain serial serial")
	assert.Nil(t, device)
}

func newDeviceClientWithDeviceLister(serial string, deviceLister func() ([]*DeviceInfo, error)) *DeviceClient {
	client := NewDeviceClient(
		&MockServer{
			Status:   wire.StatusSuccess,
			Messages: []string{serial},
		},
		DeviceWithSerial(serial),
	)
	client.deviceListFunc = deviceLister
	return client
}

func TestRunCommandNoArgs(t *testing.T) {
	s := &MockServer{
		Status:   wire.StatusSuccess,
		Messages: []string{"output"},
	}
	client := NewDeviceClient(s, AnyDevice())

	v, err := client.RunCommand("cmd")
	assert.Equal(t, "host:transport-any", s.Requests[0])
	assert.Equal(t, "shell:cmd", s.Requests[1])
	assert.NoError(t, err)
	assert.Equal(t, "output", v)
}

func TestPrepareCommandLineNoArgs(t *testing.T) {
	result, err := prepareCommandLine("cmd")
	assert.NoError(t, err)
	assert.Equal(t, "cmd", result)
}

func TestPrepareCommandLineEmptyCommand(t *testing.T) {
	_, err := prepareCommandLine("")
	assert.Equal(t, util.AssertionError, code(err))
	assert.Equal(t, "command cannot be empty", message(err))
}

func TestPrepareCommandLineBlankCommand(t *testing.T) {
	_, err := prepareCommandLine("  ")
	assert.Equal(t, util.AssertionError, code(err))
	assert.Equal(t, "command cannot be empty", message(err))
}

func TestPrepareCommandLineCleanArgs(t *testing.T) {
	result, err := prepareCommandLine("cmd", "arg1", "arg2")
	assert.NoError(t, err)
	assert.Equal(t, "cmd arg1 arg2", result)
}

func TestPrepareCommandLineArgWithWhitespaceQuotes(t *testing.T) {
	result, err := prepareCommandLine("cmd", "arg with spaces")
	assert.NoError(t, err)
	assert.Equal(t, "cmd \"arg with spaces\"", result)
}

func TestPrepareCommandLineArgWithDoubleQuoteFails(t *testing.T) {
	_, err := prepareCommandLine("cmd", "quoted\"arg")
	assert.Equal(t, util.ParseError, code(err))
	assert.Equal(t, "arg at index 0 contains an invalid double quote: quoted\"arg", message(err))
}

func code(err error) util.ErrCode {
	return err.(*util.Err).Code
}

func message(err error) string {
	return err.(*util.Err).Message
}
