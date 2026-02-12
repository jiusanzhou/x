// Copyright 2020 Zoe Blade
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package x

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"
)

// RetryFunc is the function signature for retryable operations.
// The function receives a context and should return an error if the operation fails.
// To indicate that the error is retryable, wrap it with RetryableError.
// If the returned error is not wrapped with RetryableError, the retry loop will stop.
type RetryFunc func(ctx context.Context) error

// RetryBackoff is the interface for backoff strategies.
// Each call to Next returns the next backoff duration.
// Implementations should be safe for concurrent use.
type RetryBackoff interface {
	// Next returns the next backoff duration and whether to continue retrying.
	// If the second return value is false, the retry loop should stop.
	Next() (time.Duration, bool)

	// Reset resets the backoff to its initial state.
	Reset()
}

// retryableError is a sentinel error type that indicates the error is retryable.
type retryableError struct {
	err error
}

func (e *retryableError) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *retryableError) Unwrap() error {
	return e.err
}

// RetryableError marks an error as retryable. If the error returned from the
// retry function is wrapped with RetryableError, the retry loop will continue.
// Otherwise, the retry loop will stop and return the error immediately.
func RetryableError(err error) error {
	if err == nil {
		return nil
	}
	return &retryableError{err: err}
}

// IsRetryable returns true if the error is a retryable error.
func IsRetryable(err error) bool {
	var re *retryableError
	return errors.As(err, &re)
}

// Retry executes the given function with retry logic based on the provided backoff.
// The function will be retried until:
// - It returns nil (success)
// - It returns a non-retryable error (an error not wrapped with RetryableError)
// - The context is cancelled
// - The backoff indicates no more retries
//
// Example:
//
//	backoff := x.NewEBackoff(100*time.Millisecond, 10*time.Second)
//	err := x.Retry(ctx, backoff, func(ctx context.Context) error {
//	    if err := doSomething(); err != nil {
//	        return x.RetryableError(err) // Will retry
//	    }
//	    return nil // Success
//	})
func Retry(ctx context.Context, b RetryBackoff, f RetryFunc) error {
	// Reset backoff at the start
	b.Reset()

	for {
		// Check context before executing
		if err := ctx.Err(); err != nil {
			return err
		}

		// Execute the function
		err := f(ctx)
		if err == nil {
			return nil
		}

		// Check if the error is retryable
		if !IsRetryable(err) {
			return err
		}

		// Get the next backoff duration
		delay, ok := b.Next()
		if !ok {
			// Unwrap the retryable error before returning
			var re *retryableError
			if errors.As(err, &re) {
				return re.err
			}
			return err
		}

		// Wait for the backoff duration or context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next retry
		}
	}
}

// Do is an alias for Retry that provides a more intuitive API for simple cases.
func Do(ctx context.Context, b RetryBackoff, f RetryFunc) error {
	return Retry(ctx, b, f)
}

type constantBackoff struct {
	mu       sync.Mutex
	interval time.Duration
}

// NewConstantBackoff creates a new constant backoff with the given interval.
// This backoff always returns the same duration.
//
// Example: 1s -> 1s -> 1s -> 1s
func NewConstantBackoff(interval time.Duration) RetryBackoff {
	return &constantBackoff{interval: interval}
}

func (b *constantBackoff) Next() (time.Duration, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.interval, true
}

func (b *constantBackoff) Reset() {
	// No state to reset for constant backoff
}

// exponentialBackoff implements exponential backoff.
type exponentialBackoff struct {
	mu      sync.Mutex
	initial time.Duration
	max     time.Duration
	current time.Duration
}

// NewExponentialBackoff creates a new exponential backoff.
// The backoff starts at the initial duration and doubles each time,
// up to the maximum duration.
//
// Example with initial=1s, max=30s: 1s -> 2s -> 4s -> 8s -> 16s -> 30s -> 30s
func NewExponentialBackoff(initial, max time.Duration) RetryBackoff {
	return &exponentialBackoff{
		initial: initial,
		max:     max,
		current: initial,
	}
}

