package runner_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/oligarch316/go-netx/runner"
	"github.com/oligarch316/go-netx/synctest"
	runnertest "github.com/oligarch316/go-netx/synctest/runner"
	"github.com/stretchr/testify/require"
)

const groupSize = 10

func requireGroupSize(t *testing.T, minSize int) {
	if groupSize < minSize {
		t.Skip(fmt.Sprintf("group size of %d is less than minimum size of %d for this test", groupSize, minSize))
	}
}

func setupGroup(name string, size int) (runnertest.Group, mockItemList) {
	var (
		group = runnertest.NewGroup(name)
		items = make(mockItemList, size)
	)

	for i := 0; i < size; i++ {
		items[i] = newMockItem(fmt.Sprintf("mock item %d", i))
		group.Append(items[i])
	}

	return group, items
}

func TestConcurrentGroupSuccess(t *testing.T) {
	// High level
	// Ensure basic success path behavior for group:
	// - Run() runs all item.Run() functions
	// - Close() calls all item.Close() functions
	// - Run() result channel remains open until all item.Run() calls complete
	// - Success case of nil error results from both item.Run() and item.Close()
	//   translates to no result channel signals save its closing

	requireGroupSize(t, 1)

	group, items := setupGroup("group", groupSize)

	// Call Run() and start result consumer routine
	results := group.Run().All()

	// Check results still pending
	results.RequireState(t, synctest.Pending)

	// Call Close()
	ctx, cancel := context.WithCancel(context.Background())
	group.Close(ctx)

	// Check results still pending
	results.RequireState(t, synctest.Pending)

	// Cancel the context
	cancel()

	// Check results complete
	results.RequireState(t, synctest.Complete)

	// Check ...
	results.AssertErrorSet(t)                          // ... results empty
	items.RunFlags().AssertState(t, synctest.Marked)   // ... all items ran
	items.CloseFlags().AssertState(t, synctest.Marked) // ... all items closed
}

func TestConcurrentGroupRunError(t *testing.T) {
	// High level
	// - Group result channel always includes all non-nil item.Run() errors
	// - Group result channel always closes when all item.Run() calls are
	//   complete, regardless of whether or not Close() has been called

	requireGroupSize(t, 2)

	t.Run("all before close", func(t *testing.T) {
		var (
			group, items = setupGroup("group", groupSize)
			expectedErrs = make([]interface{}, len(items))
		)

		for i, item := range items {
			expectedErrs[i] = fmt.Sprintf("%s forced run error", item)
		}

		// Call Run() and start result consumer routine
		results := group.Run().All()

		// Call Kill() on last size-1 items
		items[1:].Kill()

		// Check results still pending
		results.RequireState(t, synctest.Pending)

		// Call Kill() on first item
		items[0].Kill()

		// Check results complete
		results.RequireState(t, synctest.Complete)

		// TODO: Remove testify dependency
		// Ensure Close() is still safe to call
		require.NotPanics(t, group.CloseNow)

		// Check ...
		results.AssertErrorSet(t, expectedErrs...)           // ... expected error results
		items.RunFlags().AssertState(t, synctest.Marked)     // ... all items ran
		items.CloseFlags().AssertState(t, synctest.Unmarked) // ... no items closed
	})

	t.Run("one before close", func(t *testing.T) {
		var (
			group, items = setupGroup("group", groupSize)
			expectedErr  = "mock item 0 forced run error"
		)

		// Call Run() and start result consumer routine for 1 result
		resultChan := group.Run()
		firstResult := resultChan.Next(1)

		// Call Kill() on first item and check first result complete
		items[0].Kill()
		firstResult.RequireState(t, synctest.Complete)

		// Start result consumer routine for remaining results and check pending
		remainingResults := resultChan.All()
		remainingResults.RequireState(t, synctest.Pending)

		// Call Close() (and cancel context) and check remaining results complete
		group.CloseNow()
		remainingResults.RequireState(t, synctest.Complete)

		// Check ...
		firstResult.AssertErrorSet(t, expectedErr)             // ... expected error result
		remainingResults.AssertErrorSet(t)                     // ... no other errors
		items.RunFlags().AssertState(t, synctest.Marked)       // ... all items ran
		items[1:].CloseFlags().AssertState(t, synctest.Marked) // ... last size-1 items closed
		items[0].CloseFlag.AssertState(t, synctest.Unmarked)   // ... first item did NOT close
	})

	t.Run("one mid-close", func(t *testing.T) {
		var (
			group, items = setupGroup("group", groupSize)
			expectedErr  = "mock item 0 forced run error"
		)

		// Call Run() and start result consumer routine
		results := group.Run().All()

		// Check results still pending
		results.RequireState(t, synctest.Pending)

		// Call Close()
		ctx, cancel := context.WithCancel(context.Background())
		group.Close(ctx)

		// Check results still pending
		results.RequireState(t, synctest.Pending)

		// Call Kill() on first item
		items[0].Kill()

		// Check results still pending
		results.RequireState(t, synctest.Pending)

		// Cancel the context
		cancel()

		// Check results complete
		results.RequireState(t, synctest.Complete)

		// Check ...
		results.AssertErrorSet(t, expectedErr)             // ... expected error result
		items.RunFlags().AssertState(t, synctest.Marked)   // ... all items ran
		items.CloseFlags().AssertState(t, synctest.Marked) // ... all items closed
	})

	t.Run("one after close", func(t *testing.T) {
		var (
			group, items = setupGroup("group", groupSize-1)
			forcedErr    = errors.New("forced run after close error")
			specialItem  = newMockItem("mock item special")
		)

		items = append(items, specialItem)
		group.Append(runner.New(
			func() error {
				specialItem.Run()
				return forcedErr
			},
			specialItem.Close,
		))

		// Call Run() and start result consumer routine
		results := group.Run().All()

		// Check results still pending
		results.RequireState(t, synctest.Pending)

		// Call Close() (and cancel context)
		group.CloseNow()

		// Check results complete
		results.RequireState(t, synctest.Complete)

		// Check ...
		results.AssertErrorSet(t, forcedErr)               // ... results include post-close forced error
		items.RunFlags().AssertState(t, synctest.Marked)   // ... all items ran
		items.CloseFlags().AssertState(t, synctest.Marked) // ... all items closed
	})
}

