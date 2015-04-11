package wire

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteMessage(t *testing.T) {
	s, b := NewTestSender()
	err := SendMessageString(s, "hello")
	assert.NoError(t, err)
	assert.Equal(t, "0005hello", b.String())
}

func TestWriteEmptyMessage(t *testing.T) {
	s, b := NewTestSender()
	err := SendMessageString(s, "")
	assert.NoError(t, err)
	assert.Equal(t, "0000", b.String())
}

func NewTestSender() (Sender, *bytes.Buffer) {
	var buf bytes.Buffer
	return NewSender(&buf), &buf
}
