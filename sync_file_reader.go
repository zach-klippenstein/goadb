package goadb

import (
	"fmt"
	"io"

	"github.com/zach-klippenstein/goadb/util"
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

func newSyncFileReader(s wire.SyncScanner) (r io.ReadCloser, err error) {
	r = &syncFileReader{
		scanner: s,
	}

	// Read the header for the first chunk to consume any errors.
	if _, err = r.Read([]byte{}); err != nil {
		r.Close()
		return nil, err
	}
	return
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

	if len(buf) == 0 {
		// Read can be called with an empty buffer to read the next chunk and check for errors.
		// However, net.Conn.Read seems to return EOF when given an empty buffer, so we need to
		// handle that case ourselves.
		return 0, nil
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
	status, err := r.ReadStatus("read-chunk")
	if err != nil {
		if wire.IsAdbServerErrorMatching(err, readFileNotFoundPredicate) {
			return nil, util.Errorf(util.FileNoExistError, "no such file or directory")
		}
		return nil, err
	}

	switch status {
	case wire.StatusSyncData:
		return r.ReadBytes()
	case wire.StatusSyncDone:
		return nil, io.EOF
	default:
		return nil, fmt.Errorf("expected chunk id '%s' or '%s', but got '%s'",
			wire.StatusSyncData, wire.StatusSyncDone, []byte(status))
	}
}

// readFileNotFoundPredicate returns true if s is the adb server error message returned
// when trying to open a file that doesn't exist.
func readFileNotFoundPredicate(s string) bool {
	return s == "No such file or directory"
}
