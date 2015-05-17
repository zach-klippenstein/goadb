package wire

import (
	"bufio"
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zach-klippenstein/goadb/util"
)

func TestReadStatusOkay(t *testing.T) {
	s := NewScannerString("OKAYd")
	status, err := s.ReadStatus()
	assert.NoError(t, err)
	assert.True(t, status.IsSuccess())
	assertNotEof(t, s)
}

func TestReadIncompleteStatus(t *testing.T) {
	s := NewScannerString("oka")
	_, err := s.ReadStatus()
	assert.Equal(t, errIncompleteMessage("status", 3, 4), err)
	assertEof(t, s)
}

func TestReadLength(t *testing.T) {
	s := NewScannerString("000a")
	l, err := s.readLength()
	assert.NoError(t, err)
	assert.Equal(t, 10, l)
	assertEof(t, s)
}

func TestReadIncompleteLength(t *testing.T) {
	s := NewScannerString("aaa")
	_, err := s.readLength()
	assert.Equal(t, errIncompleteMessage("length", 3, 4), err)
	assertEof(t, s)
}

func TestReadMessage(t *testing.T) {
	s := NewScannerString("0005hello")
	msg, err := ReadMessageString(s)
	assert.NoError(t, err)
	assert.Len(t, msg, 5)
	assert.Equal(t, "hello", msg)
	assertEof(t, s)
}

func TestReadMessageWithExtraData(t *testing.T) {
	s := NewScannerString("0005hellothere")
	msg, err := ReadMessageString(s)
	assert.NoError(t, err)
	assert.Len(t, msg, 5)
	assert.Equal(t, "hello", msg)
	assertNotEof(t, s)
}

func TestReadLongerMessage(t *testing.T) {
	s := NewScannerString("001b192.168.56.101:5555	device\n")
	msg, err := ReadMessageString(s)
	assert.NoError(t, err)
	assert.Len(t, msg, 27)
	assert.Equal(t, "192.168.56.101:5555	device\n", msg)
	assertEof(t, s)
}

func TestReadEmptyMessage(t *testing.T) {
	s := NewScannerString("0000")
	msg, err := ReadMessageString(s)
	assert.NoError(t, err)
	assert.Equal(t, "", msg)
	assertEof(t, s)
}

func TestReadIncompleteMessage(t *testing.T) {
	s := NewScannerString("0005hel")
	msg, err := ReadMessageString(s)
	assert.Error(t, err)
	assert.Equal(t, errIncompleteMessage("message data", 3, 5), err)
	assert.Equal(t, "hel\000\000", msg)
	assertEof(t, s)
}

func NewScannerString(str string) *realScanner {
	return NewScanner(NewEofBuffer(str)).(*realScanner)
}

// NewEofBuffer returns a bytes.Buffer of str that returns an EOF error
// at the end of input, instead of just returning 0 bytes read.
func NewEofBuffer(str string) *TestReader {
	limitReader := io.LimitReader(bytes.NewBufferString(str), int64(len(str)))
	bufReader := bufio.NewReader(limitReader)
	return &TestReader{bufReader}
}

func assertEof(t *testing.T, s *realScanner) {
	msg, err := s.ReadMessage()
	assert.True(t, util.HasErrCode(err, util.ConnectionResetError))
	assert.Nil(t, msg)
}

func assertNotEof(t *testing.T, s *realScanner) {
	n, err := s.reader.Read(make([]byte, 1))
	assert.Equal(t, 1, n)
	assert.NoError(t, err)
}

// TestReader is a wrapper around a bufio.Reader that implements io.Closer.
type TestReader struct {
	*bufio.Reader
}

func (b *TestReader) Close() error {
	// No-op.
	return nil
}
