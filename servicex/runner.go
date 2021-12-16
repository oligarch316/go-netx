package servicex

import (
	"context"
	"net"

	"github.com/oligarch316/go-netx"
)

// Runner TODO.
type Runner struct {
	listener net.Listener
	service  netx.Service
}

// NewRunner TODO.
func NewRunner(l net.Listener, svc netx.Service) *Runner {
	return &Runner{
		listener: l,
		service:  svc,
	}
}

// Addr TODO.
func (r Runner) Addr() net.Addr { return r.listener.Addr() }

// ID TODO.
func (r Runner) ID() netx.ServiceID { return r.service.ID() }

// Run TODO.
func (r *Runner) Run() error { return r.service.Serve(r.listener) }

// Close TODO.
func (r *Runner) Close(ctx context.Context) error { return r.service.Close(ctx) }
