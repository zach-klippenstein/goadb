// +build windows

package adb

import (
	"errors"
	"strings"
)

func access(path string) error {
	if strings.Contains(path, ".exe") {
		return nil
	}
	return errors.New("not an executable")
}
