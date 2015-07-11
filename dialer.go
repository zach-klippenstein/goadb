package goadb

import (
	"fmt"
	"net"
	"runtime"

	"github.com/zach-klippenstein/goadb/wire"
)

const (
	// Default port the adb server listens on.
	AdbPort = 5037
)

/*
Dialer knows how to create connections to an adb server.
*/
type Dialer interface {
	Dial() (*wire.Conn, error)
}

/*
NewDialer creates a new Dialer.

If host is "" or port is 0, "localhost:5037" is used.
*/
func NewDialer(host string, port int) Dialer {
	if host == "" {
		host = "localhost"
	}
	if port == 0 {
		port = AdbPort
	}
	return &netDialer{host, port}
}

type netDialer struct {
	Host string
	Port int
}

func (d *netDialer) String() string {
	return fmt.Sprintf("netDialer(%s:%d)", d.Host, d.Port)
}

// Dial connects to the adb server on the host and port set on the netDialer.
// The zero-value will connect to the default, localhost:5037.
func (d *netDialer) Dial() (*wire.Conn, error) {
	host := d.Host
	port := d.Port

	netConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}

	conn := &wire.Conn{
		Scanner: wire.NewScanner(netConn),
		Sender:  wire.NewSender(netConn),
	}

	// Prevent leaking the network connection, not sure if TCPConn does this itself.
	runtime.SetFinalizer(netConn, func(conn *net.TCPConn) {
		conn.Close()
	})

	return conn, nil
}

// TODO(zach): Make this unexported.
func roundTripSingleResponse(d Dialer, req string) ([]byte, error) {
	conn, err := d.Dial()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return conn.RoundTripSingleResponse([]byte(req))
}
