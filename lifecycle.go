package x

import "context"

// Lifecycle defines common lifecycle hooks for pluggable components.
// Implement any subset — callers should type-assert before calling.
type Lifecycle interface {
	// Init performs any deferred initialization.
	Init(ctx context.Context) error

	// Close releases resources held by the component.
	Close(ctx context.Context) error
}

// HealthChecker can report its own health.
type HealthChecker interface {
	// Healthy returns nil when the component is operating normally.
	Healthy(ctx context.Context) error
}

// Reloadable can hot-reload its configuration without full restart.
type Reloadable interface {
	// Reload applies a new configuration. Implementations should be
	// safe for concurrent calls.
	Reload(cfg TypedLazyConfig) error
}

// --- helpers for callers ---

// TryInit calls Init if v implements Lifecycle, otherwise returns nil.
func TryInit(ctx context.Context, v any) error {
	if lc, ok := v.(Lifecycle); ok {
		return lc.Init(ctx)
	}
	return nil
}

// TryClose calls Close if v implements Lifecycle, otherwise returns nil.
func TryClose(ctx context.Context, v any) error {
	if lc, ok := v.(Lifecycle); ok {
		return lc.Close(ctx)
	}
	return nil
}

// TryHealthy calls Healthy if v implements HealthChecker, otherwise returns nil.
func TryHealthy(ctx context.Context, v any) error {
	if hc, ok := v.(HealthChecker); ok {
		return hc.Healthy(ctx)
	}
	return nil
}

// TryReload calls Reload if v implements Reloadable, otherwise returns nil.
func TryReload(v any, cfg TypedLazyConfig) error {
	if r, ok := v.(Reloadable); ok {
		return r.Reload(cfg)
	}
	return nil
}
