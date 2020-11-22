package serverx

import (
	"context"
	"net"
)

// ServiceID TODO.
type ServiceID interface{}

// Service TODO.
type Service interface {
	ID() ServiceID
	Serve(net.Listener) error
	Close(context.Context) error
}

type serviceRunner struct {
	Service
	l net.Listener
}

func (sr serviceRunner) Run() error { return sr.Serve(sr.l) }
