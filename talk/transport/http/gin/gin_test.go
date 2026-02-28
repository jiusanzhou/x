package gin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
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

// --- Test types for new features ---

type pathParamRequest struct {
	NodeName  string `json:"nodeName" path:"nodeName"`
	ModelName string `json:"modelName" path:"modelName"`
}

type queryParamRequest struct {
	Status string `json:"status" query:"status"`
	Node   string `json:"node" query:"node"`
}

type idStructRequest struct {
	ID string `json:"id" path:"id"`
}

func TestServer_ArbitraryPathParams(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":0"}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	ep := &talk.Endpoint{
		Name:        "CleanupModel",
		Path:        "/nodes/{nodeName}/models/{modelName}",
		Method:      "DELETE",
		RequestType: reflect.TypeOf(pathParamRequest{}),
		Handler: func(ctx context.Context, req any) (any, error) {
			r, ok := req.(pathParamRequest)
			if !ok {
				return nil, fmt.Errorf("expected pathParamRequest, got %T", req)
			}
			return map[string]string{"nodeName": r.NodeName, "modelName": r.ModelName}, nil
		},
	}
	server.registerEndpoint(ep)

	req := httptest.NewRequest(http.MethodDelete, "/nodes/gpu-001/models/llama-70b", nil)
	w := httptest.NewRecorder()
	server.engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var result map[string]string
	json.NewDecoder(w.Body).Decode(&result)

	if result["nodeName"] != "gpu-001" {
		t.Errorf("nodeName = %q, want %q", result["nodeName"], "gpu-001")
	}
	if result["modelName"] != "llama-70b" {
		t.Errorf("modelName = %q, want %q", result["modelName"], "llama-70b")
	}
}

func TestServer_QueryParams(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":0"}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	ep := &talk.Endpoint{
		Name:        "ListTasks",
		Path:        "/tasks",
		Method:      "GET",
		RequestType: reflect.TypeOf(queryParamRequest{}),
		Handler: func(ctx context.Context, req any) (any, error) {
			r, ok := req.(queryParamRequest)
			if !ok {
				return nil, fmt.Errorf("expected queryParamRequest, got %T", req)
			}
			return map[string]string{"status": r.Status, "node": r.Node}, nil
		},
	}
	server.registerEndpoint(ep)

	req := httptest.NewRequest(http.MethodGet, "/tasks?status=running&node=gpu-01", nil)
	w := httptest.NewRecorder()
	server.engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var result map[string]string
	json.NewDecoder(w.Body).Decode(&result)

	if result["status"] != "running" {
		t.Errorf("status = %q, want %q", result["status"], "running")
	}
	if result["node"] != "gpu-01" {
		t.Errorf("node = %q, want %q", result["node"], "gpu-01")
	}
}

func TestServer_StructInstantiationNoBody(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":0"}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	ep := &talk.Endpoint{
		Name:        "GetItem",
		Path:        "/items/{id}",
		Method:      "GET",
		RequestType: reflect.TypeOf(idStructRequest{}),
		Handler: func(ctx context.Context, req any) (any, error) {
			r, ok := req.(idStructRequest)
			if !ok {
				return nil, fmt.Errorf("expected idStructRequest, got %T", req)
			}
			return map[string]string{"id": r.ID}, nil
		},
	}
	server.registerEndpoint(ep)

	req := httptest.NewRequest(http.MethodGet, "/items/abc-123", nil)
	w := httptest.NewRecorder()
	server.engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var result map[string]string
	json.NewDecoder(w.Body).Decode(&result)

	if result["id"] != "abc-123" {
		t.Errorf("id = %q, want %q", result["id"], "abc-123")
	}
}

func TestServer_BackwardCompatStringID(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":0"}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	// No RequestType — backward compatible simple string extraction
	ep := &talk.Endpoint{
		Name:   "GetItem",
		Path:   "/items/{id}",
		Method: "GET",
		Handler: func(ctx context.Context, req any) (any, error) {
			id, _ := req.(string)
			return &testResponse{Message: "found", ID: id}, nil
		},
	}
	server.registerEndpoint(ep)

	req := httptest.NewRequest(http.MethodGet, "/items/456", nil)
	w := httptest.NewRecorder()
	server.engine.ServeHTTP(w, req)

	var result testResponse
	json.NewDecoder(w.Body).Decode(&result)

	if result.ID != "456" {
		t.Errorf("ID = %q, want %q", result.ID, "456")
	}
}

