// TODO(z): Write SyncSender.SendBytes().
package wire

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"
)

const (
	// Chunks cannot be longer than 64k.
	MaxChunkSize = 64 * 1024
)

/*
SyncConn is a connection to the adb server in sync mode.
Assumes the connection has been put into sync mode (by sending "sync" in transport mode).

The adb sync protocol is defined at
https://android.googlesource.com/platform/system/core/+/master/adb/SYNC.TXT.

Unlike the normal adb protocol (implemented in Conn), the sync protocol is binary.
Lengths are binary-encoded (little-endian) instead of hex.

Notes on Encoding

Length headers and other integers are encoded in little-endian, with 32 bits.

File mode seems to be encoded as POSIX file mode.

Modification time seems to be the Unix timestamp format, i.e. seconds since Epoch UTC.
*/
type SyncConn struct {
	SyncScanner
	SyncSender
}

func (c *SyncConn) Close() error {
	return c.SyncScanner.Close()
}

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

type SyncSender interface {
	// SendOctetString sends a 4-byte string.
	SendOctetString(string) error
	SendInt32(int32) error
	SendFileMode(os.FileMode) error
	SendTime(time.Time) error

	// Sends len(bytes) as an octet, followed by bytes.
	SendString(str string) error
}

type realSyncScanner struct {
	io.Reader
}

type realSyncSender struct {
	io.Writer
}

func NewSyncScanner(r io.Reader) SyncScanner {
	return &realSyncScanner{r}
}

func NewSyncSender(w io.Writer) SyncSender {
	return &realSyncSender{w}
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

func (s *realSyncSender) SendOctetString(str string) error {
	if len(str) != 4 {
		return fmt.Errorf("octet string must be exactly 4 bytes: '%s'", str)
	}
	return writeFully(s.Writer, []byte(str))
}

func (s *realSyncScanner) ReadInt32() (int32, error) {
	var value int32
	err := binary.Read(s.Reader, binary.LittleEndian, &value)
	return value, err
}

func (s *realSyncSender) SendInt32(val int32) error {
	return binary.Write(s.Writer, binary.LittleEndian, val)
}

func (s *realSyncScanner) ReadFileMode() (os.FileMode, error) {
	var value uint32
	err := binary.Read(s.Reader, binary.LittleEndian, &value)
	return os.FileMode(value), err
}

func (s *realSyncSender) SendFileMode(mode os.FileMode) error {
	return binary.Write(s.Writer, binary.LittleEndian, mode)
}

func (s *realSyncScanner) ReadTime() (time.Time, error) {
	seconds, err := s.ReadInt32()
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(int64(seconds), 0).UTC(), nil
}

func (s *realSyncSender) SendTime(t time.Time) error {
	return s.SendInt32(int32(t.Unix()))
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

func (s *realSyncSender) SendString(str string) error {
	length := len(str)
	if length > MaxChunkSize {
		// This limit might not apply to filenames, but it's big enough
		// that I don't think it will be a problem.
		return fmt.Errorf("str must be <= %d in length", MaxChunkSize)
	}

	if err := s.SendInt32(int32(length)); err != nil {
		return err
	}
	return writeFully(s.Writer, []byte(str))
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
