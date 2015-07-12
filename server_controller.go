package goadb

import (
	"os/exec"

	"github.com/zach-klippenstein/goadb/util"
)

/*
StartServer ensures there is a server running.

Currently implemented by just running
	adb start-server
*/
func StartServer() error {
	cmd := exec.Command("adb", "start-server")
	err := cmd.Run()
	return util.WrapErrorf(err, util.ServerNotAvailable, "error starting server: %s", err)
}
