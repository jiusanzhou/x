package x

// Contains reports whether target is present in items.
func Contains[V comparable](items []V, target V) bool {
	return ContainsFunc(items, func(v V) bool {
		return target == v
	})
}

// ContainsFunc reports whether any element in items satisfies the predicate fn.
func ContainsFunc[V any](items []V, fn func(V) bool) bool {
	for _, v := range items {
		if fn(v) {
			return true
		}
	}
	return false
}

// Filter returns a new slice containing only the elements of items for which fn returns true.
func Filter[T any](items []T, fn func(T) bool) []T {
	res := []T{}
	for _, v := range items {
		if fn(v) {
			res = append(res, v)
		}
	}
	return res
}

// Map returns a new slice containing the results of applying fn to each element of items.
func Map[T any, R any](items []T, fn func(T) R) []R {
	res := []R{}
	for _, v := range items {
		res = append(res, fn(v))
	}
	return res
}
