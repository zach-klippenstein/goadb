// A simple tool for sending raw messages to an adb server.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/zach-klippenstein/goadb"
	"github.com/zach-klippenstein/goadb/wire"
)

var port = flag.Int("p", adb.AdbPort, "`port` the adb server is listening on")

func main() {
	flag.Parse()

	fmt.Println("using port", *port)

	printServerVersion()

	for {
		line := readLine()
		err := doCommand(line)
		if err != nil {
			fmt.Println("error:", err)
		}
	}
}

func printServerVersion() {
	err := doCommand("host:version")
	if err != nil {
		log.Fatal(err)
	}
}

func readLine() string {
	fmt.Print("> ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
	return strings.TrimSpace(line)
}

func doCommand(cmd string) error {
	server, err := adb.NewWithConfig(adb.ServerConfig{
		Port: *port,
	})
	if err != nil {
		log.Fatal(err)
	}

	conn, err := server.Dial()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if err := wire.SendMessageString(conn, cmd); err != nil {
		return err
	}

	status, err := conn.ReadStatus("")
	if err != nil {
		return err
	}

	var msg string
	for err == nil {
		msg, err = wire.ReadMessageString(conn)
		if err == nil {
			fmt.Printf("%s> %s\n", status, msg)
		}
	}

	if err != io.EOF {
		return err
	}

	return nil
}
