package x

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetryableError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		err := RetryableError(nil)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("wraps error", func(t *testing.T) {
		original := errors.New("test error")
		wrapped := RetryableError(original)
		if wrapped.Error() != "test error" {
			t.Errorf("expected 'test error', got %q", wrapped.Error())
		}
	})

	t.Run("unwraps correctly", func(t *testing.T) {
		original := errors.New("test error")
		wrapped := RetryableError(original)
		if !errors.Is(wrapped, original) {
			t.Error("expected wrapped error to contain original")
		}
	})
}

func TestIsRetryable(t *testing.T) {
	t.Run("retryable error", func(t *testing.T) {
		err := RetryableError(errors.New("test"))
		if !IsRetryable(err) {
			t.Error("expected error to be retryable")
		}
	})

	t.Run("non-retryable error", func(t *testing.T) {
		err := errors.New("test")
		if IsRetryable(err) {
			t.Error("expected error to not be retryable")
		}
	})

	t.Run("nil error", func(t *testing.T) {
		if IsRetryable(nil) {
			t.Error("expected nil to not be retryable")
		}
	})
}

func TestRetry(t *testing.T) {
	t.Run("succeeds on first attempt", func(t *testing.T) {
		var attempts int32
		err := Retry(context.Background(), NewConstantBackoff(time.Millisecond), func(ctx context.Context) error {
			atomic.AddInt32(&attempts, 1)
			return nil
		})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if attempts != 1 {
			t.Errorf("expected 1 attempt, got %d", attempts)
		}
	})

	t.Run("retries on retryable error", func(t *testing.T) {
		var attempts int32
		backoff := WithMaxRetries(3, NewConstantBackoff(time.Millisecond))
		err := Retry(context.Background(), backoff, func(ctx context.Context) error {
			n := atomic.AddInt32(&attempts, 1)
			if n < 3 {
				return RetryableError(errors.New("retry"))
			}
			return nil
		})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if attempts != 3 {
			t.Errorf("expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("stops on non-retryable error", func(t *testing.T) {
		var attempts int32
		expectedErr := errors.New("fatal error")
		err := Retry(context.Background(), NewConstantBackoff(time.Millisecond), func(ctx context.Context) error {
			atomic.AddInt32(&attempts, 1)
			return expectedErr
		})
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected %v, got %v", expectedErr, err)
		}
		if attempts != 1 {
			t.Errorf("expected 1 attempt, got %d", attempts)
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := Retry(ctx, NewConstantBackoff(time.Second), func(ctx context.Context) error {
			return RetryableError(errors.New("retry"))
		})
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})

	t.Run("respects context timeout during wait", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		var attempts int32
		err := Retry(ctx, NewConstantBackoff(100*time.Millisecond), func(ctx context.Context) error {
			atomic.AddInt32(&attempts, 1)
			return RetryableError(errors.New("retry"))
		})
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("expected context.DeadlineExceeded, got %v", err)
		}
	})

	t.Run("returns unwrapped error when max retries exceeded", func(t *testing.T) {
		originalErr := errors.New("original")
		backoff := WithMaxRetries(2, NewConstantBackoff(time.Millisecond))
		err := Retry(context.Background(), backoff, func(ctx context.Context) error {
			return RetryableError(originalErr)
		})
		if !errors.Is(err, originalErr) {
			t.Errorf("expected %v, got %v", originalErr, err)
		}
	})
}

