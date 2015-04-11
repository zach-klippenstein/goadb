/*
The wire package implements the low-level part of the client/server wire protocol.

The protocol spec can be found at
https://android.googlesource.com/platform/system/core/+/master/adb/OVERVIEW.TXT.

For most cases, usage looks something like:
	conn := wire.Dial()
	conn.SendMessage(data)
	conn.ReadStatus() == "OKAY" || "FAIL"
	conn.ReadMessage()
	conn.Close()

For some messages, the server will return more than one message (but still a single
status). Generally, after calling ReadStatus once, you should call ReadMessage until
it returns an io.EOF error.

For most commands, the server will close the connection after sending the response.
You should still always call Close() when you're done with the connection.
*/
package wire

import (
	"fmt"
	"io"
	"net"
)

const (
	// Default port the adb server listens on.
	AdbPort = 5037

	// The official implementation of adb imposes an undocumented 255-byte limit
	// on messages.
	MaxMessageLength = 255
)

// Conn is a connection to an adb server.
type Conn struct {
	Scanner
	Sender
	io.Closer
}

// Dial connects to the adb server on the default port, AdbPort.
func Dial() (*Conn, error) {
	return DialPort(AdbPort)
}

// Dial connects to the adb server on port.
func DialPort(port int) (*Conn, error) {
	return DialAddr(fmt.Sprintf("localhost:%d", port))
}

// Dial connects to the adb server at address.
func DialAddr(address string) (*Conn, error) {
	netConn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	return &Conn{
		Scanner: NewScanner(netConn),
		Sender:  NewSender(netConn),
		Closer:  netConn,
	}, nil
}

var _ io.Closer = &Conn{}
