package wire

import (
	"fmt"
	"io"
)

// Sender sends messages to the server.
type Sender interface {
	SendMessage(msg []byte) error
	NewSyncSender() SyncSender
}

type realSender struct {
	writer io.Writer
}

func NewSender(w io.Writer) Sender {
	return &realSender{w}
}

func SendMessageString(s Sender, msg string) error {
	return s.SendMessage([]byte(msg))
}

func (s *realSender) SendMessage(msg []byte) error {
	if len(msg) > MaxMessageLength {
		return fmt.Errorf("message length exceeds maximum: %d", len(msg))
	}

	lengthAndMsg := fmt.Sprintf("%04x%s", len(msg), msg)
	return writeFully(s.writer, []byte(lengthAndMsg))
}

func (s *realSender) NewSyncSender() SyncSender {
	return NewSyncSender(s.writer)
}

var _ Sender = &realSender{}
