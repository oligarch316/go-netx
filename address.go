package netx

import (
    "net"
    "sort"
)

// AbstractAddr TODO.
type AbstractAddr string

// Network TODO.
func (aa AbstractAddr) Network() string { return string(aa) }

func (aa AbstractAddr) String() string { return string(aa) }

// AddrSorter TODO.
type AddrSorter []AddrComparer

// Less TODO.
func (as AddrSorter) Less(x, y net.Addr) bool {
    for _, cmp := range as {
        if less, equal := cmp(x, y); !equal {
            return less
        }
    }
    return false
}

// Sort TODO.
func (as AddrSorter) Sort(addrs []net.Addr) {
    sort.Slice(addrs, func(i, j int) bool { return as.Less(addrs[i], addrs[j]) })
}

// Stable TODO.
func (as AddrSorter) Stable(addrs []net.Addr) {
    sort.SliceStable(addrs, func(i, j int) bool { return as.Less(addrs[i], addrs[j]) })
}

// AddrComparer TODO.
type AddrComparer func(x, y net.Addr) (less, equal bool)

// ByLexAddress TODO.
func ByLexAddress(x, y net.Addr) (bool, bool) {
    xStr, yStr := x.String(), y.String()
    return xStr < yStr, xStr == yStr
}

// ByLexNetwork TODO.
func ByLexNetwork(x, y net.Addr) (bool, bool) {
    xNet, yNet := x.Network(), y.Network()
    return xNet < yNet, xNet == yNet
}

// ByPriorityAddress TODO.
func ByPriorityAddress(addresses ...string) AddrComparer {
    pMap := newPriorityMap(addresses)
    return func(x, y net.Addr) (bool, bool) { return pMap.compare(x.String(), y.String()) }
}

// ByPriorityNetwork TODO.
func ByPriorityNetwork(networks ...string) AddrComparer {
    pMap := newPriorityMap(networks)
    return func(x, y net.Addr) (bool, bool) { return pMap.compare(x.Network(), y.Network()) }
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