func (b *exponentialBackoff) Next() (time.Duration, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	result := b.current
	b.current *= 2
	if b.current > b.max {
		b.current = b.max
	}
	return result, true
}

func (b *exponentialBackoff) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.current = b.initial
}

// fibonacciBackoff implements Fibonacci backoff.
type fibonacciBackoff struct {
	mu      sync.Mutex
	initial time.Duration
	max     time.Duration
	prev    time.Duration
	current time.Duration
	started bool
}

// NewFibonacciBackoff creates a new Fibonacci backoff.
// The backoff follows the Fibonacci sequence, where each duration is the
// sum of the two previous durations, capped at max.
//
// Example with initial=1s, max=30s: 1s -> 1s -> 2s -> 3s -> 5s -> 8s -> 13s -> 21s -> 30s
func NewFibonacciBackoff(initial, max time.Duration) RetryBackoff {
	return &fibonacciBackoff{
		initial: initial,
		max:     max,
		prev:    0,
		current: initial,
		started: false,
	}
}

func (b *fibonacciBackoff) Next() (time.Duration, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.started {
		b.started = true
		return b.current, true
	}

	next := b.prev + b.current
	if next > b.max {
		next = b.max
	}
	b.prev = b.current
	b.current = next
	return b.current, true
}

func (b *fibonacciBackoff) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.prev = 0
	b.current = b.initial
	b.started = false
}

// withMaxRetries wraps a backoff to limit the number of retries.
type withMaxRetries struct {
	wrapped    RetryBackoff
	maxRetries uint64
	attempts   uint64
	mu         sync.Mutex
}

// WithMaxRetries wraps a backoff to limit the number of retries.
// After maxRetries retries, Next() will return false.
// Note: maxRetries is the number of *retries*, not attempts. Attempts = retries + 1.
//
// Example with maxRetries=3:
//
//	attempt 1 -> fail -> retry 1 -> fail -> retry 2 -> fail -> retry 3 -> fail -> stop
func WithMaxRetries(maxRetries uint64, b RetryBackoff) RetryBackoff {
	return &withMaxRetries{
		wrapped:    b,
		maxRetries: maxRetries,
	}
}

func (b *withMaxRetries) Next() (time.Duration, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.attempts >= b.maxRetries {
		return 0, false
	}
	b.attempts++
	return b.wrapped.Next()
}

func (b *withMaxRetries) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.attempts = 0
	b.wrapped.Reset()
}

// withCappedDuration wraps a backoff to cap the maximum duration.
type withCappedDuration struct {
	wrapped RetryBackoff
	cap     time.Duration
}

// WithCappedDuration wraps a backoff to cap individual backoff values.
// If the wrapped backoff returns a duration greater than cap, it returns cap instead.
//
// Example with cap=5s and Fibonacci(1s):
//
//	1s -> 1s -> 2s -> 3s -> 5s -> 5s -> 5s (capped from 8s, 13s, etc.)
func WithCappedDuration(cap time.Duration, b RetryBackoff) RetryBackoff {
	return &withCappedDuration{
		wrapped: b,
		cap:     cap,
	}
}

func (b *withCappedDuration) Next() (time.Duration, bool) {
	val, ok := b.wrapped.Next()
	if !ok {
		return 0, false
	}
	if val > b.cap {
		val = b.cap
	}
	return val, true
}

func (b *withCappedDuration) Reset() {
	b.wrapped.Reset()
}

// withMaxDuration wraps a backoff to limit the total retry duration.
type withMaxDuration struct {
	wrapped     RetryBackoff
	maxDuration time.Duration
	startTime   time.Time
	started     bool
	mu          sync.Mutex
}

// WithMaxDuration wraps a backoff to limit the total retry time.
// After maxDuration has elapsed since the first Next() call, Next() will return false.
//
// Example with maxDuration=10s:
//
//	(0s) 1s -> (1s) 1s -> (2s) 2s -> (4s) 3s -> (7s) 3s (would exceed 10s, so stop)
func WithMaxDuration(maxDuration time.Duration, b RetryBackoff) RetryBackoff {
	return &withMaxDuration{
		wrapped:     b,
		maxDuration: maxDuration,
	}
}

