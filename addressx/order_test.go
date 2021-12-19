package addressx_test

import (
	"math/rand"
	"net"
	"testing"

	"github.com/oligarch316/go-netx/addressx"
	"github.com/stretchr/testify/require"
)

func TestAddrSort(t *testing.T) {
	var (
		sorted = []net.Addr{
			testAddr{"nP1", "aP1"},
			testAddr{"nP1", "aP2"},
			testAddr{"nP1", "a1"},
			testAddr{"nP1", "a2"},

			testAddr{"nP2", "aP1"},
			testAddr{"nP2", "a5"},

			testAddr{"nB", "aP1"},
			testAddr{"nB", "aP2"},

			testAddr{"nP3", "a5"},

			testAddr{"nA", "a3"},
			testAddr{"nA", "a4"},

			testAddr{"nB", "a1"},
		}

		cmpList = addressx.Ordering{
			addressx.ByPriorityNetwork("nP1", "nP2"),
			addressx.ByPriorityAddress("aP1", "aP2"),
			addressx.ByPriorityNetwork("nP3"),
			addressx.ByLexNetwork,
			addressx.ByLexAddress,
		}

		iterations = 5
		nAddrs     = len(sorted)
	)

	for i := 0; i < iterations; i++ {
		// copy
		actual := make([]net.Addr, nAddrs)
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
