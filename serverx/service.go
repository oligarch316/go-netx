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
		servicesWG  = runner.NewWaitGroup(len(s.dependants))
		listenersWG = runner.NewWaitGroup(s.ml.Len())

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
				listenersWG.Wait()
				return s.svc.Close(ctx)
			},
		)

		res = []runner.Item{servicesWG, listenersWG, serviceRunner}
	)

	for _, item := range s.ml.Runners() {
		listenRunner := runner.New(
			// Run() error
			func() error {
				defer listenersWG.Done()
				return item.Run()
			},

			// Close(context.Context) error
			func(ctx context.Context) error {
				servicesWG.Wait()
				return item.Close(ctx)
			},
		)

		res = append(res, listenRunner)
	}

	s.signal = servicesWG.Done

	return res
}

func (s *service) signalDone() error {
	if s.signal != nil {
		s.signal()
		return nil
	}

	return fmt.Errorf("serverx: signalDone() called on uninitialized service (%s)", s.svc.ID())
}
