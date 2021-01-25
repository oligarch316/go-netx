package synctest

import (
	"github.com/stretchr/testify/assert"
)

type result struct {
	name  string
	value interface{}
}

func (r result) String() string { return string(r.name) }

func (r result) Value() interface{} { return r.value }

func (r result) FailAssertion() func(AssertT, string) bool {
	return func(t AssertT, message string) bool {
		t.Helper()
		return assert.Fail(t, message, r.String())
	}
}

func (r result) CompareAssertion(check assert.ComparisonAssertionFunc) func(AssertT, interface{}) bool {
	return func(t AssertT, expected interface{}) bool {
		t.Helper()
		return check(t, expected, r.Value(), r.String())
	}
}

func (r result) ValueAssertion(check assert.ValueAssertionFunc) func(AssertT) bool {
	return func(t AssertT) bool {
		t.Helper()
		return check(t, r.Value(), r.String())
	}
}

// ResultSignal TODO
type ResultSignal struct {
	*result
	Signal
}

// GoResultSignal TODO.
func GoResultSignal(name string, f func() interface{}) ResultSignal {
	r := &result{name: name + " result"}
	return ResultSignal{
		result: r,
		Signal: GoSignal(name, func() { r.value = f() }),
	}
}

// AssertEqual TODO.
func (rs ResultSignal) AssertEqual(t AssertT, expected interface{}) bool {
	t.Helper()
	return rs.CompareAssertion(assert.Equal)(t, expected)
}

// RequireEqual TODO.
func (rs ResultSignal) RequireEqual(t RequireT, expected interface{}) {
	t.Helper()
	if !rs.AssertEqual(t, expected) {
		t.FailNow()
	}
}
