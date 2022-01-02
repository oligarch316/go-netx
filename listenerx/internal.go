package listenerx

import (
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

func (ia *internalListener) Accept() (net.Conn, error) {
	conn, err := ia.Listener.Accept()
	if err != nil && err.Error() == "closed" {
		err = net.ErrClosed
	}
	return conn, err
}
