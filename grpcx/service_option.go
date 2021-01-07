package grpcx

import "google.golang.org/grpc"

// ServiceOption TODO.
type ServiceOption func(*ServiceParams)

// WithGRPCServerOptions TODO.
func WithGRPCServerOptions(opts ...grpc.ServerOption) ServiceOption {
	return func(p *ServiceParams) { p.GRPCServerOptions = append(p.GRPCServerOptions, opts...) }
}

// WithHandlers TODO.
func WithHandlers(hs ...Handler) ServiceOption {
	return func(p *ServiceParams) { p.Handlers = append(p.Handlers, hs...) }
}

// WithHandlerFuncs TODO.
func WithHandlerFuncs(fs ...func(*grpc.Server)) ServiceOption {
	hs := make([]Handler, len(fs))
	for i, f := range fs {
		hs[i] = HandlerFunc(f)
	}
	return WithHandlers(hs...)
}
