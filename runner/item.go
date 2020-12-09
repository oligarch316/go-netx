package runner

import "context"

// Item TODO.
type Item interface {
	Run() error
	Close(context.Context) error
}

type runner struct {
	doRun   func() error
	doClose func(context.Context) error
}

func (r runner) Run() error { return r.doRun() }

func (r runner) Close(ctx context.Context) error { return r.doClose(ctx) }

// New TODO.
func New(runFunc func() error, closeFunc func(context.Context) error) Item {
	return runner{doRun: runFunc, doClose: closeFunc}
}
