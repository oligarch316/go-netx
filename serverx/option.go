package serverx

import (
	"github.com/oligarch316/go-netx"
	"github.com/oligarch316/go-netx/servicex"
)

// WithListeners TODO.
func WithListeners(id servicex.ID, ls ...netx.Listener) Option {
	return func(p *Params) { p.appendListeners(id, ls...) }
}

// WithDependencies TODO.
func WithDependencies(id servicex.ID, deps ...servicex.ID) Option {
	return func(p *Params) { p.appendDependencies(id, deps...) }
}

// WithIgnoreAll TODO.
func WithIgnoreAll() Option {
	return func(p *Params) {
		p.IgnoreParams = IgnoreParams{
			IgnoreMissingListeners:    true,
			IgnoreDuplicateServices:   true,
			IgnoreMissingDependencies: true,
		}
	}
}
