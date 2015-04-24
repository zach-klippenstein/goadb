package wire

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"
)

type SyncSender interface {
	// SendOctetString sends a 4-byte string.
	SendOctetString(string) error
	SendInt32(int32) error
	SendFileMode(os.FileMode) error
	SendTime(time.Time) error

	// Sends len(bytes) as an octet, followed by bytes.
	SendString(str string) error
}

type realSyncSender struct {
	io.Writer
}

func NewSyncSender(w io.Writer) SyncSender {
	return &realSyncSender{w}
}

func (s *realSyncSender) SendOctetString(str string) error {
	if len(str) != 4 {
		return fmt.Errorf("octet string must be exactly 4 bytes: '%s'", str)
	}
	return writeFully(s.Writer, []byte(str))
}

func (s *realSyncSender) SendInt32(val int32) error {
	return binary.Write(s.Writer, binary.LittleEndian, val)
}

func (s *realSyncSender) SendFileMode(mode os.FileMode) error {
	return binary.Write(s.Writer, binary.LittleEndian, mode)
}

func (s *realSyncSender) SendTime(t time.Time) error {
	return s.SendInt32(int32(t.Unix()))
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
