package noiseconn

import (
	"net"

	"github.com/flynn/noise"
)

type Listener struct {
	net.Listener
	config noise.Config
	opts   Options
}

var _ net.Listener = (*Listener)(nil)

func NewListener(inner net.Listener, config noise.Config) *Listener {
	return NewListenerWithOptions(inner, config, Options{})
}

func (l *Listener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return NewConnWithOptions(conn, l.config, l.opts)
}

func NewListenerWithOptions(inner net.Listener, config noise.Config, opts Options) *Listener {
	return &Listener{
		Listener: inner,
		config:   config,
		opts:     opts,
	}
}
