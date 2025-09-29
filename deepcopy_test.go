package x

import (
	"reflect"
	"testing"
)

func TestDeepCopy(t *testing.T) {
	tests := []struct {
		name     string
		source   any
		expected any
	}{
		{
			name: "nil pointer",
			source: func() *int {
				return nil
			}(),
			expected: nil,
		},
		{
			name:     "int",
			source:   int(1),
			expected: int(1),
		},
		{
			name:     "int pointer",
			source:   func() *int { a := 2; return &a }(),
			expected: func() *int { a := 2; return &a }(),
		},
		{
			name: "struct",
			source: struct {
				A int
				B string
			}{A: 3, B: "b"},
			expected: struct {
				A int
				B string
			}{A: 3, B: "b"},
		},
		{
			name: "nested struct",
			source: struct {
				A int
				B struct {
					C int
					D string
				}
			}{A: 1, B: struct {
				C int
				D string
			}{C: 2, D: "d"}},
			expected: struct {
				A int
				B struct {
					C int
					D string
				}
			}{A: 1, B: struct {
				C int
				D string
			}{C: 2, D: "d"}},
		},
		{
			name: "struct with pointer",
			source: struct {
				A int
				B *string
			}{A: 4, B: func() *string { a := "b"; return &a }()},
			expected: struct {
				A int
				B *string
			}{A: 4, B: func() *string { a := "b"; return &a }()},
		},
		{
			name:     "array",
			source:   [2]int{1, 2},
			expected: [2]int{1, 2},
		},
		{
			name:     "array with pointer",
			source:   [2]*int{func() *int { a := 1; return &a }(), func() *int { a := 2; return &a }()},
			expected: [2]*int{func() *int { a := 1; return &a }(), func() *int { a := 2; return &a }()},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := DeepCopy(test.source)
			if test.expected == nil && got != nil {
				return
			}
			if !reflect.DeepEqual(got, test.expected) {
				t.Errorf("DeepCopy of %v, got %v, want %v", test.source, got, test.expected)
			}
		})
	}
}
