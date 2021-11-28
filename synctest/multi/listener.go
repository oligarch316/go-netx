package multitest

import (
	"fmt"

	"github.com/oligarch316/go-netx"
	"github.com/oligarch316/go-netx/multi"
	"github.com/oligarch316/go-netx/synctest"
	runnertest "github.com/oligarch316/go-netx/synctest/runner"
)

// Listener TODO.
type Listener struct {
	name string
	*multi.Listener
}

// NewListener TODO.
func NewListener(name string, ls ...netx.Listener) Listener {
	return WrapListener(name, multi.NewListener(ls...))
}

// WrapListener TODO.
func WrapListener(name string, l *multi.Listener) Listener {
	return Listener{name: name, Listener: l}
}

func (l Listener) String() string { return l.name }

// Accept TODO.
func (l Listener) Accept(n int) *AcceptSignal {
	name := fmt.Sprintf("%s Accept(%d)", l, n)
	res := new(AcceptSignal)
	res.ErrorSignal = synctest.GoSignalError(name, func() error {
		for i := 0; i < n; i++ {
			conn, err := l.Listener.Accept()
			if err != nil {
				return err
			}
			res.conns = append(res.conns, conn)
		}
		return nil
	})
	return res
}

// Runners TODO.
func (l Listener) Runners() []runnertest.Item {
	res := make([]runnertest.Item, l.Len())
	for i, item := range l.Listener.Runners() {
		res[i] = runnertest.Wrap(fmt.Sprintf("%s runner %d", l, i), item)
	}
	return res
}

// RunnerGroup TODO.
func (l Listener) RunnerGroup() runnertest.Group {
	return runnertest.NewGroup(fmt.Sprintf("%s runner group", l), l.Listener.Runners()...)
}
