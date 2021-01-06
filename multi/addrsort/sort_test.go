package addrsort_test

import (
	"math/rand"
	"testing"

	"github.com/oligarch316/go-netx/multi"
	"github.com/oligarch316/go-netx/multi/addrsort"
	"github.com/stretchr/testify/require"
)

func TestAddrSort(t *testing.T) {
	var (
		sorted = []multi.Addr{
			addr("nP1", "aP1"),
			addr("nP1", "aP2"),
			addr("nP1", "a1"),
			addr("nP1", "a2"),

			addr("nP2", "aP1"),
			addr("nP2", "a5"),

			addr("nB", "aP1"),
			addr("nB", "aP2"),

			addr("nP3", "a5"),

			addr("nA", "a3"),
			addr("nA", "a4"),

			addr("nB", "a1"),
		}

		cmpList = addrsort.CompareList{
			addrsort.ByPriorityNetwork("nP1", "nP2"),
			addrsort.ByPriorityAddress("aP1", "aP2"),
			addrsort.ByPriorityNetwork("nP3"),
			addrsort.ByLexNetwork,
			addrsort.ByLexAddress,
		}

		iterations = 5
		nAddrs     = len(sorted)
	)

	for i := 0; i < iterations; i++ {
		// copy
		actual := make([]multi.Addr, nAddrs)
		copy(actual, sorted)

		// shuffle
		rand.Shuffle(nAddrs, func(i, j int) {
			actual[i], actual[j] = actual[j], actual[i]
		})

		// sort
		cmpList.Stable(actual)

		// check
		require.Equal(t, sorted, actual)
	}
}
