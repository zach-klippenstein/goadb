package goadb

import (
	"os/exec"
	"strings"

	"github.com/zach-klippenstein/goadb/util"
)

/*
StartServer ensures there is a server running.
*/
func StartServer() error {
	cmd := exec.Command("adb", "start-server")
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))
	return util.WrapErrorf(err, util.ServerNotAvailable, "error starting server: %s\noutput:\n%s", err, outputStr)
}
