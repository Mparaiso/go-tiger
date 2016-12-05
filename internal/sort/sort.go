package sort

// InsertSort sorts an array of integers using insertion sort
// @see https://en.wikipedia.org/wiki/Insertion_sort
func InsertSort(a []int, predicate func(a, b int) bool) []int {
	b := make([]int, len(a), len(a))
	copy(b, a)
	if len(b) <= 1 {
		return b
	}
	for i := 1; i < len(a); i++ {
		lowest := i
		for j := i - 1; j >= 0; j-- {
			if predicate(b[i], b[j]) {
				lowest = j
			}
		}
		if i > lowest {
			b = append(b[:i], b[i+1:]...)
			b = append(b[:lowest], append([]int{a[i]}, b[lowest:]...)...)
		}
	}
	return b
}

// InsertSort2 sorts an array of integers using insertion sort
// @see https://en.wikipedia.org/wiki/Insertion_sort
func InsertSort2(array []int, predicate func(a, b int) bool) []int {
	b := make([]int, len(array))
	copy(b, array)
	for i := 1; i < len(b); i++ {
		j := i
		for j > 0 && predicate(b[j], b[j-1]) {
			b[j], b[j-1] = b[j-1], b[j]
			j = j - 1

		}
	}
	return b
}
