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

const (
	// NetworkName TODO.
	NetworkName = "multi"

	addrMerger netx.AbstractAddr = NetworkName
)

var (
	errListenMergerClosed = errors.New("listener closed")
	errListenRunnerClosed = errors.New("runner closed")
)

type listenMerger struct {
	connChan  chan net.Conn
	closeChan chan struct{}
	closeOnce sync.Once
}

func newListenMerger() listenMerger {
	return listenMerger{
		connChan:  make(chan net.Conn),
		closeChan: make(chan struct{}),
	}
}

func (lm *listenMerger) runner(l net.Listener) runner.Item {
	return &mergeRunner{
		l:  l,
		merger: lm,

		doneChan:  make(chan struct{}),
		closeChan: make(chan struct{}),
	}
}

func (lm *listenMerger) Accept() (net.Conn, error) {
	select {
	case conn := <-lm.connChan:
		return conn, nil
	case <-lm.closeChan:
		return nil, errListenMergerClosed
	}
}

func (*listenMerger) Addr() net.Addr { return addrMerger }

func (lm *listenMerger) Close() error {
	lm.closeOnce.Do(func() { close(lm.closeChan) })
	return nil
}

type mergeRunner struct {
	l  net.Listener
	merger *listenMerger

	doneChan  chan struct{}
	closeChan chan struct{}
}

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
				return nil
			case <-mr.merger.closeChan:
				// Target merger was closed
				return nil
			}
		}

		delay.Reset()

		select {
		case mr.merger.connChan <- conn:
			// Connection was successfully handed off, continue
		case <-mr.closeChan:
			// Runner was closed
			if err := conn.Close(); err != nil {
				// TODO: Log/track/surface this somehow
			}
			return fmt.Errorf("unprocessed connection: %w", errListenRunnerClosed)
		case <-mr.merger.closeChan:
			// Target merger was closed
			if err := conn.Close(); err != nil {
				// TODO: Log/track/surface this somehow
			}
			return fmt.Errorf("unprocessed connection: %w", errListenMergerClosed)
		}
	}
}

func (mr *mergeRunner) Close(ctx context.Context) error {
	err := mr.l.Close()

	select {
	case <-mr.doneChan:
	case <-ctx.Done():
		// TODO: Log/track/surface this somehow
		close(mr.closeChan)
		<-mr.doneChan
	}

	return err
}

// Listener TODO.
type Listener struct {
	listenMerger
	set
}

// NewListener TODO.
func NewListener(ls ...netx.Listener) *Listener {
	res := &Listener{
		listenMerger: newListenMerger(),
		set:      newSet(),
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
