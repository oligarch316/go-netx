package httpx

import (
	"net/http"

	"github.com/oligarch316/go-netx"
	"github.com/oligarch316/go-netx/listenerx/multi/addrsort"
	"github.com/oligarch316/go-netx/serverx"
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

// MuxHandler TODO.
type MuxHandler interface{ Register(*http.ServeMux) }

// MuxHandlerFunc TODO.
type MuxHandlerFunc func(*http.ServeMux)

// Register TODO.
func (mhf MuxHandlerFunc) Register(mux *http.ServeMux) { mhf(mux) }

// WithHTTPServerOptions TODO.
func WithHTTPServerOptions(opts ...func(*http.Server)) ServiceOption {
	return func(p *ServiceParams) { p.HTTPServerOptions = append(p.HTTPServerOptions, opts...) }
}

// WithMuxHandlers TODO.
func WithMuxHandlers(mhs ...MuxHandler) ServiceOption {
	mux := http.NewServeMux()
	for _, mh := range mhs {
		mh.Register(mux)
	}
	return WithHTTPServerOptions(func(s *http.Server) { s.Handler = mux })
}

// WithMuxHandlerFuncs TODO.
func WithMuxHandlerFuncs(fs ...func(*http.ServeMux)) ServiceOption {
	mhs := make([]MuxHandler, len(fs))
	for i, f := range fs {
		mhs[i] = MuxHandlerFunc(f)
	}
	return WithMuxHandlers(mhs...)
}

// ----- Transport Options

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
