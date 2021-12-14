package serverx

import (
	"context"
	"errors"
	"fmt"

	"github.com/oligarch316/go-netx/runner"
	"github.com/oligarch316/go-netx/servicex"
)

var (
	errNoSuchService     = errors.New("serverx: no such service")
	errMissingListener   = errors.New("serverx: missing listener")
	errDuplicateService  = errors.New("serverx: duplicate service")
	errMissingDependency = errors.New("serverx: missing dependency")
)

// Option TODO.
type Option func(*Params)

// Params TODO.
type Params struct {
	IgnoreParams
	ServiceParams
}

// IgnoreParams TODO.
type IgnoreParams struct {
	IgnoreMissingListeners    bool
	IgnoreDuplicateServices   bool
	IgnoreMissingDependencies bool
}

// Server TODO.
type Server struct {
	params   Params
	runGroup *runner.Group
}

// NewServer TODO.
func NewServer(opts ...Option) (*Server, error) {
	params := Params{
		IgnoreParams: IgnoreParams{
			IgnoreMissingListeners:    false,
			IgnoreDuplicateServices:   false,
			IgnoreMissingDependencies: false,
		},
		ServiceParams: newServiceParams(),
	}

	for _, opt := range opts {
		opt(&params)
	}

	return &Server{params: params}, findDependencyCycles(params.ServiceParams)
}

// Close TODO.
func (s *Server) Close(ctx context.Context) {
	if s.runGroup != nil {
		s.runGroup.Close(ctx)
	}
}

// DialSet TODO.
func (s *Server) DialSet(id servicex.ID) (*DialSet, error) {
	if svcParams, ok := s.params.services[id]; ok {
		return &DialSet{svcParams.listener.Set}, nil
	}

	return nil, fmt.Errorf("%w: %s", errNoSuchService, id)
}

// Serve TODO.
func (s *Server) Serve(svcs ...servicex.Service) (<-chan error, error) {
	svcMap := make(map[servicex.ID]*service)

	// Combine arguments with associated params to build finalized services
	for _, svc := range svcs {
		id := svc.ID()

		svcParams, ok := s.params.services[id]
		if !ok {
			if s.params.IgnoreMissingListeners {
				continue
			}

			return nil, fmt.Errorf("%w: %s", errMissingListener, id)
		}

		if _, exists := svcMap[id]; exists && !s.params.IgnoreDuplicateServices {
			return nil, fmt.Errorf("%w: %s", errDuplicateService, id)
		}

		svcMap[id] = newService(svc, svcParams.listener)
	}

	// Build service dependencies
	for id, svc := range svcMap {
		for depID := range s.params.services[id].deps {
			depSvc, ok := svcMap[depID]
			if !ok {
				if s.params.IgnoreMissingDependencies {
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
