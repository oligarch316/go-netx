package httpx

import (
	"fmt"
	"net/http"

	"github.com/oligarch316/go-netx"
	"github.com/oligarch316/go-netx/serverx"
	"github.com/oligarch316/go-netx/servicex"
)

const (
	transportSchemeHTTP  = "http"
	transportSchemeHTTPS = "https"
)

// TransportOption TODO.
type TransportOption func(*TransportParams)

// TransportParams TODO.
type TransportParams struct {
	HostName, SchemeName, SchemeTLSName *string
	HTTPTransportOptions                []func(*http.Transport)
}

func defaultTransportParams() TransportParams {
	schemeName := servicex.DefaultDialKey

	return TransportParams{
		HostName:             nil,
		SchemeTLSName:        nil,
		SchemeName:           &schemeName,
		HTTPTransportOptions: nil,
	}
}

func (tp TransportParams) build() *http.Transport {
	res := http.DefaultTransport.(*http.Transport).Clone()
	for _, opt := range tp.HTTPTransportOptions {
		opt(res)
	}
	return res
}

// Transport TODO.
type Transport interface {
	CloseIdleConnections()
	RegisterProtocol(string, http.RoundTripper)
	RoundTrip(*http.Request) (*http.Response, error)
}

// LoadTransport TODO.
func LoadTransport(svr *serverx.Server, opts ...TransportOption) (Transport, error) {
	dialer, err := svr.Dialer(ID)
	if err != nil {
		return nil, err
	}
	return NewTransport(dialer, opts...), nil
}

// NewTransport TODO.
func NewTransport(dialer netx.Dialer, opts ...TransportOption) Transport {
	params := defaultTransportParams()
	for _, opt := range opts {
		opt(&params)
	}

	var (
		baseTransport = params.build()
		hooks         = dialHooks{
			dial:        baseTransport.Dial,
			dialContext: baseTransport.DialContext,
		}
	)

	baseTransport.DialContext = wrapDialContext(dialer, hooks.DialContext)

	// TODO: Only in go 1.14
	// if baseTransport.DialTLSContext != nil || baseTransport.DialTLS != nil {
	//     tlsHooks := dialHooks{
	//         dial: baseTransport.DialTLS,
	//         dialContext: baseTransport.DialTLSContext,
	//     }
	//
	//     baseTransport.DialTLSContext = wrapDialContext(dialer, hooks.DialContext)
	// }

	var res Transport = baseTransport

	if params.HostName != nil {
		res = &transportHost{
			hostName:  *params.HostName,
			Transport: baseTransport,
		}
	}

	if params.SchemeName != nil {
		res.RegisterProtocol(*params.SchemeName, &transportScheme{
			schemeName:   *params.SchemeName,
			schemeTarget: transportSchemeHTTP,
			Transport:    baseTransport,
		})
	}

	if params.SchemeTLSName != nil {
		res.RegisterProtocol(*params.SchemeTLSName, &transportScheme{
			schemeName:   *params.SchemeTLSName,
			schemeTarget: transportSchemeHTTPS,
			Transport:    baseTransport,
		})
	}

	return res
}

type transportHost struct {
	hostName string
	*http.Transport
}

func (th *transportHost) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host == th.hostName {
		req.URL.Host = dialLocalHostKey
	}
	return th.Transport.RoundTrip(req)
}

type transportScheme struct {
	schemeName   string
	schemeTarget string
	*http.Transport
}

func (ts *transportScheme) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host != "" {
		return nil, fmt.Errorf("httpx: invalid host '%s' for scheme '%s', empty host required", req.URL.Host, ts.schemeName)
	}

	req.URL.Host = dialLocalHostKey
	req.URL.Scheme = ts.schemeTarget

	return ts.Transport.RoundTrip(req)
}
