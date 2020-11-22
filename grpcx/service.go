package grpcx

import (
    "context"
    "errors"
    "net"
    "sync/atomic"

    "github.com/oligarch316/go-netx/serverx"
    "google.golang.org/grpc"
)

type serviceID struct{}

var (
    // ServiceID TODO.
    ServiceID serverx.ServiceID = serviceID{}

    errMultipleServe = errors.New("serve function called multiple times")
)

// Service TODO.
type Service struct {
    Handlers []Handler
    GRPCServerOptions []grpc.ServerOption

    svr *grpc.Server
    serveFlag uint32
}

// NewService TODO.
func NewService(opts ...Option) *Service {
    res := new(Service)
    for _, opt := range opts {
        opt(res)
    }
    return res
}

// ID TODO.
func (s Service) ID() serverx.ServiceID { return ServiceID }

// Serve TODO.
func (s *Service) Serve(l net.Listener) error {
    if !atomic.CompareAndSwapUint32(&s.serveFlag, 0, 1) {
        return errMultipleServe
    }

    s.svr = grpc.NewServer(s.GRPCServerOptions...)

    for _, h := range s.Handlers {
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
