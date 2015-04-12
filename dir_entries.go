package goadb

import (
	"fmt"

	"github.com/zach-klippenstein/goadb/wire"
)

// DirEntries iterates over directory entries.
type DirEntries struct {
	scanner wire.SyncScanner

	// Called when finished iterating (successfully or not).
	doneHandler func()

	currentEntry *DirEntry
	err          error
}

func (entries *DirEntries) Next() bool {
	if entries.err != nil {
		return false
	}

	entry, done, err := readNextDirListEntry(entries.scanner)
	if err != nil {
		entries.err = err
		entries.onDone()
		return false
	}

	entries.currentEntry = entry
	if done {
		entries.onDone()
		return false
	}

	return true
}

func (entries *DirEntries) Entry() *DirEntry {
	return entries.currentEntry
}

func (entries *DirEntries) Err() error {
	return entries.err
}

func (entries *DirEntries) onDone() {
	if entries.doneHandler != nil {
		entries.doneHandler()
	}
}

func readNextDirListEntry(s wire.SyncScanner) (entry *DirEntry, done bool, err error) {
	id, err := s.ReadOctetString()
	if err != nil {
		return
	}

	if id == "DONE" {
		done = true
		return
	} else if id != "DENT" {
		err = fmt.Errorf("expected dir entry ID 'DENT', but got '%s'", id)
		return
	}

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
	name, err := s.ReadString()
	if err != nil {
		err = fmt.Errorf("error reading file name: %v", err)
		return
	}

	done = false
	entry = &DirEntry{
		Name:       name,
		Mode:       mode,
		Size:       size,
		ModifiedAt: mtime,
	}
	return
}
