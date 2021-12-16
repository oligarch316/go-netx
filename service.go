package netx

import (
	"context"
	"fmt"
	"net"
)

// ServiceID TODO.
type ServiceID fmt.Stringer

// Service TODO.
type Service interface {
	ID() ServiceID
	Serve(net.Listener) error
	Close(context.Context) error
}
