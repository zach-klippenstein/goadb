package goadb

import (
	"bufio"
	"fmt"
	"strings"
)

// Device represents a connected Android device.
type Device struct {
	// Always set.
	Serial string

	// Product, device, and model are not set in the short form.
	Product string
	Model   string
	Device  string

	// Only set for devices connected via USB.
	Usb string
}

// IsUsb returns true if the device is connected via USB.
func (d *Device) IsUsb() bool {
	return d.Usb != ""
}

func newDevice(serial string, attrs map[string]string) (*Device, error) {
	if serial == "" {
		return nil, fmt.Errorf("device serial cannot be blank")
	}

	return &Device{
		Serial:  serial,
		Product: attrs["product"],
		Model:   attrs["model"],
		Device:  attrs["device"],
		Usb:     attrs["usb"],
	}, nil
}

func parseDeviceList(list string, lineParseFunc func(string) (*Device, error)) ([]*Device, error) {
	var devices []*Device
	scanner := bufio.NewScanner(strings.NewReader(list))

	for scanner.Scan() {
		device, err := lineParseFunc(scanner.Text())
		if err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return devices, nil
}

func parseDeviceShort(line string) (*Device, error) {
	fields := strings.Fields(line)
	if len(fields) != 2 {
		return nil, fmt.Errorf("malformed device line, expected 2 fields but found %d", len(fields))
	}

	return newDevice(fields[0], map[string]string{})
}

func parseDeviceLong(line string) (*Device, error) {
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return nil, fmt.Errorf("malformed device line, expected at least 5 fields but found %d", len(fields))
	}

	attrs := parseDeviceAttributes(fields[2:])
	return newDevice(fields[0], attrs)
}

func parseDeviceAttributes(fields []string) map[string]string {
	attrs := map[string]string{}
	for _, field := range fields {
		key, val := parseKeyVal(field)
		attrs[key] = val
	}
	return attrs
}

// Parses a key:val pair and returns key, val.
func parseKeyVal(pair string) (string, string) {
	split := strings.Split(pair, ":")
	return split[0], split[1]
}
