package serverx

import (
	"context"
	"fmt"

	"github.com/oligarch316/go-netx"
	"github.com/oligarch316/go-netx/multi"
	"github.com/oligarch316/go-netx/runner"
	"github.com/oligarch316/go-netx/servicex"
)

type dependencySet map[netx.ServiceID]struct{}

func (ds dependencySet) Append(ids ...netx.ServiceID) {
	for _, id := range ids {
		ds[id] = struct{}{}
	}
}

type serviceParamData struct {
	listener *multi.Listener
	deps     dependencySet
}

// ServiceParams TODO.
type ServiceParams struct {
	services map[netx.ServiceID]*serviceParamData
}

func newServiceParams() ServiceParams {
	return ServiceParams{services: make(map[netx.ServiceID]*serviceParamData)}
}

func (sp *ServiceParams) requireData(id netx.ServiceID) *serviceParamData {
	res, ok := sp.services[id]
	if !ok {
		res = &serviceParamData{
			listener: multi.NewListener(),
			deps:     make(dependencySet),
		}
		sp.services[id] = res
	}
	return res
}

func (sp *ServiceParams) appendListeners(id netx.ServiceID, ls ...netx.Listener) {
	sp.requireData(id).listener.Append(ls...)
}

func (sp *ServiceParams) appendDependencies(id netx.ServiceID, depIDs ...netx.ServiceID) {
	sp.requireData(id).deps.Append(depIDs...)
}

type waitGroup interface {
	Done()
	Wait()
}

type noopWaitGroup struct{}

func (noopWaitGroup) Done() {}
func (noopWaitGroup) Wait() {}

type service struct {
	svc netx.Service
	ml  *multi.Listener

	dependants, requirements []*service
	dependantDoneSignal      func()
}

func newService(svc netx.Service, ml *multi.Listener) *service {
	return &service{svc: svc, ml: ml}
}

func (s *service) signalRequirements() {
	for _, req := range s.requirements {
		req.dependantDoneSignal()
	}
}

func (s *service) DependOn(svc *service) {
	s.requirements = append(s.requirements, svc)
	svc.dependants = append(svc.dependants, s)
}

func (s *service) Runners() []runner.Item {
	var res []runner.Item

	// ----- "Glue" runners
	// > wait groups for synchronization purposes
	var (
		dependantsWG waitGroup = noopWaitGroup{}
		listenersWG  waitGroup = noopWaitGroup{}
	)

	if len(s.dependants) > 0 {
		rnr := runner.NewWaitGroup(len(s.dependants))
		dependantsWG = rnr
		res = append(res, newServerRunner(s.svc.ID(), "dependants wait group", rnr))
	}

	if s.ml.Len() > 0 {
		rnr := runner.NewWaitGroup(s.ml.Len())
		listenersWG = rnr
		res = append(res, newServerRunner(s.svc.ID(), "listeners wait group", rnr))
	}

	s.dependantDoneSignal = dependantsWG.Done

	// ----- Service runner
	// > svc.Serve(...) and svc.Close(...) wrapped with glue logic
	var (
		baseServiceRunner    = servicex.NewRunner(s.ml, s.svc)
		wrappedServiceRunner = runner.New(
			// Run
			func() error {
				defer s.signalRequirements()
				return baseServiceRunner.Run()
			},

			// Close
			func(ctx context.Context) error {
				listenersWG.Wait()
				return baseServiceRunner.Close(ctx)
			},
		)
	)

	res = append(res, newServerRunner(s.svc.ID(), "service", wrappedServiceRunner))

	// ----- Listener runners
	// > ml.Runners() wrapped with glue logic
	for _, item := range s.ml.Runners() {
		var (
			baseListenRunner    = item
			wrappedListenRunner = runner.New(
				// Run
				func() error {
					defer listenersWG.Done()
					return baseListenRunner.Run()
				},

				// Close
				func(ctx context.Context) error {
					dependantsWG.Wait()
					return baseListenRunner.Close(ctx)
				},
			)
		)

		listenerName := fmt.Sprintf("listener (%s)", baseListenRunner.Addr())
		res = append(res, newServerRunner(s.svc.ID(), listenerName, wrappedListenRunner))
	}

	return res
}
