package addressx_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/oligarch316/go-netx/addressx"
	"github.com/stretchr/testify/assert"
)

type testAddr struct{ network, address string }

func (ta testAddr) Network() string { return ta.network }

func (ta testAddr) String() string { return ta.address }

// func testAddr{network, address string) net.Addr {
// 	return net.Addr{Addr: testAddr{network: network, address: address}}
// }

func TestAddrCompare(t *testing.T) {
	type (
		input       struct{ x, y net.Addr }
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

	run := func(t *testing.T, cpr addressx.Comparer, expect expectation, inputs []input) {
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
			cpr := addressx.ByLexAddress

			subtests := []testData{
				{
					name:   "is less",
					expect: isLess,
					inputs: []input{
						{x: testAddr{"nB", "a1"}, y: testAddr{"nA", "a2"}},
					},
				},
				{
					name:   "not less",
					expect: notLess,
					inputs: []input{
						{x: testAddr{"nA", "a2"}, y: testAddr{"nB", "a1"}},
					},
				},
				{
					name:   "are equal",
					expect: areEqual,
					inputs: []input{
						{x: testAddr{"nA", "a1"}, y: testAddr{"nB", "a1"}},
					},
				},
			}

			for _, item := range subtests {
				subtest := item
				t.Run(subtest.name, func(t *testing.T) { run(t, cpr, subtest.expect, subtest.inputs) })
			}
		})

		t.Run("network", func(t *testing.T) {
			cpr := addressx.ByLexNetwork

			subtests := []testData{
				{
					name:   "is less",
					expect: isLess,
					inputs: []input{
						{x: testAddr{"nA", "a2"}, y: testAddr{"nB", "a1"}},
					},
				},
				{
					name:   "not less",
					expect: notLess,
					inputs: []input{
						{x: testAddr{"nB", "a1"}, y: testAddr{"nA", "a2"}},
					},
				},
				{
					name:   "are equal",
					expect: areEqual,
					inputs: []input{
						{x: testAddr{"nA", "a1"}, y: testAddr{"nA", "a2"}},
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
			cpr := addressx.ByPriorityAddress("a4", "a3", "a2")

			subtests := []testData{
				{
					name:   "is less",
					expect: isLess,
					inputs: []input{
						// Both have priority value
						{x: testAddr{"nB", "a4"}, y: testAddr{"nA", "a3"}},

						// One has priority value
						{x: testAddr{"nB", "a2"}, y: testAddr{"nA", "a1"}},
					},
				},
				{
					name:   "not less",
					expect: notLess,
					inputs: []input{
						// Both have priority value
						{x: testAddr{"nA", "a3"}, y: testAddr{"nB", "a4"}},

						// One has priority value
						{x: testAddr{"nA", "a1"}, y: testAddr{"nB", "a2"}},
					},
				},
				{
					name:   "are equal",
					expect: areEqual,
					inputs: []input{
						// Both have priority value
						{x: testAddr{"nA", "a4"}, y: testAddr{"nB", "a4"}},

						// Neither has priority value
						{x: testAddr{"nA", "a10"}, y: testAddr{"nB", "a11"}},
					},
				},
			}

			for _, item := range subtests {
				subtest := item
				t.Run(subtest.name, func(t *testing.T) { run(t, cpr, subtest.expect, subtest.inputs) })
			}
		})

		t.Run("network", func(t *testing.T) {
			cpr := addressx.ByPriorityNetwork("nD", "nC", "nB")

			subtests := []testData{
				{
					name:   "is less",
					expect: isLess,
					inputs: []input{
						// Both have priority value
						{x: testAddr{"nD", "a2"}, y: testAddr{"nC", "a1"}},

						// One has priority value
						{x: testAddr{"nB", "a2"}, y: testAddr{"nA", "a1"}},
					},
				},
				{
					name:   "not less",
					expect: notLess,
					inputs: []input{
						// Both have priority value
						{x: testAddr{"nC", "a1"}, y: testAddr{"nD", "a2"}},

						// One has priority value
						{x: testAddr{"nA", "a1"}, y: testAddr{"nB", "a2"}},
					},
				},
				{
					name:   "are equal",
					expect: areEqual,
					inputs: []input{
						// Both have priority value
						{x: testAddr{"nD", "a1"}, y: testAddr{"nD", "a2"}},

						// Neither has priority value
						{x: testAddr{"nX", "a1"}, y: testAddr{"nY", "a2"}},
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
