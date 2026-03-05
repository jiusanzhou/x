package talk

import (
	"context"
	"testing"
)

type groupTestService struct{}

func (s *groupTestService) GetItem(ctx context.Context, id string) (string, error) {
	return id, nil
}

func (s *groupTestService) CreateItem(ctx context.Context, req *struct{ Name string }) (*struct{ Name string }, error) {
	return req, nil
}

func TestGroup_Register(t *testing.T) {
	server := NewServer(&mockTransport{}, WithPathPrefix("/api/v1"))

	group := server.Group("/admin")
	if err := group.Register(&groupTestService{}); err != nil {
		t.Fatalf("Register error: %v", err)
	}

	endpoints := server.Endpoints()
	if len(endpoints) == 0 {
		t.Fatal("expected endpoints")
	}

	for _, ep := range endpoints {
		if ep.Path[:len("/api/v1/admin")] != "/api/v1/admin" {
			t.Errorf("endpoint %s path = %q, expected prefix /api/v1/admin", ep.Name, ep.Path)
		}
	}
}

func TestGroup_Middleware(t *testing.T) {
	calls := []string{}

	serverMW := func(next EndpointFunc) EndpointFunc {
		return func(ctx context.Context, req any) (any, error) {
			calls = append(calls, "server")
			return next(ctx, req)
		}
	}

	groupMW := func(next EndpointFunc) EndpointFunc {
		return func(ctx context.Context, req any) (any, error) {
			calls = append(calls, "group")
			return next(ctx, req)
		}
	}

	server := NewServer(&mockTransport{}, WithServerMiddleware(serverMW))
	group := server.Group("/admin", groupMW)

	ep := NewEndpoint("Test", func(ctx context.Context, req any) (any, error) {
		calls = append(calls, "handler")
		return "ok", nil
	})

	group.RegisterEndpoints(ep)

	endpoints := server.Endpoints()
	if len(endpoints) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(endpoints))
	}

	// Should have 2 middleware: server + group
	if len(endpoints[0].Middleware) != 2 {
		t.Fatalf("expected 2 middleware, got %d", len(endpoints[0].Middleware))
	}

	// Execute and verify order: server → group → handler
	h := endpoints[0].WrappedHandler()
	h(context.Background(), nil)

	expected := []string{"server", "group", "handler"}
	if len(calls) != len(expected) {
		t.Fatalf("calls = %v, want %v", calls, expected)
	}
	for i, v := range expected {
		if calls[i] != v {
			t.Errorf("calls[%d] = %q, want %q", i, calls[i], v)
		}
	}
}

func TestGroup_NestedGroup(t *testing.T) {
	server := NewServer(&mockTransport{}, WithPathPrefix("/api"))

	admin := server.Group("/admin")
	superAdmin := admin.Group("/super")

	ep := NewEndpoint("Test", func(ctx context.Context, req any) (any, error) {
		return nil, nil
	})

	superAdmin.RegisterEndpoints(ep)

		endpoints := server.Endpoints()
	if len(endpoints) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(endpoints))
	}

	if endpoints[0].Path != "/api/admin/super" {
		t.Errorf("path = %q, want /api/admin/super", endpoints[0].Path)
	}
}

func TestGroup_NestedMiddleware(t *testing.T) {
	calls := []string{}

	mw := func(name string) MiddlewareFunc {
		return func(next EndpointFunc) EndpointFunc {
			return func(ctx context.Context, req any) (any, error) {
				calls = append(calls, name)
				return next(ctx, req)
			}
		}
	}

	server := NewServer(&mockTransport{}, WithServerMiddleware(mw("server")))
	admin := server.Group("/admin", mw("admin"))
	super := admin.Group("/super", mw("super"))

	ep := NewEndpoint("Test", func(ctx context.Context, req any) (any, error) {
		calls = append(calls, "handler")
		return nil, nil
	})
	super.RegisterEndpoints(ep)

	h := server.Endpoints()[0].WrappedHandler()
	h(context.Background(), nil)

	expected := []string{"server", "admin", "super", "handler"}
	if len(calls) != len(expected) {
		t.Fatalf("calls = %v, want %v", calls, expected)
	}
	for i, v := range expected {
		if calls[i] != v {
			t.Errorf("calls[%d] = %q, want %q", i, calls[i], v)
		}
	}
}
