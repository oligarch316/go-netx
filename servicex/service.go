package servicex

import (
	"context"
	"fmt"
	"net"

	"github.com/oligarch316/go-netx"
	"github.com/oligarch316/go-netx/multi/addrsort"
)

var (
	// DefaultDialKeyName TODO.
	DefaultDialKeyName = "localapp"

	// DefaultDialNetworkPriority TODO.
	DefaultDialNetworkPriority = addrsort.ByPriorityAddress(netx.InternalNetwork, "unix", "tcp")
)

// ID TODO.
type ID fmt.Stringer

// Service TODO.
type Service interface {
	ID() ID
	Serve(net.Listener) error
	Close(context.Context) error
}

type runner struct {
	listener net.Listener
	service  Service
}

// NewRunner TODO.
func NewRunner(l net.Listener, svc Service) *runner {
	return &runner{
		listener: l,
		service:  svc,
	}
}

func (r *runner) Run() error                      { return r.service.Serve(r.listener) }
func (r *runner) Close(ctx context.Context) error { return r.service.Close(ctx) }

func (r runner) Addr() net.Addr { return r.listener.Addr() }
func (r runner) ID() ID         { return r.service.ID() }
