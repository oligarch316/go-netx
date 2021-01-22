package runner

import (
	"context"
	"fmt"

	"github.com/oligarch316/go-netx/runner"
	"github.com/oligarch316/go-netx/synctest"
)

// GroupResults TODO.
type GroupResults struct {
	name      string
	c         <-chan error
	handleErr func(error) interface{}
}

func (gr GroupResults) String() string { return gr.name }

// Next TODO.
func (gr GroupResults) Next(n int) synctest.SetSignal {
	var (
		name = fmt.Sprintf("%s (%d)", gr, n)
		f    = func() (res []interface{}) {
			for i := 0; i < n; i++ {
				err, more := <-gr.c
				if !more {
					break
				}
				res = append(res, gr.handleErr(err))
			}
			return
		}
	)

	return synctest.GoSetSignal(name, f)
}

// All TODO.
func (gr GroupResults) All() synctest.SetSignal {
	var (
		name = fmt.Sprintf("%s (all)", gr)
		f    = func() (res []interface{}) {
			for err := range gr.c {
				res = append(res, gr.handleErr(err))
			}
			return
		}
	)

	return synctest.GoSetSignal(name, f)
}

// Group TODO.
type Group struct {
	name    string
	wrapped *runner.Group
}

// NewGroup TODO.
func NewGroup(name string, items ...runner.Item) Group {
	return WrapGroup(name, runner.NewGroup(items...))
}

// WrapGroup TODO.
func WrapGroup(name string, group *runner.Group) Group {
	return Group{name: name, wrapped: group}
}

func (g Group) String() string { return g.name }

// Append TODO.
func (g Group) Append(items ...runner.Item) { g.wrapped.Append(items...) }

// Run TODO.
func (g Group) Run() { g.wrapped.Run() }

// Close TODO.
func (g Group) Close(ctx context.Context) { g.wrapped.Close(ctx) }

// CloseNow TODO.
func (g Group) CloseNow() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	g.Close(ctx)
}

// Results TODO.
func (g Group) Results(handler func(error) interface{}) GroupResults {
	return GroupResults{
		name:      g.name + " <-Results()",
		c:         g.wrapped.Results(),
		handleErr: handler,
	}
}

// ErrorResults TODO.
func (g Group) ErrorResults() GroupResults {
	return g.Results(func(err error) interface{} { return err })
}

// StringResults TODO.
func (g Group) StringResults() GroupResults {
	return g.Results(func(err error) interface{} {
		if err != nil {
			return err.Error()
		}
		return "<nil>"
	})
}
