package wire

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"
)

type SyncScanner interface {
	// ReadOctetString reads a 4-byte string.
	ReadOctetString() (string, error)
	ReadInt32() (int32, error)
	ReadFileMode() (os.FileMode, error)
	ReadTime() (time.Time, error)

	// Reads an octet length, followed by length bytes.
	ReadString() (string, error)

	// Reads an octet length, and returns a reader that will read length
	// bytes (see io.LimitReader). The returned reader should be fully
	// read before reading anything off the Scanner again.
	ReadBytes() (io.Reader, error)

	// Closes the underlying reader.
	Close() error
}

type realSyncScanner struct {
	io.Reader
}

func NewSyncScanner(r io.Reader) SyncScanner {
	return &realSyncScanner{r}
}

func RequireOctetString(s SyncScanner, expected string) error {
	actual, err := s.ReadOctetString()
	if err != nil {
		return fmt.Errorf("expected to read '%s', got err: %v", expected, err)
	}
	if actual != expected {
		return fmt.Errorf("expected to read '%s', got '%s'", expected, actual)
	}
	return nil
}

func (s *realSyncScanner) ReadOctetString() (string, error) {
	octet := make([]byte, 4)
	n, err := io.ReadFull(s.Reader, octet)
	if err != nil && err != io.ErrUnexpectedEOF {
		return "", err
	} else if err == io.ErrUnexpectedEOF {
		return "", incompleteMessage("octet", n, 4)
	}

	return string(octet), nil
}
func (s *realSyncScanner) ReadInt32() (int32, error) {
	var value int32
	err := binary.Read(s.Reader, binary.LittleEndian, &value)
	return value, err
}
func (s *realSyncScanner) ReadFileMode() (filemode os.FileMode, err error) {
	var value uint32
	err = binary.Read(s.Reader, binary.LittleEndian, &value)
	if err == nil {
		filemode = ParseFileModeFromAdb(value)
	}
	return
}
func (s *realSyncScanner) ReadTime() (time.Time, error) {
	seconds, err := s.ReadInt32()
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(int64(seconds), 0).UTC(), nil
}

func (s *realSyncScanner) ReadString() (string, error) {
	length, err := s.ReadInt32()
	if err != nil {
		return "", err
	}

	bytes := make([]byte, length)
	n, err := io.ReadFull(s.Reader, bytes)
	if err != nil && err != io.ErrUnexpectedEOF {
		return "", err
	} else if err == io.ErrUnexpectedEOF {
		return "", incompleteMessage("bytes", n, int(length))
	}

	return string(bytes), nil
}
func (s *realSyncScanner) ReadBytes() (io.Reader, error) {
	length, err := s.ReadInt32()
	if err != nil {
		return nil, err
	}

	return io.LimitReader(s.Reader, int64(length)), nil
}

func (s *realSyncScanner) Close() error {
	if closer, ok := s.Reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
