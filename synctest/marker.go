package synctest

import (
	"sync/atomic"
	"testing"
)

const (
	// Marked TODO.
	Marked MarkerState = 1

	// Unmarked TODO.
	Unmarked MarkerState = 0
)

// MarkerState TODO.
type MarkerState uint32

func (ms MarkerState) String() string {
	if ms == 0 {
		return "unmarked"
	}
	return "marked"
}

// Marker TODO.
type Marker struct {
	name string
	val  MarkerState
}

// NewMarker TODO.
func NewMarker(name string) *Marker {
	return &Marker{
		name: name,
		val:  Unmarked,
	}
}

func (m *Marker) String() string { return m.name }

// Mark TODO.
func (m *Marker) Mark() { atomic.StoreUint32((*uint32)(&m.val), uint32(Marked)) }

// State TODO.
func (m *Marker) State() MarkerState {
	return MarkerState(atomic.LoadUint32((*uint32)(&m.val)))
}

func (m *Marker) checkState(expected MarkerState) (*report, bool) {
	if actual := m.State(); expected != actual {
		return &report{
			name: m.name,
			info: "state",
			diff: simpleDiff{
				expected: expected,
				actual:   actual,
			}.String(),
		}, false
	}
	return nil, true
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
