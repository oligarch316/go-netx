package synctest

import (
	"github.com/stretchr/testify/assert"
)

type setResult struct{ result }

func (sr setResult) setValue() []interface{} {
	res, _ := sr.value.([]interface{})
	return res
}

// SetSignal TODO.
type SetSignal struct {
	*setResult
	Signal
}

// GoSetSignal TODO.
func GoSetSignal(name string, f func() []interface{}) SetSignal {
	r := new(setResult)
	r.name = name + " result set"
	return SetSignal{
		setResult: r,
		Signal:    GoSignal(name, func() { r.value = f() }),
	}
}

// Value TODO.
func (ss SetSignal) Value() []interface{} { return ss.setValue() }

// AssertEqual TODO.
func (ss SetSignal) AssertEqual(t AssertT, expected ...interface{}) bool {
	t.Helper()
	if len(expected) < 1 {
		return ss.ValueAssertion(assert.Empty)(t)
	}
	return ss.CompareAssertion(assert.ElementsMatch)(t, expected)
}

// RequireEqual TODO.
func (ss SetSignal) RequireEqual(t RequireT, expected ...interface{}) {
	t.Helper()
	if !ss.AssertEqual(t, expected) {
		t.FailNow()
	}
}
