package goadb

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zach-klippenstein/goadb/wire"
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
	assert.EqualError(t, err, "expected chunk id 'DATA', but got 'ATAD'")
}

func TestReadMultipleCalls(t *testing.T) {
	s := wire.NewSyncScanner(strings.NewReader(
		"DATA\006\000\000\000hello DATA\005\000\000\000worldDONE"))
	reader := newSyncFileReader(s)

	firstByte := make([]byte, 1)
	_, err := io.ReadFull(reader, firstByte)
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
	reader := newSyncFileReader(s)

	buf := make([]byte, 20)
	_, err := io.ReadFull(reader, buf)
	assert.Equal(t, io.ErrUnexpectedEOF, err)
	assert.Equal(t, "hello world\000", string(buf[:12]))
}
