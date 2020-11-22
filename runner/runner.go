package runner

import (
    "context"
    "sync"
)

// Item TODO.
type Item interface {
    Run() error
    Close(context.Context) error
}

// Group TODO.
type Group struct {
    items []Item

    closeQ chan chan context.Context
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
        wg sync.WaitGroup
    )

    // size * 2 | once each for run error and (possible) close error
    g.resultQ = make(chan error, size*2)

    // size | one close signal per item
    g.closeQ = make(chan chan context.Context, size)
    defer close(g.closeQ)

    for _, item := range g.items {
        var (
            // closeChan is buffered because we may or may not ever read from it
            closeChan = make(chan context.Context, 1)
            doneChan = make(chan error)
        )

        g.closeQ <- closeChan

        go func(i Item) { doneChan <- i.Run() }(item)

        wg.Add(1)
        go func(i Item) {
            select {
            case res := <-doneChan:
                g.resultQ <- res
            case ctx := <-closeChan:
                g.resultQ <- i.Close(ctx)
                g.resultQ <- (<-doneChan)
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
