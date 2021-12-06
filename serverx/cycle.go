package serverx

import (
	"sort"
	"strings"
)

type idList []ServiceID

func (il idList) Len() int           { return len(il) }
func (il idList) Swap(i, j int)      { il[i], il[j] = il[j], il[i] }
func (il idList) Less(i, j int) bool { return il[i].String() < il[j].String() }

type cycleError []ServiceID

func (ce cycleError) complete() bool { return len(ce) > 1 && ce[0] == ce[len(ce)-1] }

func (ce cycleError) Error() string {
	var (
		n        = len(ce)
		reversed = make([]string, n)
	)

	for i := 0; i < n; i++ {
		reversed[i] = ce[n-(i+1)].String()
	}

	return strings.Join(reversed, " → ")
}

func findDependencyCycles(params serviceParams) error {
	var (
		visited = make(map[ServiceID]bool)
		recurse func(ServiceID) cycleError
	)

	recurse = func(svcID ServiceID) cycleError {
		if complete, exists := visited[svcID]; exists {
			if complete {
				return nil
			}

			return cycleError{svcID}
		}

		visited[svcID] = false
		defer func() { visited[svcID] = true }()

		if param, ok := params[svcID]; ok {
			depIDs := make(idList, 0, len(param.deps))
			for depID := range param.deps {
				depIDs = append(depIDs, depID)
			}

			sort.Stable(depIDs)

			for _, depID := range depIDs {
				if cycle := recurse(depID); cycle != nil {
					if cycle.complete() {
						return cycle
					}
					return append(cycle, svcID)
				}
			}
		}

		return nil
	}

	svcIDs := make(idList, 0, len(params))
	for svcID := range params {
		svcIDs = append(svcIDs, svcID)
	}

	sort.Stable(svcIDs)

	for _, svcID := range svcIDs {
		if err := recurse(svcID); err != nil {
			return err
		}
	}

	return nil
}
