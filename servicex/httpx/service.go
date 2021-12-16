package httpx

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync/atomic"

	"github.com/oligarch316/go-netx"
)

type namespace struct{}

func (n namespace) String() string { return "httpx" }

// ID TODO.
var ID netx.ServiceID = namespace{}

var errServiceClosed = errors.New("httpx: service closed")

// ServiceOption TODO.
type ServiceOption func(*ServiceParams)

// ServiceParams TODO.
type ServiceParams struct {
	// TODO: observ/logging/event handling injection

	HTTPServerOptions []func(*http.Server)
}

func (sp ServiceParams) build() *http.Server {
	res := new(http.Server)
	for _, opt := range sp.HTTPServerOptions {
		opt(res)
	}
	return res
}

// Service TODO.
type Service struct {
	svr       *http.Server
	closeFlag uint32
}

// NewService TODO.
func NewService(opts ...ServiceOption) *Service {
	var params ServiceParams
	for _, opt := range opts {
		opt(&params)
	}
	return &Service{svr: params.build()}
}

// ID TODO.
func (Service) ID() netx.ServiceID { return ID }

// Serve TODO.
func (s *Service) Serve(l net.Listener) error {
	if atomic.LoadUint32(&s.closeFlag) != 0 {
		return errServiceClosed
	}

	if err := s.svr.Serve(l); err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Close TODO.
func (s *Service) Close(ctx context.Context) error {
	atomic.StoreUint32(&s.closeFlag, 1)

	if err := s.svr.Shutdown(ctx); err != nil {
		// TODO: Log/track/surface this err somehow

		return s.svr.Close()
	}

	return nil
}
