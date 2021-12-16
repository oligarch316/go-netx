package serverx

import (
	"testing"

	"github.com/oligarch316/go-netx"
	"github.com/stretchr/testify/assert"
)

type testID string

func (ti testID) String() string { return string(ti) }

var (
	idA = testID("A")
	idB = testID("B")
	idC = testID("C")
	idD = testID("D")
)

type depMap map[netx.ServiceID][]netx.ServiceID

func TestServerDependencyCycles(t *testing.T) {
	subtests := []struct {
		name     string
		deps     depMap
		expected error
	}{
		{
			name:     "no dependencies",
			deps:     nil,
			expected: nil,
		},
		{
			name: "no cycles",
			deps: depMap{
				idA: {idB, idC},
				idB: {idD},
				idC: {idD},
			},
			expected: nil,
		},
		{
			name: "single element cycle",
			deps: depMap{
				idA: {idA},
			},
			expected: cycleError{idA, idA},
		},
		{
			name: "multi element cycle",
			deps: depMap{
				idA: {idB},
				idB: {idC},
				idC: {idD},
				idD: {idA},
			},
			expected: cycleError{idA, idD, idC, idB, idA},
		},
		{
			name: "inner cycle",
			deps: depMap{
				idA: {idB},
				idB: {idC},
				idC: {idB},
				idD: {idC},
			},
			expected: cycleError{idB, idC, idB},
		},
	}

	for _, item := range subtests {
		subtest := item

		t.Run(subtest.name, func(t *testing.T) {
			t.Parallel()

			params := newServiceParams()
			for k, v := range subtest.deps {
				params.appendDependencies(k, v...)
			}

			assert.Equal(t, subtest.expected, findDependencyCycles(params))
		})
	}
}
