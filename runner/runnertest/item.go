package runnertest

import (
	"context"

	"github.com/oligarch316/go-netx/runner"
)

// Item TODO.
type Item struct{ item runner.Item }

// WrapItem TODO.
func WrapItem(item runner.Item) Item { return Item{item} }

// Run TODO.
func (i Item) Run() *ErrorSignal { return GoErrorSignal("run", i.item.Run) }

// Close TODO.
func (i Item) Close(ctx context.Context) *ErrorSignal {
	return GoErrorSignal("close", func() error { return i.item.Close(ctx) })
}
