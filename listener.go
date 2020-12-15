package netx

import (
	"context"
	"net"

	"google.golang.org/grpc/test/bufconn"
)

const (
	// NetworkInternal TODO.
	NetworkInternal = "internal"

	internalAddr AbstractAddr = NetworkInternal
	internalSize = 256
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
	if network == NetworkInternal {
		return &listenerInternal{ Listener: bufconn.Listen(internalSize) }, nil
	}

	l, err := net.Listen(network, address)
	return &listenerBasic{ Listener: l }, err
}

type listenerBasic struct {
	net.Listener
	dialer net.Dialer
}

func (lb *listenerBasic) Dial() (net.Conn, error) {
	addr := lb.Addr()
	return lb.dialer.Dial(addr.Network(), addr.String())
}

func (lb *listenerBasic) DialContext(ctx context.Context) (net.Conn, error) {
	addr := lb.Addr()
	return lb.dialer.DialContext(ctx, addr.Network(), addr.String())
}

type listenerInternal struct{ *bufconn.Listener }

func (listenerInternal) Addr() net.Addr { return internalAddr }

func (li listenerInternal) DialContext(_ context.Context) (net.Conn, error) {
	// TODO: Ignoring context here is unmannerly, will be finicky to implement correctly though

	return li.Dial()
}
