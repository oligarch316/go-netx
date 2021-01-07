package grpcx

import (
	"context"
	"errors"
	"net"
	"sync/atomic"

	"github.com/oligarch316/go-netx/serverx"
	"google.golang.org/grpc"
)

var errServiceClosed = errors.New("grpcx: service closed")

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

func (sp ServiceParams) build() *grpc.Server {
	res := grpc.NewServer(sp.GRPCServerOptions...)
	for _, h := range sp.Handlers {
		h.Register(res)
	}
	return res
}

// Service TODO.
type Service struct {
	svr       *grpc.Server
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
func (s Service) ID() serverx.ServiceID { return ID }

// Serve TODO.
func (s *Service) Serve(l net.Listener) error {
	if atomic.LoadUint32(&s.closeFlag) != 0 {
		return errServiceClosed
	}

	return s.svr.Serve(l)
}

// Close TODO.
func (s *Service) Close(ctx context.Context) error {
	atomic.StoreUint32(&s.closeFlag, 1)

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
