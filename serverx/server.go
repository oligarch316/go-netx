package serverx

import (
	"context"
	"errors"
	"fmt"

	"github.com/oligarch316/go-netx"
	"github.com/oligarch316/go-netx/listenerx/multi"
	"github.com/oligarch316/go-netx/runner"
)

var (
	errNoSuchService     = errors.New("serverx: no such service")
	errMissingListener   = errors.New("serverx: missing listener")
	errDuplicateService  = errors.New("serverx: duplicate service")
	errMissingDependency = errors.New("serverx: missing dependency")
)

type serviceParam struct {
	dependencies map[netx.ServiceID]struct{}
	listeners    []netx.Listener
	listenerOpts []multi.ListenerOption
}

type serviceData struct {
	dependencies map[netx.ServiceID][]netx.ServiceID
	listeners    map[netx.ServiceID]*multi.Listener
}

// Option TODO.
type Option func(*Params)

// Params TODO.
type Params struct {
	Ignore   IgnoreParams
	Services ServiceParams
}

// IgnoreParams TODO.
type IgnoreParams struct {
	DuplicateServices   bool
	MissingDependencies bool
	MissingListeners    bool
}

// ServiceParams TODO.
type ServiceParams map[netx.ServiceID]*serviceParam

func (sp ServiceParams) build() serviceData {
	res := serviceData{
		dependencies: make(map[netx.ServiceID][]netx.ServiceID),
		listeners:    make(map[netx.ServiceID]*multi.Listener),
	}

	for id, param := range sp {
		deps := make([]netx.ServiceID, 0)
		for depID := range param.dependencies {
			deps = append(deps, depID)
		}

		res.dependencies[id] = deps
		res.listeners[id] = multi.NewListener(param.listeners, param.listenerOpts...)
	}

	return res
}

func (sp ServiceParams) lookup(id netx.ServiceID) *serviceParam {
	res, ok := sp[id]
	if !ok {
		res = &serviceParam{dependencies: make(map[netx.ServiceID]struct{})}
		sp[id] = res
	}
	return res
}

func (sp ServiceParams) AppendListeners(id netx.ServiceID, ls ...netx.Listener) {
	param := sp.lookup(id)
	param.listeners = append(param.listeners, ls...)
}

func (sp ServiceParams) AppendListenerOpts(id netx.ServiceID, opts ...multi.ListenerOption) {
	param := sp.lookup(id)
	param.listenerOpts = append(param.listenerOpts, opts...)
}

func (sp ServiceParams) AppendDependencies(id netx.ServiceID, depIDs ...netx.ServiceID) {
	param := sp.lookup(id)
	for _, id := range depIDs {
		param.dependencies[id] = struct{}{}
	}
}

// Server TODO.
type Server struct {
	ignore   IgnoreParams
	services serviceData
	runGroup *runner.Group
}

// NewServer TODO.
func NewServer(opts ...Option) (*Server, error) {
	params := Params{
		Ignore: IgnoreParams{
			MissingListeners:    false,
			DuplicateServices:   false,
			MissingDependencies: false,
		},
		Services: make(ServiceParams),
	}

	for _, opt := range opts {
		opt(&params)
	}

	res := &Server{
		ignore:   params.Ignore,
		services: params.Services.build(),
	}

	return res, cycleCheck(res.services.dependencies)
}

// Close TODO.
func (s *Server) Close(ctx context.Context) {
	if s.runGroup != nil {
		s.runGroup.Close(ctx)
	}
}

// Dialer TODO.
func (s *Server) Dialer(id netx.ServiceID) (*multi.Dialer, error) {
	if ml, ok := s.services.listeners[id]; ok {
		return ml.Dialer, nil
	}

	return nil, fmt.Errorf("%w: %s", errNoSuchService, id)
}

// Serve TODO.
func (s *Server) Serve(svcs ...netx.Service) (<-chan error, error) {
	svcMap := make(map[netx.ServiceID]*service)

	// Combine arguments with associated params to build finalized services
	for _, svc := range svcs {
		id := svc.ID()

		ml, ok := s.services.listeners[id]
		if !ok {
			if s.ignore.MissingListeners {
				continue
			}

			return nil, fmt.Errorf("%w: %s", errMissingListener, id)
		}

		if _, exists := svcMap[id]; exists && !s.ignore.DuplicateServices {
			return nil, fmt.Errorf("%w: %s", errDuplicateService, id)
		}

		svcMap[id] = newService(svc, ml)
	}

	// Build service dependencies
	for id, svc := range svcMap {
		for _, depID := range s.services.dependencies[id] {
			depSvc, ok := svcMap[depID]
			if !ok {
				if s.ignore.MissingDependencies {
					continue
				}

				return nil, fmt.Errorf("%w: %s → %s", errMissingDependency, id, depID)
			}

			svc.DependOn(depSvc)
		}
	}

	// Build the run group from services
	s.runGroup = runner.NewGroup()
	for _, svc := range svcMap {
		s.runGroup.Append(svc.Runners()...)
	}

	// Start the run group
	return s.runGroup.Run(), nil
}
