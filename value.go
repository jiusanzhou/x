//go:build go1.18

package x

type Value[T comparable] struct {
	val T

	cond bool
	zero T
}

func (v *Value[T]) Value() T {
	return v.val
}

// Or V(a).Or(-1)
func (v *Value[T]) Or(r T) *Value[T] {
	// check if is else
	// check if val is nil or zero value
	if !v.cond || v.zero == v.val {
		v.val = r
	}
	return v
}

// If V(a).If(true).Or(-1)
func (v *Value[T]) If(b bool) *Value[T] {
	v.cond = b
	return v
}

// Ifn V(a).Ifn(func() bool { return true }).Or(-1)
func (v *Value[T]) Ifn(fn func() bool) *Value[T] {
	v.cond = fn()
	return v
}

// Unwrap the value from (value, error)
// if err != nil, return v
func (v *Value[T]) Unwrap(mv T, err error) *Value[T] {
	if err == nil {
		v.val = mv
	}
	return v
}

// V create the value for expression
func V[T comparable](v T) *Value[T] {
	return &Value[T]{
		val:  v,
		cond: true,
	}
}

// Unwrap the value from (value, error)
// if err != nil, return v
func Unwrap[T comparable](mv T, err error) *Value[T] {
	return V(mv).Unwrap(mv, err)
}