func (b *withMaxDuration) Next() (time.Duration, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.started {
		b.started = true
		b.startTime = time.Now()
	}

	// Check if we've exceeded the max duration
	elapsed := time.Since(b.startTime)
	if elapsed >= b.maxDuration {
		return 0, false
	}

	val, ok := b.wrapped.Next()
	if !ok {
		return 0, false
	}

	// Cap the delay if it would exceed the remaining time
	remaining := b.maxDuration - elapsed
	if val > remaining {
		val = remaining
	}

	return val, true
}

func (b *withMaxDuration) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.started = false
	b.wrapped.Reset()
}

// withJitter wraps a backoff to add random jitter.
type withJitter struct {
	wrapped RetryBackoff
	jitter  time.Duration
	rand    *rand.Rand
	mu      sync.Mutex
}

// WithJitter wraps a backoff to add random jitter to each backoff value.
// The jitter is added or subtracted from the backoff value, within the range [-jitter, +jitter].
// The resulting value is always non-negative.
//
// Example with jitter=100ms and Constant(1s):
//
//	~900ms -> ~1050ms -> ~980ms -> ~1100ms (varies randomly)
func WithJitter(jitter time.Duration, b RetryBackoff) RetryBackoff {
	return &withJitter{
		wrapped: b,
		jitter:  jitter,
		rand:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (b *withJitter) Next() (time.Duration, bool) {
	val, ok := b.wrapped.Next()
	if !ok {
		return 0, false
	}

	b.mu.Lock()
	// Generate random jitter in range [-jitter, +jitter]
	delta := time.Duration(b.rand.Int63n(int64(2*b.jitter+1))) - b.jitter
	b.mu.Unlock()

	val += delta
	if val < 0 {
		val = 0
	}
	return val, true
}

func (b *withJitter) Reset() {
	b.wrapped.Reset()
}

// withJitterPercent wraps a backoff to add percentage-based random jitter.
type withJitterPercent struct {
	wrapped RetryBackoff
	percent uint64
	rand    *rand.Rand
	mu      sync.Mutex
}

// WithJitterPercent wraps a backoff to add percentage-based random jitter.
// The jitter is calculated as a percentage of the backoff value.
// For example, with percent=10, the jitter range is [-10%, +10%] of the backoff value.
//
// Example with percent=10 and Constant(1s):
//
//	~950ms -> ~1080ms -> ~920ms -> ~1030ms (varies randomly, within +/- 10%)
func WithJitterPercent(percent uint64, b RetryBackoff) RetryBackoff {
	return &withJitterPercent{
		wrapped: b,
		percent: percent,
		rand:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (b *withJitterPercent) Next() (time.Duration, bool) {
	val, ok := b.wrapped.Next()
	if !ok {
		return 0, false
	}

	b.mu.Lock()
	// Calculate jitter based on percentage
	jitter := time.Duration(float64(val) * float64(b.percent) / 100)
	// Generate random jitter in range [-jitter, +jitter]
	var delta time.Duration
	if jitter > 0 {
		delta = time.Duration(b.rand.Int63n(int64(2*jitter+1))) - jitter
	}
	b.mu.Unlock()

	val += delta
	if val < 0 {
		val = 0
	}
	return val, true
}

func (b *withJitterPercent) Reset() {
	b.wrapped.Reset()
}

// Convenience functions for common patterns

// Constant executes the function with constant backoff.
func Constant(ctx context.Context, interval time.Duration, f RetryFunc) error {
	return Retry(ctx, NewConstantBackoff(interval), f)
}

// Exponential executes the function with exponential backoff.
func Exponential(ctx context.Context, initial, max time.Duration, f RetryFunc) error {
	return Retry(ctx, NewExponentialBackoff(initial, max), f)
}

// Fibonacci executes the function with Fibonacci backoff.
func Fibonacci(ctx context.Context, initial, max time.Duration, f RetryFunc) error {
	return Retry(ctx, NewFibonacciBackoff(initial, max), f)
}
