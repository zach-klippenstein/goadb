package wire

const (
	// The official implementation of adb imposes an undocumented 255-byte limit
	// on messages.
	MaxMessageLength = 255
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
it returns an io.EOF error.

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

	if err = ReadStatusFailureAsError(conn, string(req)); err != nil {
		return nil, err
	}

	return conn.ReadMessage()
}

func (conn *Conn) Close() error {
	if err := conn.Sender.Close(); err != nil {
		return err
	}
	return conn.Scanner.Close()
}
