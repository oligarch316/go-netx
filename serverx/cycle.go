package serverx

import (
	"fmt"
	"sort"
	"strings"
)

// Cycle TODO.
type Cycle []ServiceID

func (c Cycle) Error() string { return fmt.Sprintf("dependency cycle: %s", c.String()) }

func (c Cycle) String() string {
	nItems := len(c)
	if nItems < 1 {
		return "<empty>"
	}

	temp := make([]string, nItems+1)
	temp[nItems] = c[0].String()

	for i, id := range c {
		temp[i] = id.String()
	}

	return strings.Join(temp, " → ")
}

func (c Cycle) normalize() Cycle {
	nItems := len(c)

	switch nItems {
	case 0:
		return nil
	case 1:
		return c
	}

	var (
		bestIdx = 0
		bestVal = c[0].String()
	)

	for i := 1; i < nItems; i++ {
		if curVal := c[i].String(); curVal < bestVal {
			bestIdx, bestVal = i, curVal
		}
	}

	return Cycle(append(c[bestIdx:], c[:bestIdx]...))
}

// CycleList TODO.
type CycleList []Cycle

func (cl CycleList) Error() string {
	switch nItems := len(cl); nItems {
	case 0:
		return "dependency cycles: <empty>"
	case 1:
		return cl[0].Error()
	default:
		return fmt.Sprintf("%d dependency cycles, including: %s", nItems, cl[0].String())
	}
}

func (cl CycleList) normalize() CycleList {
	var res CycleList

	for _, c := range cl {
		res = append(res, c.normalize())
	}

	sort.Slice(res, func(i, j int) bool {
		iC, jC := res[i], res[j]

		if iLen, jLen := len(iC), len(jC); iLen != jLen {
			return iLen < jLen
		}

		for idx, iItem := range iC {
			if iStr, jStr := iItem.String(), jC[idx].String(); iStr != jStr {
				return iStr < jStr
			}
		}

		return false
	})

	return res
}

func findDependencyCycles(params serviceParams) CycleList {
	// Short Summary | DFS, Detect back-edges, Memoize partial cycles detected per ServiceID

	var (
		// Presence in map => visited
		// "false" value => search in progress
		// "true" value => search completed
		gCompleted = make(map[ServiceID]bool)

		// Updated upon search complete with search results
		gMemo = make(map[ServiceID]CycleList)

		// Last value in a (partial) cycle will be the vistited but incomplete "trigger" id
		triggerOf = func(c Cycle) ServiceID { return c[len(c)-1] }

		gCycles CycleList
		search  func(ServiceID) CycleList
	)

	search = func(svcID ServiceID) CycleList {
		if completed, visited := gCompleted[svcID]; visited {
			// Previously visited ...

			if !completed {
				// ... but search incomplete ==> cycle "trigger" encountered (back-edge)
				return CycleList{{svcID}}
			}

			// ... and search complete ==> use previous results stored in gMemo

			var res CycleList

			for _, memCycle := range gMemo[svcID] {
				if !gCompleted[triggerOf(memCycle)] {
					// Persist only those results where the "trigger" remains incomplete
					res = append(res, memCycle)
				}
			}

			// Update stored results to the filtered list
			gMemo[svcID] = res
			return res
		}

		// Not previously visited

		var res CycleList

		// Mark visited
		gCompleted[svcID] = false

		if param, ok := params[svcID]; ok {
			for depID := range param.deps {
				for _, partialCycle := range search(depID) {
					// Recursively search each dependency

					if triggerOf(partialCycle) == svcID {
						// Any cycles "triggered" by "me" are complete, add them to global results
						gCycles = append(gCycles, partialCycle)
						continue
					}

					// Augment all other cycles and pass them up
					res = append(res, append(Cycle{svcID}, partialCycle...))
				}
			}
		}

		// Mark search complete
		gCompleted[svcID] = true

		// Store results
		gMemo[svcID] = res
		return res
	}

	for svcID := range params {
		search(svcID)
	}

	return gCycles.normalize()
}
