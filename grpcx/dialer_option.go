package grpcx

import "google.golang.org/grpc"

// DialerOption TODO.
type DialerOption func(*DialerParams)

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

// WithResolveNetworkPriority TODO.
func WithResolveNetworkPriority(networks ...string) DialerOption {
	return func(p *DialerParams) { p.Resolver.NetworkPriority = networks }
}