func TestDo(t *testing.T) {
	var attempts int32
	err := Do(context.Background(), NewConstantBackoff(time.Millisecond), func(ctx context.Context) error {
		n := atomic.AddInt32(&attempts, 1)
		if n < 2 {
			return RetryableError(errors.New("retry"))
		}
		return nil
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestConstantBackoff(t *testing.T) {
	b := NewConstantBackoff(100 * time.Millisecond)

	for i := 0; i < 5; i++ {
		d, ok := b.Next()
		if !ok {
			t.Error("expected ok to be true")
		}
		if d != 100*time.Millisecond {
			t.Errorf("expected 100ms, got %v", d)
		}
	}

	b.Reset()
	d, _ := b.Next()
	if d != 100*time.Millisecond {
		t.Errorf("expected 100ms after reset, got %v", d)
	}
}

func TestExponentialBackoff(t *testing.T) {
	b := NewExponentialBackoff(100*time.Millisecond, 1*time.Second)

	expected := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		400 * time.Millisecond,
		800 * time.Millisecond,
		1 * time.Second,
		1 * time.Second,
	}

	for i, exp := range expected {
		d, ok := b.Next()
		if !ok {
			t.Errorf("iteration %d: expected ok to be true", i)
		}
		if d != exp {
			t.Errorf("iteration %d: expected %v, got %v", i, exp, d)
		}
	}

	b.Reset()
	d, _ := b.Next()
	if d != 100*time.Millisecond {
		t.Errorf("expected 100ms after reset, got %v", d)
	}
}

func TestFibonacciBackoff(t *testing.T) {
	b := NewFibonacciBackoff(100*time.Millisecond, 1*time.Second)

	expected := []time.Duration{
		100 * time.Millisecond,
		100 * time.Millisecond,
		200 * time.Millisecond,
		300 * time.Millisecond,
		500 * time.Millisecond,
		800 * time.Millisecond,
		1 * time.Second,
		1 * time.Second,
	}

	for i, exp := range expected {
		d, ok := b.Next()
		if !ok {
			t.Errorf("iteration %d: expected ok to be true", i)
		}
		if d != exp {
			t.Errorf("iteration %d: expected %v, got %v", i, exp, d)
		}
	}

	b.Reset()
	d, _ := b.Next()
	if d != 100*time.Millisecond {
		t.Errorf("expected 100ms after reset, got %v", d)
	}
}

func TestWithMaxRetries(t *testing.T) {
	b := WithMaxRetries(3, NewConstantBackoff(100*time.Millisecond))

	for i := uint64(0); i < 3; i++ {
		d, ok := b.Next()
		if !ok {
			t.Errorf("iteration %d: expected ok to be true", i)
		}
		if d != 100*time.Millisecond {
			t.Errorf("iteration %d: expected 100ms, got %v", i, d)
		}
	}

	_, ok := b.Next()
	if ok {
		t.Error("expected ok to be false after max retries")
	}

	b.Reset()
	d, ok := b.Next()
	if !ok || d != 100*time.Millisecond {
		t.Error("expected reset to work")
	}
}

func TestWithCappedDuration(t *testing.T) {
	b := WithCappedDuration(500*time.Millisecond, NewExponentialBackoff(100*time.Millisecond, 10*time.Second))

	expected := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		400 * time.Millisecond,
		500 * time.Millisecond,
		500 * time.Millisecond,
	}

	for i, exp := range expected {
		d, ok := b.Next()
		if !ok {
			t.Errorf("iteration %d: expected ok to be true", i)
		}
		if d != exp {
			t.Errorf("iteration %d: expected %v, got %v", i, exp, d)
		}
	}
}

func TestWithMaxDuration(t *testing.T) {
	b := WithMaxDuration(500*time.Millisecond, NewConstantBackoff(100*time.Millisecond))

	count := 0
	for {
		_, ok := b.Next()
		if !ok {
			break
		}
		count++
		time.Sleep(100 * time.Millisecond)
		if count > 10 {
			t.Fatal("expected backoff to stop")
		}
	}

	if count < 3 || count > 6 {
		t.Errorf("expected around 5 iterations, got %d", count)
	}
}

func TestWithJitter(t *testing.T) {
	b := WithJitter(50*time.Millisecond, NewConstantBackoff(100*time.Millisecond))

	for i := 0; i < 10; i++ {
		d, ok := b.Next()
		if !ok {
			t.Error("expected ok to be true")
		}
		if d < 50*time.Millisecond || d > 150*time.Millisecond {
			t.Errorf("expected duration between 50ms and 150ms, got %v", d)
		}
	}
}

func TestWithJitterPercent(t *testing.T) {
	b := WithJitterPercent(10, NewConstantBackoff(100*time.Millisecond))

	for i := 0; i < 10; i++ {
		d, ok := b.Next()
		if !ok {
			t.Error("expected ok to be true")
		}
		if d < 90*time.Millisecond || d > 110*time.Millisecond {
			t.Errorf("expected duration between 90ms and 110ms, got %v", d)
		}
	}
}

func TestConvenienceFunctions(t *testing.T) {
	t.Run("Constant", func(t *testing.T) {
		var attempts int32
		err := Constant(context.Background(), time.Millisecond, func(ctx context.Context) error {
			n := atomic.AddInt32(&attempts, 1)
			if n < 3 {
				return RetryableError(errors.New("retry"))
			}
			return nil
		})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("Exponential", func(t *testing.T) {
		var attempts int32
		err := Exponential(context.Background(), time.Millisecond, 100*time.Millisecond, func(ctx context.Context) error {
			n := atomic.AddInt32(&attempts, 1)
			if n < 3 {
				return RetryableError(errors.New("retry"))
			}
			return nil
		})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("Fibonacci", func(t *testing.T) {
		var attempts int32
		err := Fibonacci(context.Background(), time.Millisecond, 100*time.Millisecond, func(ctx context.Context) error {
			n := atomic.AddInt32(&attempts, 1)
			if n < 3 {
				return RetryableError(errors.New("retry"))
			}
			return nil
		})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func TestBackoffReset(t *testing.T) {
	t.Run("exponential reset", func(t *testing.T) {
		b := NewExponentialBackoff(100*time.Millisecond, 10*time.Second)
		b.Next()
		b.Next()
		b.Next()
		b.Reset()
		d, _ := b.Next()
		if d != 100*time.Millisecond {
			t.Errorf("expected 100ms after reset, got %v", d)
		}
	})

	t.Run("fibonacci reset", func(t *testing.T) {
		b := NewFibonacciBackoff(100*time.Millisecond, 10*time.Second)
		b.Next()
		b.Next()
		b.Next()
		b.Reset()
		d, _ := b.Next()
		if d != 100*time.Millisecond {
			t.Errorf("expected 100ms after reset, got %v", d)
		}
	})

	t.Run("withMaxRetries reset", func(t *testing.T) {
		b := WithMaxRetries(2, NewConstantBackoff(100*time.Millisecond))
		b.Next()
		b.Next()
		_, ok := b.Next()
		if ok {
			t.Error("expected exhausted after 2 retries")
		}
		b.Reset()
		_, ok = b.Next()
		if !ok {
			t.Error("expected ok after reset")
		}
	})
}

func TestCombinedMiddleware(t *testing.T) {
	b := WithMaxRetries(5,
		WithCappedDuration(200*time.Millisecond,
			WithJitterPercent(10,
				NewExponentialBackoff(100*time.Millisecond, 10*time.Second))))

	for i := 0; i < 5; i++ {
		d, ok := b.Next()
		if !ok {
			t.Errorf("iteration %d: expected ok to be true", i)
		}
		if d > 220*time.Millisecond {
			t.Errorf("iteration %d: expected duration <= 220ms (cap+jitter), got %v", i, d)
		}
	}

	_, ok := b.Next()
	if ok {
		t.Error("expected ok to be false after max retries")
	}
}
