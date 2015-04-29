package wire

import (
	"fmt"
	"net"
	"runtime"
)

/*
Dialer knows how to create connections to an adb server.
*/
type Dialer interface {
	Dial() (*Conn, error)
}

type netDialer struct {
	Host string
	Port int
}

func NewDialer(host string, port int) Dialer {
	return &netDialer{host, port}
}

// Dial connects to the adb server on the host and port set on the netDialer.
// The zero-value will connect to the default, localhost:5037.
func (d *netDialer) Dial() (*Conn, error) {
	host := d.Host
	if host == "" {
		host = "localhost"
	}

	netConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, d.Port))
	if err != nil {
		return nil, err
	}

	conn := &Conn{
		Scanner: NewScanner(netConn),
		Sender:  NewSender(netConn),
	}

	// Prevent leaking the network connection, not sure if TCPConn does this itself.
	runtime.SetFinalizer(netConn, func(conn *net.TCPConn) {
		conn.Close()
	})

	return conn, nil
}

func RoundTripSingleResponse(d Dialer, req string) ([]byte, error) {
	conn, err := d.Dial()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return conn.RoundTripSingleResponse([]byte(req))
}
