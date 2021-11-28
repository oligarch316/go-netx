package runner_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/oligarch316/go-netx/synctest"
	runnertest "github.com/oligarch316/go-netx/synctest/runner"
)

const waitGroupSize = 10

func TestConcurrentWaitGroupBasic(t *testing.T) {
	// High level
	// - Run() and Wait() MUST block while # Done() calls < size
	// - Run() and Wait() MUST NOT block once # Done() calls == size

	wg := runnertest.NewWaitGroup("wait group", waitGroupSize)

	// Start Run() and Wait() routines
	runSig := wg.Run()
	waitSig := wg.Wait()

	// Call Done() size-1 times
	wg.Done(waitGroupSize-1).RequireState(t, synctest.Complete)

	// Check Run() and Wait() still pending
	runSig.RequireState(t, synctest.Pending)
	runSig.RequireState(t, synctest.Pending)

	// Call Done() once more
	wg.Done(1).RequireState(t, synctest.Complete)

	// Check Run() and Wait() complete
	runSig.RequireState(t, synctest.Complete)
	waitSig.RequireState(t, synctest.Complete)

	// Check no Run() error
	runSig.AssertError(t, nil)
}

func TestConcurrentWaitGroupClose(t *testing.T) {
	// High level
	// - Run() and Wait() MUST block while Close() HAS NOT been called
	// - Close(), Run() and Wait() MUST block while Close() context HAS NOT "expired"
	// - Close(), Run() and Wait() MUST NOT block once Close() context HAS "expired"
	// - Done() MUST NOT block once Run() HAS completed

	var (
		wg          = runnertest.NewWaitGroup("wait group", waitGroupSize)
		expectedErr = fmt.Sprintf("wait group closed with 1 out of %d items pending", waitGroupSize)
	)

	// Start Run() and Wait() routines
	runSig := wg.Run()
	waitSig := wg.Wait()

	// Call done size-1 times
	wg.Done(waitGroupSize-1).RequireState(t, synctest.Complete)

	// Check Run() and Wait() still pending
	runSig.RequireState(t, synctest.Pending)
	waitSig.RequireState(t, synctest.Pending)

	// Call Close()
	ctx, cancel := context.WithCancel(context.Background())
	closeSig := wg.Close(ctx)

	// Check Run(), Wait() and Close() still pending
	runSig.RequireState(t, synctest.Pending)
	waitSig.RequireState(t, synctest.Pending)
	closeSig.RequireState(t, synctest.Pending)

	// Cancel the context
	cancel()

	// Check Run(), Wait() and Close() complete
	closeSig.RequireState(t, synctest.Complete)
	runSig.RequireState(t, synctest.Complete)
	waitSig.RequireState(t, synctest.Complete)

	// Check for expected Run() error and no Close() error
	runSig.AssertError(t, expectedErr)
	closeSig.AssertError(t, nil)

	// Ensure a final Done() still completes
	wg.Done(1).AssertState(t, synctest.Complete)
}

func TestConcurrentWaitGroupDone(t *testing.T) {
	// High level
	// - Done() SHOULD block while Run() HAS NOT been called
	// - Run() and Wait() MUST NOT block once # Done() calls == size, even if
	//   all Done() calls are made before Run() was called

	wg := runnertest.NewWaitGroup("wait group", waitGroupSize)

	// Start a Wait() routine
	waitSig := wg.Wait()

	// Call Done() size times
	doneSig := wg.Done(waitGroupSize)

	// Check Wait() and Done() still pending
	waitSig.RequireState(t, synctest.Pending)
	doneSig.RequireState(t, synctest.Pending)

	// Start a Run() routine
	runSig := wg.Run()

	// Check Wait(), Done() and Run complete
	waitSig.RequireState(t, synctest.Complete)
	doneSig.RequireState(t, synctest.Complete)
	runSig.RequireState(t, synctest.Complete)

	// Check no run error
	runSig.AssertError(t, nil)
}
