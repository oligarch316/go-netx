package grpcx

import (
	"github.com/oligarch316/go-netx"
	"github.com/oligarch316/go-netx/serverx"
	"google.golang.org/grpc"
)

// ----- serverx.Server Options

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

// ----- Service Options

// Option TODO.
type Option func(*Service)

// WithHandlers TODO.
func WithHandlers(hs ...Handler) Option {
	return func(s *Service) { s.Handlers = append(s.Handlers, hs...) }
}

// WithHandlerFuncs TODO.
func WithHandlerFuncs(fs ...func(*grpc.Server)) Option {
	var hs []Handler
	for _, f := range fs {
		hs = append(hs, HandlerFunc(f))
	}

	return WithHandlers(hs...)
}

// WithGRPCServerOptions TODO.
func WithGRPCServerOptions(opts ...grpc.ServerOption) Option {
	return func(s *Service) { s.GRPCServerOptions = append(s.GRPCServerOptions, opts...) }
}
