package synctest

import (
	"fmt"
	"sync/atomic"
	"testing"
)

// const (
// 	// Marked TODO.
// 	Marked MarkerState = 1

// 	// Unmarked TODO.
// 	Unmarked MarkerState = 0
// )

// // MarkerState TODO.
// type MarkerState uint32

// func (ms MarkerState) String() string {
// 	if ms == 0 {
// 		return "unmarked"
// 	}
// 	return "marked"
// }

const (
	// Marked TODO.
	Marked MarkerState = true

	// Unmarked TODO.
	Unmarked MarkerState = false
)

// MarkerState TODO.
type MarkerState bool

var markerStateName = map[bool]string{
	true:  "marked",
	false: "unmarked",
}

func (ms MarkerState) String() string { return markerStateName[bool(ms)] }

// Marker TODO.
type Marker struct {
	Reporter
	// val MarkerState
	val uint32
}

// NewMarker TODO.
func NewMarker(name string) *Marker {
	return &Marker{
		Reporter: Reporter(name),
		// val:      Unmarked,
		val: 0,
	}
}

// Mark TODO.
func (m *Marker) Mark() { atomic.StoreUint32(&m.val, 1) }

// func (m *Marker) Mark() { atomic.StoreUint32((*uint32)(&m.val), uint32(Marked)) }

// State TODO.
func (m *Marker) State() MarkerState {
	// return MarkerState(atomic.LoadUint32((*uint32)(&m.val)))
	return MarkerState(atomic.LoadUint32(&m.val) == 1)
}

func (m *Marker) checkState(expected MarkerState) (string, bool) {
	if actual := m.State(); expected != actual {
		diff := simpleDiff{
			expected: expected,
			actual:   actual,
		}
		return m.Report("state", diff.String()), false
	}
	return "", true
}

// AssertState TODO.
func (m *Marker) AssertState(t *testing.T, expected MarkerState) bool {
	t.Helper()
	if report, ok := m.checkState(expected); !ok {
		t.Error(report)
		return false
	}
	return true
}

// RequireState TODO.
func (m *Marker) RequireState(t *testing.T, expected MarkerState) {
	t.Helper()
	if report, ok := m.checkState(expected); !ok {
		t.Fatal(report)
	}
}

// MarkerList TODO.
type MarkerList []*Marker

func (ml MarkerList) checkStates(expected MarkerState) (string, bool) {
	var good, bad []interface{}
	for _, marker := range ml {
		if expected == marker.State() {
			good = append(good, marker)
			continue
		}
		bad = append(bad, marker)
	}

	if len(bad) == 0 {
		return "", true
	}

	actualDiff := complexDiff{
		complexDiffSection{
			title:      fmt.Sprintf("actual | %s", MarkerState(!expected)),
			itemPrefix: "-",
			items:      bad,
		},
		complexDiffSection{
			title:      fmt.Sprintf("actual | %s", expected),
			itemPrefix: "•",
			items:      good,
		},
	}

	diff := fmt.Sprintf("expected: %s\n%s", expected, actualDiff)

	return Report("marker list", "states", diff), false
}

// AssertState TODO.
func (ml MarkerList) AssertState(t *testing.T, expected MarkerState) bool {
	t.Helper()
	res := true
	for _, marker := range ml {
		res = marker.AssertState(t, expected) && res
	}
	return res
}

// RequireState TODO.
func (ml MarkerList) RequireState(t *testing.T, expected MarkerState) {
	t.Helper()
	if !ml.AssertState(t, expected) {
		t.FailNow()
	}
}
