package adb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func ParseDeviceList(t *testing.T) {
	devs, err := parseDeviceList(`192.168.56.101:5555	device
05856558`, parseDeviceShort)

	assert.NoError(t, err)
	assert.Len(t, devs, 2)
	assert.Equal(t, "192.168.56.101:5555", devs[0].Serial)
	assert.Equal(t, "05856558", devs[1].Serial)
}

func TestParseDeviceShort(t *testing.T) {
	dev, err := parseDeviceShort("192.168.56.101:5555	device\n")
	assert.NoError(t, err)
	assert.Equal(t, &DeviceInfo{
		Serial: "192.168.56.101:5555"}, dev)
}

func TestParseDeviceLong(t *testing.T) {
	dev, err := parseDeviceLong("SERIAL    device product:PRODUCT model:MODEL device:DEVICE\n")
	assert.NoError(t, err)
	assert.Equal(t, &DeviceInfo{
		Serial:     "SERIAL",
		Product:    "PRODUCT",
		Model:      "MODEL",
		DeviceInfo: "DEVICE"}, dev)
}

func TestParseDeviceLongUnauthorized(t *testing.T) {
	dev, err := parseDeviceLong("SERIAL    unauthorized usb:1234 transport_id:8")
	assert.NoError(t, err)
	assert.Equal(t, &DeviceInfo{
		Serial: "SERIAL",
		Usb: "1234"}, dev)
}

func TestParseDeviceLongUsb(t *testing.T) {
	dev, err := parseDeviceLong("SERIAL    device usb:1234 product:PRODUCT model:MODEL device:DEVICE \n")
	assert.NoError(t, err)
	assert.Equal(t, &DeviceInfo{
		Serial:     "SERIAL",
		Product:    "PRODUCT",
		Model:      "MODEL",
		DeviceInfo: "DEVICE",
		Usb:        "1234"}, dev)
}
