package runner_test

import (
	rtest "github.com/oligarch316/go-netx/runner/runnertest"
)

type mockItem struct {
	name                          string
	killChan, closeChan, doneChan chan struct{}

	DidRun, DidClose *rtest.Flag
	ForceCloseError  bool
}
