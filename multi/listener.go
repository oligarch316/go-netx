package multi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/oligarch316/go-netx"
	"github.com/oligarch316/go-netx/runner"
)

// NetworkName TODO.
const NetworkName = "multi"

var (
	errMergeListenerClosed = errors.New("multi: listener closed")
	errMergeRunnerClosed   = errors.New("multi: runner closed")
)

type mergeAddr struct{}

func (mergeAddr) Network() string { return NetworkName }
func (mergeAddr) String() string  { return NetworkName }

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
		return nil, errMergeListenerClosed
	}
}

func (ml *mergeListener) Close() error {
	ml.closeOnce.Do(func() { close(ml.closeChan) })
	return nil
}

func (ml *mergeListener) runner(l net.Listener) runner.Item {
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

func (mr *mergeRunner) Run() error {
	defer close(mr.doneChan)

	delay := newRetryDelay()

	for {
		conn, err := mr.l.Accept()
		if err != nil {
			// TODO: check for interface{ Temporary() bool } instead of net.Error here
			ne, ok := err.(net.Error)
			if !ok || !ne.Temporary() {
				// TODO: What do we expect as a result of mr.l.Close()?
				// Can we replace that expected error with nil here?

				// Upon further reading, we may need our very own signal here just
				// to indicate the error is expected and comes from mr.l.Close() called
				// from ml.Close().
				// Meaning something like mr.closeChan becomes mr.forceCloseChan? and we
				// introduce some sort of mr.closingChan here? To do:
				// select {
				// case <-mr.closing:
				// 	=> All good, return nil
				// default:
				// 	=> Not good, return that error
				// }
				// And then mr.Close() does `close(mr.closingChan)` 1st thing?
				//
				// Another thought, this can be accomplished by just wrapping the listener to
				// do the same logic and returning a defined and exported known error. And if
				// we're doing that, maybe netx itself could own that? Actually no, bad idea,
				// but wrapping the netx.listener for this still might be a good move.
				// Maybe there's a clean way to wrap all errors such that the Temporary() checks
				// and even retry logic are made smoother as well?

				return err
			}

			// TODO: Log/track/surface this somehow

			select {
			case <-delay.Sleep():
				// Retry delay has elapsed, continue
				continue
			case <-mr.closeChan:
				// Runner was closed
				return nil
			case <-mr.mergeL.closeChan:
				// Target merge listener was closed
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

// Listener TODO.
type Listener struct {
	mergeListener
	Set
}

// NewListener TODO.
func NewListener(ls ...netx.Listener) *Listener {
	res := &Listener{
		mergeListener: newMergeListener(),
		Set:           newSet(),
	}

	res.Append(ls...)
	return res
}

// Runners TODO.
func (l *Listener) Runners() []runner.Item {
	res := make([]runner.Item, l.Len())

	for i, item := range l.listeners {
		res[i] = l.runner(item)
	}

	return res
}

func (l *Listener) String() string {
	return fmt.Sprintf("multi listener %d", l.ID())
}
