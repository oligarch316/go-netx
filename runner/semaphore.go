package runner

import (
	"context"
	"fmt"
)

// Semaphore TODO.
type Semaphore struct {
	count                        int
	semChan, closeChan, doneChan chan struct{}
}

// NewSemaphore TODO.
func NewSemaphore(count int) *Semaphore {
	return &Semaphore{
		count:     count,
		semChan:   make(chan struct{}, count),
		closeChan: make(chan struct{}),
		doneChan:  make(chan struct{}),
	}
}

// Done TODO.
func (s *Semaphore) Done() { s.semChan <- struct{}{} }

// Wait TODO.
func (s *Semaphore) Wait() { <-s.doneChan }

// Run TODO.
func (s *Semaphore) Run() error {
	var res error

L:
	for i := 0; i < s.count; i++ {
		select {
		case <-s.semChan:
			continue
		case <-s.closeChan:
			res = fmt.Errorf("semaphore closed with %d out of %d items pending", s.count-i, s.count)
			break L
		}
	}

	close(s.doneChan)
	return res
}

// Close TODO.
func (s *Semaphore) Close(ctx context.Context) error {
	select {
	case <-s.doneChan:
	case <-ctx.Done():
		close(s.closeChan)
	}

	return nil
}
