package netx

import (
	"context"
	"net"
)

// Dialer TODO.
type Dialer interface {
	Dial() (net.Conn, error)
	DialContext(context.Context) (net.Conn, error)
}

// Listener TODO.
type Listener interface {
	Dialer
	net.Listener
}

// Listen TODO.
func Listen(network, address string) (Listener, error) {
	if network == InternalNetwork {
		return NewInternal(internalDefaultSize), nil
	}

	l, err := net.Listen(network, address)
	return New(l), err
}

type basicListener struct {
	net.Listener
	dialer net.Dialer
}

// New TODO.
func New(l net.Listener) Listener {
	return &basicListener{Listener: l}
}

func (bl *basicListener) Dial() (net.Conn, error) {
	addr := bl.Addr()
	return bl.dialer.Dial(addr.Network(), addr.String())
}

func (bl *basicListener) DialContext(ctx context.Context) (net.Conn, error) {
	addr := bl.Addr()
	return bl.dialer.DialContext(ctx, addr.Network(), addr.String())
}
