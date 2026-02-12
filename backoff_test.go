package x

import (
	"testing"
	"time"
)

func TestBackoff_Basic(t *testing.T) {
	b := NewBackOffWithJitter(100*time.Millisecond, 1*time.Second, 0)

	if delay := b.Get("test"); delay != 0 {
		t.Errorf("Get() on new key = %v, want 0", delay)
	}

	b.Next("test", time.Now())
	delay := b.Get("test")
	if delay != 100*time.Millisecond {
		t.Errorf("Get() after first Next() = %v, want %v", delay, 100*time.Millisecond)
	}

	b.Next("test", time.Now())
	delay = b.Get("test")
	if delay != 200*time.Millisecond {
		t.Errorf("Get() after second Next() = %v, want %v", delay, 200*time.Millisecond)
	}
}

func TestBackoff_MaxDuration(t *testing.T) {
	b := NewBackOffWithJitter(100*time.Millisecond, 300*time.Millisecond, 0)

	for i := 0; i < 10; i++ {
		b.Next("test", time.Now())
	}

	delay := b.Get("test")
	if delay > 300*time.Millisecond {
		t.Errorf("Get() exceeded max duration: %v > %v", delay, 300*time.Millisecond)
	}
}

func TestBackoff_Reset(t *testing.T) {
	b := NewBackOffWithJitter(100*time.Millisecond, 1*time.Second, 0)

	b.Next("test", time.Now())
	b.Next("test", time.Now())
	b.Reset("test")

	if delay := b.Get("test"); delay != 0 {
		t.Errorf("Get() after Reset() = %v, want 0", delay)
	}
}

func TestBackoff_DeleteEntry(t *testing.T) {
	b := NewBackOffWithJitter(100*time.Millisecond, 1*time.Second, 0)

	b.Next("test", time.Now())
	b.DeleteEntry("test")

	if delay := b.Get("test"); delay != 0 {
		t.Errorf("Get() after DeleteEntry() = %v, want 0", delay)
	}
}

func TestBackoff_IsInBackOffSince(t *testing.T) {
	b := NewBackOffWithJitter(100*time.Millisecond, 1*time.Second, 0)

	if b.IsInBackOffSince("test", time.Now()) {
		t.Error("IsInBackOffSince() should be false for unknown key")
	}

	b.Next("test", time.Now())
	eventTime := time.Now()

	if !b.IsInBackOffSince("test", eventTime) {
		t.Error("IsInBackOffSince() should be true immediately after Next()")
	}
}

func TestBackoff_IsInBackOffSinceUpdate(t *testing.T) {
	b := NewBackOffWithJitter(100*time.Millisecond, 1*time.Second, 0)

	if b.IsInBackOffSinceUpdate("test", time.Now()) {
		t.Error("IsInBackOffSinceUpdate() should be false for unknown key")
	}

	b.Next("test", time.Now())

	if !b.IsInBackOffSinceUpdate("test", time.Now()) {
		t.Error("IsInBackOffSinceUpdate() should be true immediately after Next()")
	}
}

func TestBackoff_GC(t *testing.T) {
	b := NewBackOffWithJitter(1*time.Millisecond, 1*time.Millisecond, 0)

	b.Next("test1", time.Now())
	b.Next("test2", time.Now())

	time.Sleep(5 * time.Millisecond)
	b.GC()

	if delay := b.Get("test1"); delay != 0 {
		t.Errorf("Get() after GC = %v, want 0", delay)
	}
}

func TestBackoff_WithJitter(t *testing.T) {
	b := NewBackOffWithJitter(100*time.Millisecond, 1*time.Second, 0.5)

	b.Next("test", time.Now())
	delay1 := b.Get("test")

	if delay1 < 100*time.Millisecond || delay1 > 150*time.Millisecond {
		t.Errorf("Get() with jitter = %v, expected between 100ms and 150ms", delay1)
	}
}

func TestBackoff_MultipleKeys(t *testing.T) {
	b := NewBackOffWithJitter(100*time.Millisecond, 1*time.Second, 0)

	b.Next("key1", time.Now())
	b.Next("key2", time.Now())
	b.Next("key1", time.Now())

	if delay := b.Get("key1"); delay != 200*time.Millisecond {
		t.Errorf("Get(key1) = %v, want %v", delay, 200*time.Millisecond)
	}
	if delay := b.Get("key2"); delay != 100*time.Millisecond {
		t.Errorf("Get(key2) = %v, want %v", delay, 100*time.Millisecond)
	}
}
