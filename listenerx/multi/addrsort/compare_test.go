package addrsort_test

import (
	"fmt"
	"testing"

	"github.com/oligarch316/go-netx/listenerx/multi"
	"github.com/oligarch316/go-netx/listenerx/multi/addrsort"
	"github.com/stretchr/testify/assert"
)

type testAddr struct{ network, address string }

func (ta testAddr) Network() string { return ta.network }

func (ta testAddr) String() string { return ta.address }

func addr(network, address string) multi.Addr {
	return multi.Addr{Addr: testAddr{network: network, address: address}}
}

func TestAddrCompare(t *testing.T) {
	type (
		input       struct{ x, y multi.Addr }
		expectation func(*testing.T, bool, bool, string) bool
		testData    struct {
			name   string
			expect expectation
			inputs []input
		}
	)

	var (
		isLess = func(t *testing.T, l, e bool, info string) bool {
			res := assert.True(t, l, "%s: 'less' should be true", info)
			return assert.False(t, e, "%s: 'equal' should be false", info) && res
		}
		notLess = func(t *testing.T, l, e bool, info string) bool {
			res := assert.False(t, l, "%s: 'less' should be false", info)
			return assert.False(t, e, "%s: 'equal' should be false", info) && res
		}
		areEqual = func(t *testing.T, _, e bool, info string) bool {
			return assert.True(t, e, "%s: 'equal' should be true", info)
		}
	)

	run := func(t *testing.T, cpr addrsort.Comparer, expect expectation, inputs []input) {
		for _, input := range inputs {
			info := fmt.Sprintf(
				"%s|%s, %s|%s",
				input.x.Network(), input.x.String(),
				input.y.Network(), input.y.String(),
			)

			less, equal := cpr(input.x, input.y)
			expect(t, less, equal, info)
		}
	}

	t.Run("lexographic", func(t *testing.T) {
		t.Run("address", func(t *testing.T) {
			cpr := addrsort.ByLexAddress

			subtests := []testData{
				{
					name:   "is less",
					expect: isLess,
					inputs: []input{
						{x: addr("nB", "a1"), y: addr("nA", "a2")},
					},
				},
				{
					name:   "not less",
					expect: notLess,
					inputs: []input{
						{x: addr("nA", "a2"), y: addr("nB", "a1")},
					},
				},
				{
					name:   "are equal",
					expect: areEqual,
					inputs: []input{
						{x: addr("nA", "a1"), y: addr("nB", "a1")},
					},
				},
			}

			for _, item := range subtests {
				subtest := item
				t.Run(subtest.name, func(t *testing.T) { run(t, cpr, subtest.expect, subtest.inputs) })
			}
		})

		t.Run("network", func(t *testing.T) {
			cpr := addrsort.ByLexNetwork

			subtests := []testData{
				{
					name:   "is less",
					expect: isLess,
					inputs: []input{
						{x: addr("nA", "a2"), y: addr("nB", "a1")},
					},
				},
				{
					name:   "not less",
					expect: notLess,
					inputs: []input{
						{x: addr("nB", "a1"), y: addr("nA", "a2")},
					},
				},
				{
					name:   "are equal",
					expect: areEqual,
					inputs: []input{
						{x: addr("nA", "a1"), y: addr("nA", "a2")},
					},
				},
			}

			for _, item := range subtests {
				subtest := item
				t.Run(subtest.name, func(t *testing.T) { run(t, cpr, subtest.expect, subtest.inputs) })
			}
		})
	})

	t.Run("priority", func(t *testing.T) {
		t.Run("address", func(t *testing.T) {
			cpr := addrsort.ByPriorityAddress("a4", "a3", "a2")

			subtests := []testData{
				{
					name:   "is less",
					expect: isLess,
					inputs: []input{
						// Both have priority value
						{x: addr("nB", "a4"), y: addr("nA", "a3")},

						// One has priority value
						{x: addr("nB", "a2"), y: addr("nA", "a1")},
					},
				},
				{
					name:   "not less",
					expect: notLess,
					inputs: []input{
						// Both have priority value
						{x: addr("nA", "a3"), y: addr("nB", "a4")},

						// One has priority value
						{x: addr("nA", "a1"), y: addr("nB", "a2")},
					},
				},
				{
					name:   "are equal",
					expect: areEqual,
					inputs: []input{
						// Both have priority value
						{x: addr("nA", "a4"), y: addr("nB", "a4")},

						// Neither has priority value
						{x: addr("nA", "a10"), y: addr("nB", "a11")},
					},
				},
			}

			for _, item := range subtests {
				subtest := item
				t.Run(subtest.name, func(t *testing.T) { run(t, cpr, subtest.expect, subtest.inputs) })
			}
		})

		t.Run("network", func(t *testing.T) {
			cpr := addrsort.ByPriorityNetwork("nD", "nC", "nB")

			subtests := []testData{
				{
					name:   "is less",
					expect: isLess,
					inputs: []input{
						// Both have priority value
						{x: addr("nD", "a2"), y: addr("nC", "a1")},

						// One has priority value
						{x: addr("nB", "a2"), y: addr("nA", "a1")},
					},
				},
				{
					name:   "not less",
					expect: notLess,
					inputs: []input{
						// Both have priority value
						{x: addr("nC", "a1"), y: addr("nD", "a2")},

						// One has priority value
						{x: addr("nA", "a1"), y: addr("nB", "a2")},
					},
				},
				{
					name:   "are equal",
					expect: areEqual,
					inputs: []input{
						// Both have priority value
						{x: addr("nD", "a1"), y: addr("nD", "a2")},

						// Neither has priority value
						{x: addr("nX", "a1"), y: addr("nY", "a2")},
					},
				},
			}

			for _, item := range subtests {
				subtest := item
				t.Run(subtest.name, func(t *testing.T) { run(t, cpr, subtest.expect, subtest.inputs) })
			}
		})
	})
}
