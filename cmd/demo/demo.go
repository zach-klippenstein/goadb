package main

import (
	"fmt"
	"log"

	adb "github.com/zach-klippenstein/goadb"
	"github.com/zach-klippenstein/goadb/wire"
)

func main() {
	client := &adb.HostClient{wire.Dial}
	fmt.Println("Starting server…")
	client.StartServer()

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

	fmt.Println("Killing server…")
	client.KillServer()
}
