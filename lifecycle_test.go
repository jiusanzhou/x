package x

import (
	"context"
	"errors"
	"testing"
)

type mockLifecycle struct {
	initCalled  bool
	closeCalled bool
	err         error
}

func (m *mockLifecycle) Init(ctx context.Context) error {
	m.initCalled = true
	return m.err
}
func (m *mockLifecycle) Close(ctx context.Context) error {
	m.closeCalled = true
	return m.err
}

type mockHealthChecker struct {
	err error
}

func (m *mockHealthChecker) Healthy(ctx context.Context) error {
	return m.err
}

type mockReloadable struct {
	reloaded bool
	lastCfg  TypedLazyConfig
	err      error
}

func (m *mockReloadable) Reload(cfg TypedLazyConfig) error {
	m.reloaded = true
	m.lastCfg = cfg
	return m.err
}

type plainStruct struct{}

func TestTryInit(t *testing.T) {
	ctx := context.Background()

	// implements Lifecycle
	lc := &mockLifecycle{}
	if err := TryInit(ctx, lc); err != nil {
		t.Errorf("TryInit() error = %v", err)
	}
	if !lc.initCalled {
		t.Error("TryInit() did not call Init()")
	}

	// with error
	lcErr := &mockLifecycle{err: errors.New("init fail")}
	if err := TryInit(ctx, lcErr); err == nil {
		t.Error("TryInit() should propagate error")
	}

	// does not implement Lifecycle
	if err := TryInit(ctx, &plainStruct{}); err != nil {
		t.Errorf("TryInit() on plain struct should return nil, got %v", err)
	}
}

func TestTryClose(t *testing.T) {
	ctx := context.Background()

	lc := &mockLifecycle{}
	if err := TryClose(ctx, lc); err != nil {
		t.Errorf("TryClose() error = %v", err)
	}
	if !lc.closeCalled {
		t.Error("TryClose() did not call Close()")
	}

	if err := TryClose(ctx, &plainStruct{}); err != nil {
		t.Errorf("TryClose() on plain struct should return nil, got %v", err)
	}
}

func TestTryHealthy(t *testing.T) {
	ctx := context.Background()

	hc := &mockHealthChecker{}
	if err := TryHealthy(ctx, hc); err != nil {
		t.Errorf("TryHealthy() healthy check error = %v", err)
	}

	hcErr := &mockHealthChecker{err: errors.New("unhealthy")}
	if err := TryHealthy(ctx, hcErr); err == nil {
		t.Error("TryHealthy() should propagate error")
	}

	if err := TryHealthy(ctx, &plainStruct{}); err != nil {
		t.Errorf("TryHealthy() on plain struct should return nil, got %v", err)
	}
}

func TestTryReload(t *testing.T) {
	cfg := TypedLazyConfig{Type: "new", Name: "updated"}

	r := &mockReloadable{}
	if err := TryReload(r, cfg); err != nil {
		t.Errorf("TryReload() error = %v", err)
	}
	if !r.reloaded {
		t.Error("TryReload() did not call Reload()")
	}
	if r.lastCfg.Type != "new" {
		t.Errorf("TryReload() cfg.Type = %q, want %q", r.lastCfg.Type, "new")
	}

	if err := TryReload(&plainStruct{}, cfg); err != nil {
		t.Errorf("TryReload() on plain struct should return nil, got %v", err)
	}
}
