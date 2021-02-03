package synctest

import (
	"fmt"
	"testing"
	"time"
)

// DefaultSignalTimeout TODO.
const DefaultSignalTimeout = 10 * time.Millisecond

var (
	// Complete TODO.
	Complete = SignalState{val: true, timeout: DefaultSignalTimeout}

	// Pending TODO.
	Pending = SignalState{val: false, timeout: DefaultSignalTimeout}
)

// SignalState TODO.
type SignalState struct {
	val     bool
	timeout time.Duration
}

func (ss SignalState) String() string {
	if ss.val {
		return "complete"
	}
	return "pending"
}

// After TODO.
func (ss SignalState) After(timeout time.Duration) SignalState {
	return SignalState{val: ss.val, timeout: timeout}
}

// Signal TODO.
type Signal struct {
	name string
	c    chan struct{}
}

// GoSignal TODO.
func GoSignal(name string, f func()) Signal {
	res := Signal{
		name: name,
		c:    make(chan struct{}),
	}

	go func() {
		f()
		close(res.c)
	}()

	return res
}

func (s Signal) String() string { return s.name }

// State TODO.
func (s Signal) State(timeout time.Duration) SignalState {
	res := SignalState{timeout: timeout}
	select {
	case <-time.After(timeout):
	case <-s.c:
		res.val = true
	}
	return res
}

func (s Signal) checkState(expected SignalState) (*report, bool) {
	if actual := s.State(expected.timeout); expected.val != actual.val {
		return &report{
			name: s.name,
			info: fmt.Sprintf("state after %s", expected.timeout),
			diff: simpleDiff{
				expected: expected,
				actual:   actual,
			}.String(),
		}, false
	}
	return nil, true
}

// AssertState TODO.
func (s Signal) AssertState(t *testing.T, expected SignalState) bool {
	t.Helper()
	if report, ok := s.checkState(expected); !ok {
		t.Error(report)
		return false
	}
	return true
}

// RequireState TODO.
func (s Signal) RequireState(t *testing.T, expected SignalState) {
	t.Helper()
	if report, ok := s.checkState(expected); !ok {
		t.Fatal(report)
	}
}

// SignalList TODO.
type SignalList []Signal

// AssertState TODO.
func (sl SignalList) AssertState(t *testing.T, expected SignalState) bool {
	t.Helper()
	res := true
	for _, signal := range sl {
		res = signal.AssertState(t, expected) && res
	}
	return res
}

// RequireState TODO.
func (sl SignalList) RequireState(t *testing.T, expected SignalState) {
	t.Helper()
	if !sl.AssertState(t, expected) {
		t.FailNow()
	}
}
