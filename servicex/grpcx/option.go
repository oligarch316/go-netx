package grpcx

import (
	"github.com/oligarch316/go-netx"
	"github.com/oligarch316/go-netx/multi/addrsort"
	"github.com/oligarch316/go-netx/serverx"
	"google.golang.org/grpc"
)

// ----- Server Options

// WithListeners TODO.
func WithListeners(ls ...netx.Listener) serverx.Option {
	return serverx.WithListeners(ID, ls...)
}

// WithDependencies TODO.
func WithDependencies(deps ...netx.ServiceID) serverx.Option {
	return serverx.WithDependencies(ID, deps...)
}

// ----- Service Options

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

// ----- Dialer Options

// WithGRPCDialOptions TODO.
func WithGRPCDialOptions(opts ...grpc.DialOption) DialerOption {
	return func(p *DialerParams) { p.GRPCDialOptions = append(p.GRPCDialOptions, opts...) }
}

// WithResolveNoScheme TODO.
func WithResolveNoScheme(p *DialerParams) { p.Resolver.SchemeName = nil }

// WithResolveNoDNSHost TODO.
func WithResolveNoDNSHost(p *DialerParams) { p.Resolver.DNSHostName = nil }

// WithResolveSchemeName TODO.
func WithResolveSchemeName(name string) DialerOption {
	return func(p *DialerParams) { p.Resolver.SchemeName = &name }
}

// WithResolveDNSHostName TODO.
func WithResolveDNSHostName(name string) DialerOption {
	return func(p *DialerParams) { p.Resolver.DNSHostName = &name }
}

// WithResolveAddressOrder TODO.
func WithResolveAddressOrder(cmps ...addrsort.Comparer) DialerOption {
	return func(p *DialerParams) { p.Resolver.AddressOrder = cmps }
}
