package httpx

import (
	"github.com/oligarch316/go-netx"
	"github.com/oligarch316/go-netx/internal/servicex"
	"github.com/oligarch316/go-netx/serverx"
)

var id = servicex.ID{Namespace: "httpx"}

// ID TODO.
var ID serverx.ServiceID = id

// WithListeners TODO.
func WithListeners(ls ...netx.Listener) serverx.Option {
	return id.WithListeners(ls)
}

// WithDependencies TODO.
func WithDependencies(svcIDs ...serverx.ServiceID) serverx.Option {
	return id.WithDependencies(svcIDs)
}
