// +build windows

package adb

import (
	"errors"
	"strings"
)

func isExecutableOnPlatform(path string) error {
	if strings.HasSuffix(path, ".exe") || strings.HasSuffix(path, ".cmd") ||
		strings.HasSuffix(path, ".bat") {
		return nil
	}
	return errors.New("not an executable")
}
