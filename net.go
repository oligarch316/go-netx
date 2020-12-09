package netx

import (
	"context"
	"net"

	"google.golang.org/grpc/test/bufconn"
)

const addrInternal abstractAddr = "internal"

type abstractAddr string

func (aa abstractAddr) Network() string { return string(aa) }

func (aa abstractAddr) String() string { return string(aa) }

// Listener TODO.
type Listener interface {
	net.Listener
	Dial() (net.Conn, error)
	DialContext(context.Context) (net.Conn, error)
}

// BasicListener TODO.
type BasicListener struct {
	net.Listener
	dialer net.Dialer
}

// Dial TODO.
func (bl *BasicListener) Dial() (net.Conn, error) {
	addr := bl.Addr()
	return bl.dialer.Dial(addr.Network(), addr.String())
}

// DialContext TODO.
func (bl *BasicListener) DialContext(ctx context.Context) (net.Conn, error) {
	addr := bl.Addr()
	return bl.dialer.DialContext(ctx, addr.Network(), addr.String())
}

// InternalListener TODO.
type InternalListener struct{ *bufconn.Listener }

// Addr TODO.
func (il InternalListener) Addr() net.Addr { return addrInternal }

// DialContext TODO.
func (il InternalListener) DialContext(_ context.Context) (net.Conn, error) {
	// TODO: Ignoring context here is unmannerly, will be finicky to implement correctly though

	return il.Dial()
}
