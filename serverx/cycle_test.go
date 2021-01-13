package serverx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCycleNormalize(t *testing.T) {
	subtests := []struct {
		name     string
		cycles   []Cycle
		expected Cycle
	}{
		{
			name:     "nil",
			cycles:   []Cycle{nil},
			expected: nil,
		},
		{
			name:     "empty",
			cycles:   []Cycle{{}},
			expected: nil,
		},
		{
			name:     "single element",
			cycles:   []Cycle{{idA}},
			expected: Cycle{idA},
		},
		{
			name: "non-trivial",
			cycles: []Cycle{
				{idA, idB, idC, idD},
				{idD, idA, idB, idC},
				{idC, idD, idA, idB},
				{idB, idC, idD, idA},
			},
			expected: Cycle{idA, idB, idC, idD},
		},
	}

	for _, item := range subtests {
		subtest := item

		t.Run(item.name, func(t *testing.T) {
			t.Parallel()
			for _, c := range subtest.cycles {
				assert.Equal(t, subtest.expected, c.normalize())
			}
		})
	}
}

func TestCycleListNormalize(t *testing.T) {
	subtests := []struct {
		name     string
		list     CycleList
		expected CycleList
	}{
		{
			name:     "nil",
			list:     nil,
			expected: nil,
		},
		{
			name:     "empty",
			list:     CycleList{},
			expected: nil,
		},
		{
			name:     "single un-normalized cycle",
			list:     CycleList{{idC, idD, idA, idB}},
			expected: CycleList{{idA, idB, idC, idD}},
		},
		{
			name: "non-trivial",
			list: CycleList{
				{idA, idB, idD, idC},
				{idA, idB, idC, idD},
				{idA, idB},
				{idC, idD},
			},
			expected: CycleList{
				{idA, idB},
				{idC, idD},
				{idA, idB, idC, idD},
				{idA, idB, idD, idC},
			},
		},
	}

	for _, item := range subtests {
		subtest := item

		t.Run(item.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, subtest.expected, subtest.list.normalize())
		})
	}
}

func TestCyclesFindDependencies(t *testing.T) {
	type depMap map[ServiceID][]ServiceID

	subtests := []struct {
		name     string
		deps     depMap
		expected CycleList
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
			name: "simple cycles",
			deps: depMap{
				idA: {idA},
				idB: {idB},
				idC: {idC},
				idD: {idD},
			},
			expected: CycleList{
				{idA},
				{idB},
				{idC},
				{idD},
			},
		},
		{
			name: "disconnected cycles",
			deps: depMap{
				idA: {idB},
				idB: {idA},
				idC: {idD},
				idD: {idC},
			},
			expected: CycleList{
				{idA, idB},
				{idC, idD},
			},
		},
		{
			name: "connected cycles",
			deps: depMap{
				idA: {idB},
				idB: {idA, idC},
				idC: {idB},
			},
			expected: CycleList{
				{idA, idB},
				{idB, idC},
			},
		},
		{
			name: "complex cycles",
			deps: depMap{
				idA: {idB, idC},
				idB: {idA, idC},
				idC: {idA, idD},
				idD: {idA, idB},
			},
			expected: CycleList{
				{idA, idB},
				{idA, idC},
				{idA, idB, idC},
				{idA, idC, idD},
				{idB, idC, idD},
				{idA, idB, idC, idD},
			},
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
