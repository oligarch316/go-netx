package runnertest

import (
	"fmt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Signal TODO.
type Signal struct {
	name string
	c    chan struct{}
}

// WrapSignal TODO.
func WrapSignal(name string, f func()) (Signal, func()) {
	var (
		sig      = Signal{name: name, c: make(chan struct{})}
		wrappedF = func() {
			f()
			close(sig.c)
		}
	)

	return sig, wrappedF
}

// GoSignal TODO.
func GoSignal(name string, f func()) Signal {
	res, wrapped := WrapSignal(name, f)
	go wrapped()
	return res
}

func (s Signal) String() string { return s.name }

// Is TODO.
func (s Signal) Is(state State) bool {
	select {
	case <-s.c:
		return state.val
	case <-state.wait():
		return !state.val
	}
}

func (s Signal) comparison(state State) func() bool {
	return func() bool { return s.Is(state) }
}

// Assert TODO.
func (s Signal) Assert(t assert.TestingT, state State, msgAndArgs ...interface{}) bool {
	if len(msgAndArgs) < 1 {
		msgAndArgs = []interface{}{"expected %s to be %v", s, state}
	}
	return assert.Condition(t, s.comparison(state), msgAndArgs...)
}

// Require TODO.
func (s Signal) Require(t require.TestingT, state State, msgAndArgs ...interface{}) {
	if !s.Assert(t, state, msgAndArgs...) {
		t.FailNow()
	}
}

// ErrorSignal TODO.
type ErrorSignal struct {
	Signal
	Err error
}

// WrapErrorSignal TODO.
func WrapErrorSignal(name string, f func() error) (sig *ErrorSignal, wrappedF func()) {
	sig = new(ErrorSignal)
	sig.Signal, wrappedF = WrapSignal(name, func() { sig.Err = f() })
	return
}

// GoErrorSignal TODO.
func GoErrorSignal(name string, f func() error) *ErrorSignal {
	res, wrapped := WrapErrorSignal(name, f)
	go wrapped()
	return res
}

// AssertError TODO.
func (es *ErrorSignal) AssertError(t assert.TestingT, expected interface{}, msgAndArgs ...interface{}) bool {
	if len(msgAndArgs) < 1 {
		msgAndArgs = []interface{}{"%s error check", es}
	}

	switch exp := expected.(type) {
	case nil:
		return assert.NoError(t, es.Err, msgAndArgs...)
	case string:
		return assert.EqualError(t, es.Err, exp, msgAndArgs...)
	case error:
		return assert.EqualError(t, es.Err, exp.Error(), msgAndArgs...)
	default:
		return assert.Fail(t, fmt.Sprintf("invalid error expectation type: %T", exp), msgAndArgs...)
	}
}

// RequireError TODO.
func (es *ErrorSignal) RequireError(t require.TestingT, expected interface{}, msgAndArgs ...interface{}) {
	if !es.AssertError(t, expected, msgAndArgs...) {
		t.FailNow()
	}
}
