package runner

import "context"

// Item TODO.
type Item interface {
	Run() error
	Close(context.Context) error
}

type item struct {
	doRun   func() error
	doClose func(context.Context) error
}

func (i item) Run() error { return i.doRun() }

func (i item) Close(ctx context.Context) error { return i.doClose(ctx) }

// New TODO.
func New(runFunc func() error, closeFunc func(context.Context) error) Item {
	return item{doRun: runFunc, doClose: closeFunc}
}
