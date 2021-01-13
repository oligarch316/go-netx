package runner_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/oligarch316/go-netx/runner"
	rtest "github.com/oligarch316/go-netx/runner/runnertest"
)

type testWG struct {
	wg *runner.WaitGroup
	rtest.Item
}

func newWaitGroup(size int) *testWG {
	wg := runner.NewWaitGroup(size)
	return &testWG{wg: wg, Item: rtest.WrapItem(wg)}
}

func (twg *testWG) Done() rtest.Signal { return rtest.GoSignal("done", twg.wg.Done) }
func (twg *testWG) Wait() rtest.Signal { return rtest.GoSignal("wait", twg.wg.Wait) }

func TestWaitGroupBasic(t *testing.T) {
	// High level
	// - Run() and Wait() MUST block while # Done() calls < size
	// - Run() and Wait() MUST NOT block once # Done() calls == size

	var (
		size = 16
		wg   = newWaitGroup(size)
	)

	// Start a run routine
	t.Logf("beginning Run() of wait group size %d\n", size)
	runSig := wg.Run()

	// Start a wait routine
	t.Log("beginning Wait()")
	waitSig := wg.Wait()

	// Call done size-1 times (and check none hang)
	t.Logf("calling Done() %d times\n", size-1)
	for i := 0; i < size-1; i++ {
		wg.Done().Require(t, rtest.Complete)
	}

	// Check that run and wait are still pending
	runSig.Require(t, rtest.Pending)
	waitSig.Require(t, rtest.Pending)

	// Call the last done (and check it doesn't hang)
	t.Log("calling Done() 1 more time")
	wg.Done().Require(t, rtest.Complete)

	// Check that run and wait are complete
	runSig.Require(t, rtest.Complete)
	waitSig.Require(t, rtest.Complete)

	// Check no run error
	runSig.AssertError(t, nil)
}

func TestWaitGroupClose(t *testing.T) {
	// High level
	// - Run() and Wait() MUST block while Close() HAS NOT been called
	// - Close(), Run() and Wait() MUST block while Close() context HAS NOT "expired"
	// - Close(), Run() and Wait() MUST NOT block once Close() context HAS "expired"
	// - Done() MUST NOT block once Run() HAS completed

	var (
		size        = 16
		wg          = newWaitGroup(size)
		expectedErr = fmt.Sprintf("wait group closed with 1 out of %d items pending", size)
	)

	// Start a run routine (and check for expected error)
	t.Logf("beginning Run() of wait group size %d\n", size)
	runSig := wg.Run()

	// Start a wait routine
	t.Log("beginning Wait()")
	waitSig := wg.Wait()

	// Call done size-1 times (and check none hang)
	t.Logf("calling Done() %d times\n", size-1)
	for i := 0; i < size-1; i++ {
		wg.Done().Require(t, rtest.Complete)
	}

	// Check that run and wait are still pending
	runSig.Require(t, rtest.Pending)
	waitSig.Require(t, rtest.Pending)

	// Call close
	t.Log("calling Close()")
	ctx, cancel := context.WithCancel(context.Background())
	closeSig := wg.Close(ctx)

	// Check that close, run and wait are all still pending
	closeSig.Require(t, rtest.Pending)
	runSig.Require(t, rtest.Pending)
	waitSig.Require(t, rtest.Pending)

	// Cancel the context
	t.Log("canceling context")
	cancel()

	// Check that close, run and wait are all complete
	closeSig.Require(t, rtest.Complete)
	runSig.Require(t, rtest.Complete)
	waitSig.Require(t, rtest.Complete)

	// Check for expected run error and no close error
	runSig.AssertError(t, expectedErr)
	closeSig.AssertError(t, nil)

	// Call the last done (and check it doesn't hang)
	t.Log("calling Done() 1 more time")
	wg.Done().Require(t, rtest.Complete)
}

func TestWaitGroupDone(t *testing.T) {
	// High level
	// - Done() SHOULD block while Run() HAS NOT been called
	// - Run() and Wait() MUST NOT block once # Done() calls == size, even if
	//   all Done() calls are made before Run() was called

	var (
		size        = 16
		wg          = newWaitGroup(size)
		doneSignals = make([]rtest.Signal, size)
	)

	requireDoneSignals := func(state rtest.State) {
		success := true

		for _, sig := range doneSignals {
			success = sig.Assert(t, state) && success
		}

		if !success {
			t.FailNow()
		}
	}

	// Start a wait routine
	t.Log("beginning Wait()")
	waitSig := wg.Wait()

	// Call done size times
	t.Logf("calling Done() %d times\n", size)
	for i := 0; i < size; i++ {
		doneSignals[i] = wg.Done()
	}

	// Check done and wait calls are pending
	requireDoneSignals(rtest.Pending)
	waitSig.Require(t, rtest.Pending)

	// Start a run routine
	t.Logf("beginning Run() of wait group size %d\n", size)
	runSig := wg.Run()

	// Check that done, run and wait calls are complete
	requireDoneSignals(rtest.Complete)
	runSig.Require(t, rtest.Complete)
	waitSig.Require(t, rtest.Complete)

	// Check no run error
	runSig.AssertError(t, nil)
}
