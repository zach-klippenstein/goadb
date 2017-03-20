package adb

/*
isExecutable function calls the isExecutableOnPlatform function.
Implementation the isExecutableOnPlatform function is provided in
execuatble_win.go for windows and
executable_unix.go for unix.
*/
func isExecutable(path string) error {
	return isExecutableOnPlatform(path)
}
