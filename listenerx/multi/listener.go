package multi

import (
	"fmt"
	"time"

	"github.com/oligarch316/go-netx"
	"github.com/oligarch316/go-netx/addressx"
	"github.com/oligarch316/go-netx/listenerx"
	"github.com/oligarch316/go-netx/listenerx/retry"
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
		DialerParams: DialerParams{
			AddressOrdering: addressx.Ordering{
				addressx.ByPriorityNetwork(listenerx.InternalNetwork, "unix", "tcp"),
			},
			Strategy: DialStrategyFirstOnly,
		},
		RunnerParams: RunnerParams{
			AcceptRetryDelay: retry.DelayFuncExponential(5*time.Millisecond, 1*time.Second, 2),
			EventHandler:     func(RunnerEvent) {},
		},
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
