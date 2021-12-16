package serverx

import "github.com/oligarch316/go-netx"

// WithListeners TODO.
func WithListeners(id netx.ServiceID, ls ...netx.Listener) Option {
	return func(p *Params) { p.appendListeners(id, ls...) }
}

// WithDependencies TODO.
func WithDependencies(id netx.ServiceID, deps ...netx.ServiceID) Option {
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
