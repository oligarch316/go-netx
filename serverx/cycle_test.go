package serverx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type depMap map[ServiceID][]ServiceID

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

			params := make(serviceParams)
			for k, v := range subtest.deps {
				params.AddDependencies(k, v...)
			}

			assert.Equal(t, subtest.expected, findDependencyCycles(params))
		})
	}
}
