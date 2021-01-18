package runner_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/oligarch316/go-netx/runner"
	rtest "github.com/oligarch316/go-netx/runner/runnertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const groupSize = 10

func requireGroupSize(t *testing.T, minSize int) {
	if groupSize < minSize {
		t.Skip(fmt.Sprintf("group size of %d is less than minimum size of %d for this test", groupSize, minSize))
	}
}

type groupResultSignal struct {
	*rtest.ErrorSignal
	Closed bool
}

func goGroupResultSignal(group *runner.Group) *groupResultSignal {
	res := new(groupResultSignal)
	res.ErrorSignal = rtest.GoErrorSignal("result", func() error {
		err, more := <-group.Results()
		res.Closed = !more
		return err
	})
	return res
}

func (grs *groupResultSignal) RequireClosed(t require.TestingT, expected bool) {
	rtest.Bool(expected).Require(t, grs.Closed, "%s channel closed", grs)
}

type groupResultListSignal struct {
	rtest.Signal
	Errors []error
}

func goGroupResultListSignal(group *runner.Group) *groupResultListSignal {
	res := new(groupResultListSignal)
	res.Signal = rtest.GoSignal("all results", func() {
		for err := range group.Results() {
			res.Errors = append(res.Errors, err)
		}
	})
	return res
}

func (grls *groupResultListSignal) AssertErrors(t assert.TestingT, expected ...string) bool {
	if len(expected) == 0 {
		return assert.Empty(t, grls.Errors, grls.String())
	}

	actual := make([]string, len(grls.Errors))
	for i, err := range grls.Errors {
		if err == nil {
			actual[i] = "nil"
		} else {
			actual[i] = err.Error()
		}
	}

	return assert.ElementsMatch(t, expected, actual, grls.String())
}

