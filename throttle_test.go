package x

import (
	"context"
	"testing"
	"time"
)

func TestNewTokenBucketRateLimiter(t *testing.T) {
	limiter := NewTokenBucketRateLimiter(10, 5)
	if limiter == nil {
		t.Fatal("NewTokenBucketRateLimiter returned nil")
	}
	if limiter.QPS() != 10 {
		t.Errorf("QPS() = %v, want 10", limiter.QPS())
	}
}

func TestTokenBucketRateLimiter_TryAccept(t *testing.T) {
	limiter := NewTokenBucketRateLimiter(10, 5)

	accepted := 0
	for i := 0; i < 10; i++ {
		if limiter.TryAccept() {
			accepted++
		}
	}

	if accepted != 5 {
		t.Errorf("TryAccept() accepted %d, want 5 (burst size)", accepted)
	}
}

func TestTokenBucketRateLimiter_Accept(t *testing.T) {
	limiter := NewTokenBucketRateLimiter(100, 1)

	start := time.Now()
	limiter.Accept()
	limiter.Accept()
	elapsed := time.Since(start)

	if elapsed < 5*time.Millisecond {
		t.Errorf("Accept() should block for refill, elapsed = %v", elapsed)
	}
}

func TestTokenBucketRateLimiter_Wait(t *testing.T) {
	limiter := NewTokenBucketRateLimiter(100, 5)

	ctx := context.Background()
	err := limiter.Wait(ctx)
	if err != nil {
		t.Errorf("Wait() error = %v", err)
	}
}

func TestTokenBucketRateLimiter_Wait_Cancelled(t *testing.T) {
	limiter := NewTokenBucketRateLimiter(0.1, 1)
	limiter.TryAccept()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := limiter.Wait(ctx)
	if err == nil {
		t.Error("Wait() with cancelled context should return error")
	}
}

func TestTokenBucketRateLimiter_Stop(t *testing.T) {
	limiter := NewTokenBucketRateLimiter(10, 5)
	limiter.Stop()
}

func TestNewTokenBucketPassiveRateLimiter(t *testing.T) {
	limiter := NewTokenBucketPassiveRateLimiter(10, 5)
	if limiter == nil {
		t.Fatal("NewTokenBucketPassiveRateLimiter returned nil")
	}
	if limiter.QPS() != 10 {
		t.Errorf("QPS() = %v, want 10", limiter.QPS())
	}
}

func TestFakeAlwaysRateLimiter(t *testing.T) {
	limiter := NewFakeAlwaysRateLimiter()

	for i := 0; i < 100; i++ {
		if !limiter.TryAccept() {
			t.Error("FakeAlwaysRateLimiter.TryAccept() should always return true")
			break
		}
	}

	limiter.Accept()
	limiter.Stop()

	err := limiter.Wait(context.Background())
	if err != nil {
		t.Errorf("FakeAlwaysRateLimiter.Wait() error = %v", err)
	}
}

func TestFakeNeverRateLimiter(t *testing.T) {
	limiter := NewFakeNeverRateLimiter()

	if limiter.TryAccept() {
		t.Error("FakeNeverRateLimiter.TryAccept() should always return false")
	}

	err := limiter.Wait(context.Background())
	if err == nil {
		t.Error("FakeNeverRateLimiter.Wait() should return error")
	}

	done := make(chan struct{})
	go func() {
		limiter.Accept()
		close(done)
	}()

	limiter.Stop()
	<-done
}

func TestNewTokenBucketRateLimiterWithClock(t *testing.T) {
	clock := RealClock{}
	limiter := NewTokenBucketRateLimiterWithClock(10, 5, clock)
	if limiter == nil {
		t.Fatal("NewTokenBucketRateLimiterWithClock returned nil")
	}
}

func TestNewTokenBucketPassiveRateLimiterWithClock(t *testing.T) {
	clock := RealClock{}
	limiter := NewTokenBucketPassiveRateLimiterWithClock(10, 5, clock)
	if limiter == nil {
		t.Fatal("NewTokenBucketPassiveRateLimiterWithClock returned nil")
	}
}
