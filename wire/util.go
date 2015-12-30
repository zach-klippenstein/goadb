package wire

import (
	"fmt"
	"io"
	"regexp"

	"sync"

	"github.com/zach-klippenstein/goadb/util"
)

// ErrorResponseDetails is an error message returned by the server for a particular request.
type ErrorResponseDetails struct {
	Request   string
	ServerMsg string
}

// deviceNotFoundMessagePattern matches all possible error messages returned by adb servers to
// report that a matching device was not found. Used to set the util.DeviceNotFound error code on
// error values.
//
// Old servers send "device not found", and newer ones "device 'serial' not found".
var deviceNotFoundMessagePattern = regexp.MustCompile(`device( '.*')? not found`)

func adbServerError(request string, serverMsg string) error {
	var msg string
	if request == "" {
		msg = fmt.Sprintf("server error: %s", serverMsg)
	} else {
		msg = fmt.Sprintf("server error for %s request: %s", request, serverMsg)
	}

	errCode := util.AdbError
	if deviceNotFoundMessagePattern.MatchString(serverMsg) {
		errCode = util.DeviceNotFound
	}

	return &util.Err{
		Code:    errCode,
		Message: msg,
		Details: ErrorResponseDetails{
			Request:   request,
			ServerMsg: serverMsg,
		},
	}
}

// IsAdbServerErrorMatching returns true if err is an *util.Err with code AdbError and for which
// predicate returns true when passed Details.ServerMsg.
func IsAdbServerErrorMatching(err error, predicate func(string) bool) bool {
	if err, ok := err.(*util.Err); ok && err.Code == util.AdbError {
		return predicate(err.Details.(ErrorResponseDetails).ServerMsg)
	}
	return false
}

func errIncompleteMessage(description string, actual int, expected int) error {
	return &util.Err{
		Code:    util.ConnectionResetError,
		Message: fmt.Sprintf("incomplete %s: read %d bytes, expecting %d", description, actual, expected),
		Details: struct {
			ActualReadBytes int
			ExpectedBytes   int
		}{
			ActualReadBytes: actual,
			ExpectedBytes:   expected,
		},
	}
}

// writeFully writes all of data to w.
// Inverse of io.ReadFully().
func writeFully(w io.Writer, data []byte) error {
	offset := 0
	for offset < len(data) {
		n, err := w.Write(data[offset:])
		if err != nil {
			return util.WrapErrorf(err, util.NetworkError, "error writing %d bytes at offset %d", len(data), offset)
		}
		offset += n
	}
	return nil
}

// MultiCloseable wraps c in a ReadWriteCloser that can be safely closed multiple times.
func MultiCloseable(c io.ReadWriteCloser) io.ReadWriteCloser {
	return &multiCloseable{ReadWriteCloser: c}
}

type multiCloseable struct {
	io.ReadWriteCloser
	closeOnce sync.Once
	err       error
}

func (c *multiCloseable) Close() error {
	c.closeOnce.Do(func() {
		c.err = c.ReadWriteCloser.Close()
	})
	return c.err
}
