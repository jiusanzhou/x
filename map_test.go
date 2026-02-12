package x

import (
	"testing"
)

func TestKeys(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]int
		want int
	}{
		{"empty map", map[string]int{}, 0},
		{"single entry", map[string]int{"a": 1}, 1},
		{"multiple entries", map[string]int{"a": 1, "b": 2, "c": 3}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Keys(tt.m)
			if len(got) != tt.want {
				t.Errorf("Keys() len = %d, want %d", len(got), tt.want)
			}
			for _, k := range got {
				if _, ok := tt.m[k]; !ok {
					t.Errorf("Keys() returned key %q not in original map", k)
				}
			}
		})
	}
}

func TestValues(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]int
		want int
	}{
		{"empty map", map[string]int{}, 0},
		{"single entry", map[string]int{"a": 1}, 1},
		{"multiple entries", map[string]int{"a": 1, "b": 2, "c": 3}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Values(tt.m)
			if len(got) != tt.want {
				t.Errorf("Values() len = %d, want %d", len(got), tt.want)
			}
		})
	}
}

func TestRange(t *testing.T) {
	t.Run("iterate all", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		count := 0
		Range(m, func(k string, v int) bool {
			count++
			return true
		})
		if count != 3 {
			t.Errorf("Range() iterated %d times, want 3", count)
		}
	})

	t.Run("early termination", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		count := 0
		Range(m, func(k string, v int) bool {
			count++
			return false
		})
		if count != 1 {
			t.Errorf("Range() with early termination iterated %d times, want 1", count)
		}
	})

	t.Run("empty map", func(t *testing.T) {
		m := map[string]int{}
		count := 0
		Range(m, func(k string, v int) bool {
			count++
			return true
		})
		if count != 0 {
			t.Errorf("Range() on empty map iterated %d times, want 0", count)
		}
	})
}

func TestUpdateMap(t *testing.T) {
	type item struct {
		id   string
		name string
	}

	keyFn := func(i item) string { return i.id }
	convertFn := func(i item) string { return i.name }

	t.Run("add new items", func(t *testing.T) {
		original := map[string]string{}
		inputs := []item{{id: "1", name: "Alice"}, {id: "2", name: "Bob"}}

		changed, deleted := UpdateMap(original, inputs, convertFn, keyFn, false)

		if len(changed) != 2 {
			t.Errorf("changed len = %d, want 2", len(changed))
		}
		if len(deleted) != 0 {
			t.Errorf("deleted len = %d, want 0", len(deleted))
		}
	})

	t.Run("delete items", func(t *testing.T) {
		original := map[string]string{"1": "Alice", "2": "Bob", "3": "Charlie"}
		inputs := []item{{id: "1", name: "Alice"}}

		changed, deleted := UpdateMap(original, inputs, convertFn, keyFn, false)

		if len(changed) != 1 {
			t.Errorf("changed len = %d, want 1", len(changed))
		}
		if len(deleted) != 2 {
			t.Errorf("deleted len = %d, want 2", len(deleted))
		}
	})

	t.Run("update without force", func(t *testing.T) {
		original := map[string]string{"1": "OriginalAlice"}
		inputs := []item{{id: "1", name: "NewAlice"}}

		changed, _ := UpdateMap(original, inputs, convertFn, keyFn, false)

		if changed["1"] != "OriginalAlice" {
			t.Errorf("changed[1] = %q, want OriginalAlice (no force)", changed["1"])
		}
	})

	t.Run("update with force", func(t *testing.T) {
		original := map[string]string{"1": "OriginalAlice"}
		inputs := []item{{id: "1", name: "NewAlice"}}

		changed, _ := UpdateMap(original, inputs, convertFn, keyFn, true)

		if changed["1"] != "NewAlice" {
			t.Errorf("changed[1] = %q, want NewAlice (force update)", changed["1"])
		}
	})
}
