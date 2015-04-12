// TODO(z): Implement send.
package goadb

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/zach-klippenstein/goadb/wire"
)

/*
DirEntry holds information about a directory entry on a device.

Unfortunately, adb doesn't seem to set the directory bit for directories.
*/
type DirEntry struct {
	Name       string
	Mode       os.FileMode
	Size       int32
	ModifiedAt time.Time
}

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
		return nil, fmt.Errorf("expected stat ID 'STAT', but got '%s'", id)
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

	return &DirEntries{
		scanner:     conn,
		doneHandler: func() { conn.Close() },
	}, nil
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
		err = fmt.Errorf("error reading file mode: %v", err)
		return
	}
	size, err := s.ReadInt32()
	if err != nil {
		err = fmt.Errorf("error reading file size: %v", err)
		return
	}
	mtime, err := s.ReadTime()
	if err != nil {
		err = fmt.Errorf("error reading file time: %v", err)
		return
	}

	entry = &DirEntry{
		Mode:       mode,
		Size:       size,
		ModifiedAt: mtime,
	}
	return
}
