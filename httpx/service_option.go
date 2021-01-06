package httpx

import "net/http"

// ServiceOption TODO.
type ServiceOption func(*ServiceParams)

// WithHTTPServerOptions TODO.
func WithHTTPServerOptions(opts ...func(*http.Server)) ServiceOption {
	return func(p *ServiceParams) { p.HTTPServerOptions = append(p.HTTPServerOptions, opts...) }
}

// MuxHandler TODO.
type MuxHandler interface{ Register(*http.ServeMux) }

// MuxHandlerFunc TODO.
type MuxHandlerFunc func(*http.ServeMux)

// Register TODO.
func (mhf MuxHandlerFunc) Register(mux *http.ServeMux) { mhf(mux) }

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
