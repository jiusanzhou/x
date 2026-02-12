package x

import (
	"sync"
	"testing"
)

func TestSyncMap_StoreAndLoad(t *testing.T) {
	var m SyncMap[string, int]

	m.Store("key1", 42)
	m.Store("key2", 100)

	v, ok := m.Load("key1")
	if !ok || v != 42 {
		t.Errorf("Load(key1) = %v, %v; want 42, true", v, ok)
	}

	v, ok = m.Load("key2")
	if !ok || v != 100 {
		t.Errorf("Load(key2) = %v, %v; want 100, true", v, ok)
	}

	_, ok = m.Load("nonexistent")
	if ok {
		t.Error("Load(nonexistent) should return false")
	}
}

func TestSyncMap_LoadOnEmpty(t *testing.T) {
	var m SyncMap[string, int]

	_, ok := m.Load("key")
	if ok {
		t.Error("Load on empty map should return false")
	}
}

func TestSyncMap_Delete(t *testing.T) {
	var m SyncMap[string, int]

	m.Store("key", 42)
	m.Delete("key")

	_, ok := m.Load("key")
	if ok {
		t.Error("Load after Delete should return false")
	}
}

func TestSyncMap_DeleteOnEmpty(t *testing.T) {
	var m SyncMap[string, int]
	m.Delete("key")
}

func TestSyncMap_LoadOrStore(t *testing.T) {
	var m SyncMap[string, int]

	v, loaded := m.LoadOrStore("key", 42)
	if loaded || v != 42 {
		t.Errorf("LoadOrStore on new key = %v, %v; want 42, false", v, loaded)
	}

	v, loaded = m.LoadOrStore("key", 100)
	if !loaded || v != 42 {
		t.Errorf("LoadOrStore on existing key = %v, %v; want 42, true", v, loaded)
	}
}

func TestSyncMap_LoadAndDelete(t *testing.T) {
	var m SyncMap[string, int]

	m.Store("key", 42)

	v, loaded := m.LoadAndDelete("key")
	if !loaded || v != 42 {
		t.Errorf("LoadAndDelete = %v, %v; want 42, true", v, loaded)
	}

	_, loaded = m.Load("key")
	if loaded {
		t.Error("Load after LoadAndDelete should return false")
	}
}

func TestSyncMap_LoadAndDeleteOnEmpty(t *testing.T) {
	var m SyncMap[string, int]

	_, loaded := m.LoadAndDelete("key")
	if loaded {
		t.Error("LoadAndDelete on empty map should return false")
	}
}

func TestSyncMap_Range(t *testing.T) {
	var m SyncMap[string, int]

	m.Store("a", 1)
	m.Store("b", 2)
	m.Store("c", 3)

	count := 0
	m.Range(func(k string, v int) bool {
		count++
		return true
	})

	if count != 3 {
		t.Errorf("Range iterated %d times, want 3", count)
	}
}

func TestSyncMap_RangeEarlyTermination(t *testing.T) {
	var m SyncMap[string, int]

	m.Store("a", 1)
	m.Store("b", 2)
	m.Store("c", 3)

	count := 0
	m.Range(func(k string, v int) bool {
		count++
		return false
	})

	if count != 1 {
		t.Errorf("Range with early termination iterated %d times, want 1", count)
	}
}

func TestSyncMap_RangeOnEmpty(t *testing.T) {
	var m SyncMap[string, int]

	count := 0
	m.Range(func(k string, v int) bool {
		count++
		return true
	})

	if count != 0 {
		t.Errorf("Range on empty map iterated %d times, want 0", count)
	}
}

func TestSyncMap_Len(t *testing.T) {
	var m SyncMap[string, int]

	if m.Len() != 0 {
		t.Errorf("Len on empty map = %d, want 0", m.Len())
	}

	m.Store("a", 1)
	m.Store("b", 2)

	if m.Len() != 2 {
		t.Errorf("Len = %d, want 2", m.Len())
	}
}

func TestSyncMap_Keys(t *testing.T) {
	var m SyncMap[string, int]

	m.Store("a", 1)
	m.Store("b", 2)

	keys := m.Keys()
	if len(keys) != 2 {
		t.Errorf("Keys len = %d, want 2", len(keys))
	}
}

func TestSyncMap_Values(t *testing.T) {
	var m SyncMap[string, int]

	m.Store("a", 1)
	m.Store("b", 2)

	values := m.Values()
	if len(values) != 2 {
		t.Errorf("Values len = %d, want 2", len(values))
	}
}

func TestSyncMap_Clone(t *testing.T) {
	var m SyncMap[string, int]

	m.Store("a", 1)
	m.Store("b", 2)

	clone := m.Clone()

	v, ok := clone.Load("a")
	if !ok || v != 1 {
		t.Errorf("Clone.Load(a) = %v, %v; want 1, true", v, ok)
	}

	m.Store("a", 100)
	v, _ = clone.Load("a")
	if v != 1 {
		t.Errorf("Clone should be independent, got %v, want 1", v)
	}
}

func TestSyncMap_Grow(t *testing.T) {
	var m SyncMap[string, int]

	m.Grow(100)
	m.Store("key", 42)

	v, ok := m.Load("key")
	if !ok || v != 42 {
		t.Errorf("Load after Grow = %v, %v; want 42, true", v, ok)
	}
}

func TestSyncMap_Concurrent(t *testing.T) {
	var m SyncMap[int, int]
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			m.Store(i, i*2)
			m.Load(i)
		}(i)
	}

	wg.Wait()

	if m.Len() != 100 {
		t.Errorf("Len after concurrent stores = %d, want 100", m.Len())
	}
}
