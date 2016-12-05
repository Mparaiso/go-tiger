package sort_test

import (
	"github.com/Mparaiso/go-tiger/internal/sort"
	"github.com/Mparaiso/go-tiger/test"

	"testing"
)

func TestSort(t *testing.T) {

	for _, T := range []struct {
		data      []int
		result    []int
		predicate func(i, j int) bool
	}{
		{data: []int{2, 3, 0, 1}, predicate: func(a, b int) bool { return a < b }, result: []int{0, 1, 2, 3}},
		{data: []int{3, 7, 16, 2, 0, 1, 1, 4, 2}, predicate: func(a, b int) bool { return a < b }, result: []int{0, 1, 1, 2, 2, 3, 4, 7, 16}},
	} {
		test.Error(t, sort.InsertSort(T.data, T.predicate), T.result)
		test.Error(t, sort.InsertSort2(T.data, T.predicate), T.result)
	}

}
