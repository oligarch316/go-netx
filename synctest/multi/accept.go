package multitest

import (
	"fmt"
	"net"
	"testing"

	"github.com/oligarch316/go-netx/synctest"
)

type connStringChecker string

func (csc connStringChecker) Info() string      { return "actual.String() || string(actual)" }
func (csc connStringChecker) Data() interface{} { return string(csc) }

func (csc connStringChecker) Check(actual interface{}) bool {
	switch typ := actual.(type) {
	case fmt.Stringer:
		return string(csc) == typ.String()
	case string:
		return string(csc) == typ
	}
	return false
}

func newConnCheckers(expected []interface{}) ([]synctest.Checker, error) {
	res := make([]synctest.Checker, len(expected))
	for i, item := range expected {
		switch typ := item.(type) {
		case nil:
			res[i] = synctest.NilChecker
		case synctest.Checker:
			res[i] = typ
		case string:
			res[i] = connStringChecker(typ)
		default:
			return nil, fmt.Errorf("invalid expected conn type %v (%T)", expected, typ)
		}
	}
	return res, nil
}

// AcceptSignal TODO.
type AcceptSignal struct {
	*synctest.ErrorSignal
	conns []net.Conn
}

// Conns TODO.
func (as *AcceptSignal) Conns() []net.Conn { return as.conns }

func (as *AcceptSignal) checkConns(expected []synctest.Checker) (string, bool) {
	actual := make([]interface{}, len(as.conns))
	for i, item := range as.conns {
		actual[i] = item
	}

	if diff := synctest.NewSetDiff(actual, expected); !diff.AllSame() {
		return as.Report("conn set", diff), false
	}

	return "", true
}

// AssertConns TODO.
func (as *AcceptSignal) AssertConns(t *testing.T, expected ...interface{}) bool {
	t.Helper()

	checkers, err := newConnCheckers(expected)
	if err != nil {
		t.Error(err)
		return false
	}

	if report, ok := as.checkConns(checkers); !ok {
		t.Error(report)
		return false
	}

	return true
}

// RequireConns TODO.
func (as *AcceptSignal) RequireConns(t *testing.T, expected ...interface{}) {
	t.Helper()

	checkers, err := newConnCheckers(expected)
	if err != nil {
		t.Fatal(err)
		return
	}

	if report, ok := as.checkConns(checkers); !ok {
		t.Fatal(report)
	}
}
