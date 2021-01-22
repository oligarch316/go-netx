package synctest

import (
	"sync/atomic"

	"github.com/stretchr/testify/assert"
)

var (
	// Marked TODO.
	Marked = MarkerState{name: "marked", val: 1}

	// Unmarked TODO.
	Unmarked = MarkerState{name: "unmarked", val: 0}
)

// MarkerState TODO.
type MarkerState struct {
	name string
	val  uint32
}

func (ms MarkerState) String() string { return ms.name }

// Marker TODO.
type Marker struct {
	name string
	val  uint32
}

// NewMarker TODO.
func NewMarker(name string) *Marker { return &Marker{name: name, val: Unmarked.val} }

func (m Marker) String() string { return m.name }

// Mark TODO.
func (m *Marker) Mark() { atomic.StoreUint32(&m.val, Marked.val) }

// Is TODO.
func (m *Marker) Is(state MarkerState) bool { return atomic.LoadUint32(&m.val) == state.val }

// AssertState TODO.
func (m *Marker) AssertState(t AssertT, expected MarkerState) bool {
	t.Helper()
	return assert.True(t, m.Is(expected), "%s %s", m, expected)
}

// RequireState TODO.
func (m *Marker) RequireState(t RequireT, expected MarkerState) {
	t.Helper()
	if !m.AssertState(t, expected) {
		t.FailNow()
	}
}

// MarkerList TODO.
type MarkerList []*Marker

func (ml MarkerList) mapOver(f func(*Marker) bool) bool {
	res := true
	for _, marker := range ml {
		res = f(marker) && res
	}
	return res
}

// Are TODO.
func (ml MarkerList) Are(state MarkerState) bool {
	return ml.mapOver(func(m *Marker) bool { return m.Is(state) })
}

// AssertState TODO.
func (ml MarkerList) AssertState(t AssertT, expected MarkerState) bool {
	t.Helper()
	return ml.mapOver(func(f *Marker) bool { return f.AssertState(t, expected) })
}

// RequireState TODO.
func (ml MarkerList) RequireState(t RequireT, expected MarkerState) {
	t.Helper()
	if !ml.AssertState(t, expected) {
		t.FailNow()
	}
}
