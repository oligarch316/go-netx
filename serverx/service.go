package serverx

import (
	"context"
	"fmt"
	"net"

	"github.com/oligarch316/go-netx/multi"
	"github.com/oligarch316/go-netx/runner"
)

// ServiceID TODO.
type ServiceID fmt.Stringer

// Service TODO.
type Service interface {
	ID() ServiceID
	Serve(net.Listener) error
	Close(context.Context) error
}

type service struct {
	svc Service
	ml  *multi.Listener

	dependants, dependsOn []*service
	signal                func()
}

func newService(svc Service, ml *multi.Listener) *service {
	return &service{svc: svc, ml: ml}
}

func (s *service) DependOn(svc *service) {
	s.dependsOn = append(s.dependsOn, svc)
	svc.dependants = append(svc.dependants, s)
}

func (s *service) Runners() []runner.Item {
	var (
		servicesSem  = runner.NewSemaphore(len(s.dependants))
		listenersSem = runner.NewSemaphore(s.ml.Len())

		serviceRunner = runner.New(
			// Run() error
			func() error {
				defer func() {
					for _, svc := range s.dependsOn {
						// TODO: Log/track/surface the possible error here somehow
						svc.signalDone()
					}
				}()
				return s.svc.Serve(s.ml)
			},

			// Close(context.Context) error
			func(ctx context.Context) error {
				listenersSem.Wait()
				return s.svc.Close(ctx)
			},
		)

		res = []runner.Item{servicesSem, listenersSem, serviceRunner}
	)

	for _, item := range s.ml.Runners() {
		listenRunner := runner.New(
			// Run() error
			func() error {
				defer listenersSem.Done()
				return item.Run()
			},

			// Close(context.Context) error
			func(ctx context.Context) error {
				servicesSem.Wait()
				return item.Close(ctx)
			},
		)

		res = append(res, listenRunner)
	}

	s.signal = servicesSem.Done

	return res
}

func (s *service) signalDone() error {
	if s.signal != nil {
		s.signal()
		return nil
	}

	return fmt.Errorf("signalDone() called on uninitialized service (%s)", s.svc.ID())
}
