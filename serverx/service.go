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
	// Run logic
	svc Service
	ml  *multi.Listener

	// Dependency synchronization
	dependants          []*service // those which depend upon me
	dependees           []*service // those which I depend upon
	dependantDoneSignal func()     // mechanism for dependants to inform me of completion
}

func newService(svc Service, ml *multi.Listener) *service {
	return &service{svc: svc, ml: ml}
}

func (s *service) DependOn(svc *service) {
	s.dependees = append(s.dependees, svc)
	svc.dependants = append(svc.dependants, s)
}

func (s *service) Runners() []runner.Item {
	// ----- "Glue" runners
	// > wait groups for synchronization purposes
	var (
		dependantsWG = runner.NewWaitGroup(len(s.dependants))
		listenersWG  = runner.NewWaitGroup(s.ml.Len())
		res          = []runner.Item{dependantsWG, listenersWG}
	)

	// ----- Service runner
	// > svc.Serve(...) and svc.Close(...) wrapped with glue logic
	signalDependees := func() {
		for _, dependee := range s.dependees {
			// TODO: Log/track/surface the possible error here somehow
			dependee.signalDependantDone()
		}
	}

	serviceRun := func() error {
		defer signalDependees()
		return s.svc.Serve(s.ml)
	}

	serviceClose := func(ctx context.Context) error {
		listenersWG.Wait()
		return s.svc.Close(ctx)
	}

	res = append(res, runner.New(serviceRun, serviceClose))

	// ----- Listener runners
	// > ml.Runners() wrapped with glue logic
	for _, item := range s.ml.Runners() {
		var (
			origRun   = item.Run
			origClose = item.Close
		)

		wrappedRun := func() error {
			defer listenersWG.Done()
			return origRun()
		}

		wrappedClose := func(ctx context.Context) error {
			dependantsWG.Wait()
			return origClose(ctx)
		}

		res = append(res, runner.New(wrappedRun, wrappedClose))
	}

	s.dependantDoneSignal = dependantsWG.Done

	return res
}

func (s *service) signalDependantDone() error {
	if s.dependantDoneSignal != nil {
		s.dependantDoneSignal()
		return nil
	}

	return fmt.Errorf("serverx: signalDone() called on uninitialized service (%s)", s.svc.ID())
}
