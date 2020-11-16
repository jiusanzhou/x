package x

import "reflect"

// Value ...
type Value struct {
	val interface{}

	cond bool
}

// NewValue create the value for expression
func NewValue(v interface{}) *Value {
	return &Value{
		val:  v,
		cond: true,
	}
}

// V for simple to create Value like
func V(v interface{}) *Value {
	return NewValue(v)
}

// auto generate the type gen

// String ...
func (v *Value) String() (string, bool) {
	vv, ok := v.val.(string)
	return vv, ok
}

// Bool ...
func (v *Value) Bool() (bool, bool) {
	vv, ok := v.val.(bool)
	return vv, ok
}

// Int ...
func (v *Value) Int() (int, bool) {
	vv, ok := v.val.(int)
	return vv, ok
}

// Int32 ...
func (v *Value) Int32() (int32, bool) {
	vv, ok := v.val.(int32)
	return vv, ok
}

// Int64 ...
func (v *Value) Int64() (int64, bool) {
	vv, ok := v.val.(int64)
	return vv, ok
}

// Float32 ...
func (v *Value) Float32() (float32, bool) {
	vv, ok := v.val.(float32)
	return vv, ok
}

// Float64 ...
func (v *Value) Float64() (float64, bool) {
	vv, ok := v.val.(float64)
	return vv, ok
}

// Interface ...
func (v *Value) Interface() interface{} {
	return v.val
}

// ============= the scope =================

// Or Value(a).Or(-1)
func (v *Value) Or(r interface{}) *Value {
	// check if is else
	// check if val is nil or zero value
	if !v.cond || v.val == nil || reflect.ValueOf(v.val).IsZero() {
		v.val = r
	}
	return v
}

// If Value(a).If(a == 1).Or(0)
// can accept a func() bool
func (v *Value) If(r bool) *Value {
	v.cond = r
	return v
}

// Ifn can accept a function: func() bool
func (v *Value) Ifn(fn func() bool) *Value {
	v.cond = fn()
	return v
}

// Unwrap the value from (value, error)
// if err != nil, return v
func (v *Value) Unwrap(mv interface{}, err error) *Value {
	if err == nil {
		v.val = mv
	}
	return v
}
