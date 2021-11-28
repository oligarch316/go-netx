package runnertest

import (
	"context"
	"fmt"

	"github.com/oligarch316/go-netx/runner"
	"github.com/oligarch316/go-netx/synctest"
)

// Group TODO.
type Group struct {
	name string
	*runner.Group
}

// NewGroup TODO.
func NewGroup(name string, items ...runner.Item) Group {
	return WrapGroup(name, runner.NewGroup(items...))
}

// WrapGroup TODO.
func WrapGroup(name string, group *runner.Group) Group {
	return Group{name: name, Group: group}
}

func (g Group) String() string { return g.name }

// CloseNow TODO.
func (g Group) CloseNow() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	g.Close(ctx)
}

// Run TODO.
func (g Group) Run() GroupResultChannel {
	return GroupResultChannel{
		name: g.name + " result channel",
		c:    g.Group.Run(),
	}
}

// GroupResultChannel TODO.
type GroupResultChannel struct {
	name string
	c    <-chan error
}

func (grc GroupResultChannel) String() string { return grc.name }

// All TODO.
func (grc GroupResultChannel) All() *synctest.ErrorSetSignal {
	name := fmt.Sprintf("%s (read all)", grc.name)
	read := func() []error {
		var res []error
		for err := range grc.c {
			res = append(res, err)
		}
		return res
	}

	return synctest.GoSignalErrorSet(name, read)
}

// Next TODO.
func (grc GroupResultChannel) Next(n int) *synctest.ErrorSetSignal {
	name := fmt.Sprintf("%s (read %d)", grc.name, n)
	read := func() []error {
		var res []error
		for i := 0; i < n; i++ {
			err, more := <-grc.c
			if !more {
				break
			}
			res = append(res, err)
		}
		return res
	}
	return synctest.GoSignalErrorSet(name, read)
}
