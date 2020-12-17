package grpcx

import (
	"fmt"

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

// Error TODO.
type Error struct {
	Component string
	err error
}

func (e Error) Error() string {
	return fmt.Sprintf("%s %s: %s", namespace, e.Component, e.err.Error())
}

// Unwrap TODO.
func (e Error) Unwrap() error {
	return e.err
}

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
