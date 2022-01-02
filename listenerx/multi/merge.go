package multi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/oligarch316/go-netx/listenerx/retry"
)

var (
	errMergeListenerClosed = errors.New("source listener closed")
	errMergeRunnerClosed   = errors.New("runner closed")
)

const mergeAddrNetwork = "multi"

type mergeAddr struct{}

func (mergeAddr) Network() string { return mergeAddrNetwork }
func (mergeAddr) String() string  { return mergeAddrNetwork }

type mergeListener struct {
	connChan  chan net.Conn
	closeChan chan struct{}
	closeOnce sync.Once
}

func newMergeListener() *mergeListener {
	return &mergeListener{
		connChan:  make(chan net.Conn),
		closeChan: make(chan struct{}),
	}
}

func (mergeListener) Addr() net.Addr { return mergeAddr{} }

func (ml *mergeListener) Accept() (net.Conn, error) {
	select {
	case conn := <-ml.connChan:
		return conn, nil
	case <-ml.closeChan:
		return nil, net.ErrClosed
	}
}

func (ml *mergeListener) Close() error {
	ml.closeOnce.Do(func() { close(ml.closeChan) })
	return nil
}

// RunnerParams TODO.
type RunnerParams struct {
	AcceptRetryDelay retry.DelayFunc
	EventHandler     RunnerEventHandler
}

// MergeRunner TODO.
type MergeRunner struct {
	params RunnerParams

	source net.Listener
	sink   *mergeListener

	doneChan  chan struct{}
	closeChan chan struct{}
}

func newMergeRunner(params RunnerParams, source net.Listener, sink *mergeListener) *MergeRunner {
	return &MergeRunner{
		params:    params,
		source:    source,
		sink:      sink,
		doneChan:  make(chan struct{}),
		closeChan: make(chan struct{}),
	}
}

func (mr *MergeRunner) sendEvent(re RunnerEvent) { mr.params.EventHandler(re) }

// Addr TODO.
func (mr MergeRunner) Addr() net.Addr { return mr.source.Addr() }

// Run TODO.
func (mr *MergeRunner) Run() error {
	defer close(mr.doneChan)

	delay := retry.NewDelay(mr.params.AcceptRetryDelay)

	for {
		conn, err := mr.source.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				// Expected close error (happy path) => return nil
				return nil
			}

			if ne, ok := err.(net.Error); !ok || !ne.Temporary() {
				// Not nil, closed or temporary error => return err
				return err
			}

			// Temporary error => delay and retry
			attempt, delayDuration := delay.Next()
			mr.sendEvent(RunnerEventTemporaryAcceptError{
				runnerEvent:        runnerEvent{addr: mr.Addr(), err: err},
				Attempt:            attempt,
				RetryDelayDuration: delayDuration,
			})

			select {
			case <-time.After(delayDuration):
				// Retry delay has elapsed, continue
				continue
			case <-mr.closeChan:
				// Runner was closed
				return fmt.Errorf("interrupted retry: %w", errMergeRunnerClosed)
			case <-mr.sink.closeChan:
				// Target merge listener was closed
				return fmt.Errorf("interrupted retry: %w", errMergeListenerClosed)
			}
		}

		delay.Reset()

		select {
		case mr.sink.connChan <- conn:
			// Connection was successfully handed off, continue
		case <-mr.closeChan:
			// Runner was closed
			if err := conn.Close(); err != nil {
				mr.sendEvent(RunnerEventUnprocessedConnectionCloseError{
					runnerEvent: runnerEvent{addr: mr.Addr(), err: err},
				})
			}
			return fmt.Errorf("unprocessed connection: %w", errMergeRunnerClosed)
		case <-mr.sink.closeChan:
			// Target merge listener was closed
			if err := conn.Close(); err != nil {
				mr.sendEvent(RunnerEventUnprocessedConnectionCloseError{
					runnerEvent: runnerEvent{addr: mr.Addr(), err: err},
				})
			}
			return fmt.Errorf("unprocessed connection: %w", errMergeListenerClosed)
		}
	}
}

// Close TODO.
func (mr *MergeRunner) Close(ctx context.Context) error {
	if err := mr.source.Close(); err != nil {
		mr.sendEvent(RunnerEventListenerCloseError{
			runnerEvent: runnerEvent{addr: mr.Addr(), err: err},
		})

		close(mr.closeChan)
		return nil
	}

	select {
	case <-mr.doneChan:
	case <-ctx.Done():
		mr.sendEvent(RunnerEventCloseContextExpiredError{
			runnerEvent: runnerEvent{addr: mr.Addr(), err: ctx.Err()},
		})

		close(mr.closeChan)
	}
	return nil
}
