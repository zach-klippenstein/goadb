package wire

import (
	"io"
	"io/ioutil"
	"strconv"

	"github.com/zach-klippenstein/goadb/util"
)

// TODO(zach): All EOF errors returned from networoking calls should use ConnectionResetError.

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
	ReadUntilEof() ([]byte, error)

	NewSyncScanner() SyncScanner

	Close() error
}

type realScanner struct {
	reader io.ReadCloser
}

func NewScanner(r io.ReadCloser) Scanner {
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
		return "", util.WrapErrorf(err, util.NetworkError, "error reading status")
	} else if err == io.ErrUnexpectedEOF {
		return StatusCode(status), errIncompleteMessage("status", n, 4)
	}

	return StatusCode(status), nil
}

func (s *realScanner) ReadMessage() ([]byte, error) {
	var err error

	length, err := s.readLength()
	if err != nil {
		return nil, err
	}

	data := make([]byte, length)
	n, err := io.ReadFull(s.reader, data)

	if err != nil && err != io.ErrUnexpectedEOF {
		return data, util.WrapErrorf(err, util.NetworkError, "error reading message data")
	} else if err == io.ErrUnexpectedEOF {
		return data, errIncompleteMessage("message data", n, length)
	}
	return data, nil
}

func (s *realScanner) ReadUntilEof() ([]byte, error) {
	data, err := ioutil.ReadAll(s.reader)
	if err != nil {
		return nil, util.WrapErrorf(err, util.NetworkError, "error reading until EOF")
	}
	return data, nil
}

func (s *realScanner) NewSyncScanner() SyncScanner {
	return NewSyncScanner(s.reader)
}

func (s *realScanner) Close() error {
	return util.WrapErrorf(s.reader.Close(), util.NetworkError, "error closing scanner")
}

func (s *realScanner) readLength() (int, error) {
	lengthHex := make([]byte, 4)
	n, err := io.ReadFull(s.reader, lengthHex)
	if err != nil {
		return 0, errIncompleteMessage("length", n, 4)
	}

	length, err := strconv.ParseInt(string(lengthHex), 16, 64)
	if err != nil {
		return 0, util.WrapErrorf(err, util.NetworkError, "could not parse hex length %v", lengthHex)
	}

	// Clip the length to 255, as per the Google implementation.
	if length > MaxMessageLength {
		length = MaxMessageLength
	}

	return int(length), nil
}

var _ Scanner = &realScanner{}
