package grpcx

import (
	"google.golang.org/grpc"

	"github.com/oligarch316/go-netx/multi/addrsort"
)

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

// WithResolveAddressOrder TODO.
func WithResolveAddressOrder(cmps ...addrsort.Comparer) DialerOption {
	return func(p *DialerParams) { p.Resolver.AddressOrder = cmps }
}
