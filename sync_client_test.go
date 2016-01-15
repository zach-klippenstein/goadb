// TODO(z): Implement tests for sync_client functions.
package adb

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zach-klippenstein/goadb/util"
	"github.com/zach-klippenstein/goadb/wire"
)

var someTime = time.Date(2015, 5, 3, 8, 8, 8, 0, time.UTC)

func TestStatValid(t *testing.T) {
	var buf bytes.Buffer
	conn := &wire.SyncConn{wire.NewSyncScanner(&buf), wire.NewSyncSender(&buf)}

	var mode os.FileMode = 0777

	conn.SendOctetString("STAT")
	conn.SendFileMode(mode)
	conn.SendInt32(4)
	conn.SendTime(someTime)

	entry, err := stat(conn, "/thing")
	assert.NoError(t, err)
	require.NotNil(t, entry)
	assert.Equal(t, mode, entry.Mode, "expected os.FileMode %s, got %s", mode, entry.Mode)
	assert.Equal(t, int32(4), entry.Size)
	assert.Equal(t, someTime, entry.ModifiedAt)
	assert.Equal(t, "", entry.Name)
}

func TestStatBadResponse(t *testing.T) {
	var buf bytes.Buffer
	conn := &wire.SyncConn{wire.NewSyncScanner(&buf), wire.NewSyncSender(&buf)}

	conn.SendOctetString("SPAT")

	entry, err := stat(conn, "/")
	assert.Nil(t, entry)
	assert.Error(t, err)
}

func TestStatNoExist(t *testing.T) {
	var buf bytes.Buffer
	conn := &wire.SyncConn{wire.NewSyncScanner(&buf), wire.NewSyncSender(&buf)}

	conn.SendOctetString("STAT")
	conn.SendFileMode(0)
	conn.SendInt32(0)
	conn.SendTime(time.Unix(0, 0).UTC())

	entry, err := stat(conn, "/")
	assert.Nil(t, entry)
	assert.Equal(t, util.FileNoExistError, err.(*util.Err).Code)
}
