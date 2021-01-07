package httpx

import (
	"net/http"

	"github.com/oligarch316/go-netx/multi/addrsort"
)

// TransportOption TODO.
type TransportOption func(*TransportParams)

// WithHTTPTransportOptions TODO.
func WithHTTPTransportOptions(opts ...func(*http.Transport)) TransportOption {
	return func(p *TransportParams) { p.HTTPTransportOptions = append(p.HTTPTransportOptions, opts...) }
}

// WithResolveNoScheme TODO.
func WithResolveNoScheme(p *TransportParams) { p.SchemeName = nil }

// WithResolveNoTLSScheme TODO.
func WithResolveNoTLSScheme(p *TransportParams) { p.SchemeTLSName = nil }

// WithResolveNoHost TODO.
func WithResolveNoHost(p *TransportParams) { p.HostName = nil }

// WithResolveSchemeName TODO.
func WithResolveSchemeName(name string) TransportOption {
	return func(p *TransportParams) { p.SchemeName = &name }
}

// WithResolveTLSSchemeName TODO.
func WithResolveTLSSchemeName(name string) TransportOption {
	return func(p *TransportParams) { p.SchemeTLSName = &name }
}

// WithResolveHostName TODO.
func WithResolveHostName(name string) TransportOption {
	return func(p *TransportParams) { p.HostName = &name }
}

// WithResolveAddressOrder TODO.
func WithResolveAddressOrder(cmps ...addrsort.Comparer) TransportOption {
	return func(p *TransportParams) { p.AddressOrder = cmps }
}
