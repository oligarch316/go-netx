package servicex

import (
	"github.com/oligarch316/go-netx"
	"github.com/oligarch316/go-netx/serverx"
)

// ID TODO.
type ID struct{ Namespace string }

func (id ID) String() string { return id.Namespace }

// WithListeners TODO.
func (id ID) WithListeners(ls []netx.Listener) serverx.Option {
	return func(p *serverx.Params) error {
		p.AddListeners(id, ls...)
		return nil
	}
}

// WithDependencies TODO.
func (id ID) WithDependencies(svcIDs []serverx.ServiceID) serverx.Option {
	return func(p *serverx.Params) error {
		p.AddDependencies(id, svcIDs...)
		return nil
	}
}
