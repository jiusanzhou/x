package talk

import (
	"context"
	"errors"
	"testing"
)

func TestAuthMiddleware_NoEndpointInContext(t *testing.T) {
	authFn := func(ctx context.Context, req any) (string, error) {
		return "user1", nil
	}

	mw := AuthMiddleware(authFn)
	handler := mw(func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})

	// No endpoint in context → should pass through
	resp, err := handler(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Errorf("resp = %v, want ok", resp)
	}
}

func TestAuthMiddleware_AuthNone(t *testing.T) {
	authCalled := false
	authFn := func(ctx context.Context, req any) (string, error) {
		authCalled = true
		return "", nil
	}

	ep := &Endpoint{
		Name:     "Public",
		Metadata: map[string]any{"auth": "none"},
	}

	mw := AuthMiddleware(authFn)
	handler := mw(func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})

	ctx := WithEndpointContext(context.Background(), ep)
	resp, err := handler(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Errorf("resp = %v, want ok", resp)
	}
	if authCalled {
		t.Error("auth should not be called for auth=none")
	}
}

func TestAuthMiddleware_AuthToken_Success(t *testing.T) {
	authFn := func(ctx context.Context, req any) (string, error) {
		return "user-42", nil
	}

	ep := &Endpoint{
		Name:     "Protected",
		Metadata: map[string]any{"auth": "token"},
	}

	mw := AuthMiddleware(authFn)
	handler := mw(func(ctx context.Context, req any) (any, error) {
		identity, ok := IdentityFromContext(ctx)
		if !ok {
			t.Error("identity should be in context")
		}
		if identity != "user-42" {
			t.Errorf("identity = %q, want user-42", identity)
		}

		level := AuthLevelFromContext(ctx)
		if level != AuthToken {
			t.Errorf("auth level = %q, want token", level)
		}
		return "ok", nil
	})

	ctx := WithEndpointContext(context.Background(), ep)
	resp, err := handler(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Errorf("resp = %v, want ok", resp)
	}
}

func TestAuthMiddleware_AuthToken_Failure(t *testing.T) {
	authFn := func(ctx context.Context, req any) (string, error) {
		return "", errors.New("invalid token")
	}

	ep := &Endpoint{
		Name:     "Protected",
		Metadata: map[string]any{"auth": "token"},
	}

	mw := AuthMiddleware(authFn)
	handler := mw(func(ctx context.Context, req any) (any, error) {
		t.Error("hr should not be called")
		return nil, nil
	})

	ctx := WithEndpointContext(context.Background(), ep)
	_, err := handler(ctx, nil)
	if err == nil {
		t.Fatal("expected error")
	}

	talkErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if talkErr.Code != Unauthenticated {
		t.Errorf("code = %v, want Unauthenticated", talkErr.Code)
	}
}

func TestAuthMiddleware_NoAuthMetadata(t *testing.T) {
	authCalled := false
	authFn := func(ctx context.Context, req any) (string, error) {
		authCalled = true
		return "", nil
	}

	ep := &Endpoint{
		Name:     "NoAuth",
		Metadata: map[string]any{},
	}

	mw := AuthMiddleware(authFn)
	handler := mw(func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})

	ctx := WithEndpointContext(context.Background(), ep)
	resp, err := handler(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Errorf("resp = %v, want ok", resp)
	}
	if authCalled {
		t.Error("auth should not be called without auth metadata")
	}
}

func TestAuthMiddleware_AdminLevel(t *testing.T) {
	authFn := func(ctx context.Context, req any) (string, error) {
		return "admin-user", nil
	}

	ep := &Endpoint{
		Name:     "AdminOnly",
		Metadata: map[string]any{"auth": "admin"},
	}

	mw := AuthMiddleware(authFn)
	handler := mw(func(ctx context.Context, req any) (any, error) {
		level := AuthLevelFromContext(ctx)
		if level != AuthAdmin {
			t.Errorf("auth level = %q, want admin", level)
		}
		return "ok", nil
	})

	ctx := WithEndpointContext(context.Background(), ep)
	_, err := handler(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEndpointFromContext(t *testing.T) {
	ep := &Endpoint{Name: "Test"}
	ctx := WithEndpointContext(context.Background(), ep)

	got := EndpointFromContext(ctx)
	if got != ep {
		t.Errorf("got %v, want %v", got, ep)
	}

	// nil context
	got2 := EndpointFromContext(context.Background())
	if got2 != nil {
		t.Errorf("expected nil from empty context, got %v", got2)
	}
}

func TestIdentityFromContext_NotSet(t *testing.T) {
	_, ok := IdentityFromContext(context.Background())
	if ok {
		t.Error("expected ok=false for empty context")
	}
}

func TestAuthLevelFromContext_NotSet(t *testing.T) {
	level := AuthLevelFromContext(context.Background())
	if level != "" {
		t.Errorf("expected empty, got %q", level)
	}
}

func TestParseAnnotation_Auth(t *testing.T) {
	ann := parseAnnotation("@talk path=/admin method=POST auth=admin")
	if ann == nil {
		t.Fatal("expected annotation")
	}
	if ann.auth != "admin" {
		t.Errorf("auth = %q, want admin", ann.auth)
	}
	if ann.path != "/admin" {
		t.Errorf("path = %q, want /admin", ann.path)
	}
}
