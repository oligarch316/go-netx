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

type testGroupItem struct {
	name                          string
	killChan, closeChan, doneChan chan struct{}

	DidRun, DidClose *rtest.Flag
	ForceCloseError  bool
}

func newGroupItem(name string) *testGroupItem {
	return &testGroupItem{
		name:      name,
		killChan:  make(chan struct{}),
		closeChan: make(chan struct{}),
		doneChan:  make(chan struct{}),

		DidRun:   rtest.NewFlag(name + " did run"),
		DidClose: rtest.NewFlag(name + " did close"),
	}
}

func (tgi *testGroupItem) Kill() {
	close(tgi.killChan)
	<-tgi.doneChan
}

func (tgi *testGroupItem) Run() error {
	tgi.DidRun.Mark()
	defer close(tgi.doneChan)

	select {
	case <-tgi.killChan:
		return fmt.Errorf("%s forced run error", tgi)
	case <-tgi.closeChan:
		return nil
	}
}

func (tgi *testGroupItem) Close(ctx context.Context) error {
	tgi.DidClose.Mark()

	if tgi.ForceCloseError {
		return fmt.Errorf("%s forced close error", tgi)
	}

	defer close(tgi.closeChan)

	select {
	case <-tgi.doneChan:
	case <-ctx.Done():
	}

	return nil
}

func (tgi testGroupItem) String() string { return tgi.name }

type testGroupItemList []*testGroupItem

func (tgil testGroupItemList) Kill() {
	for _, item := range tgil {
		item.Kill()
	}
}

func (tgil testGroupItemList) AssertDidRun(t assert.TestingT, expected bool) bool {
	res := true
	for _, item := range tgil {
		res = item.DidRun.Assert(t, expected) && res
	}
	return res
}

func (tgil testGroupItemList) AssertDidClose(t assert.TestingT, expected bool) bool {
	res := true
	for _, item := range tgil {
		res = item.DidClose.Assert(t, expected) && res
	}
	return res
}

type testGroupResults struct {
	rtest.Signal
	Errs []error
}

func goGroupResults(group *runner.Group) *testGroupResults {
	res := new(testGroupResults)
	res.Signal = rtest.GoSignal("results", func() {
		for err := range group.Results() {
			res.Errs = append(res.Errs, err)
		}
	})
	return res
}

func (tgr testGroupResults) AssertErrors(t assert.TestingT, expected ...string) bool {
	if len(expected) == 0 {
		return assert.Empty(t, tgr.Errs, "group results")
	}

	actual := make([]string, len(tgr.Errs))
	for i, err := range tgr.Errs {
		if err == nil {
			actual[i] = "<nil>"
		} else {
			actual[i] = err.Error()
		}
	}

	return assert.ElementsMatch(t, expected, actual, "group results")
}

func setupGroup(size int) (*runner.Group, testGroupItemList) {
	var (
		group = runner.NewGroup()
		items = make(testGroupItemList, size)
	)

	for i := 0; i < size; i++ {
		items[i] = newGroupItem(fmt.Sprintf("test item %d", i))
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
	results := goGroupResults(group)

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
			expectedErrs[i] = fmt.Sprintf("test item %d forced run error", i)
		}

		t.Logf("calling Run() on group of size %d\n", groupSize)
		group.Run()

		t.Log("beginning consumption of Results()")
		results := goGroupResults(group)

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
			expectedErr  = "test item 0 forced run error"
		)

		t.Logf("calling Run() on group of size %d\n", groupSize)
		group.Run()

		t.Log("beginning consumption of Results()")
		results := goGroupResults(group)

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
			expectedErr  = "test item 0 forced run error"
		)

		t.Logf("calling Run() on group of size %d\n", groupSize)
		group.Run()

		t.Log("beginning consumption of Results()")
		results := goGroupResults(group)

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
			specialItem  = newGroupItem("test item special")
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
		results := goGroupResults(group)

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
		expectedErr  = "test item 0 forced close error"
	)

	t.Log("setting first item to force a close error")
	items[0].ForceCloseError = true

	t.Logf("calling Run() on group of size %d\n", groupSize)
	group.Run()

	t.Log("beginning consumption of Results()")
	results := goGroupResults(group)

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
			"test item 0 forced run error",
			"test item 1 forced close error",
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
	results := goGroupResults(group)

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

	var (
		group, items = setupGroup(groupSize)
		nextResult   = func() error { return <-group.Results() }
	)

	t.Logf("calling Run() on group of size %d\n", groupSize)
	group.Run()

	t.Log("calling Kill() and reading result for each item")
	for i, item := range items {
		var (
			expectedErr = fmt.Sprintf("test item %d forced run error", i)
			sig         = rtest.GoErrorSignal("read result", nextResult)
		)

		item.Kill()
		sig.Require(t, rtest.Complete)
		sig.RequireError(t, expectedErr)
	}

	var (
		finalErr    error
		finalMore   bool
		finalResult = func() { finalErr, finalMore = <-group.Results() }
	)

	t.Log("reading once more from results")
	sig := rtest.GoSignal("read final result", finalResult)

	sig.Require(t, rtest.Complete)
	assert.NoError(t, finalErr, "final error result")
	assert.False(t, finalMore, "final more result")
}
