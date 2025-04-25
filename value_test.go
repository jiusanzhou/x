package x

import (
	"reflect"
	"testing"
)

func TestInt(t *testing.T) {
	tests := []struct {
		name string
		v    *Value[int]
		want interface{}
	}{
		{"Or: Int 0", V(1).Or(-1), 1},
		{"Or: Int 1", V(0).Or(-1), -1},
		{"IfOr: Int 0", V(1).If(true).Or(-1), 1},
		{"IfOr: Int 1", V(0).If(false).Or(-1), -1},
		{"IfnOr: Int 0", V(1).Ifn(func() bool { return true }).Or(-1), 1},

		// {"Or: Str empty", V("").Or("empty"), "empty"},
		// {"Or: Str value", V("ok").Or("empty"), "ok"},
		// {"Or: Object nil", V(interface{}(nil)).Or(1), 1},
		// {"IfOr: Str empty", V("ok").If(false).Or("empty"), "empty"},
		// {"IfOr: Str value", V("ok").If(true).Or("empty"), "ok"},
	}

	fn := func(v interface{}, err error) (interface{}, error) {
		return v, err
	}

	fn(fn(nil, nil))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.Value(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name string
		v    *Value[string]
		want interface{}
	}{

		// {"Or: Object nil", V(interface{}(nil)).Or(1), 1},
		{"Or: Str empty", V("").Or("empty"), "empty"},
		{"Or: Str value", V("ok").Or("empty"), "ok"},
		{"IfOr: Str empty", V("ok").If(false).Or("empty"), "empty"},
		{"IfOr: Str value", V("ok").If(true).Or("empty"), "ok"},
	}

	fn := func(v interface{}, err error) (interface{}, error) {
		return v, err
	}

	fn(fn(nil, nil))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.Value(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAny(t *testing.T) {
	tests := []struct {
		name string
		v    *Value[any]
		want interface{}
	}{

		{"Or: Object nil", V(interface{}(nil)).Or(1), 1},
	}

	fn := func(v interface{}, err error) (interface{}, error) {
		return v, err
	}

	fn(fn(nil, nil))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.Value(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