func TestConcurrentGroupCloseError(t *testing.T) {
	// High level
	// - When item.Close() returns a non-nil error, group result channel MUST NOT wait
	//   for that same item's Run() call to complete before closing
	// - item.Close() errors MUST appear in the group result channel

	requireGroupSize(t, 1)

	var (
		group, items = setupGroup("group", groupSize)
		expectedErr  = "mock item 0 forced close error"
	)

	// Force a close error on first item
	items[0].ForceCloseError = true

	// Call Run() and start result consumer routine
	results := group.Run().All()

	// Check results still pending
	results.RequireState(t, synctest.Pending)

	// Call Close() (and cancel context)
	group.CloseNow()

	// Check results complete (despite hanging items[0].Run())
	results.RequireState(t, synctest.Complete)

	// Call Kill() on first item for cleanup
	items[0].Kill()

	// Check ...
	results.AssertErrorSet(t, expectedErr)             // ... expected error result
	items.RunFlags().AssertState(t, synctest.Marked)   // ... all items ran
	items.CloseFlags().AssertState(t, synctest.Marked) // ... all items closed
}

func TestConcurrentGroupLateResultsConsumer(t *testing.T) {
	// High level
	// - Ensure no issues arise when items complete their lifecycles before
	//   any reads from the result channel occur

	requireGroupSize(t, 2)

	var (
		group, items = setupGroup("group", groupSize)
		expectedErrs = []interface{}{
			"mock item 0 forced run error",
			"mock item 1 forced close error",
		}
	)

	// Force a close error on second item
	items[1].ForceCloseError = true

	// Call Run()
	resultChan := group.Run()

	// Call Kill() on first item
	items[0].Kill()

	// Call Close() (and cancel context)
	group.CloseNow()

	// Start result consumer routine
	results := resultChan.All()

	// Check results complete
	results.RequireState(t, synctest.Complete)

	// Call Kill() on second item for cleanup
	items[1].Kill()

	// Check ...
	results.AssertErrorSet(t, expectedErrs...)             // ... expected error results
	items.RunFlags().AssertState(t, synctest.Marked)       // ... all items ran
	items[1:].CloseFlags().AssertState(t, synctest.Marked) // ... last size-1 items closed

	// NOTE:
	// Don't make any assertions about items[0].CloseFlag Marked/Unmarked.
	// Without reading the result from Kill() before calling group.Close()
	// whether items[0].Close() is called remains ambiguous.
	// Specific behavior for this case is tested in RunError/one_before_close.
}

func TestConcurrentGroupTimelyResults(t *testing.T) {
	// High level
	// - Ensure errors from (and closing of) result channel manifest as they
	//   occur

	requireGroupSize(t, 1)

	group, items := setupGroup("group", groupSize)

	// Call Run()
	resultChan := group.Run()

	// For each item one by one ...
	for _, item := range items {
		expectedErr := fmt.Sprintf("%s forced run error", item)

		// ... start consumer routine for 1 result
		result := resultChan.Next(1)

		// ... check single result still pending
		result.RequireState(t, synctest.Pending)

		// ... call Kill() on item
		item.Kill()

		// ... check single result complete with expected error
		result.RequireState(t, synctest.Complete)
		result.AssertErrorSet(t, expectedErr)
	}

	// Start result consumer routine for remaining results
	remainder := resultChan.All()

	// Check remaining results complete and empty
	remainder.RequireState(t, synctest.Complete)
	remainder.AssertErrorSet(t)
}
