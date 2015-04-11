package wire

import (
	"fmt"
	"io"
	"strconv"
)

// StatusCodes are returned by the server. If the code indicates failure, the
// next message will be the error.
type StatusCode string

const (
	StatusSuccess StatusCode = "OKAY"
	StatusFailure            = "FAIL"
	StatusNone               = ""
)

func (status StatusCode) IsSuccess() bool {
	return status == StatusSuccess
}

/*
Scanner reads tokens from a server.
See Conn for more details.
*/
type Scanner interface {
	ReadStatus() (StatusCode, error)
	ReadMessage() ([]byte, error)
}

type realScanner struct {
	reader io.Reader
}

func NewScanner(r io.Reader) Scanner {
	return &realScanner{r}
}

func ReadMessageString(s Scanner) (string, error) {
	msg, err := s.ReadMessage()
	if err != nil {
		return string(msg), err
	}
	return string(msg), nil
}

func (s *realScanner) ReadStatus() (StatusCode, error) {
	status := make([]byte, 4)
	n, err := io.ReadFull(s.reader, status)
	if err != nil && err != io.ErrUnexpectedEOF {
		return "", err
	} else if err == io.ErrUnexpectedEOF {
		return StatusCode(status), incompleteMessage("status", n, 4)
	}

	return StatusCode(status), nil
}

func (s *realScanner) ReadMessage() ([]byte, error) {
	length, err := s.readLength()
	if err != nil {
		return nil, err
	}

	data := make([]byte, length)
	n, err := io.ReadFull(s.reader, data)
	if err != nil && err != io.ErrUnexpectedEOF {
		return data, fmt.Errorf("error reading message data: %v", err)
	} else if err == io.ErrUnexpectedEOF {
		return data, incompleteMessage("message data", n, length)
	}
	return data, nil
}

func (s *realScanner) readLength() (int, error) {
	lengthHex := make([]byte, 4)
	n, err := io.ReadFull(s.reader, lengthHex)
	if err != nil && err != io.ErrUnexpectedEOF {
		return 0, err
	} else if err == io.ErrUnexpectedEOF {
		return 0, incompleteMessage("length", n, 4)
	}

	length, err := strconv.ParseInt(string(lengthHex), 16, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid hex length: %v", err)
	}

	// Clip the length to 255, as per the Google implementation.
	if length > MaxMessageLength {
		length = MaxMessageLength
	}

	return int(length), nil
}

func incompleteMessage(description string, actual int, expected int) error {
	return fmt.Errorf("incomplete %s: read %d bytes, expecting %d", description, actual, expected)
}

var _ Scanner = &realScanner{}
