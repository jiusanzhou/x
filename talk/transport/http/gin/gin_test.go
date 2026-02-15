package gin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"go.zoe.im/x"
	"go.zoe.im/x/talk"
	"go.zoe.im/x/talk/codec"
	thttp "go.zoe.im/x/talk/transport/http"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type testResponse struct {
	Message string `json:"message"`
	ID      string `json:"id,omitempty"`
}

func TestNewServer(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":8080"}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	if server.String() != "http/gin" {
		t.Errorf("String() = %q, want %q", server.String(), "http/gin")
	}
}

func TestServer_Engine(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":8080"}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	engine := server.Engine()
	if engine == nil {
		t.Error("Engine() should not return nil")
	}
}

func TestServer_SetCodec(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":8080"}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	customCodec := codec.MustGet("json")
	server.SetCodec(customCodec)

	if server.codec != customCodec {
		t.Error("codec not set correctly")
	}
}

func TestServer_Handler(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":0"}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	endpoints := []*talk.Endpoint{
		{
			Name:   "GetItem",
			Path:   "/items/{id}",
			Method: "GET",
			Handler: func(ctx context.Context, req any) (any, error) {
				id, _ := req.(string)
				return &testResponse{Message: "Found", ID: id}, nil
			},
		},
		{
			Name:   "CreateItem",
			Path:   "/items",
			Method: "POST",
			Handler: func(ctx context.Context, req any) (any, error) {
				return &testResponse{Message: "Created"}, nil
			},
		},
		{
			Name:   "ErrorEndpoint",
			Path:   "/error",
			Method: "GET",
			Handler: func(ctx context.Context, req any) (any, error) {
				return nil, talk.NewError(talk.NotFound, "item not found")
			},
		},
	}

	for _, ep := range endpoints {
		server.registerEndpoint(ep)
	}

	t.Run("GET /items/:id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/items/123", nil)
		w := httptest.NewRecorder()

		server.engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}

		var result testResponse
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("decode failed: %v", err)
		}

		if result.ID != "123" {
			t.Errorf("ID = %q, want %q", result.ID, "123")
		}
	})

	t.Run("POST /items", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/items", nil)
		w := httptest.NewRecorder()

		server.engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("GET /error returns error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/error", nil)
		w := httptest.NewRecorder()

		server.engine.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestServer_ServeAndShutdown(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":0"}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	endpoints := []*talk.Endpoint{
		{
			Name:   "Hello",
			Path:   "/hello",
			Method: "GET",
			Handler: func(ctx context.Context, req any) (any, error) {
				return &testResponse{Message: "Hello, World!"}, nil
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = server.Serve(ctx, endpoints)
	if err != nil && err != context.DeadlineExceeded {
		t.Errorf("Serve returned unexpected error: %v", err)
	}
}

func TestServer_NotSupportInvoke(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":8080"}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	err = server.Invoke(context.Background(), "test", nil, nil)
	if err == nil {
		t.Error("expected error from server Invoke")
	}

	talkErr, ok := err.(*talk.Error)
	if !ok {
		t.Error("expected *talk.Error")
	}
	if talkErr.Code != talk.Unimplemented {
		t.Errorf("expected Unimplemented error, got %v", talkErr.Code)
	}
}

func TestServer_NotSupportInvokeStream(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":8080"}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	_, err = server.InvokeStream(context.Background(), "test", nil)
	if err == nil {
		t.Error("expected error from server InvokeStream")
	}

	talkErr, ok := err.(*talk.Error)
	if !ok {
		t.Error("expected *talk.Error")
	}
	if talkErr.Code != talk.Unimplemented {
		t.Errorf("expected Unimplemented error, got %v", talkErr.Code)
	}
}

func TestServer_ShutdownWithoutServe(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":8080"}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	err = server.Shutdown(context.Background())
	if err != nil {
		t.Errorf("Shutdown should not error when server not started: %v", err)
	}
}

func TestServer_Close(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":8080"}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	err = server.Close()
	if err != nil {
		t.Errorf("Close should not error: %v", err)
	}
}

func TestConvertPathParams(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/users/{id}", "/users/:id"},
		{"/users/{id}/posts/{postId}", "/users/:id/posts/:postId"},
		{"/simple", "/simple"},
		{"/{a}/{b}/{c}", "/:a/:b/:c"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := convertPathParams(tt.input)
			if result != tt.expected {
				t.Errorf("convertPathParams(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFactoryRegistration(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Type:   "gin",
		Config: json.RawMessage(`{"addr": ":8080"}`),
	}

	server, err := thttp.ServerFactory.Create(cfg)
	if err != nil {
		t.Fatalf("ServerFactory.Create failed: %v", err)
	}
	if server == nil {
		t.Error("ServerFactory.Create returned nil")
	}
}

func TestServer_SSEStreaming(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":0"}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	eventCount := 0
	endpoints := []*talk.Endpoint{
		{
			Name:       "WatchEvents",
			Path:       "/events",
			Method:     "GET",
			StreamMode: talk.StreamServerSide,
			StreamHandler: func(ctx context.Context, req any, stream talk.Stream) error {
				for i := 0; i < 3; i++ {
					eventCount++
					if err := stream.Send(map[string]int{"count": i}); err != nil {
						return err
					}
				}
				return nil
			},
		},
	}

	for _, ep := range endpoints {
		server.registerEndpoint(ep)
	}

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	w := httptest.NewRecorder()

	server.engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/event-stream") {
		t.Errorf("Content-Type = %q, want prefix %q", contentType, "text/event-stream")
	}

	if eventCount != 3 {
		t.Errorf("eventCount = %d, want 3", eventCount)
	}
}

func TestServer_AllHTTPMethods(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":0"}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		ep := &talk.Endpoint{
			Name:   fmt.Sprintf("Test%s", method),
			Path:   fmt.Sprintf("/test-%s", method),
			Method: method,
			Handler: func(ctx context.Context, req any) (any, error) {
				return map[string]string{"method": method}, nil
			},
		}
		server.registerEndpoint(ep)
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, fmt.Sprintf("/test-%s", method), nil)
			w := httptest.NewRecorder()

			server.engine.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
			}
		})
	}
}

func TestServer_CustomMiddleware(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":0"}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	// Add custom middleware
	middlewareCalled := false
	server.Engine().Use(func(c *gin.Context) {
		middlewareCalled = true
		c.Next()
	})

	endpoints := []*talk.Endpoint{
		{
			Name:   "Test",
			Path:   "/test",
			Method: "GET",
			Handler: func(ctx context.Context, req any) (any, error) {
				return map[string]string{"status": "ok"}, nil
			},
		},
	}

	for _, ep := range endpoints {
		server.registerEndpoint(ep)
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	server.engine.ServeHTTP(w, req)

	if !middlewareCalled {
		t.Error("custom middleware was not called")
	}
}
