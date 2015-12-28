package wire

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zach-klippenstein/goadb/util"
)

var (
	someTime = time.Date(2015, 04, 12, 20, 7, 51, 0, time.UTC)
	// The little-endian encoding of someTime.Unix()
	someTimeEncoded = []byte{151, 208, 42, 85}
)

func TestSyncSendOctetString(t *testing.T) {
	var buf bytes.Buffer
	s := NewSyncSender(&buf)
	err := s.SendOctetString("helo")
	assert.NoError(t, err)
	assert.Equal(t, "helo", buf.String())
}

func TestSyncSendOctetStringTooLong(t *testing.T) {
	var buf bytes.Buffer
	s := NewSyncSender(&buf)
	err := s.SendOctetString("hello")
	assert.Equal(t, util.AssertionErrorf("octet string must be exactly 4 bytes: 'hello'"), err)
}

func TestSyncReadTime(t *testing.T) {
	s := NewSyncScanner(bytes.NewReader(someTimeEncoded))
	decoded, err := s.ReadTime()
	assert.NoError(t, err)
	assert.Equal(t, someTime, decoded)
}

func TestSyncSendTime(t *testing.T) {
	var buf bytes.Buffer
	s := NewSyncSender(&buf)
	err := s.SendTime(someTime)
	assert.NoError(t, err)
	assert.Equal(t, someTimeEncoded, buf.Bytes())
}

func TestSyncReadString(t *testing.T) {
	s := NewSyncScanner(strings.NewReader("\005\000\000\000hello"))
	str, err := s.ReadString()
	assert.NoError(t, err)
	assert.Equal(t, "hello", str)
}

func TestSyncReadStringTooShort(t *testing.T) {
	s := NewSyncScanner(strings.NewReader("\005\000\000\000h"))
	_, err := s.ReadString()
	assert.Equal(t, errIncompleteMessage("bytes", 1, 5), err)
}

func TestSyncSendBytes(t *testing.T) {
	var buf bytes.Buffer
	s := NewSyncSender(&buf)
	err := s.SendBytes([]byte("hello"))
	assert.NoError(t, err)
	assert.Equal(t, "\005\000\000\000hello", buf.String())
}

func TestSyncReadBytes(t *testing.T) {
	s := NewSyncScanner(strings.NewReader("\005\000\000\000helloworld"))

	reader, err := s.ReadBytes()
	assert.NoError(t, err)
	assert.NotNil(t, reader)

	str, err := ioutil.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, "hello", string(str))
}
