package listenerx

import (
	"net"

	"github.com/oligarch316/go-netx"
)

// New TODO.
func New(network, address string) (netx.Listener, error) {
	if network == InternalNetwork {
		return NewInternal(internalDefaultSize), nil
	}

	l, err := net.Listen(network, address)
	return NewBasic(l), err
}
