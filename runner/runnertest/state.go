package runnertest

import (
	"fmt"
	"time"
)

// DefaultTimeout TODO.
const DefaultTimeout = 10 * time.Millisecond

var (
	// Complete TODO.
	Complete = State{val: true, timeout: DefaultTimeout}

	// Pending TODO.
	Pending = State{val: false, timeout: DefaultTimeout}
)

// State TODO.
type State struct {
	val     bool
	timeout time.Duration
}

func (s State) wait() <-chan time.Time { return time.After(s.timeout) }

// After TODO.
func (s State) After(timeout time.Duration) State {
	return State{val: s.val, timeout: timeout}
}

func (s State) String() string {
	if s.val {
		return "complete"
	}
	return "pending"
}

// Format TODO.
func (s State) Format(fs fmt.State, verb rune) {
	if verb == 'v' {
		fmt.Fprintf(fs, "%s (timeout %s)", s.String(), s.timeout)
		return
	}
	fmt.Fprintf(fs, string([]rune{'%', verb}), s.String())
}
