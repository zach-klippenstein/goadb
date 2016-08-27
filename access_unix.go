// +build darwin freebsd linux netbsd openbsd

package adb

import "golang.org/x/sys/unix"

func isExecutableOnPlatform(path string) error {
	return unix.Access(path, unix.X_OK)
}
