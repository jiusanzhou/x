package x

import (
	"context"
	"sync/atomic"
	"testing"
)

func TestNewGraceRunner(t *testing.T) {
	runner := NewGraceRunner()
	if runner == nil {
		t.Fatal("NewGraceRunner() returned nil")
	}
	if runner.ctx == nil {
		t.Error("GraceRunner.ctx is nil")
	}
	if runner.cancel == nil {
		t.Error("GraceRunner.cancel is nil")
	}
	if runner.cleanups == nil {
		t.Error("GraceRunner.cleanups is nil")
	}
}

func TestGraceRunner_Context(t *testing.T) {
	runner := NewGraceRunner()
	ctx := runner.Context()

	if ctx == nil {
		t.Fatal("Context() returned nil")
	}

	select {
	case <-ctx.Done():
		t.Error("Context should not be cancelled initially")
	default:
	}
}

func TestGraceRunner_RegisterCleanup(t *testing.T) {
	runner := NewGraceRunner()

	if len(runner.cleanups) != 0 {
		t.Errorf("initial cleanups len = %d, want 0", len(runner.cleanups))
	}

	runner.RegisterCleanup(func() {})
	if len(runner.cleanups) != 1 {
		t.Errorf("after RegisterCleanup cleanups len = %d, want 1", len(runner.cleanups))
	}

	runner.RegisterCleanup(func() {})
	runner.RegisterCleanup(func() {})
	if len(runner.cleanups) != 3 {
		t.Errorf("after 3 RegisterCleanup cleanups len = %d, want 3", len(runner.cleanups))
	}
}

func TestGraceRunner_runCleanups_Order(t *testing.T) {
	runner := NewGraceRunner()
	var order []int

	runner.RegisterCleanup(func() { order = append(order, 1) })
	runner.RegisterCleanup(func() { order = append(order, 2) })
	runner.RegisterCleanup(func() { order = append(order, 3) })

	runner.runCleanups()

	if len(order) != 3 {
		t.Fatalf("runCleanups executed %d functions, want 3", len(order))
	}
	if order[0] != 3 || order[1] != 2 || order[2] != 1 {
		t.Errorf("runCleanups order = %v, want [3, 2, 1] (LIFO)", order)
	}
}

func TestGraceRunWithCleanup(t *testing.T) {
	var cleanupCalled int32

	err := GraceRunWithCleanup(
		func(register func(CleanupFunc)) {
			register(func() { atomic.AddInt32(&cleanupCalled, 1) })
		},
		func() error {
			return nil
		},
	)

	if err != nil {
		t.Errorf("GraceRunWithCleanup() error = %v", err)
	}

	if atomic.LoadInt32(&cleanupCalled) != 1 {
		t.Errorf("cleanup called %d times, want 1", cleanupCalled)
	}
}

func TestGraceRunWithCleanup_MultipleCleanups(t *testing.T) {
	var order []int

	GraceRunWithCleanup(
		func(register func(CleanupFunc)) {
			register(func() { order = append(order, 1) })
			register(func() { order = append(order, 2) })
			register(func() { order = append(order, 3) })
		},
		func() error {
			return nil
		},
	)

	if len(order) != 3 {
		t.Fatalf("cleanups executed %d times, want 3", len(order))
	}
	if order[0] != 3 || order[1] != 2 || order[2] != 1 {
		t.Errorf("cleanup order = %v, want [3, 2, 1]", order)
	}
}

func TestGreRunner_ContextCancelledOnReturn(t *testing.T) {
	runner := NewGraceRunner()
	ctx := runner.Context()

	done := make(chan struct{})
	go func() {
		runner.Run(func() error {
			return nil
		})
		close(done)
	}()

	<-done

	select {
	case <-ctx.Done():
	default:
		t.Error("Context should be cancelled after Run returns")
	}
}

func TestGraceRunner_Run_ReturnsError(t *testing.T) {
	runner := NewGraceRunner()
	expectedErr := context.DeadlineExceeded

	err := runner.Run(func() error {
		return expectedErr
	})

	if err != expectedErr {
		t.Errorf("Run() error = %v, want %v", err, expectedErr)
	}
}
