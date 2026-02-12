package x

import (
	"errors"
	"testing"
)

func TestErrors_Error(t *testing.T) {
	tests := []struct {
		name   string
		errors Errors
		want   string
	}{
		{"empty", Errors{}, ""},
		{"single error", Errors{errors.New("error1")}, "error1"},
		{"multiple errors", Errors{errors.New("error1"), errors.New("error2")}, "error1; error2"},
		{"three errors", Errors{errors.New("a"), errors.New("b"), errors.New("c")}, "a; b; c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.errors.Error(); got != tt.want {
				t.Errorf("Errors.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestErrors_IsNil(t *testing.T) {
	tests := []struct {
		name   string
		errors Errors
		want   bool
	}{
		{"empty", Errors{}, true},
		{"with error", Errors{errors.New("error")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.errors.IsNil(); got != tt.want {
				t.Errorf("Errors.IsNil() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrors_Add(t *testing.T) {
	var errs Errors

	errs.Add(nil)
	if len(errs) != 0 {
		t.Error("Add(nil) should not add to errors")
	}

	errs.Add(errors.New("error1"))
	if len(errs) != 1 {
		t.Errorf("Add() len = %d, want 1", len(errs))
	}

	errs.Add(errors.New("error2"), nil, errors.New("error3"))
	if len(errs) != 3 {
		t.Errorf("Add() with mixed nil len = %d, want 3", len(errs))
	}
}

func TestNewErrors(t *testing.T) {
	tests := []struct {
		name  string
		input []error
		wantN int
	}{
		{"all nil", []error{nil, nil}, 0},
		{"mixed", []error{errors.New("a"), nil, errors.New("b")}, 2},
		{"all valid", []error{errors.New("a"), errors.New("b")}, 2},
		{"empty", []error{}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewErrors(tt.input...)
			if len(got) != tt.wantN {
				t.Errorf("NewErrors() len = %d, want %d", len(got), tt.wantN)
			}
		})
	}
}

func TestErrors_ImplementsError(t *testing.T) {
	var _ error = Errors{}
}