func TestServer_EndpointMiddleware(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":0"}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	mwCalled := false
	ep := &talk.Endpoint{
		Name:   "Test",
		Path:   "/test-mw",
		Method: "GET",
		Handler: func(ctx context.Context, req any) (any, error) {
			return &testResponse{Message: "original"}, nil
		},
		Middleware: []talk.MiddlewareFunc{
			func(next talk.EndpointFunc) talk.EndpointFunc {
				return func(ctx context.Context, req any) (any, error) {
					mwCalled = true
					resp, err := next(ctx, req)
					if r, ok := resp.(*testResponse); ok {
						r.Message = "modified-by-middleware"
					}
					return resp, err
				}
			},
		},
	}
	server.registerEndpoint(ep)

	req := httptest.NewRequest(http.MethodGet, "/test-mw", nil)
	w := httptest.NewRecorder()
	server.engine.ServeHTTP(w, req)

	var result testResponse
	json.NewDecoder(w.Body).Decode(&result)

	if !mwCalled {
		t.Error("middleware was not called")
	}
	if result.Message != "modified-by-middleware" {
		t.Errorf("Message = %q, want %q", result.Message, "modified-by-middleware")
	}
}

// jsonOnlyQueryRequest has only json tags (no query tags) to test fallback.
type jsonOnlyQueryRequest struct {
	Page int    `json:"page"`
	Size int    `json:"size"`
	Name string `json:"name,omitempty"`
}

func TestServer_QueryParamsJSONFallback(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":0"}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	router := gin.New()
	server.engine = router

	ep := &talk.Endpoint{
		Name:        "ListItems",
		Path:        "/items",
		Method:      "GET",
		RequestType: reflect.TypeOf(jsonOnlyQueryRequest{}),
		Handler: func(ctx context.Context, req any) (any, error) {
			r, ok := req.(jsonOnlyQueryRequest)
			if !ok {
				return nil, fmt.Errorf("expected jsonOnlyQueryRequest, got %T", req)
			}
			return map[string]any{"page": r.Page, "size": r.Size, "name": r.Name}, nil
		},
	}
	server.registerEndpoint(ep)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/items?page=2&size=20&name=hello", nil)
	router.ServeHTTP(w, req)

	var result map[string]any
	json.NewDecoder(w.Body).Decode(&result)

	if result["page"] != float64(2) {
		t.Errorf("page = %v, want 2", result["page"])
	}
	if result["size"] != float64(20) {
		t.Errorf("size = %v, want 20", result["size"])
	}
	if result["name"] != "hello" {
		t.Errorf("name = %v, want %q", result["name"], "hello")
	}
}

func TestServer_PostBodyNotOverriddenByQueryJSONFallback(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":0"}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	router := gin.New()
	server.engine = router

	ep := &talk.Endpoint{
		Name:        "CreateItem",
		Path:        "/items",
		Method:      "POST",
		RequestType: reflect.TypeOf(jsonOnlyQueryRequest{}),
		Handler: func(ctx context.Context, req any) (any, error) {
			r, ok := req.(jsonOnlyQueryRequest)
			if !ok {
				return nil, fmt.Errorf("expected jsonOnlyQueryRequest, got %T", req)
			}
			return map[string]any{"page": r.Page, "size": r.Size, "name": r.Name}, nil
		},
	}
	server.registerEndpoint(ep)

	w := httptest.NewRecorder()
	// POST with body AND query params that share the same json tag names.
	// Body values must NOT be overridden by query params via json fallback.
	body := strings.NewReader(`{"page":5,"size":50,"name":"from-body"}`)
	req, _ := http.NewRequest("POST", "/items?page=99&size=99&name=from-query", body)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var result map[string]any
	json.NewDecoder(w.Body).Decode(&result)

	// Body values should win — query params must NOT override via json fallback
	if result["page"] != float64(5) {
		t.Errorf("page = %v, want 5 (from body, not query)", result["page"])
	}
	if result["size"] != float64(50) {
		t.Errorf("size = %v, want 50 (from body, not query)", result["size"])
	}
	if result["name"] != "from-body" {
		t.Errorf("name = %v, want %q (from body, not query)", result["name"], "from-body")
	}
}
