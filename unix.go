package adb

/*
Exported Access function calls private access function.
Implementation for private access function is provided in
access_win.go for windows and
access_unix.go for unix
*/
func Access(path string) error {
	return access(path)
}
