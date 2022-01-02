package multi

import (
	"fmt"
	"net"
	"time"
)

// RunnerEventHandler TODO.
type RunnerEventHandler func(RunnerEvent)

type runnerEvent struct {
	addr net.Addr
	err  error
}

func (re runnerEvent) errString(message string) string {
	return fmt.Sprintf("%s: %s", message, re.err)
}

func (re runnerEvent) Addr() net.Addr { return re.addr }
func (re runnerEvent) Unwrap() error  { return re.err }

// RunnerEvent TODO.
type RunnerEvent interface {
	Addr() net.Addr
	Unwrap() error
	error
}

type (
	// RunnerEventCloseContextExpiredError TODO.
	RunnerEventCloseContextExpiredError struct{ runnerEvent }

	// RunnerEventListenerCloseError TODO.
	RunnerEventListenerCloseError struct{ runnerEvent }

	// RunnerEventUnprocessedConnectionCloseError TODO.
	RunnerEventUnprocessedConnectionCloseError struct{ runnerEvent }

	// RunnerEventTemporaryAcceptError TODO.
	RunnerEventTemporaryAcceptError struct {
		Attempt            int
		RetryDelayDuration time.Duration
		runnerEvent
	}
)

func (e RunnerEventCloseContextExpiredError) Error() string {
	return e.errString("close context expired error")
}

func (e RunnerEventListenerCloseError) Error() string {
	return e.errString("listener close error")
}

func (e RunnerEventUnprocessedConnectionCloseError) Error() string {
	return e.errString("unprocessed connection close error")
}

func (e RunnerEventTemporaryAcceptError) Error() string {
	return e.errString("temporary accept error")
}
