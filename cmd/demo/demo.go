// An app demonstrating most of the library's features.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	adb "github.com/zach-klippenstein/goadb"
	"github.com/zach-klippenstein/goadb/util"
)

var port = flag.Int("p", adb.AdbPort, "")

func main() {
	flag.Parse()

	client := adb.NewHostClient(adb.ClientConfig{})

	fmt.Println("Starting server…")
	adb.StartServer()

	serverVersion, err := client.GetServerVersion()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Server version:", serverVersion)

	devices, err := client.ListDevices()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Devices:")
	for _, device := range devices {
		fmt.Printf("\t%+v\n", *device)
	}

	PrintDeviceInfoAndError(adb.AnyDevice())
	PrintDeviceInfoAndError(adb.AnyLocalDevice())
	PrintDeviceInfoAndError(adb.AnyUsbDevice())

	serials, err := client.ListDeviceSerials()
	if err != nil {
		log.Fatal(err)
	}
	for _, serial := range serials {
		PrintDeviceInfoAndError(adb.DeviceWithSerial(serial))
	}

	fmt.Println()
	fmt.Println("Watching for device state changes.")
	watcher, err := adb.NewDeviceWatcher(adb.ClientConfig{})
	for event := range watcher.C() {
		fmt.Printf("\t[%s]%+v\n", time.Now(), event)
	}
	if watcher.Err() != nil {
		printErr(watcher.Err())
	}

	//fmt.Println("Killing server…")
	//client.KillServer()
}

func printErr(err error) {
	switch err := err.(type) {
	case *util.Err:
		fmt.Println(err.Error())
		if err.Cause != nil {
			fmt.Print("caused by ")
			printErr(err.Cause)
		}
	default:
		fmt.Println("error:", err)
	}
}

func PrintDeviceInfoAndError(descriptor adb.DeviceDescriptor) {
	device := adb.NewDeviceClient(adb.ClientConfig{}, descriptor)
	if err := PrintDeviceInfo(device); err != nil {
		log.Println(err)
	}
}

func PrintDeviceInfo(device *adb.DeviceClient) error {
	serialNo, err := device.GetSerial()
	if err != nil {
		return err
	}
	devPath, err := device.GetDevicePath()
	if err != nil {
		return err
	}
	state, err := device.GetState()
	if err != nil {
		return err
	}

	fmt.Println(device)
	fmt.Printf("\tserial no: %s\n", serialNo)
	fmt.Printf("\tdevPath: %s\n", devPath)
	fmt.Printf("\tstate: %s\n", state)

	cmdOutput, err := device.RunCommand("pwd")
	if err != nil {
		fmt.Println("\terror running command:", err)
	}
	fmt.Printf("\tcmd output: %s\n", cmdOutput)

	stat, err := device.Stat("/sdcard")
	if err != nil {
		fmt.Println("\terror stating /sdcard:", err)
	}
	fmt.Printf("\tstat \"/sdcard\": %+v\n", stat)

	fmt.Println("\tfiles in \"/\":")
	entries, err := device.ListDirEntries("/")
	if err != nil {
		fmt.Println("\terror listing files:", err)
	} else {
		for entries.Next() {
			fmt.Printf("\t%+v\n", *entries.Entry())
		}
		if entries.Err() != nil {
			fmt.Println("\terror listing files:", err)
		}
	}

	fmt.Println("\tnon-existent file:")
	stat, err = device.Stat("/supercalifragilisticexpialidocious")
	if err != nil {
		fmt.Println("\terror:", err)
	} else {
		fmt.Printf("\tstat: %+v\n", stat)
	}

	fmt.Print("\tload avg: ")
	loadavgReader, err := device.OpenRead("/proc/loadavg")
	if err != nil {
		fmt.Println("\terror opening file:", err)
	} else {
		loadAvg, err := ioutil.ReadAll(loadavgReader)
		if err != nil {
			fmt.Println("\terror reading file:", err)
		} else {
			fmt.Println(string(loadAvg))
		}
	}

	return nil
}
