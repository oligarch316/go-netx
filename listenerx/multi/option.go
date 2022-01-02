package multi

import (
	"github.com/oligarch316/go-netx/addressx"
	"github.com/oligarch316/go-netx/listenerx/retry"
)

// WithRunnerEventHandler TODO.
func WithRunnerEventHandler(handler RunnerEventHandler) ListenerOption {
	return func(p *ListenerParams) { p.Runner.EventHandler = handler }
}

// WithRunnerRetryDelay TODO.
func WithRunnerRetryDelay(delayFunc retry.DelayFunc) ListenerOption {
	return func(p *ListenerParams) { p.Runner.AcceptRetryDelay = delayFunc }
}

// WithDialerAddressOrdering TODO.
func WithDialerAddressOrdering(ordering addressx.Ordering) ListenerOption {
	return func(p *ListenerParams) { p.Dialer.AddressOrdering = ordering }
}

// WithDialerStrategy TODO.
func WithDialerStrategy(strategy DialStrategy) ListenerOption {
	return func(p *ListenerParams) { p.Dialer.Strategy = strategy }
}
