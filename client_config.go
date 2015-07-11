package goadb

var (
	defaultDialer Dialer = NewDialer("", 0)
)

type ClientConfig struct {
	Dialer Dialer
}

func (c ClientConfig) sanitized() ClientConfig {
	if c.Dialer == nil {
		c.Dialer = defaultDialer
	}
	return c
}
