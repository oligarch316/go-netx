package serverx

import (
	"context"

	"github.com/oligarch316/go-netx"
	"github.com/oligarch316/go-netx/multi"
	"github.com/oligarch316/go-netx/runner"
	"github.com/oligarch316/go-netx/servicex"
)

type dependencySet map[servicex.ID]struct{}

func (ds dependencySet) Append(ids ...servicex.ID) {
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
	services map[servicex.ID]*serviceParamData
}

func newServiceParams() ServiceParams {
	return ServiceParams{services: make(map[servicex.ID]*serviceParamData)}
}

func (sp *ServiceParams) requireData(id servicex.ID) *serviceParamData {
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

func (sp *ServiceParams) appendListeners(id servicex.ID, ls ...netx.Listener) {
	sp.requireData(id).listener.Append(ls...)
}

func (sp *ServiceParams) appendDependencies(id servicex.ID, depIDs ...servicex.ID) {
	sp.requireData(id).deps.Append(depIDs...)
}

type service struct {
	svc servicex.Service
	ml  *multi.Listener

	dependants, requirements []*service
	dependantDoneSignal      func()
}

func newService(svc servicex.Service, ml *multi.Listener) *service {
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
	// ----- "Glue" runners
	// > wait groups for synchronization purposes
	var (
		dependantsWG = runner.NewWaitGroup(len(s.dependants))
		listenersWG  = runner.NewWaitGroup(s.ml.Len())
		res          = []runner.Item{dependantsWG, listenersWG}
	)

	s.dependantDoneSignal = dependantsWG.Done

	// ----- Service runner
	// > svc.Serve(...) and svc.Close(...) wrapped with glue logic
	var (
		baseServiceRunner = servicex.NewRunner(s.ml, s.svc)

		wrappedServiceRun = func() error {
			defer s.signalRequirements()
			return baseServiceRunner.Run()
		}

		wrappedServiceClose = func(ctx context.Context) error {
			listenersWG.Wait()
			return baseServiceRunner.Close(ctx)
		}
	)

	res = append(res, runner.New(wrappedServiceRun, wrappedServiceClose))

	// ----- Listener runners
	// > ml.Runners() wrapped with glue logic
	for _, item := range s.ml.Runners() {
		var (
			baseRunner = item

			wrappedRun = func() error {
				defer listenersWG.Done()
				return baseRunner.Run()
			}

			wrappedClose = func(ctx context.Context) error {
				dependantsWG.Wait()
				return baseRunner.Close(ctx)
			}
		)

		res = append(res, runner.New(wrappedRun, wrappedClose))
	}

	return res
}
