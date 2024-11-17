package x

// Contains retutrn if we have target value
func Contains[V comparable](items []V, target V) bool {
	return ContainsFunc(items, func(v V) bool {
		return target == v
	})
}

// ContainsFunc return if any item called the fn return true
func ContainsFunc[V any](items []V, fn func(V) bool) bool {
	for _, v := range items {
		if fn(v) {
			return true
		}
	}
	return false
}

// Filter use to filter items of array, return value is true
func Filter[T any](items []T, fn func(T) bool) []T {
	res := []T{}
	for _, v := range items {
		if fn(v) {
			res = append(res, v)
		}
	}
	return res
}

// Map call func of each items
func Map[T any, R any](items []T, fn func(T) R) []R {
	res := []R{}
	for _, v := range items {
		res = append(res, fn(v))
	}
	return res
}
