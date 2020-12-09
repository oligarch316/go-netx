package netx

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	"github.com/oligarch316/go-netx/runner"
)

// ----- multiSet
// - Managed list of Listeners
// - Provides and consumes MultiAddrs for dialing

var multiID uint32

// ErrInvalidMultiAddr TODO.
var ErrInvalidMultiAddr = errors.New("invalid multi address")

// MultiAddr TODO.
type MultiAddr struct {
	net.Addr

	id  uint32
	idx int
}

type multiSet struct {
	id        uint32
	listeners []Listener
}

func newMultiSet() multiSet {
	return multiSet{id: atomic.AddUint32(&multiID, 1)}
}

func (ms multiSet) lookup(addr MultiAddr) (Listener, error) {
	switch {
	case addr.id != ms.id:
		return nil, fmt.Errorf("%w: address id '%d' does not match listener id '%d'", ErrInvalidMultiAddr, addr.id, ms.id)
	case addr.idx >= len(ms.listeners):
		return nil, fmt.Errorf("%w: address index '%d' out of bounds", ErrInvalidMultiAddr, addr.idx)
	}

	return ms.listeners[addr.idx], nil
}

func (ms multiSet) Addrs() []MultiAddr {
	var res []MultiAddr

	for idx, l := range ms.listeners {
		res = append(res, MultiAddr{
			Addr: l.Addr(),
			id:   ms.id,
			idx:  idx,
		})
	}

	return res
}

func (ms *multiSet) Append(ls ...Listener) { ms.listeners = append(ms.listeners, ls...) }

func (ms multiSet) Dial(addr MultiAddr) (net.Conn, error) {
	l, err := ms.lookup(addr)
	if err != nil {
		return nil, err
	}

	return l.Dial()
}

func (ms multiSet) DialContext(ctx context.Context, addr MultiAddr) (net.Conn, error) {
	l, err := ms.lookup(addr)
	if err != nil {
		return nil, err
	}

	return l.DialContext(ctx)
}

func (ms multiSet) ID() uint32 { return ms.id }

func (ms multiSet) Len() int { return len(ms.listeners) }

// ----- multiListener
// - Provides a channel fed net.Listener interface
// - Builds multiRunner producers for the above channel from Listeners

const addrMulti abstractAddr = "multi"

var (
	errMLAcceptListenerClosed = errors.New("listener closed")

	errMLRunRunnerClosed   = errors.New("runner closed")
	errMLRunListenerClosed = errors.New("listener closed")
)

type multiListener struct {
	connChan  chan net.Conn
	closeChan chan struct{}
	closeOnce sync.Once
}

func newMultiListener() multiListener {
	return multiListener{
		connChan:  make(chan net.Conn),
		closeChan: make(chan struct{}),
	}
}

func (ml *multiListener) runner(l net.Listener) runner.Item {
	return &multiRunner{
		l:  l,
		ml: ml,

		doneChan:  make(chan struct{}),
		closeChan: make(chan struct{}),
	}
}

func (ml *multiListener) Accept() (net.Conn, error) {
	select {
	case conn := <-ml.connChan:
		return conn, nil
	case <-ml.closeChan:
		return nil, errMLAcceptListenerClosed
	}
}

func (ml *multiListener) Addr() net.Addr { return addrMulti }

func (ml *multiListener) Close() error {
	ml.closeOnce.Do(func() { close(ml.closeChan) })
	return nil
}

// ----- multiRunner
// - Implements the runner.Item interface
// - Feeds Accpet() connections from a net.Listener into its parent multiListener

type multiRunner struct {
	l  net.Listener
	ml *multiListener

	doneChan  chan struct{}
	closeChan chan struct{}
}

func (mr *multiRunner) Run() error {
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
			case <-mr.ml.closeChan:
				// Target multiListener was closed
				return nil
			}
		}

		delay.Reset()

		select {
		case mr.ml.connChan <- conn:
			// Connection was successfully handed off, continue
		case <-mr.closeChan:
			// Runner was closed
			if err := conn.Close(); err != nil {
				// TODO: Log/track/surface this somehow
			}
			return fmt.Errorf("unprocessed connection: %w", errMLRunRunnerClosed)
		case <-mr.ml.closeChan:
			// Target multiListener was closed
			if err := conn.Close(); err != nil {
				// TODO: Log/track/surface this somehow
			}
			return fmt.Errorf("unprocessed connection: %w", errMLRunListenerClosed)
		}
	}
}

func (mr *multiRunner) Close(ctx context.Context) error {
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

// ----- MultiListener
// - Exports a multiListener/multiSet amalgam to ...
//
// --> Provide multiSet Addr/Dial functionality
// --> Provide multiListener net.Listener implementation
// --> Create (its own) multiListener runners from (its own) multiSet Listeners

// MultiListener TODO.
type MultiListener struct {
	multiListener
	multiSet
}

// NewMultiListener TODO.
func NewMultiListener(ls ...Listener) *MultiListener {
	res := &MultiListener{
		multiListener: newMultiListener(),
		multiSet:      newMultiSet(),
	}

	res.Append(ls...)
	return res
}

// Runners TODO.
func (ml *MultiListener) Runners() []runner.Item {
	res := make([]runner.Item, ml.Len())

	for i, l := range ml.listeners {
		res[i] = ml.runner(l)
	}

	return res
}

func (ml *MultiListener) String() string {
	return fmt.Sprintf("multi listener %d", ml.ID())
}
