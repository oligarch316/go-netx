package serverx

import (
	"context"
	"sync"

	"github.com/oligarch316/go-netx"
	"github.com/oligarch316/go-netx/runner"
)

// Option TODO.
type Option func(*Server) error

// Server TODO.
type Server struct {
	listenerMap map[ServiceID]*netx.MultiListener

    listenGroup, serviceGroup *runner.Group

    serveOnce, closeOnce sync.Once

	listenersDone chan struct{}
	resultChan    chan error
}

// NewServer TODO.
func NewServer(opts ...Option) (*Server, error) {
	res := &Server{
		listenerMap:   make(map[ServiceID]*netx.MultiListener),

        listenGroup: runner.NewGroup(),
        serviceGroup: runner.NewGroup(),

		listenersDone: make(chan struct{}),
		resultChan:    make(chan error),
	}

	for _, opt := range opts {
		if err := opt(res); err != nil {
			return nil, err
		}
	}

	return res, nil
}

// TODO (Thought): If we made a whole separate Params type and did our appending
// through that, we could ensure appending only occurs during NewServer and
// avoid worries of appending after Serve()/Close() is called

// AppendListeners TODO.
func (s *Server) AppendListeners(svcID ServiceID, ls ...netx.Listener) {
	if ml, ok := s.listenerMap[svcID]; ok {
		ml.Append(ls...)
		return
	}

	s.listenerMap[svcID] = netx.NewMultiListener(ls...)
}

// Close TODO.
func (s *Server) Close(ctx context.Context) {
    // TODO: This once.Do is very likely unnecessary
	s.closeOnce.Do(func() { s.close(ctx) })
}

// Serve TODO.
func (s *Server) Serve(svcs ...Service) <-chan error {
	s.serveOnce.Do(func() { s.serve(svcs) })
	return s.resultChan
}

// Name this run instead?
func (s *Server) serve(svcs []Service) {
	for _, svc := range svcs {
		ml, ok := s.listenerMap[svc.ID()]
		if !ok {
			// TODO: Log/track/surface this somehow
			continue
		}

		s.listenGroup.Append(ml.Runners()...)
		s.serviceGroup.Append(serviceRunner{
			Service: svc,
			l:       ml,
		})
	}

	go s.report()
    go s.serviceGroup.Run()
	go s.listenGroup.Run()
}

func (s *Server) fanIn(src <-chan error) {
	for err := range src {
		if err != nil {
			s.resultChan <- err
		}
	}
}

func (s *Server) report() {
    // Consume listen runner results until complete
    s.fanIn(s.listenGroup.Results())

    // Signal listening completed
    close(s.listenersDone)

    // Consume service runner results until complete
    s.fanIn(s.serviceGroup.Results())

    // Signal serve completed
    close(s.resultChan)
}

func (s *Server) close(ctx context.Context) {
    // Trigger close of listen runners
    s.listenGroup.Close(ctx)

    // Wait for 1st of...
    select {
    case <-s.listenersDone:
        // ...listen runner completion
    case <-ctx.Done():
        // ...context expiration

        // TODO: Log/track/surface this somehow
    }

    // Trigger close of service runners
    s.serviceGroup.Close(ctx)
}
