package synctest

import (
	"errors"
	"fmt"
	"sort"
	"testing"
)

type errorChecker interface {
	info() string
	data() interface{}
	check(error) bool
}

type errorIsChecker struct{ error }

func (eic errorIsChecker) info() string { return "errors.Is(actual, expected)" }

func (eic errorIsChecker) data() interface{} { return eic.error }

func (eic errorIsChecker) check(actual error) bool { return errors.Is(actual, eic.error) }

type errorStringChecker string

func (esc errorStringChecker) info() string { return "actual.Error()" }

func (esc errorStringChecker) data() interface{} { return string(esc) }

func (esc errorStringChecker) check(actual error) bool { return string(esc) == actual.Error() }

// ErrorSignal TODO.
type ErrorSignal struct {
	Signal
	error
}

// GoSignalError TODO.
func GoSignalError(name string, f func() error) *ErrorSignal {
	res := new(ErrorSignal)
	res.Signal = GoSignal(name, func() { res.error = f() })
	return res
}

// Unwrap TODO.
func (er *ErrorSignal) Unwrap() error { return er.error }

func (er *ErrorSignal) checkError(expected errorChecker) (*report, bool) {
	if !expected.check(er.error) {
		return &report{
			name: er.name,
			info: expected.info(),
			diff: simpleDiff{
				expected: expected.data(),
				actual:   er.error,
			}.String(),
		}, false
	}
	return nil, true
}

// AssertErrorIs TODO.
func (er *ErrorSignal) AssertErrorIs(t *testing.T, expected error) bool {
	t.Helper()
	if report, ok := er.checkError(errorIsChecker{expected}); !ok {
		t.Error(report)
		return false
	}
	return true
}

// RequireErrorIs TODO.
func (er *ErrorSignal) RequireErrorIs(t *testing.T, expected error) {
	t.Helper()
	if report, ok := er.checkError(errorIsChecker{expected}); !ok {
		t.Fatal(report)
	}
}

// AssertErrorString TODO.
func (er *ErrorSignal) AssertErrorString(t *testing.T, expected string) bool {
	t.Helper()
	if report, ok := er.checkError(errorStringChecker(expected)); !ok {
		t.Error(report)
		return false
	}
	return true
}

// RequireErrorString TODO.
func (er *ErrorSignal) RequireErrorString(t *testing.T, expected string) {
	t.Helper()
	if report, ok := er.checkError(errorStringChecker(expected)); !ok {
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

func (ess *ErrorSetSignal) checkErrorSet(expected []errorChecker) (*report, bool) {
	var (
		extra = make([]interface{}, 0, len(ess.errs))
		same  = make([]interface{}, 0, len(ess.errs))
	)

L:
	for _, actualItem := range ess.errs {
		for i, expectedItem := range expected {
			if expectedItem.check(actualItem) {
				// Add to same list
				same = append(same, actualItem)

				// Delete from expected list
				lastIdx := len(expected) - 1
				expected[i] = expected[lastIdx]
				expected = expected[:lastIdx]

				// Continue with next actual
				continue L
			}
		}

		// Add to extra list
		extra = append(extra, actualItem)
	}

	var (
		nExtra   = len(extra)
		nMissing = len(expected)
	)

	if nMissing == 0 && nExtra == 0 {
		return nil, true
	}

	var diff complexDiff

	if nMissing > 0 {
		missingMap := make(map[string][]interface{})
		for _, item := range expected {
			info := item.info()
			missingMap[info] = append(missingMap[info], item.data())
		}

		for info, items := range missingMap {
			diff = append(diff, complexDiffSection{
				title:      fmt.Sprintf("missing %s", info),
				itemPrefix: "-",
				items:      items,
			})
		}
		sort.Sort(diff)
	}

	if nExtra > 0 {
		diff = append(diff, complexDiffSection{
			title:      "extra",
			itemPrefix: "+",
			items:      extra,
		})
	}

	if len(same) > 0 {
		diff = append(diff, complexDiffSection{
			title:      "same",
			itemPrefix: "•",
			items:      same,
		})
	}

	return &report{
		name: ess.name,
		info: "error set",
		diff: diff.String(),
	}, false
}

func (ess *ErrorSetSignal) buildCheckers(expected []interface{}) ([]errorChecker, error) {
	res := make([]errorChecker, len(expected))
	for i, item := range expected {
		switch typ := item.(type) {
		case nil:
			res[i] = errorIsChecker{nil}
		case error:
			res[i] = errorIsChecker{typ}
		case string:
			res[i] = errorStringChecker(typ)
		default:
			return nil, fmt.Errorf("invalid error set expectation type: %T", typ)
		}
	}
	return res, nil
}

// AssertErrorSet TODO.
func (ess *ErrorSetSignal) AssertErrorSet(t *testing.T, expected ...interface{}) bool {
	t.Helper()

	checkers, err := ess.buildCheckers(expected)
	if err != nil {
		t.Error(err)
		return false
	}

	if report, ok := ess.checkErrorSet(checkers); !ok {
		t.Error(report)
		return false
	}

	return true
}

// RequireErrorSet TODO.
func (ess *ErrorSetSignal) RequireErrorSet(t *testing.T, expected ...interface{}) {
	t.Helper()

	checkers, err := ess.buildCheckers(expected)
	if err != nil {
		t.Fatal(err)
		return
	}

	if report, ok := ess.checkErrorSet(checkers); !ok {
		t.Fatal(report)
	}
}