func setupGroup(size int) (*runner.Group, mockItemList) {
	var (
		group = runner.NewGroup()
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

	group, items := setupGroup(groupSize)

	t.Logf("calling Run() on group of size %d\n", groupSize)
	group.Run()

	t.Log("beginning consumption of Results()")
	results := goGroupResultListSignal(group)

	results.Require(t, rtest.Pending)

	t.Log("calling Close(...) on group")
	ctx, cancel := context.WithCancel(context.Background())
	group.Close(ctx)

	results.Require(t, rtest.Pending)

	t.Log("canceling Close(...) context")
	cancel()

	results.Require(t, rtest.Complete)

	results.AssertErrors(t)
	items.AssertDidRun(t, true)
	items.AssertDidClose(t, true)
}

func TestConcurrentGroupRunError(t *testing.T) {
	// High level
	// - Group results channel always includes all non-nil item.Run() errors
	// - Group results channel always closes when all item.Run() calls are
	//   complete, regarless of whether or not Close() has been called

	requireGroupSize(t, 2)

	t.Run("all before close", func(t *testing.T) {
		var (
			group, items = setupGroup(groupSize)
			expectedErrs = make([]string, groupSize)
		)

		for i := 0; i < groupSize; i++ {
			expectedErrs[i] = fmt.Sprintf("mock item %d forced run error", i)
		}

		t.Logf("calling Run() on group of size %d\n", groupSize)
		group.Run()

		t.Log("beginning consumption of Results()")
		results := goGroupResultListSignal(group)

		t.Logf("calling Kill() on last %d items\n", groupSize-1)
		items[1:].Kill()

		results.Require(t, rtest.Pending)

		t.Log("calling Kill() on first item")
		items[0].Kill()

		results.Require(t, rtest.Complete)

		t.Log("calling Close(...) on group and canceling context")
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		require.NotPanics(t, func() { group.Close(ctx) })

		results.AssertErrors(t, expectedErrs...)
		items.AssertDidRun(t, true)
		items.AssertDidClose(t, false)
	})

	t.Run("one before close", func(t *testing.T) {
		var (
			group, items = setupGroup(groupSize)
			expectedErr  = "mock item 0 forced run error"
		)

		t.Logf("calling Run() on group of size %d\n", groupSize)
		group.Run()

		t.Log("beginning consumption of Results()")
		results := goGroupResultListSignal(group)

		t.Log("calling Kill() on first item")
		items[0].Kill()

		results.Require(t, rtest.Pending)

		t.Log("calling Close(...) on group and canceling context")
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		group.Close(ctx)

		results.Require(t, rtest.Complete)

		results.AssertErrors(t, expectedErr)
		items.AssertDidRun(t, true)
		items[:1].AssertDidClose(t, false)
		items[1:].AssertDidClose(t, true)
	})

	t.Run("one mid-close", func(t *testing.T) {
		var (
			group, items = setupGroup(groupSize)
			expectedErr  = "mock item 0 forced run error"
		)

		t.Logf("calling Run() on group of size %d\n", groupSize)
		group.Run()

		t.Log("beginning consumption of Results()")
		results := goGroupResultListSignal(group)

		results.Require(t, rtest.Pending)

		t.Log("calling Close(...) on group")
		ctx, cancel := context.WithCancel(context.Background())
		group.Close(ctx)

		results.Require(t, rtest.Pending)

		t.Log("calling Kill() on first item")
		items[0].Kill()

		results.Require(t, rtest.Pending)

		t.Log("canceling Close(...) context")
		cancel()

		results.Require(t, rtest.Complete)

		results.AssertErrors(t, expectedErr)
		items.AssertDidRun(t, true)
		items.AssertDidClose(t, true)
	})

	t.Run("one after close", func(t *testing.T) {
		var (
			group, items = setupGroup(groupSize - 1)
			forcedErr    = "forced run after close error"
			specialItem  = newMockItem("mock item special")
		)

		items = append(items, specialItem)
		group.Append(runner.New(
			func() error {
				specialItem.Run()
				return errors.New(forcedErr)
			},
			specialItem.Close,
		))

		t.Logf("calling Run() on group of size %d\n", groupSize)
		group.Run()

		t.Log("beginning consumption of Results()")
		results := goGroupResultListSignal(group)

		results.Require(t, rtest.Pending)

		t.Log("calling Close(...) on group and canceling context")
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		group.Close(ctx)

		results.Require(t, rtest.Complete)

		results.AssertErrors(t, forcedErr)
		items.AssertDidRun(t, true)
		items.AssertDidClose(t, true)
	})
}

func TestConcurrentGroupCloseError(t *testing.T) {
	// High level
	// - When item.Close() returns a non-nil error, group results MUST NOT wait
	//   for that same item's Run() call to complete before closing
	// - item.Close() errors MUST appear in the group results channel

	requireGroupSize(t, 1)

	var (
		group, items = setupGroup(groupSize)
		expectedErr  = "mock item 0 forced close error"
	)

	t.Log("setting first item to force a close error")
	items[0].ForceCloseError = true

	t.Logf("calling Run() on group of size %d\n", groupSize)
	group.Run()

	t.Log("beginning consumption of Results()")
	results := goGroupResultListSignal(group)

	results.Require(t, rtest.Pending)

	t.Log("calling Close(...) on group and canceling context")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	group.Close(ctx)

	results.Require(t, rtest.Complete)

	t.Log("calling Kill() on first item for cleanup")
	items[0].Kill()

	results.AssertErrors(t, expectedErr)
	items.AssertDidRun(t, true)
	items.AssertDidClose(t, true)
}

func TestConcurrentGroupLateResultsConsumer(t *testing.T) {
	// High level
	// - Ensure no issues arise when items complete their lifecycles before
	//   any reads from the Results() channel occur

	requireGroupSize(t, 2)

	var (
		group, items = setupGroup(groupSize)
		ctx, cancel  = context.WithCancel(context.Background())
		expectedErrs = []string{
			"mock item 0 forced run error",
			"mock item 1 forced close error",
		}
	)

	t.Log("setting second item to force a close error")
	items[1].ForceCloseError = true

	t.Logf("calling Run() on group of size %d\n", groupSize)
	group.Run()

	t.Log("calling Kill() on first item")
	items[0].Kill()

	t.Logf("calling Close(...) and canceling context")
	group.Close(ctx)
	cancel()

	t.Log("beginning consumption of Results()")
	results := goGroupResultListSignal(group)

	results.Require(t, rtest.Complete)

	t.Log("calling Kill() on second item for cleanup")
	items[1].Kill()

	results.AssertErrors(t, expectedErrs...)
	items.AssertDidRun(t, true)
	items[:1].AssertDidClose(t, false)
	items[1:].AssertDidClose(t, true)
}

func TestConcurrentGroupTimelyResults(t *testing.T) {
	// High level
	// - Ensure errors from (and closing of) results channel manifest as they
	//   occur

	requireGroupSize(t, 1)

	group, items := setupGroup(groupSize)

	t.Logf("calling Run() on group of size %d\n", groupSize)
	group.Run()

	t.Log("calling Kill() and reading result for each item")
	for i, item := range items {
		var (
			expectedErr = fmt.Sprintf("mock item %d forced run error", i)
			result      = goGroupResultSignal(group)
		)

		item.Kill()
		result.Require(t, rtest.Complete)
		result.RequireClosed(t, false)
		result.AssertError(t, expectedErr)
	}

	t.Log("reading once more from results")
	result := goGroupResultSignal(group)

	result.Require(t, rtest.Complete)
	result.AssertError(t, nil)
	result.RequireClosed(t, true)
}
