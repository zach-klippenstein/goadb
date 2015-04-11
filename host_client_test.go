package goadb

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zach-klippenstein/goadb/wire"
)

func TestGetServerVersion(t *testing.T) {
	client := &HostClient{mockDialer(&MockServer{
		Status:   wire.StatusSuccess,
		Messages: []string{"000a"},
	})}

	v, err := client.GetServerVersion()
	assert.NoError(t, err)
	assert.Equal(t, 10, v)
}

func mockDialer(s *MockServer) Dialer {
	return func() (*wire.Conn, error) {
		return &wire.Conn{s, s, s}, nil
	}
}

type MockServer struct {
	Status   wire.StatusCode
	Messages []string

	nextMsgIndex int
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

func (s *MockServer) SendMessage(msg []byte) error {
	return nil
}

func (s *MockServer) Close() error {
	return nil
}
