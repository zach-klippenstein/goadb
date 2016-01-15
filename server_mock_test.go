package adb

import (
	"io"
	"strings"

	"github.com/zach-klippenstein/goadb/util"
	"github.com/zach-klippenstein/goadb/wire"
)

// MockServer implements Server, Scanner, and Sender.
type MockServer struct {
	// Each time an operation is performed, if this slice is non-empty, the head element
	// of this slice is returned and removed from the slice. If the head is nil, it is removed
	// but not returned.
	Errs []error

	Status string

	// Messages are returned from read calls in order, each preceded by a length header.
	Messages     []string
	nextMsgIndex int

	// Each message passed to a send call is appended to this slice.
	Requests []string

	// Each time an operation is performed, its name is appended to this slice.
	Trace []string
}

var _ Server = &MockServer{}

func (s *MockServer) Dial() (*wire.Conn, error) {
	s.logMethod("Dial")
	if err := s.getNextErrToReturn(); err != nil {
		return nil, err
	}
	return wire.NewConn(s, s), nil
}

func (s *MockServer) Start() error {
	s.logMethod("Start")
	return nil
}

func (s *MockServer) ReadStatus(req string) (string, error) {
	s.logMethod("ReadStatus")
	if err := s.getNextErrToReturn(); err != nil {
		return "", err
	}
	return s.Status, nil
}

func (s *MockServer) ReadMessage() ([]byte, error) {
	s.logMethod("ReadMessage")
	if err := s.getNextErrToReturn(); err != nil {
		return nil, err
	}
	if s.nextMsgIndex >= len(s.Messages) {
		return nil, util.WrapErrorf(io.EOF, util.NetworkError, "")
	}

	s.nextMsgIndex++
	return []byte(s.Messages[s.nextMsgIndex-1]), nil
}

func (s *MockServer) ReadUntilEof() ([]byte, error) {
	s.logMethod("ReadUntilEof")
	if err := s.getNextErrToReturn(); err != nil {
		return nil, err
	}

	var data []string
	for ; s.nextMsgIndex < len(s.Messages); s.nextMsgIndex++ {
		data = append(data, s.Messages[s.nextMsgIndex])
	}
	return []byte(strings.Join(data, "")), nil
}

func (s *MockServer) SendMessage(msg []byte) error {
	s.logMethod("SendMessage")
	if err := s.getNextErrToReturn(); err != nil {
		return err
	}
	s.Requests = append(s.Requests, string(msg))
	return nil
}

func (s *MockServer) NewSyncScanner() wire.SyncScanner {
	s.logMethod("NewSyncScanner")
	return nil
}

func (s *MockServer) NewSyncSender() wire.SyncSender {
	s.logMethod("NewSyncSender")
	return nil
}

func (s *MockServer) Close() error {
	s.logMethod("Close")
	if err := s.getNextErrToReturn(); err != nil {
		return err
	}
	return nil
}

func (s *MockServer) getNextErrToReturn() (err error) {
	if len(s.Errs) > 0 {
		err = s.Errs[0]
		s.Errs = s.Errs[1:]
	}
	return
}

func (s *MockServer) logMethod(name string) {
	s.Trace = append(s.Trace, name)
}
