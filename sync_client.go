// TODO(z): Implement send.
package goadb

import (
	"io"
	"os"
	"time"

	"github.com/zach-klippenstein/goadb/util"
	"github.com/zach-klippenstein/goadb/wire"
)

var zeroTime = time.Unix(0, 0).UTC()

func stat(conn *wire.SyncConn, path string) (*DirEntry, error) {
	if err := conn.SendOctetString("STAT"); err != nil {
		return nil, err
	}
	if err := conn.SendString(path); err != nil {
		return nil, err
	}

	id, err := conn.ReadOctetString()
	if err != nil {
		return nil, err
	}
	if id != "STAT" {
		return nil, util.Errorf(util.AssertionError, "expected stat ID 'STAT', but got '%s'", id)
	}

	return readStat(conn)
}

func listDirEntries(conn *wire.SyncConn, path string) (entries *DirEntries, err error) {
	if err = conn.SendOctetString("LIST"); err != nil {
		return
	}
	if err = conn.SendString(path); err != nil {
		return
	}

	return &DirEntries{scanner: conn}, nil
}

func receiveFile(conn *wire.SyncConn, path string) (io.ReadCloser, error) {
	if err := conn.SendOctetString("RECV"); err != nil {
		return nil, err
	}
	if err := conn.SendString(path); err != nil {
		return nil, err
	}

	return newSyncFileReader(conn), nil
}

func readStat(s wire.SyncScanner) (entry *DirEntry, err error) {
	mode, err := s.ReadFileMode()
	if err != nil {
		err = util.WrapErrf(err, "error reading file mode: %v", err)
		return
	}
	size, err := s.ReadInt32()
	if err != nil {
		err = util.WrapErrf(err, "error reading file size: %v", err)
		return
	}
	mtime, err := s.ReadTime()
	if err != nil {
		err = util.WrapErrf(err, "error reading file time: %v", err)
		return
	}

	// adb doesn't indicate when a file doesn't exist, but will return all zeros.
	// Theoretically this could be an actual file, but that's very unlikely.
	if mode == os.FileMode(0) && size == 0 && mtime == zeroTime {
		return nil, util.Errorf(util.FileNoExistError, "file doesn't exist")
	}

	entry = &DirEntry{
		Mode:       mode,
		Size:       size,
		ModifiedAt: mtime,
	}
	return
}
