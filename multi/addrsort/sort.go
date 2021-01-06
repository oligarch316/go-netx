package addrsort

import (
	"sort"

	"github.com/oligarch316/go-netx/multi"
)

// CompareList TODO.
type CompareList []Comparer

// Less TODO.
func (cl CompareList) Less(x, y multi.Addr) bool {
	for _, cmp := range cl {
		if less, equal := cmp(x, y); !equal {
			return less
		}
	}
	return false
}

// Sort TODO.
func (cl CompareList) Sort(addrs []multi.Addr) {
	sort.Slice(addrs, func(i, j int) bool { return cl.Less(addrs[i], addrs[j]) })
}

// Stable TODO.
func (cl CompareList) Stable(addrs []multi.Addr) {
	sort.SliceStable(addrs, func(i, j int) bool { return cl.Less(addrs[i], addrs[j]) })
}

// Sort TODO.
func Sort(addrs []multi.Addr, cmps ...Comparer) { CompareList(cmps).Sort(addrs) }

// Stable TODO.
func Stable(addrs []multi.Addr, cmps ...Comparer) { CompareList(cmps).Stable(addrs) }
