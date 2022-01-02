package serverx

import (
	"github.com/oligarch316/go-netx"
	"github.com/oligarch316/go-netx/listenerx/multi"
)

// WithListeners TODO.
func WithListeners(id netx.ServiceID, ls ...netx.Listener) Option {
	return func(p *Params) { p.Services.AppendListeners(id, ls...) }
}

// WithListenerOpts TODO.
func WithListenerOpts(id netx.ServiceID, opts ...multi.ListenerOption) Option {
	return func(p *Params) { p.Services.AppendListenerOpts(id, opts...) }
}

// WithDependencies TODO.
func WithDependencies(id netx.ServiceID, deps ...netx.ServiceID) Option {
	return func(p *Params) { p.Services.AppendDependencies(id, deps...) }
}

// WithIgnoreAll TODO.
func WithIgnoreAll() Option {
	return func(p *Params) {
		p.Ignore = IgnoreParams{
			DuplicateServices:   true,
			MissingDependencies: true,
			MissingListeners:    true,
		}
	}
}
