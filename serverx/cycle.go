package serverx

import (
	"sort"
	"strings"

	"github.com/oligarch316/go-netx"
)

type cycleIDList []netx.ServiceID

func (cil cycleIDList) Len() int           { return len(cil) }
func (cil cycleIDList) Swap(i, j int)      { cil[i], cil[j] = cil[j], cil[i] }
func (cil cycleIDList) Less(i, j int) bool { return cil[i].String() < cil[j].String() }

type cycleError []netx.ServiceID

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

func cycleCheck(depMap map[netx.ServiceID][]netx.ServiceID) error {
	var (
		visited = make(map[netx.ServiceID]bool)
		recurse func(netx.ServiceID) cycleError
	)

	recurse = func(svcID netx.ServiceID) cycleError {
		if complete, exists := visited[svcID]; exists {
			if complete {
				return nil
			}

			return cycleError{svcID}
		}

		visited[svcID] = false
		defer func() { visited[svcID] = true }()

		depIDs := cycleIDList(depMap[svcID])
		sort.Stable(depIDs)

		for _, depID := range depIDs {
			if err := recurse(depID); err != nil {
				if err.complete() {
					return err
				}
				return append(err, svcID)
			}
		}

		return nil
	}

	rootIDs := make(cycleIDList, 0, len(depMap))
	for rootID := range depMap {
		rootIDs = append(rootIDs, rootID)
	}

	sort.Stable(rootIDs)

	for _, rootID := range rootIDs {
		if err := recurse(rootID); err != nil {
			return err
		}
	}

	return nil
}
