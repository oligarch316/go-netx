package runner_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	realrunner "github.com/oligarch316/go-netx/runner"
	"github.com/oligarch316/go-netx/synctest"
	"github.com/oligarch316/go-netx/synctest/runner"
	"github.com/stretchr/testify/require"
)

const groupSize = 10

func requireGroupSize(t *testing.T, minSize int) {
	if groupSize < minSize {
		t.Skip(fmt.Sprintf("group size of %d is less than minimum size of %d for this test", groupSize, minSize))
	}
}

func setupGroup(name string, size int) (runner.Group, mockItemList) {
	var (
		group = runner.NewGroup(name)
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
	// - Results() channel remains open until all item.Run() calls complete
	// - Success case of nil error results from both item.Run() and item.Close()
	//   translates to no Results() channel signals save its closing

	requireGroupSize(t, 1)

	group, items := setupGroup("group", groupSize)

	// Call Run() and start Results() consumer routine
	group.Run()
	results := group.ErrorResults().All()

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
	results.AssertEqual(t)                             // ... results empty
	items.RunFlags().AssertState(t, synctest.Marked)   // ... all items ran
	items.CloseFlags().AssertState(t, synctest.Marked) // ... all items closed
}

func TestConcurrentGroupRunError(t *testing.T) {
	// High level
	// - Group results channel always includes all non-nil item.Run() errors
	// - Group results channel always closes when all item.Run() calls are
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

		// Call Run() and start Results() consumer routine
		group.Run()
		results := group.StringResults().All()

		// Call Kill() on last size-1 items
		items[1:].Kill()

		// Check results still pending
		results.RequireState(t, synctest.Pending)

		// Call Kill() on first item
		items[0].Kill()

		// Check results complete
		results.RequireState(t, synctest.Complete)

		// Ensure Close() is still safe to call
		require.NotPanics(t, group.CloseNow)

		// Check ...
		results.AssertEqual(t, expectedErrs...)              // ... expected error results
		items.RunFlags().AssertState(t, synctest.Marked)     // ... all items ran
		items.CloseFlags().AssertState(t, synctest.Unmarked) // ... no items closed
	})

	t.Run("one before close", func(t *testing.T) {
		var (
			group, items = setupGroup("group", groupSize)
			expectedErr  = "mock item 0 forced run error"
		)

		// Call Run() and start Results() consumer routine
		group.Run()
		results := group.StringResults().All()

		// Call Kill() on first item
		items[0].Kill()

		// Check results still pending
		results.RequireState(t, synctest.Pending)

		// Call Close() (and cancel context)
		group.CloseNow()

		// Check results complete
		results.RequireState(t, synctest.Complete)

		// Check ...
		results.AssertEqual(t, expectedErr)                    // ... expected error result
		items.RunFlags().AssertState(t, synctest.Marked)       // ... all items ran
		items[1:].CloseFlags().AssertState(t, synctest.Marked) // ... last size-1 items closed
		items[0].CloseFlag.AssertState(t, synctest.Unmarked)   // ... first item did NOT close
	})

	t.Run("one mid-close", func(t *testing.T) {
		var (
			group, items = setupGroup("group", groupSize)
			expectedErr  = "mock item 0 forced run error"
		)

		// Call Run() and start Results() consumer routine
		group.Run()
		results := group.StringResults().All()

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
		results.AssertEqual(t, expectedErr)                // ... expected error result
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
		group.Append(realrunner.New(
			func() error {
				specialItem.Run()
				return forcedErr
			},
			specialItem.Close,
		))

		// Call Run() and start Results() consumer routine
		group.Run()
		results := group.ErrorResults().All()

		// Check results still pending
		results.RequireState(t, synctest.Pending)

		// Call Close() (and cancel context)
		group.CloseNow()

		// Check results complete
		results.RequireState(t, synctest.Complete)

		// Check ...
		results.AssertEqual(t, forcedErr)                  // ... results include post-close forced error
		items.RunFlags().AssertState(t, synctest.Marked)   // ... all items ran
		items.CloseFlags().AssertState(t, synctest.Marked) // ... all items closed
	})
}

func TestConcurrentGroupCloseError(t *testing.T) {
	// High level
	// - When item.Close() returns a non-nil error, group results MUST NOT wait
	//   for that same item's Run() call to complete before closing
	// - item.Close() errors MUST appear in the group results channel

	requireGroupSize(t, 1)

	var (
		group, items = setupGroup("group", groupSize)
		expectedErr  = "mock item 0 forced close error"
	)

	// Force a close error on first item
	items[0].ForceCloseError = true

	// Call Run() and start Results() consumer routine
	group.Run()
	results := group.StringResults().All()

	// Check results still pending
	results.RequireState(t, synctest.Pending)

	// Call Close() (and cancel context)
	group.CloseNow()

	// Check results complete (despite hanging items[0].Run())
	results.RequireState(t, synctest.Complete)

	// Call Kill() on first item for cleanup
	items[0].Kill()

	// Check ...
	results.AssertEqual(t, expectedErr)                // ... expected error result
	items.RunFlags().AssertState(t, synctest.Marked)   // ... all items ran
	items.CloseFlags().AssertState(t, synctest.Marked) // ... all items closed
}

func TestConcurrentGroupLateResultsConsumer(t *testing.T) {
	// High level
	// - Ensure no issues arise when items complete their lifecycles before
	//   any reads from the Results() channel occur

	requireGroupSize(t, 2)

	// var (
	// 	group, items = setupGroup(groupSize)
	// 	ctx, cancel  = context.WithCancel(context.Background())
	// 	expectedErrs = []string{
	// 		"mock item 0 forced run error",
	// 		"mock item 1 forced close error",
	// 	}
	// )

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
	group.Run()

	// Call Kill() on first item
	items[0].Kill()

	// Call Close() (and cancel context)
	group.CloseNow()

	// Start Results() consumer routine
	results := group.StringResults().All()

	// Check results complete
	results.RequireState(t, synctest.Complete)

	// Call Kill() on second item for cleanup
	items[1].Kill()

	// Check ...
	results.AssertEqual(t, expectedErrs...)                // ... expected error results
	items.RunFlags().AssertState(t, synctest.Marked)       // ... all items ran
	items[1:].CloseFlags().AssertState(t, synctest.Marked) // ... last size-1 items closed
	items[0].CloseFlag.AssertState(t, synctest.Unmarked)   // ... first item did NOT close
}

func TestConcurrentGroupTimelyResults(t *testing.T) {
	// High level
	// - Ensure errors from (and closing of) results channel manifest as they
	//   occur

	t.Skip("ergonomics fix required, see group.go TODO")

	requireGroupSize(t, 1)

	var (
		group, items = setupGroup("group", groupSize)
		resultChan   = group.StringResults()
	)

	// Call Run()
	group.Run()

	// For each item one by one ...
	for _, item := range items {
		t.Logf("processing %s\n", item)

		expectedErr := fmt.Sprintf("%s forced run error", item)

		// ... start single result consumer routine
		result := resultChan.Next(1)

		// ... check single result still pending
		result.RequireState(t, synctest.Pending)

		// ... call Kill() on item
		item.Kill()

		// ... check single result complete with expected error
		result.RequireState(t, synctest.Complete)
		result.AssertEqual(t, expectedErr)
	}

	// Start remainder result consumer routine
	remainder := resultChan.All()

	// Check remainder results complete and empty
	remainder.RequireState(t, synctest.Complete)
	remainder.AssertEqual(t)
}
