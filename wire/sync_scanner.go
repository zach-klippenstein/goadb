package wire

import (
	"encoding/binary"
	"io"
	"os"
	"time"

	"github.com/zach-klippenstein/goadb/util"
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
		return util.WrapErrorf(err, util.NetworkError, "expected to read '%s'", expected)
	}
	if actual != expected {
		return util.AssertionErrorf("expected to read '%s', got '%s'", expected, actual)
	}
	return nil
}

func (s *realSyncScanner) ReadOctetString() (string, error) {
	octet := make([]byte, 4)
	n, err := io.ReadFull(s.Reader, octet)

	if err != nil && err != io.ErrUnexpectedEOF {
		return "", util.WrapErrorf(err, util.NetworkError, "error reading octet string from sync scanner")
	} else if err == io.ErrUnexpectedEOF {
		return "", errIncompleteMessage("octet", n, 4)
	}

	return string(octet), nil
}
func (s *realSyncScanner) ReadInt32() (int32, error) {
	var value int32
	err := binary.Read(s.Reader, binary.LittleEndian, &value)
	return value, util.WrapErrorf(err, util.NetworkError, "error reading int from sync scanner")
}
func (s *realSyncScanner) ReadFileMode() (os.FileMode, error) {
	var value uint32
	err := binary.Read(s.Reader, binary.LittleEndian, &value)
	if err != nil {
		return 0, util.WrapErrorf(err, util.NetworkError, "error reading filemode from sync scanner")
	}
	return ParseFileModeFromAdb(value), nil

}
func (s *realSyncScanner) ReadTime() (time.Time, error) {
	seconds, err := s.ReadInt32()
	if err != nil {
		return time.Time{}, util.WrapErrorf(err, util.NetworkError, "error reading time from sync scanner")
	}

	return time.Unix(int64(seconds), 0).UTC(), nil
}

func (s *realSyncScanner) ReadString() (string, error) {
	length, err := s.ReadInt32()
	if err != nil {
		return "", util.WrapErrorf(err, util.NetworkError, "error reading length from sync scanner")
	}

	bytes := make([]byte, length)
	n, rawErr := io.ReadFull(s.Reader, bytes)
	if rawErr != nil && rawErr != io.ErrUnexpectedEOF {
		return "", util.WrapErrorf(rawErr, util.NetworkError, "error reading string from sync scanner")
	} else if rawErr == io.ErrUnexpectedEOF {
		return "", errIncompleteMessage("bytes", n, int(length))
	}

	return string(bytes), nil
}
func (s *realSyncScanner) ReadBytes() (io.Reader, error) {
	length, err := s.ReadInt32()
	if err != nil {
		return nil, util.WrapErrorf(err, util.NetworkError, "error reading bytes from sync scanner")
	}

	return io.LimitReader(s.Reader, int64(length)), nil
}

func (s *realSyncScanner) Close() error {
	if closer, ok := s.Reader.(io.Closer); ok {
		return util.WrapErrorf(closer.Close(), util.NetworkError, "error closing sync scanner")
	}
	return nil
}
