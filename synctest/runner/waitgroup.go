package runnertest

import (
	"fmt"

	"github.com/oligarch316/go-netx/runner"
	"github.com/oligarch316/go-netx/synctest"
)

// WaitGroup TODO.
type WaitGroup struct {
	Item
	wrapped *runner.WaitGroup
}

// NewWaitGroup TODO.
func NewWaitGroup(name string, size int) WaitGroup {
	return WrapWaitGroup(name, runner.NewWaitGroup(size))
}

// WrapWaitGroup TODO.
func WrapWaitGroup(name string, wg *runner.WaitGroup) WaitGroup {
	return WaitGroup{
		wrapped: wg,
		Item:    Wrap(name, wg),
	}
}

// Done TODO.
func (wg WaitGroup) Done(n int) synctest.Signal {
	var (
		name = fmt.Sprintf("%s Done(%d)", wg.name, n)
		f    = func() {
			for i := 0; i < n; i++ {
				wg.wrapped.Done()
			}
		}
	)
	return synctest.GoSignal(name, f)
}

// Wait TODO.
func (wg WaitGroup) Wait() synctest.Signal {
	return synctest.GoSignal(wg.name+" Wait()", wg.wrapped.Wait)
}
