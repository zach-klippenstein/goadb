package wire

import "github.com/zach-klippenstein/goadb/internal/errors"

const (
	// The official implementation of adb imposes an undocumented 1-megabyte limit
	// on payload size.
	MaxPayloadSize = 1024 * 1024
)

/*
Conn is a normal connection to an adb server.

For most cases, usage looks something like:
	conn := wire.Dial()
	conn.SendMessage(data)
	conn.ReadStatus() == StatusSuccess || StatusFailure
	conn.ReadMessage()
	conn.Close()

For some messages, the server will return more than one message (but still a single
status). Generally, after calling ReadStatus once, you should call ReadMessage until
it returns an io.EOF error. Note: the protocol docs seem to suggest that connections will be
kept open for multiple commands, but this is not the case. The official client closes
a connection immediately after its read the response, in most cases. The docs might be
referring to the connection between the adb server and the device, but I haven't confirmed
that.

For most commands, the server will close the connection after sending the response.
You should still always call Close() when you're done with the connection.
*/
type Conn struct {
	Scanner
	Sender
}

func NewConn(scanner Scanner, sender Sender) *Conn {
	return &Conn{scanner, sender}
}

// NewSyncConn returns connection that can operate in sync mode.
// The connection must already have been switched (by sending the sync command
// to a specific device), or the return connection will return an error.
func (c *Conn) NewSyncConn() *SyncConn {
	return &SyncConn{
		SyncScanner: c.Scanner.NewSyncScanner(),
		SyncSender:  c.Sender.NewSyncSender(),
	}
}

// RoundTripSingleResponse sends a message to the server, and reads a single
// message response. If the reponse has a failure status code, returns it as an error.
func (conn *Conn) RoundTripSingleResponse(req []byte) (resp []byte, err error) {
	if err = conn.SendMessage(req); err != nil {
		return nil, err
	}

	if _, err = conn.ReadStatus(string(req)); err != nil {
		return nil, err
	}

	return conn.ReadMessage()
}

func (conn *Conn) Close() error {
	errs := struct {
		SenderErr  error
		ScannerErr error
	}{
		SenderErr:  conn.Sender.Close(),
		ScannerErr: conn.Scanner.Close(),
	}

	if errs.ScannerErr != nil || errs.SenderErr != nil {
		return &errors.Err{
			Code:    errors.NetworkError,
			Message: "error closing connection",
			Details: errs,
		}
	}
	return nil
}
