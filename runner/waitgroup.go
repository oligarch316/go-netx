package runner

import (
	"context"
	"fmt"
)

// WaitGroup TODO.
type WaitGroup struct {
	size                          int
	itemChan, closeChan, doneChan chan struct{}
}

// NewWaitGroup TODO.
func NewWaitGroup(size int) *WaitGroup {
	return &WaitGroup{
		size:      size,
		itemChan:  make(chan struct{}),
		closeChan: make(chan struct{}),
		doneChan:  make(chan struct{}),
	}
}

// Done TODO.
func (wg *WaitGroup) Done() { <-wg.itemChan }

// Wait TODO.
func (wg *WaitGroup) Wait() { <-wg.doneChan }

// Run TODO.
func (wg *WaitGroup) Run() error {
	var (
		res error
		msg struct{}
	)

L:
	for i := 0; i < wg.size; i++ {
		select {
		case wg.itemChan <- msg:
			continue
		case <-wg.closeChan:
			res = fmt.Errorf("wait group closed with %d out of %d items pending", wg.size-i, wg.size)
			break L
		}
	}

	close(wg.itemChan)
	close(wg.doneChan)
	return res
}

// Close TODO.
func (wg *WaitGroup) Close(ctx context.Context) error {
	select {
	case <-wg.doneChan:
	case <-ctx.Done():
		close(wg.closeChan)
	}

	return nil
}
