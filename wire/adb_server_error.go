package wire

import (
	"fmt"
)

type AdbServerError struct {
	Request   string
	ServerMsg string
}

var _ error = &AdbServerError{}

func (e *AdbServerError) Error() string {
	if e.Request == "" {
		return fmt.Sprintf("server error: %s", e.ServerMsg)
	} else {
		return fmt.Sprintf("server error for %s request: %s", e.Request, e.ServerMsg)
	}
}

func incompleteMessage(description string, actual int, expected int) error {
	return fmt.Errorf("incomplete %s: read %d bytes, expecting %d", description, actual, expected)
}
