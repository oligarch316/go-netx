package grpcx

import (
	"context"
	"errors"
	"net"
	"sync/atomic"

	"google.golang.org/grpc"
)

var errServiceMultipleServe = errors.New("already running")

type (
	// Handler TODO.
	Handler interface{ Register(*grpc.Server) }

	// HandlerFunc TODO.
	HandlerFunc func(*grpc.Server)
)

// Register TODO.
func (hf HandlerFunc) Register(s *grpc.Server) { hf(s) }

// ServiceParams TODO.
type ServiceParams struct {
	Handlers          []Handler
	GRPCServerOptions []grpc.ServerOption
}

// Service TODO.
type Service struct {
	identity

	params    ServiceParams
	svr       *grpc.Server
	serveFlag uint32
}

// NewService TODO.
func NewService(opts ...ServiceOption) *Service {
	res := new(Service)
	for _, opt := range opts {
		opt(&res.params)
	}
	return res
}

// Serve TODO.
func (s *Service) Serve(l net.Listener) error {
	if !atomic.CompareAndSwapUint32(&s.serveFlag, 0, 1) {
		return Error{ Component: "service", err: errServiceMultipleServe }
	}

	defer func() { s.serveFlag = 0 }()

	s.svr = grpc.NewServer(s.params.GRPCServerOptions...)

	for _, h := range s.params.Handlers {
		h.Register(s.svr)
	}

	return s.svr.Serve(l)
}

// Close TODO.
func (s *Service) Close(ctx context.Context) error {
	doneChan := make(chan struct{})

	go func() {
		s.svr.GracefulStop()
		close(doneChan)
	}()

	select {
	case <-doneChan:
		return nil
	case <-ctx.Done():
		s.svr.Stop()
		return ctx.Err()
	}
}
