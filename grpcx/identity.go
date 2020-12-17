package grpcx

import (
	"github.com/oligarch316/go-netx"
	"github.com/oligarch316/go-netx/serverx"
)

const namespace = "grpcx"

type (
	serviceID struct{}
	identity  struct{}
)

func (serviceID) String() string { return namespace }

func (identity) ID() serverx.ServiceID { return ID }

// ID TODO.
var ID serverx.ServiceID = serviceID{}

// WithListeners TODO.
func WithListeners(ls ...netx.Listener) serverx.Option {
	return func(p *serverx.Params) error {
		p.AddListeners(ID, ls...)
		return nil
	}
}

// WithDependencies TODO.
func WithDependencies(svcIDs ...serverx.ServiceID) serverx.Option {
	return func(p *serverx.Params) error {
		p.AddDependencies(ID, svcIDs...)
		return nil
	}
}
