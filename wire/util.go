package wire

import (
	"fmt"
	"io"

	"github.com/zach-klippenstein/goadb/util"
)

// ErrorResponseDetails is an error message returned by the server for a particular request.
type ErrorResponseDetails struct {
	Request   string
	ServerMsg string
}

// Reads the status, and if failure, reads the message and returns it as an error.
// If the status is success, doesn't read the message.
// req is just used to populate the AdbError, and can be nil.
func ReadStatusFailureAsError(s Scanner, req string) error {
	status, err := s.ReadStatus()
	if err != nil {
		return util.WrapErrorf(err, util.NetworkError, "error reading status for %s", req)
	}

	if !status.IsSuccess() {
		msg, err := s.ReadMessage()
		if err != nil {
			return util.WrapErrorf(err, util.NetworkError,
				"server returned error for %s, but couldn't read the error message", req)
		}

		return adbServerError(req, string(msg))
	}

	return nil
}

func adbServerError(request string, serverMsg string) error {
	var msg string
	if request == "" {
		msg = fmt.Sprintf("server error: %s", serverMsg)
	} else {
		msg = fmt.Sprintf("server error for %s request: %s", request, serverMsg)
	}

	errCode := util.AdbError
	if serverMsg == "device not found" {
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
