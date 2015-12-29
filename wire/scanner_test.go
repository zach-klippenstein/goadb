package wire

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zach-klippenstein/goadb/util"
)

func TestReadStatusOkay(t *testing.T) {
	s := newEofReader("OKAYd")
	status, err := readStatusFailureAsError(s, "", readHexLength)
	assert.NoError(t, err)
	assert.False(t, isFailureStatus(status))
	assertNotEof(t, s)
}

func TestReadIncompleteStatus(t *testing.T) {
	s := newEofReader("oka")
	_, err := readStatusFailureAsError(s, "", readHexLength)
	assert.EqualError(t, err, "NetworkError: error reading status for ")
	assert.Equal(t, errIncompleteMessage("", 3, 4), err.(*util.Err).Cause)
	assertEof(t, s)
}

func TestReadFailureIncompleteStatus(t *testing.T) {
	s := newEofReader("FAIL")
	_, err := readStatusFailureAsError(s, "req", readHexLength)
	assert.EqualError(t, err, "NetworkError: server returned error for req, but couldn't read the error message")
	assert.Error(t, err.(*util.Err).Cause)
	assertEof(t, s)
}

func TestReadFailureEmptyStatus(t *testing.T) {
	s := newEofReader("FAIL0000")
	_, err := readStatusFailureAsError(s, "", readHexLength)
	assert.EqualError(t, err, "AdbError: server error:  ({Request: ServerMsg:})")
	assert.NoError(t, err.(*util.Err).Cause)
	assertEof(t, s)
}

func TestReadFailureStatus(t *testing.T) {
	s := newEofReader("FAIL0004fail")
	_, err := readStatusFailureAsError(s, "", readHexLength)
	assert.EqualError(t, err, "AdbError: server error: fail ({Request: ServerMsg:fail})")
	assert.NoError(t, err.(*util.Err).Cause)
	assertEof(t, s)
}

func TestReadMessage(t *testing.T) {
	s := newEofReader("0005hello")
	msg, err := readMessage(s, readHexLength)
	assert.NoError(t, err)
	assert.Len(t, msg, 5)
	assert.Equal(t, "hello", string(msg))
	assertEof(t, s)
}

func TestReadMessageWithExtraData(t *testing.T) {
	s := newEofReader("0005hellothere")
	msg, err := readMessage(s, readHexLength)
	assert.NoError(t, err)
	assert.Len(t, msg, 5)
	assert.Equal(t, "hello", string(msg))
	assertNotEof(t, s)
}

func TestReadLongerMessage(t *testing.T) {
	s := newEofReader("001b192.168.56.101:5555	device\n")
	msg, err := readMessage(s, readHexLength)
	assert.NoError(t, err)
	assert.Len(t, msg, 27)
	assert.Equal(t, "192.168.56.101:5555	device\n", string(msg))
	assertEof(t, s)
}

func TestReadEmptyMessage(t *testing.T) {
	s := newEofReader("0000")
	msg, err := readMessage(s, readHexLength)
	assert.NoError(t, err)
	assert.Equal(t, "", string(msg))
	assertEof(t, s)
}

func TestReadIncompleteMessage(t *testing.T) {
	s := newEofReader("0005hel")
	msg, err := readMessage(s, readHexLength)
	assert.Error(t, err)
	assert.Equal(t, errIncompleteMessage("message data", 3, 5), err)
	assert.Equal(t, "hel\000\000", string(msg))
	assertEof(t, s)
}

func TestReadLength(t *testing.T) {
	s := newEofReader("000a")
	l, err := readHexLength(s)
	assert.NoError(t, err)
	assert.Equal(t, 10, l)
	assertEof(t, s)
}

func TestReadLengthIncompleteLength(t *testing.T) {
	s := newEofReader("aaa")
	_, err := readHexLength(s)
	assert.Equal(t, errIncompleteMessage("length", 3, 4), err)
	assertEof(t, s)
}

func assertEof(t *testing.T, r io.Reader) {
	msg, err := readMessage(r, readHexLength)
	assert.True(t, util.HasErrCode(err, util.ConnectionResetError))
	assert.Nil(t, msg)
}

func assertNotEof(t *testing.T, r io.Reader) {
	n, err := r.Read(make([]byte, 1))
	assert.Equal(t, 1, n)
	assert.NoError(t, err)
}

// newEofBuffer returns a bytes.Buffer of str that returns an EOF error
// at the end of input, instead of just returning 0 bytes read.
func newEofReader(str string) io.ReadCloser {
	limitReader := io.LimitReader(bytes.NewBufferString(str), int64(len(str)))
	bufReader := bufio.NewReader(limitReader)
	return ioutil.NopCloser(bufReader)
}
