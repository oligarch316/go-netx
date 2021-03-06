package serverx

import (
	"context"
	"errors"
	"fmt"

	"github.com/oligarch316/go-netx/runner"
)

var errNoSuchServiceID = errors.New("serverx: no such service id")

// Server TODO.
type Server struct {
	svcParams serviceParams
	runGroup  *runner.Group
}

// NewServer TODO.
func NewServer(opts ...Option) (*Server, error) {
	var (
		res    = &Server{svcParams: make(serviceParams)}
		params = Params{ParamsService: res.svcParams}
	)

	for _, opt := range opts {
		if err := opt(&params); err != nil {
			return nil, err
		}
	}

	return res, findDependencyCycles(res.svcParams)
}

// Close TODO.
func (s *Server) Close(ctx context.Context) {
	if s.runGroup != nil {
		s.runGroup.Close(ctx)
	}
}

// DialSet TODO.
func (s *Server) DialSet(id ServiceID) (*DialSet, error) {
	if ml, ok := s.svcParams.mlOk(id); ok {
		return &DialSet{ml.Set}, nil
	}
	return nil, fmt.Errorf("%w: %s", errNoSuchServiceID, id)
}

// Serve TODO.
func (s *Server) Serve(svcs ...Service) <-chan error {
	svcMap := make(map[ServiceID]*service)

	// Combine arguments with associated multi listeners to build finalized services
	for _, svc := range svcs {
		svcID := svc.ID()

		ml, ok := s.svcParams.mlOk(svcID)
		if !ok {
			// Skip those services with no listeners available

			// TODO: Log/track/surface this somehow
			continue
		}

		if _, exists := svcMap[svcID]; exists {
			// TODO: Log/track/surface this somehow
		}

		svcMap[svcID] = newService(svc, ml)
	}

	// Build service dependencies
	for id, svc := range svcMap {
		for depID := range s.svcParams[id].deps {
			depSvc, ok := svcMap[depID]
			if !ok {
				// TODO: Log/track/surface this somehow
				continue
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
	s.runGroup.Run()

	// Return the run group's error channel
	return s.runGroup.Results()
}
