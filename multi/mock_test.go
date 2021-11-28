package multi_test

import (
	"context"
	"fmt"
	"net"
)

type mockAddr string

func (mockAddr) Network() string   { return "mock network" }
func (ma mockAddr) String() string { return string(ma) + " address" }

type mockConn struct {
	name string
	net.Conn
}

func (mc mockConn) String() string { return mc.name }

type mockError struct {
	msg       string
	temporary bool
}

func (me mockError) Error() string { return me.msg }

type mockListener struct {
	name string

	connChan  chan net.Conn
	errChan   chan error
	closeChan chan struct{}

	ForceCloseError, ForceDialError bool
}

func newMockListener(name string) mockListener {
	return mockListener{
		name:      name,
		connChan:  make(chan net.Conn),
		errChan:   make(chan error),
		closeChan: make(chan struct{}),
	}
}

func (ml mockListener) sendConn(c net.Conn) { ml.connChan <- c }
func (ml mockListener) sendError(err error) { ml.errChan <- err }

func (ml mockListener) String() string { return ml.name }

func (ml mockListener) Addr() net.Addr { return mockAddr(ml.name) }

func (ml mockListener) Accept() (net.Conn, error) {
	select {
	case <-ml.closeChan:
	default:
		select {
		case <-ml.closeChan:
		case res := <-ml.connChan:
			return res, nil
		case err := <-ml.errChan:
			return nil, err
		}
	}
	return nil, fmt.Errorf("listener closed")
}

func (ml mockListener) Close() error {
	if ml.ForceCloseError {
		return fmt.Errorf("%s forced close error", ml)
	}

	close(ml.closeChan)
	return nil
}

func (ml mockListener) Dial() (net.Conn, error) {
	if ml.ForceDialError {
		return nil, fmt.Errorf("%s forced dial error", ml)
	}

	res := mockConn{name: fmt.Sprintf("%s dial conn", ml)}
	ml.sendConn(res)
	return res, nil
}

func (ml mockListener) DialContext(_ context.Context) (net.Conn, error) { return ml.Dial() }
