package synctest

import (
	"fmt"

	"github.com/stretchr/testify/assert"
)

type errorResult struct{ result }

func (er errorResult) errorValue() error {
	res, _ := er.value.(error)
	return res
}

func (er errorResult) ErrorAssertion(check assert.ErrorAssertionFunc) func(AssertT) bool {
	return func(t AssertT) bool {
		t.Helper()
		return check(t, er.errorValue(), er.String())
	}
}

func (er errorResult) ErrorStringAssertion() func(AssertT, string) bool {
	return func(t AssertT, expected string) bool {
		t.Helper()
		return assert.EqualError(t, er.errorValue(), expected, er.String())
	}
}

// ErrorSignal TODO.
type ErrorSignal struct {
	*errorResult
	Signal
}

// GoErrorSignal TODO.
func GoErrorSignal(name string, f func() error) ErrorSignal {
	er := new(errorResult)
	er.name = name + " error"
	return ErrorSignal{
		errorResult: er,
		Signal:      GoSignal(name, func() { er.value = f() }),
	}
}

// Value TODO.
func (es ErrorSignal) Value() error { return es.errorValue() }

// AssertEqual TODO.
func (es ErrorSignal) AssertEqual(t AssertT, expected interface{}) bool {
	t.Helper()
	switch exp := expected.(type) {
	case nil:
		return es.ErrorAssertion(assert.NoError)(t)
	case string:
		return es.ErrorStringAssertion()(t, exp)
	case error:
		return es.ErrorStringAssertion()(t, exp.Error())
	default:
		return es.FailAssertion()(t, fmt.Sprintf("invalid error expectation type: %T", exp))
	}
}

// RequireEqual TODO.
func (es ErrorSignal) RequireEqual(t RequireT, expected interface{}) {
	t.Helper()
	if !es.AssertEqual(t, expected) {
		t.FailNow()
	}
}
