package goadb

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zach-klippenstein/goadb/util"
	"github.com/zach-klippenstein/goadb/wire"
	"io/ioutil"
)

func TestReadNextChunk(t *testing.T) {
	s := wire.NewSyncScanner(strings.NewReader(
		"DATA\006\000\000\000hello DATA\005\000\000\000worldDONE"))

	// Read 1st chunk
	reader, err := readNextChunk(s)
	assert.NoError(t, err)
	assert.Equal(t, int64(6), reader.(*io.LimitedReader).N)
	buf := make([]byte, 10)
	n, err := reader.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, 6, n)
	assert.Equal(t, "hello ", string(buf[:6]))

	// Read 2nd chunk
	reader, err = readNextChunk(s)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), reader.(*io.LimitedReader).N)
	buf = make([]byte, 10)
	n, err = reader.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, "world", string(buf[:5]))

	// Read DONE
	_, err = readNextChunk(s)
	assert.Equal(t, io.EOF, err)
}
func TestReadNextChunkInvalidChunkId(t *testing.T) {
	s := wire.NewSyncScanner(strings.NewReader(
		"ATAD\006\000\000\000hello "))

	// Read 1st chunk
	_, err := readNextChunk(s)
	assert.EqualError(t, err, "AssertionError: expected chunk id 'DATA' or 'DONE', but got 'ATAD'")
}

func TestReadMultipleCalls(t *testing.T) {
	s := wire.NewSyncScanner(strings.NewReader(
		"DATA\006\000\000\000hello DATA\005\000\000\000worldDONE"))
	reader, err := newSyncFileReader(s)
	assert.NoError(t, err)

	firstByte := make([]byte, 1)
	_, err = io.ReadFull(reader, firstByte)
	assert.NoError(t, err)
	assert.Equal(t, "h", string(firstByte))

	restFirstChunkBytes := make([]byte, 5)
	_, err = io.ReadFull(reader, restFirstChunkBytes)
	assert.NoError(t, err)
	assert.Equal(t, "ello ", string(restFirstChunkBytes))

	secondChunkBytes := make([]byte, 5)
	_, err = io.ReadFull(reader, secondChunkBytes)
	assert.NoError(t, err)
	assert.Equal(t, "world", string(secondChunkBytes))

	_, err = io.ReadFull(reader, make([]byte, 5))
	assert.Equal(t, io.EOF, err)
}

func TestReadAll(t *testing.T) {
	s := wire.NewSyncScanner(strings.NewReader(
		"DATA\006\000\000\000hello DATA\005\000\000\000worldDONE"))
	reader, err := newSyncFileReader(s)
	assert.NoError(t, err)

	buf := make([]byte, 20)
	_, err = io.ReadFull(reader, buf)
	assert.Equal(t, io.ErrUnexpectedEOF, err)
	assert.Equal(t, "hello world\000", string(buf[:12]))
}

func TestReadError(t *testing.T) {
	s := wire.NewSyncScanner(strings.NewReader(
		"FAIL\004\000\000\000fail"))
	_, err := newSyncFileReader(s)
	assert.EqualError(t, err, "AdbError: server error for read-chunk request: fail ({Request:read-chunk ServerMsg:fail})")
}

func TestReadEmpty(t *testing.T) {
	s := wire.NewSyncScanner(strings.NewReader(
		"DONE"))
	r, err := newSyncFileReader(s)
	assert.NoError(t, err)

	// Multiple read calls that return EOF is a valid case.
	for i := 0; i < 5; i++ {
		data, err := ioutil.ReadAll(r)
		assert.NoError(t, err)
		assert.Empty(t, data)
	}
}

func TestReadErrorNotFound(t *testing.T) {
	s := wire.NewSyncScanner(strings.NewReader(
		"FAIL\031\000\000\000No such file or directory"))
	_, err := newSyncFileReader(s)
	assert.True(t, util.HasErrCode(err, util.FileNoExistError))
	assert.EqualError(t, err, "FileNoExistError: no such file or directory")
}
