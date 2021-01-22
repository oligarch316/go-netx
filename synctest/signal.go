package synctest

import (
	"fmt"
	"time"

	"github.com/stretchr/testify/assert"
)

// DefaultTimeout TODO.
const DefaultTimeout = 10 * time.Millisecond

var (
	// Complete TODO.
	Complete = SignalState{
		name:    "complete",
		val:     true,
		timeout: DefaultTimeout,
	}

	// Pending TODO.
	Pending = SignalState{
		name:    "pending",
		val:     false,
		timeout: DefaultTimeout,
	}
)

// SignalState TODO.
type SignalState struct {
	name    string
	val     bool
	timeout time.Duration
}

func (ss SignalState) String() string {
	return fmt.Sprintf("%s (timeout %s)", ss.name, ss.timeout)
}

// After TODO.
func (ss SignalState) After(timeout time.Duration) SignalState {
	return SignalState{name: ss.name, val: ss.val, timeout: timeout}
}

// Signal TODO.
type Signal struct {
	name string
	c    chan struct{}
}

// GoSignal TODO.
func GoSignal(name string, f func()) Signal {
	res := Signal{name: name, c: make(chan struct{})}
	go func() {
		f()
		close(res.c)
	}()
	return res
}

func (s Signal) String() string { return s.name }

// Is TODO.
func (s Signal) Is(state SignalState) bool {
	select {
	case <-s.c:
		return state.val
	case <-time.After(state.timeout):
		return !state.val
	}
}

// AssertState TODO.
func (s Signal) AssertState(t AssertT, expected SignalState) bool {
	t.Helper()
	return assert.True(t, s.Is(expected), "%s %s", s, expected)
}

// RequireState TODO.
func (s Signal) RequireState(t RequireT, expected SignalState) {
	t.Helper()
	if !s.AssertState(t, expected) {
		t.FailNow()
	}
}

// SignalList TODO.
type SignalList []Signal

func (sl SignalList) mapOver(f func(Signal) bool) bool {
	res := true
	for _, signal := range sl {
		res = f(signal) && res
	}
	return res
}

// Are TODO.
func (sl SignalList) Are(state SignalState) bool {
	return sl.mapOver(func(s Signal) bool { return s.Is(state) })
}

// AssertState TODO.
func (sl SignalList) AssertState(t AssertT, expected SignalState) bool {
	t.Helper()
	return sl.mapOver(func(s Signal) bool { return s.AssertState(t, expected) })
}

// RequireState TODO.
func (sl SignalList) RequireState(t RequireT, expected SignalState) {
	t.Helper()
	if !sl.AssertState(t, expected) {
		t.FailNow()
	}
}
