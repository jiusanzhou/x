package x

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestIntervalSet(t *testing.T) {
	var count int32
	iv := IntervalSet(func(t time.Time) {
		atomic.AddInt32(&count, 1)
	}, 10*time.Millisecond)

	time.Sleep(55 * time.Millisecond)
	iv.Stop()

	c := atomic.LoadInt32(&count)
	if c < 4 || c > 6 {
		t.Errorf("expected ~5 calls, got %d", c)
	}
}

func TestIntervalSetWithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var count int32

	iv := IntervalSetWithContext(ctx, func(t time.Time) {
		atomic.AddInt32(&count, 1)
	}, 10*time.Millisecond)

	time.Sleep(35 * time.Millisecond)
	cancel()
	<-iv.Done()

	c := atomic.LoadInt32(&count)
	if c < 2 || c > 4 {
		t.Errorf("expected ~3 calls, got %d", c)
	}
}

func TestIntervalStop(t *testing.T) {
	var count int32
	iv := IntervalSet(func(t time.Time) {
		atomic.AddInt32(&count, 1)
	}, 10*time.Millisecond)

	iv.Stop()
	iv.Stop()

	time.Sleep(30 * time.Millisecond)
	c := atomic.LoadInt32(&count)
	if c > 1 {
		t.Errorf("expected 0-1 calls after stop, got %d", c)
	}
}

func TestIntervalFunc(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Millisecond)
	defer cancel()

	var count int32
	err := IntervalFunc(ctx, 10*time.Millisecond, func(ctx context.Context) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}

	c := atomic.LoadInt32(&count)
	if c < 4 || c > 6 {
		t.Errorf("expected ~5 calls, got %d", c)
	}
}

func TestIntervalFuncError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("test error")

	var count int32
	err := IntervalFunc(ctx, 10*time.Millisecond, func(ctx context.Context) error {
		c := atomic.AddInt32(&count, 1)
		if c >= 3 {
			return expectedErr
		}
		return nil
	})

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected test error, got %v", err)
	}

	c := atomic.LoadInt32(&count)
	if c != 3 {
		t.Errorf("expected 3 calls, got %d", c)
	}
}

func TestIntervalFuncImmediate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	var count int32
	err := IntervalFuncImmediate(ctx, 10*time.Millisecond, func(ctx context.Context) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}

	c := atomic.LoadInt32(&count)
	if c < 2 || c > 4 {
		t.Errorf("expected ~3 calls (immediate + interval), got %d", c)
	}
}

func TestIntervalDone(t *testing.T) {
	iv := IntervalSet(func(t time.Time) {}, 10*time.Millisecond)

	select {
	case <-iv.Done():
		t.Error("Done() should not be closed before Stop()")
	default:
	}

	iv.Stop()

	select {
	case <-iv.Done():
	case <-time.After(100 * time.Millisecond):
		t.Error("Done() should be closed after Stop()")
	}
}
