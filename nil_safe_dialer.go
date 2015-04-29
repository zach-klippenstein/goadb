package goadb

import "github.com/zach-klippenstein/goadb/wire"

type nilSafeDialer struct {
	wire.Dialer
}

func (d nilSafeDialer) Dial() (*wire.Conn, error) {
	if d.Dialer == nil {
		d.Dialer = wire.NewDialer("", AdbPort)
	}

	return d.Dialer.Dial()
}
