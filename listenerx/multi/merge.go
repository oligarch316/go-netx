package multi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
)

var (
	errMergeListenerClosed = errors.New("multi: listener closed")
	errMergeRunnerClosed   = errors.New("multi: runner closed")
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

func newMergeListener() mergeListener {
	return mergeListener{
		connChan:  make(chan net.Conn),
		closeChan: make(chan struct{}),
	}
}

func (*mergeListener) Addr() net.Addr { return mergeAddr{} }

func (ml *mergeListener) Accept() (net.Conn, error) {
	select {
	case conn := <-ml.connChan:
		return conn, nil
	case <-ml.closeChan:
		// TODO: Change to net.ErrClosed for compatible go versions
		return nil, errMergeListenerClosed
	}
}

func (ml *mergeListener) Close() error {
	ml.closeOnce.Do(func() { close(ml.closeChan) })
	return nil
}

func (ml *mergeListener) runner(l net.Listener) *mergeRunner {
	return &mergeRunner{
		l:      l,
		mergeL: ml,

		doneChan:  make(chan struct{}),
		closeChan: make(chan struct{}),
	}
}

type mergeRunner struct {
	l      net.Listener
	mergeL *mergeListener

	doneChan  chan struct{}
	closeChan chan struct{}
}

func (mr mergeRunner) Addr() net.Addr { return mr.l.Addr() }

func (mr *mergeRunner) Run() error {
	defer close(mr.doneChan)

	delay := newRetryDelay()

	for {
		conn, err := mr.l.Accept()
		if err != nil {
			ne, ok := err.(net.Error)
			if !ok || !ne.Temporary() {
				// TODO: What do we expect as a result of mr.l.Close()?
				// Can we replace that expected error with nil here?
				return err
			}

			// TODO: Log/track/surface this somehow

			select {
			case <-delay.Sleep():
				// Retry delay has elapsed, continue
				continue
			case <-mr.closeChan:
				// Runner was closed

				// TODO: Should this be a non-nil error indicating forced (non-happy-path) closure? Probably...
				return nil
			case <-mr.mergeL.closeChan:
				// Target merge listener was closed

				// TODO: Should this be a non-nil error indicating forced (non-happy-path) closure? Probably...
				return nil
			}
		}

		delay.Reset()

		select {
		case mr.mergeL.connChan <- conn:
			// Connection was successfully handed off, continue
		case <-mr.closeChan:
			// Runner was closed
			if err := conn.Close(); err != nil {
				// TODO: Log/track/surface this somehow
			}
			return fmt.Errorf("unprocessed connection: %w", errMergeRunnerClosed)
		case <-mr.mergeL.closeChan:
			// Target merge listener was closed
			if err := conn.Close(); err != nil {
				// TODO: Log/track/surface this somehow
			}
			return fmt.Errorf("unprocessed connection: %w", errMergeListenerClosed)
		}
	}
}

func (mr *mergeRunner) Close(ctx context.Context) error {
	if err := mr.l.Close(); err != nil {
		// TODO: Log/track/surface this somehow
		close(mr.closeChan)
		return nil
	}

	select {
	case <-mr.doneChan:
	case <-ctx.Done():
		// TODO: Log/track/surface this somehow
		close(mr.closeChan)
	}
	return nil
}
