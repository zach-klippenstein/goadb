package goadb

import (
	"fmt"
	"io"
	"net"
	"runtime"

	"github.com/zach-klippenstein/goadb/util"
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

	address := fmt.Sprintf("%s:%d", host, port)
	netConn, err := net.Dial("tcp", address)
	if err != nil {
		// Attempt to start the server and try again.
		if err = StartServer(); err != nil {
			return nil, util.WrapErrorf(err, util.ServerNotAvailable, "error starting server")
		}

		address = fmt.Sprintf("%s:%d", host, port)
		netConn, err = net.Dial("tcp", address)
		if err != nil {
			return nil, util.WrapErrorf(err, util.ServerNotAvailable, "error dialing %s", address)
		}
	}

	// net.Conn can't be closed more than once, but wire.Conn will try to close both sender and scanner
	// so we need to wrap it to make it safe.
	safeConn := wire.MultiCloseable(netConn)

	// Prevent leaking the network connection, not sure if TCPConn does this itself.
	// Note that the network connection may still be in use after the conn isn't (scanners/senders
	// can give their underlying connections to other scanner/sender types), so we can't
	// set the finalizer on conn.
	runtime.SetFinalizer(safeConn, func(conn io.ReadWriteCloser) {
		conn.Close()
	})

	return &wire.Conn{
		Scanner: wire.NewScanner(safeConn),
		Sender:  wire.NewSender(safeConn),
	}, nil
}

func roundTripSingleResponse(d Dialer, req string) ([]byte, error) {
	conn, err := d.Dial()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return conn.RoundTripSingleResponse([]byte(req))
}
