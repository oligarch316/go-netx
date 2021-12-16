package listenerx

import (
	"context"
	"net"

	"github.com/oligarch316/go-netx"
)

type basicListener struct {
	net.Listener
	dialer net.Dialer
}

// NewBasic TODO.
func NewBasic(l net.Listener) netx.Listener {
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
