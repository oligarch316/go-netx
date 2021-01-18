package runner_test

import (
	"context"
	"fmt"

	rtest "github.com/oligarch316/go-netx/runner/runnertest"
	"github.com/stretchr/testify/assert"
)

type mockItem struct {
	name                          string
	killChan, closeChan, doneChan chan struct{}

	DidRun, DidClose *rtest.Flag
	ForceCloseError  bool
}

func newMockItem(name string) *mockItem {
	return &mockItem{
		name:      name,
		killChan:  make(chan struct{}),
		closeChan: make(chan struct{}),
		doneChan:  make(chan struct{}),

		DidRun:   rtest.NewFlag(name + " did run"),
		DidClose: rtest.NewFlag(name + " did close"),
	}
}

func (mi mockItem) String() string { return mi.name }

func (mi *mockItem) Kill() {
	close(mi.killChan)
	<-mi.doneChan
}

func (mi *mockItem) Run() error {
	mi.DidRun.Mark()
	defer close(mi.doneChan)

	select {
	case <-mi.killChan:
		return fmt.Errorf("%s forced run error", mi)
	case <-mi.closeChan:
		return nil
	}
}

func (mi *mockItem) Close(ctx context.Context) error {
	mi.DidClose.Mark()

	if mi.ForceCloseError {
		return fmt.Errorf("%s forced close error", mi)
	}

	defer close(mi.closeChan)

	select {
	case <-mi.doneChan:
	case <-ctx.Done():
	}

	return nil
}

type mockItemList []*mockItem

func (mil mockItemList) Kill() {
	for _, item := range mil {
		item.Kill()
	}
}

func (mil mockItemList) AssertDidRun(t assert.TestingT, expected bool) bool {
	res := true
	for _, item := range mil {
		res = item.DidRun.Assert(t, expected) && res
	}
	return res
}

func (mil mockItemList) AssertDidClose(t assert.TestingT, expected bool) bool {
	res := true
	for _, item := range mil {
		res = item.DidClose.Assert(t, expected) && res
	}
	return res
}
