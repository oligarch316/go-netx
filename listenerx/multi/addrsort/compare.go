package addrsort

import "github.com/oligarch316/go-netx/listenerx/multi"

// Comparer TODO.
type Comparer func(x, y multi.Addr) (less, equal bool)

// ----- Lexographic

// ByLexAddress TODO.
func ByLexAddress(x, y multi.Addr) (bool, bool) {
	xStr, yStr := x.String(), y.String()
	return xStr < yStr, xStr == yStr
}

// ByLexNetwork TODO.
func ByLexNetwork(x, y multi.Addr) (bool, bool) {
	xNet, yNet := x.Network(), y.Network()
	return xNet < yNet, xNet == yNet
}

// ----- Priority

// ByPriorityAddress TODO.
func ByPriorityAddress(addresses ...string) Comparer {
	pMap := newPriorityMap(addresses)
	return func(x, y multi.Addr) (bool, bool) { return pMap.compare(x.String(), y.String()) }
}

// ByPriorityNetwork TODO.
func ByPriorityNetwork(networks ...string) Comparer {
	pMap := newPriorityMap(networks)
	return func(x, y multi.Addr) (bool, bool) { return pMap.compare(x.Network(), y.Network()) }
}

type priorityMap map[string]int

func newPriorityMap(items []string) priorityMap {
	res := make(priorityMap)
	for i, item := range items {
		res[item] = i
	}
	return res
}

func (pm priorityMap) compare(x, y string) (bool, bool) {
	var (
		xVal, xOk = pm[x]
		yVal, yOk = pm[y]
	)

	switch {
	case xOk && yOk:
		return xVal < yVal, xVal == yVal
	case xOk || yOk:
		return xOk, false
	default:
		return false, true
	}
}
