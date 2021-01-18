package runnertest

import (
	"sync/atomic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var boolToAssertion = map[bool]assert.BoolAssertionFunc{
	true:  assert.True,
	false: assert.False,
}

// Bool TODO.
type Bool bool

// Assert TODO.
func (b Bool) Assert(t assert.TestingT, value bool, msgAndArgs ...interface{}) bool {
	return boolToAssertion[bool(b)](t, value, msgAndArgs...)
}

// Require TODO.
func (b Bool) Require(t require.TestingT, value bool, msgAndArgs ...interface{}) {
	if !b.Assert(t, value, msgAndArgs...) {
		t.FailNow()
	}
}

// Flag TODO.
type Flag struct {
	name   string
	marker uint32
}

// NewFlag TODO.
func NewFlag(name string) *Flag { return &Flag{name: name} }

func (f Flag) String() string { return f.name }

// Mark TODO.
func (f *Flag) Mark() { atomic.StoreUint32(&f.marker, 1) }

// IsMarked TODO.
func (f *Flag) IsMarked() bool { return atomic.LoadUint32(&f.marker) == 1 }

// Assert TODO.
func (f *Flag) Assert(t assert.TestingT, expected bool, msgAndArgs ...interface{}) bool {
	if len(msgAndArgs) < 1 {
		msgAndArgs = []interface{}{f.String()}
	}
	return Bool(expected).Assert(t, f.IsMarked(), msgAndArgs...)
}

// Require TODO.
func (f *Flag) Require(t require.TestingT, expected bool, msgAndArgs ...interface{}) {
	if !f.Assert(t, expected, msgAndArgs...) {
		t.FailNow()
	}
}
