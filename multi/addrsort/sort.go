package addrsort

import (
	"sort"

	"github.com/oligarch316/go-netx/multi"
)

// Ordering TODO.
type Ordering []Comparer

// Less TODO.
func (o Ordering) Less(x, y multi.Addr) bool {
	for _, cmp := range o {
		if less, equal := cmp(x, y); !equal {
			return less
		}
	}
	return false
}

// Sort TODO.
func (o Ordering) Sort(addrs []multi.Addr) {
	sort.Slice(addrs, func(i, j int) bool { return o.Less(addrs[i], addrs[j]) })
}

// Stable TODO.
func (o Ordering) Stable(addrs []multi.Addr) {
	sort.SliceStable(addrs, func(i, j int) bool { return o.Less(addrs[i], addrs[j]) })
}

// Sort TODO.
func Sort(addrs []multi.Addr, cmps ...Comparer) { Ordering(cmps).Sort(addrs) }

// Stable TODO.
func Stable(addrs []multi.Addr, cmps ...Comparer) { Ordering(cmps).Stable(addrs) }
