package goadb

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zach-klippenstein/goadb/wire"
)

func TestGetServerVersion(t *testing.T) {
	s := &MockServer{
		Status:   wire.StatusSuccess,
		Messages: []string{"000a"},
	}
	client := NewHostClientDialer(s)

	v, err := client.GetServerVersion()
	assert.Equal(t, "host:version", s.Requests[0])
	assert.NoError(t, err)
	assert.Equal(t, 10, v)
}

type MockServer struct {
	Status wire.StatusCode

	// Messages are sent in order, each preceded by a length header.
	Messages []string

	// Each request is appended to this slice.
	Requests []string

	nextMsgIndex int
}

func (s *MockServer) Dial() (*wire.Conn, error) {
	return wire.NewConn(s, s), nil
}

func (s *MockServer) ReadStatus() (wire.StatusCode, error) {
	return s.Status, nil
}

func (s *MockServer) ReadMessage() ([]byte, error) {
	if s.nextMsgIndex >= len(s.Messages) {
		return nil, io.EOF
	}

	s.nextMsgIndex++
	return []byte(s.Messages[s.nextMsgIndex-1]), nil
}

func (s *MockServer) ReadUntilEof() ([]byte, error) {
	var data []string
	for ; s.nextMsgIndex < len(s.Messages); s.nextMsgIndex++ {
		data = append(data, s.Messages[s.nextMsgIndex])
	}
	return []byte(strings.Join(data, "")), nil
}

func (s *MockServer) SendMessage(msg []byte) error {
	s.Requests = append(s.Requests, string(msg))
	return nil
}

func (s *MockServer) NewSyncScanner() wire.SyncScanner {
	return nil
}

func (s *MockServer) NewSyncSender() wire.SyncSender {
	return nil
}

func (s *MockServer) Close() error {
	return nil
}
