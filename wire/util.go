package wire

import (
	"fmt"
	"io"
)

// Reads the status, and if failure, reads the message and returns it as an error.
// If the status is success, doesn't read the message.
// req is just used to populate the AdbError, and can be nil.
func ReadStatusFailureAsError(s Scanner, req string) error {
	status, err := s.ReadStatus()
	if err != nil {
		return fmt.Errorf("error reading status for %s: %+v", req, err)
	}

	if !status.IsSuccess() {
		msg, err := s.ReadMessage()
		if err != nil {
			return fmt.Errorf("server returned error for %s, but couldn't read the error message: %+v", err)
		}

		return &AdbError{
			Request:   req,
			ServerMsg: string(msg),
		}
	}

	return nil
}

// writeFully writes all of data to w.
// Inverse of io.ReadFully().
func writeFully(w io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := w.Write(data)
		if err != nil {
			return err
		}
		data = data[n:]
	}
	return nil
}
