package synctest

import (
	"github.com/stretchr/testify/assert"
)

type setResult struct{ result }

func (sr setResult) Value() []interface{} {
	res, _ := sr.value.([]interface{})
	return res
}

// SetAssertion TODO.
func (sr setResult) SetAssertion(check func(AssertT, []interface{}, ...interface{}) bool) func(AssertT) bool {
	return func(t AssertT) bool {
		t.Helper()
		return check(t, sr.Value(), sr.String())
	}
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
