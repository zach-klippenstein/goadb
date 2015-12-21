package main

import (
	"fmt"
	"os"

	"github.com/zach-klippenstein/goadb"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	serial          = kingpin.Flag("serial", "Connect to device by serial number.").Short('s').String()
	shellCommand    = kingpin.Command("shell", "Run a shell command on the device.")
	shellCommandArg = shellCommand.Arg("command", "Command to run on device.").Strings()
	devicesCommand  = kingpin.Command("devices", "List devices.")
	devicesLongFlag = devicesCommand.Flag("long", "Include extra detail about devices.").Short('l').Bool()
)

func main() {
	var exitCode int

	switch kingpin.Parse() {
	case "devices":
		exitCode = listDevices(*devicesLongFlag)
	case "shell":
		exitCode = runShellCommand(*shellCommandArg, parseDevice())
	}

	os.Exit(exitCode)
}

func parseDevice() goadb.DeviceDescriptor {
	if *serial != "" {
		return goadb.DeviceWithSerial(*serial)
	}

	return goadb.AnyDevice()
}

func listDevices(long bool) int {
	client := goadb.NewHostClient(goadb.ClientConfig{})
	devices, err := client.ListDevices()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
	}

	for _, device := range devices {
		if long {
			fmt.Printf("%s\t\t%s\t%s\n", device.Serial, device.Product, device.Model)
		} else {
			fmt.Println(device.Serial)
		}
	}

	return 0
}

func runShellCommand(commandAndArgs []string, device goadb.DeviceDescriptor) int {
	if len(commandAndArgs) == 0 {
		fmt.Fprintln(os.Stderr, "error: no command")
		kingpin.Usage()
		return 1
	}

	command := commandAndArgs[0]
	var args []string

	if len(commandAndArgs) > 1 {
		args = commandAndArgs[1:]
	}

	client := goadb.NewDeviceClient(goadb.ClientConfig{}, device)
	output, err := client.RunCommand(command, args...)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	fmt.Print(output)
	return 0
}
