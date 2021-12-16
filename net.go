package netx

import (
	"context"
	"fmt"
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

// ServiceID TODO.
type ServiceID fmt.Stringer

// Service TODO.
type Service interface {
	ID() ServiceID
	Serve(net.Listener) error
	Close(context.Context) error
}
