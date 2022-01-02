package multi

import (
	"fmt"

	"github.com/oligarch316/go-netx"
)

// ListenerOption TODO.
type ListenerOption func(*ListenerParams)

// ListenerParams TODO.
type ListenerParams struct {
	DialerParams
	RunnerParams
}

func defaultListenerParams() ListenerParams {
	return ListenerParams{
		// TODO
	}
}

// Listener TODO.
type Listener struct {
	*Dialer
	*mergeListener

	runnerParams RunnerParams
}

// NewListener TODO.
func NewListener(ls []netx.Listener, opts ...ListenerOption) *Listener {
	params := defaultListenerParams()
	for _, opt := range opts {
		opt(&params)
	}

	return &Listener{
		Dialer:        newDialer(params.DialerParams, ls),
		mergeListener: newMergeListener(),
		runnerParams:  params.RunnerParams,
	}
}

// Runners TODO.
func (l *Listener) Runners() []*MergeRunner {
	res := make([]*MergeRunner, l.Len())

	for i, item := range l.set.listeners {
		res[i] = newMergeRunner(l.runnerParams, item, l.mergeListener)
	}

	return res
}

func (l *Listener) String() string { return fmt.Sprintf("multi listener %d", l.set.id) }
