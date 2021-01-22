package runner

import (
	"context"
	"sync"
)

// TODO:
// Already I ran into the call Results() before calling Run() silent issue, and I wrote this.
// The result channel needs to be a return value from Run(), really obvious.

// Group TODO.
type Group struct {
	items []Item

	closeQ  chan chan context.Context
	resultQ chan error
}

// NewGroup TODO.
func NewGroup(items ...Item) *Group {
	res := new(Group)
	res.Append(items...)
	return res
}

// Append TODO.
func (g *Group) Append(items ...Item) { g.items = append(g.items, items...) }

// Run TODO.
func (g *Group) Run() {
	var (
		size = len(g.items)
		wg   sync.WaitGroup
	)

	// size | one (or less) of [run OR close] error per item
	g.resultQ = make(chan error, size)

	// size | one close signal per item
	g.closeQ = make(chan chan context.Context, size)
	defer close(g.closeQ)

	for _, item := range g.items {
		var (
			// closeChan is buffered because we may or may not ever read from it
			closeChan = make(chan context.Context, 1)
			doneChan  = make(chan error)
		)

		g.closeQ <- closeChan

		go func(i Item) { doneChan <- i.Run() }(item)

		wg.Add(1)
		go func(i Item) {
			select {
			case err := <-doneChan:
				// Expectations:
				// - item.Run() SHOULD block forever until a call to item.Close()
				// - a well behaved item.Run() SHOULD always return a non-nil
				//   error if item.Close() has yet to be called

				// Thus:
				// Send pre-close Run() results regardless of nil/non-nil value.
				// There must always be a consumable indication that Run() has
				// completed before Close().
				g.resultQ <- err
			case ctx := <-closeChan:
				// Expectation:
				// item.Close() SHOULD return a non-nil error if and only if
				// item.Run() cannot be relied upon to unblock.

				if err := i.Close(ctx); err != nil {
					// Thus if Close() error != nil:
					// Send the (non-nil) Close() error and abandon the Run()
					// routine as an orphan.
					g.resultQ <- err
					break
				}

				// Thus if Close() error == nil
				// Re-wait for Run() result and send it only if it's non-nil
				if err := <-doneChan; err != nil {
					g.resultQ <- err
				}
			}

			wg.Done()
		}(item)
	}

	go func() {
		wg.Wait()
		close(g.resultQ)
	}()
}

// Close TODO.
func (g *Group) Close(ctx context.Context) {
	for closeChan := range g.closeQ {
		closeChan <- ctx
	}
}

// Results TODO.
func (g *Group) Results() <-chan error { return g.resultQ }
