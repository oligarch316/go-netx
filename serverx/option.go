package serverx

import (
	"github.com/oligarch316/go-netx"
	"github.com/oligarch316/go-netx/multi"
)

type (
	// Option TODO.
	Option func(*Params) error

	// ParamsService TODO.
	ParamsService interface {
		AddListeners(ServiceID, ...netx.Listener)
		AddDependencies(ServiceID, ...ServiceID)
	}

	// Params TODO.
	Params struct {
		ParamsService

		// TODO: Any other various top level server parameters (logging for example)
	}
)

type serviceParam struct {
	ml   *multi.Listener
	deps map[ServiceID]struct{}
}

func (sp *serviceParam) addListeners(ls ...netx.Listener) {
	if sp.ml == nil {
		sp.ml = multi.NewListener(ls...)
		return
	}

	sp.ml.Append(ls...)
}

func (sp *serviceParam) addDependencies(depIDs ...ServiceID) {
	if sp.deps == nil {
		sp.deps = make(map[ServiceID]struct{})
	}

	for _, depID := range depIDs {
		sp.deps[depID] = struct{}{}
	}
}

type serviceParams map[ServiceID]*serviceParam

func (sps serviceParams) AddListeners(id ServiceID, ls ...netx.Listener) {
	sps.paramOrNew(id).addListeners(ls...)
}

func (sps serviceParams) AddDependencies(id ServiceID, depIDs ...ServiceID) {
	sps.paramOrNew(id).addDependencies(depIDs...)
}

func (sps serviceParams) mlOk(id ServiceID) (*multi.Listener, bool) {
	if param, ok := sps[id]; ok && param.ml != nil {
		return param.ml, true
	}
	return nil, false
}

func (sps serviceParams) paramOrNew(id ServiceID) *serviceParam {
	param, ok := sps[id]
	if !ok {
		param = new(serviceParam)
		sps[id] = param
	}
	return param
}
