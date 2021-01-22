package runner

import (
	"context"

	"github.com/oligarch316/go-netx/runner"
	"github.com/oligarch316/go-netx/synctest"
)

// Item TODO.
type Item struct {
	name    string
	wrapped runner.Item
}

// New TODO.
func New(name string, runFunc func() error, closeFunc func(context.Context) error) Item {
	return Wrap(name, runner.New(runFunc, closeFunc))
}

// Wrap TODO.
func Wrap(name string, item runner.Item) Item {
	return Item{name: name, wrapped: item}
}

// String TODO.
func (i Item) String() string { return i.name }

// Run TODO.
func (i Item) Run() synctest.ErrorSignal {
	return synctest.GoErrorSignal(i.name+" Run()", i.wrapped.Run)
}

// Close TODO.
func (i Item) Close(ctx context.Context) synctest.ErrorSignal {
	return synctest.GoErrorSignal(i.name+" Close()", func() error {
		return i.wrapped.Close(ctx)
	})
}

// CloseNow TODO.
func (i Item) CloseNow() synctest.ErrorSignal {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return i.Close(ctx)
}
