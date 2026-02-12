package x

import (
	"context"
	"sync"
	"time"
)

// Interval represents a repeatedly called function with fixed time delay.
type Interval struct {
	cancel context.CancelFunc
	done   chan struct{}
	once   sync.Once
}

// IntervalSet repeatedly calls a function with a fixed time delay between each call.
// Returns an Interval that can be stopped with Stop().
func IntervalSet(fn func(t time.Time), delay time.Duration) *Interval {
	ctx, cancel := context.WithCancel(context.Background())
	iv := &Interval{
		cancel: cancel,
		done:   make(chan struct{}),
	}

	go func() {
		defer close(iv.done)
		ticker := time.NewTicker(delay)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				fn(t)
			}
		}
	}()

	return iv
}

// IntervalSetWithContext is like IntervalSet but accepts a parent context.
func IntervalSetWithContext(ctx context.Context, fn func(t time.Time), delay time.Duration) *Interval {
	ctx, cancel := context.WithCancel(ctx)
	iv := &Interval{
		cancel: cancel,
		done:   make(chan struct{}),
	}

	go func() {
		defer close(iv.done)
		ticker := time.NewTicker(delay)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				fn(t)
			}
		}
	}()

	return iv
}

// Stop cancels the repeating action and waits for it to complete.
func (iv *Interval) Stop() {
	iv.once.Do(func() {
		iv.cancel()
		<-iv.done
	})
}

// StopAsync cancels the repeating action without waiting.
func (iv *Interval) StopAsync() {
	iv.cancel()
}

// Done returns a channel that is closed when the interval stops.
func (iv *Interval) Done() <-chan struct{} {
	return iv.done
}

// IntervalFunc runs a function at regular intervals until the context is cancelled
// or the function returns an error.
func IntervalFunc(ctx context.Context, delay time.Duration, fn func(ctx context.Context) error) error {
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := fn(ctx); err != nil {
				return err
			}
		}
	}
}

// IntervalFuncImmediate is like IntervalFunc but runs the function immediately before starting the interval.
func IntervalFuncImmediate(ctx context.Context, delay time.Duration, fn func(ctx context.Context) error) error {
	if err := fn(ctx); err != nil {
		return err
	}
	return IntervalFunc(ctx, delay, fn)
}
