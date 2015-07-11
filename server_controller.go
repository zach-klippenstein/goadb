package goadb

import "os/exec"

/*
StartServer ensures there is a server running.

Currently implemented by just running
	adb start-server
*/
func StartServer() error {
	cmd := exec.Command("adb", "start-server")
	return cmd.Run()
}
