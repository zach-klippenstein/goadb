package goadb

import (
	"fmt"
	"io"

	"github.com/zach-klippenstein/goadb/wire"
)

// syncFileReader wraps a SyncConn that has requested to receive a file.
type syncFileReader struct {
	// Reader used to read data from the adb connection.
	scanner wire.SyncScanner

	// Reader for the current chunk only.
	chunkReader io.Reader
}

var _ io.ReadCloser = &syncFileReader{}

func newSyncFileReader(s wire.SyncScanner) io.ReadCloser {
	return &syncFileReader{
		scanner: s,
	}
}

func (r *syncFileReader) Read(buf []byte) (n int, err error) {
	if r.chunkReader == nil {
		chunkReader, err := readNextChunk(r.scanner)
		if err != nil {
			// If this is EOF, we've read the last chunk.
			// Either way, we want to pass it up to the caller.
			return 0, err
		}
		r.chunkReader = chunkReader
	}

	n, err = r.chunkReader.Read(buf)
	if err == io.EOF {
		// End of current chunk, don't return an error, the next chunk will be
		// read on the next call to this method.
		r.chunkReader = nil
		return n, nil
	}

	return n, err
}

func (r *syncFileReader) Close() error {
	return r.scanner.Close()
}

// readNextChunk creates an io.LimitedReader for the next chunk of data,
// and returns io.EOF if the last chunk has been read.
func readNextChunk(r wire.SyncScanner) (io.Reader, error) {
	id, err := r.ReadOctetString()
	if err != nil {
		return nil, err
	}

	switch id {
	case "DATA":
		return r.ReadBytes()
	case "DONE":
		return nil, io.EOF
	default:
		return nil, fmt.Errorf("expected chunk id 'DATA', but got '%s'", id)
	}
}
