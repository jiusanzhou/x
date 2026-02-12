package x

import (
	"testing"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name   string
		items  []int
		target int
		want   bool
	}{
		{"found in middle", []int{1, 2, 3, 4, 5}, 3, true},
		{"found at start", []int{1, 2, 3}, 1, true},
		{"found at end", []int{1, 2, 3}, 3, true},
		{"not found", []int{1, 2, 3}, 4, false},
		{"empty slice", []int{}, 1, false},
		{"single element found", []int{42}, 42, true},
		{"single element not found", []int{42}, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Contains(tt.items, tt.target); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContains_Strings(t *testing.T) {
	items := []string{"apple", "banana", "cherry"}
	if !Contains(items, "banana") {
		t.Error("Contains() should find 'banana'")
	}
	if Contains(items, "grape") {
		t.Error("Contains() should not find 'grape'")
	}
}

func TestContainsFunc(t *testing.T) {
	type person struct {
		name string
		age  int
	}
	people := []person{
		{"Alice", 30},
		{"Bob", 25},
		{"Charlie", 35},
	}

	tests := []struct {
		name string
		fn   func(person) bool
		want bool
	}{
		{"find by name", func(p person) bool { return p.name == "Bob" }, true},
		{"find by age", func(p person) bool { return p.age > 30 }, true},
		{"not found by name", func(p person) bool { return p.name == "Dave" }, false},
		{"not found by age", func(p person) bool { return p.age > 40 }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsFunc(people, tt.fn); got != tt.want {
				t.Errorf("ContainsFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainsFunc_EmptySlice(t *testing.T) {
	var items []int
	if ContainsFunc(items, func(i int) bool { return true }) {
		t.Error("ContainsFunc() on empty slice should return false")
	}
}

func TestFilter(t *testing.T) {
	tests := []struct {
		name  string
		items []int
		fn    func(int) bool
		want  []int
	}{
		{"filter evens", []int{1, 2, 3, 4, 5, 6}, func(n int) bool { return n%2 == 0 }, []int{2, 4, 6}},
		{"filter odds", []int{1, 2, 3, 4, 5}, func(n int) bool { return n%2 != 0 }, []int{1, 3, 5}},
		{"filter none", []int{1, 2, 3}, func(n int) bool { return n > 10 }, []int{}},
		{"filter all", []int{1, 2, 3}, func(n int) bool { return n > 0 }, []int{1, 2, 3}},
		{"empty slice", []int{}, func(n int) bool { return true }, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Filter(tt.items, tt.fn)
			if len(got) != len(tt.want) {
				t.Errorf("Filter() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Filter()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestMap(t *testing.T) {
	t.Run("double integers", func(t *testing.T) {
		items := []int{1, 2, 3, 4, 5}
		got := Map(items, func(n int) int { return n * 2 })
		want := []int{2, 4, 6, 8, 10}
		if len(got) != len(want) {
			t.Errorf("Map() length = %v, want %v", len(got), len(want))
			return
		}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("Map()[%d] = %v, want %v", i, got[i], want[i])
			}
		}
	})

	t.Run("int to string", func(t *testing.T) {
		items := []int{1, 2, 3}
		got := Map(items, func(n int) string {
			return string(rune('A' + n - 1))
		})
		want := []string{"A", "B", "C"}
		if len(got) != len(want) {
			t.Errorf("Map() length = %v, want %v", len(got), len(want))
			return
		}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("Map()[%d] = %v, want %v", i, got[i], want[i])
			}
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		items := []int{}
		got := Map(items, func(n int) int { return n * 2 })
		if len(got) != 0 {
			t.Errorf("Map() on empty slice should return empty slice, got length %v", len(got))
		}
	})
}
