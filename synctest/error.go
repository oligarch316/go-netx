package synctest

import (
	"errors"
	"fmt"
	"testing"
)

type errorIsChecker struct{ error }

func (eic errorIsChecker) Info() string      { return "errors.Is(actual, expected)" }
func (eic errorIsChecker) Data() interface{} { return eic.error }

func (eic errorIsChecker) Check(actual interface{}) bool {
	actualErr, ok := actual.(error)
	return ok && errors.Is(actualErr, eic.error)
}

type errorStringChecker string

func (esc errorStringChecker) Info() string      { return "expected == actual.Error()" }
func (esc errorStringChecker) Data() interface{} { return string(esc) }

func (esc errorStringChecker) Check(actual interface{}) bool {
	actualErr, ok := actual.(error)
	return ok && actualErr != nil && string(esc) == actualErr.Error()
}

func newErrorChecker(expected interface{}) (res Checker, err error) {
	switch typ := expected.(type) {
	case nil:
		res = NilChecker
	case Checker:
		res = typ
	case error:
		res = errorIsChecker{typ}
	case string:
		res = errorStringChecker(typ)
	default:
		err = fmt.Errorf("invalid expected error type %v (%T)", expected, typ)
	}
	return
}

// ErrorSignal TODO.
type ErrorSignal struct {
	Signal
	err error
}

// GoSignalError TODO.
func GoSignalError(name string, f func() error) *ErrorSignal {
	res := new(ErrorSignal)
	res.Signal = GoSignal(name, func() { res.err = f() })
	return res
}

// Error TODO.
func (er *ErrorSignal) Error() error { return er.err }

func (er *ErrorSignal) checkError(expected Checker) (string, bool) {
	if !expected.Check(er.err) {
		diff := simpleDiff{
			expected: expected.Data(),
			actual:   er.err,
		}
		return er.Report(expected.Info(), diff.String()), false
	}
	return "", true
}

// AssertError TODO.
func (er *ErrorSignal) AssertError(t *testing.T, expected interface{}) bool {
	t.Helper()

	checker, err := newErrorChecker(expected)
	if err != nil {
		t.Error(err)
		return false
	}

	if report, ok := er.checkError(checker); !ok {
		t.Error(report)
		return false
	}

	return true
}

// RequireError TODO.
func (er *ErrorSignal) RequireError(t *testing.T, expected interface{}) {
	t.Helper()

	checker, err := newErrorChecker(expected)
	if err != nil {
		t.Fatal(err)
		return
	}

	if report, ok := er.checkError(checker); !ok {
		t.Fatal(report)
	}
}

// ErrorSetSignal TODO.
type ErrorSetSignal struct {
	Signal
	errs []error
}

// GoSignalErrorSet TODO.
func GoSignalErrorSet(name string, f func() []error) *ErrorSetSignal {
	res := new(ErrorSetSignal)
	res.Signal = GoSignal(name, func() { res.errs = f() })
	return res
}

// Errors TODO.
func (ess *ErrorSetSignal) Errors() []error { return ess.errs }

func (ess *ErrorSetSignal) checkErrors(expected []Checker) (string, bool) {
	actual := make([]interface{}, len(ess.errs))
	for i, item := range ess.errs {
		actual[i] = item
	}

	if diff := NewSetDiff(actual, expected); !diff.AllSame() {
		return ess.Report("error set", diff.String()), false
	}

	return "", true
}

// AssertErrors TODO.
func (ess *ErrorSetSignal) AssertErrors(t *testing.T, expected ...interface{}) bool {
	t.Helper()

	var (
		checkers = make([]Checker, len(expected))
		err      error
	)

	for i, item := range expected {
		if checkers[i], err = newErrorChecker(item); err != nil {
			t.Error(err)
			return false
		}
	}

	if report, ok := ess.checkErrors(checkers); !ok {
		t.Error(report)
		return false
	}

	return true
}

// RequireErrors TODO.
func (ess *ErrorSetSignal) RequireErrors(t *testing.T, expected ...interface{}) {
	t.Helper()

	var (
		checkers = make([]Checker, len(expected))
		err      error
	)

	for i, item := range expected {
		if checkers[i], err = newErrorChecker(item); err != nil {
			t.Fatal(err)
			return
		}
	}

	if report, ok := ess.checkErrors(checkers); !ok {
		t.Fatal(report)
	}
}
