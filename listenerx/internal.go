package listenerx

import (
	"context"
	"net"

	"github.com/oligarch316/go-netx"
	"google.golang.org/grpc/test/bufconn"
)

const (
	// InternalNetwork TODO.
	InternalNetwork = "internal"

	internalDefaultSize = 256
)

type internalAddr struct{}

func (internalAddr) Network() string { return InternalNetwork }
func (internalAddr) String() string  { return InternalNetwork }

type internalListener struct{ *bufconn.Listener }

// NewInternal TODO.
func NewInternal(size int) netx.Listener {
	return &internalListener{Listener: bufconn.Listen(size)}
}

func (internalListener) Addr() net.Addr { return internalAddr{} }

func (li internalListener) DialContext(_ context.Context) (net.Conn, error) {
	// TODO: Ignoring context here is unmannerly, will be finicky to implement correctly though

	return li.Dial()
}
