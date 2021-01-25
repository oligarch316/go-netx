package runner_test

import (
	"context"
	"fmt"

	"github.com/oligarch316/go-netx/synctest"
)

type mockItem struct {
	name                          string
	killChan, closeChan, doneChan chan struct{}

	RunFlag, CloseFlag *synctest.Marker
	ForceCloseError    bool
}

func newMockItem(name string) *mockItem {
	return &mockItem{
		name:      name,
		killChan:  make(chan struct{}),
		closeChan: make(chan struct{}),
		doneChan:  make(chan struct{}),

		RunFlag:   synctest.NewMarker(name + " Run() flag"),
		CloseFlag: synctest.NewMarker(name + " Close() flag"),
	}
}

func (mi mockItem) String() string { return mi.name }

func (mi *mockItem) Kill() {
	close(mi.killChan)

	// Necessary to ensure our kill signal doesn't complete/race with race mi.Close(...) during tests
	<-mi.doneChan
}

func (mi *mockItem) Run() error {
	mi.RunFlag.Mark()
	defer close(mi.doneChan)

	select {
	case <-mi.killChan:
		return fmt.Errorf("%s forced run error", mi)
	case <-mi.closeChan:
		return nil
	}
}

func (mi *mockItem) Close(ctx context.Context) error {
	mi.CloseFlag.Mark()

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

func (mil mockItemList) RunFlags() (res synctest.MarkerList) {
	for _, item := range mil {
		res = append(res, item.RunFlag)
	}
	return
}

func (mil mockItemList) CloseFlags() (res synctest.MarkerList) {
	for _, item := range mil {
		res = append(res, item.CloseFlag)
	}
	return
}
