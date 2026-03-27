package std

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

	"go.zoe.im/x"
	"go.zoe.im/x/talk"
	"go.zoe.im/x/talk/codec"
)

type testResponse struct {
	Message string `json:"message"`
	ID      string `json:"id,omitempty"`
}

func TestServer_ServeAndShutdown(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":0"}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	if server.String() != "http/std" {
		t.Errorf("String() = %q, want %q", server.String(), "http/std")
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
			Name:   "E",
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

	ts := httptest.NewServer(server.mux)
	defer ts.Close()

	t.Run("GET /items/{id}", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/items/123")
		if err != nil {
			t.Fatalf("GET failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}

		var result testResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("decode failed: %v", err)
		}

		if result.ID != "123" {
			t.Errorf("ID = %q, want %q", result.ID, "123")
		}
	})

	t.Run("POST /items", func(t *testing.T) {
		resp, err := http.Post(ts.URL+"/items", "application/json", nil)
		if err != nil {
			t.Fatalf("POST failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("GET /error returns error", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/error")
		if err != nil {
			t.Fatalf("GET failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}
	})
}

func TestClient_Invoke(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&testResponse{Message: "success"})
	}))
	defer ts.Close()

	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(fmt.Sprintf(`{"addr": %q}`, ts.URL)),
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	if client.String() != "http/std/client" {
		t.Errorf("String() = %q, want %q", client.String(), "http/std/client")
	}

	var resp testResponse
	err = client.Invoke(context.Background(), "/test", nil, &resp)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if resp.Message != "success" {
		t.Errorf("Message = %q, want %q", resp.Message, "success")
	}
}

func TestClient_InvokeError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&talk.Error{
			Code:    talk.InvalidArgument,
			Message: "bad request",
		})
	}))
	defer ts.Close()

	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(fmt.Sprintf(`{"addr": %q}`, ts.URL)),
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	var resp testResponse
	err = client.Invoke(context.Background(), "/test", nil, &resp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	talkErr, ok := err.(*talk.Error)
	if !ok {
		t.Fatalf("expected talk.Error, got %T", err)
	}

	if talkErr.Code != talk.InvalidArgument {
		t.Errorf("Code = %v, want %v", talkErr.Code, talk.InvalidArgument)
	}
}

func TestClient_InvokeStream(t *testing.T) {
	events := []testResponse{
		{Message: "event1"},
		{Message: "event2"},
		{Message: "event3"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		for _, event := range events {
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}))
	defer ts.Close()

	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(fmt.Sprintf(`{"addr": %q}`, ts.URL)),
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	stream, err := client.InvokeStream(context.Background(), "/stream", nil)
	if err != nil {
		t.Fatalf("InvokeStream failed: %v", err)
	}
	defer stream.Close()

	for i, expected := range events {
		var msg testResponse
		if err := stream.Recv(&msg); err != nil {
			t.Fatalf("Recv[%d] failed: %v", i, err)
		}
		if msg.Message != expected.Message {
			t.Errorf("Recv[%d].Message = %q, want %q", i, msg.Message, expected.Message)
		}
	}
}

func TestCodecIntegration(t *testing.T) {
	c := codec.MustGet("json")
	if c.Name() != "json" {
		t.Errorf("Name() = %q, want %q", c.Name(), "json")
	}
	if c.ContentType() != "application/json" {
		t.Errorf("ContentType() = %q, want %q", c.ContentType(), "application/json")
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

	ts := httptest.NewServer(server.mux)
	defer ts.Close()

	req, _ := http.NewRequest("DELETE", ts.URL+"/nodes/gpu-001/models/llama-70b", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)

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

	ts := httptest.NewServer(server.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/tasks?status=running&node=gpu-01")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)

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

	ts := httptest.NewServer(server.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/items/abc-123")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)

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

	ts := httptest.NewServer(server.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/items/456")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var result testResponse
	json.NewDecoder(resp.Body).Decode(&result)

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
		Path:   "/test",
		Method: "GET",
		Handler: func(ctx context.Context, req any) (any, error) {
			return &testResponse{Message: "original"}, nil
		},
		Middleware: []talk.MiddlewareFunc{
			func(next talk.EndpointFunc) talk.EndpointFunc {
				return func(ctx context.Context, req any) (any, error) {
					mwCalled = true
					resp, err := next(ctx, req)
					// Modify response
					if r, ok := resp.(*testResponse); ok {
						r.Message = "modified-by-middleware"
					}
					return resp, err
				}
			},
		},
	}
	server.registerEndpoint(ep)

	ts := httptest.NewServer(server.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/test")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var result testResponse
	json.NewDecoder(resp.Body).Decode(&result)

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

	ts := httptest.NewServer(server.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/items?page=2&size=20&name=hello")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)

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

	ts := httptest.NewServer(server.mux)
	defer ts.Close()

	// POST with body AND query params that share the same json tag names.
	// Body values must NOT be overridden by query params via json fallback.
	body := strings.NewReader(`{"page":5,"size":50,"name":"from-body"}`)
	req, _ := http.NewRequest("POST", ts.URL+"/items?page=99&size=99&name=from-query", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)

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

func TestServer_WithServeMuxStillStartsServer(t *testing.T) {
	mux := http.NewServeMux()
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":0"}`),
	}

	server, err := NewServer(cfg, WithServeMux(mux))
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	endpoints := []*talk.Endpoint{
		{
			Name:   "Hello",
			Path:   "/hello",
			Method: "GET",
			Handler: func(ctx context.Context, req any) (any, error) {
				return &testResponse{Message: "hello from external mux"}, nil
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Serve should NOT block forever; it should start the HTTP server
	// and return when context is cancelled.
	err = server.Serve(ctx, endpoints)
	if err != nil && err != context.DeadlineExceeded {
		t.Errorf("Serve returned unexpected error: %v", err)
	}

	// Verify externalMux is true but server was still created
	if !server.externalMux {
		t.Error("expected externalMux to be true")
	}
	if server.server == nil {
		t.Error("expected server to be created even with externalMux")
	}
}

func TestServer_WithHTTPServerDoesNotStart(t *testing.T) {
	httpServer := &http.Server{}
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":0"}`),
	}

	server, err := NewServer(cfg, WithHTTPServer(httpServer))
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	endpoints := []*talk.Endpoint{
		{
			Name:   "Hello",
			Path:   "/hello",
			Method: "GET",
			Handler: func(ctx context.Context, req any) (any, error) {
				return &testResponse{Message: "hello"}, nil
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// With externalServer, Serve should block on ctx.Done() without starting
	err = server.Serve(ctx, endpoints)
	if err != nil {
		t.Errorf("Serve returned unexpected error: %v", err)
	}
}

func TestServer_TransportAccessor(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":0"}`),
	}

	transport, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	srv := talk.NewServer(transport)
	got := srv.Transport()
	if got != transport {
		t.Errorf("Transport() returned different transport")
	}
}

func TestDeriveClientPath(t *testing.T) {
	tests := []struct {
		name       string
		wantMethod string
		wantPath   string
	}{
		{"GetUser", "GET", "/user/{id}"},
		{"ListUsers", "GET", "/users"},
		{"CreateUser", "POST", "/user"},
		{"UpdateUser", "PUT", "/user/{id}"},
		{"DeleteUser", "DELETE", "/user/{id}"},
		{"WatchUsers", "GET", "/users/watch"},
		{"DoSomething", "POST", "/do-something"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, path := deriveClientPath(tt.name)
			if method != tt.wantMethod {
				t.Errorf("method = %q, want %q", method, tt.wantMethod)
			}
			if path != tt.wantPath {
				t.Errorf("path = %q, want %q", path, tt.wantPath)
			}
		})
	}
}

func TestClient_InvokeWithMethodName(t *testing.T) {
	var capturedMethod, capturedPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&testResponse{Message: "ok"})
	}))
	defer ts.Close()

	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(fmt.Sprintf(`{"addr": %q}`, ts.URL)),
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	var resp testResponse
	err = client.Invoke(context.Background(), "CreateUser", map[string]string{"name": "alice"}, &resp)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if capturedMethod != "POST" {
		t.Errorf("method = %q, want POST", capturedMethod)
	}
	if capturedPath != "/user" {
		t.Errorf("path = %q, want /user", capturedPath)
	}
}

func TestClient_InvokeDirectPath(t *testing.T) {
	var capturedMethod, capturedPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&testResponse{Message: "ok"})
	}))
	defer ts.Close()

	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(fmt.Sprintf(`{"addr": %q}`, ts.URL)),
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	var resp testResponse
	err = client.Invoke(context.Background(), "/custom/path", nil, &resp)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	// Direct paths should use POST
	if capturedMethod != "POST" {
		t.Errorf("method = %q, want POST", capturedMethod)
	}
	if capturedPath != "/custom/path" {
		t.Errorf("path = %q, want /custom/path", capturedPath)
	}
}

func TestClient_InvokeWithPathPrefix(t *testing.T) {
	var capturedMethod, capturedPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&testResponse{Message: "ok"})
	}))
	defer ts.Close()

	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(fmt.Sprintf(`{"addr": %q}`, ts.URL)),
	}

	client, err := NewClient(cfg, WithClientPathPrefix("/api/v1"))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	var resp testResponse
	err = client.Invoke(context.Background(), "ListUsers", nil, &resp)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if capturedMethod != "GET" {
		t.Errorf("method = %q, want GET", capturedMethod)
	}
	if capturedPath != "/api/v1/users" {
		t.Errorf("path = %q, want /api/v1/users", capturedPath)
	}
}

func TestClient_InvokeGetWithID(t *testing.T) {
	var capturedPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&testResponse{Message: "ok"})
	}))
	defer ts.Close()

	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(fmt.Sprintf(`{"addr": %q}`, ts.URL)),
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	var resp testResponse
	err = client.Invoke(context.Background(), "GetUser", "abc-123", &resp)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if capturedPath != "/user/abc-123" {
		t.Errorf("path = %q, want /user/abc-123", capturedPath)
	}
}

func TestExtractResource(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"GetUser", "user"},
		{"ListOrders", "orders"},
		{"CreateItem", "item"},
		{"UpdateTask", "task"},
		{"DeleteEntry", "entry"},
		{"WatchEvents", "events"},
		{"DoSomething", "dosomething"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractResource(tt.name)
			if got != tt.want {
				t.Errorf("extractResource(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"DoSomething", "do-something"},
		{"RunTask", "run-task"},
		{"hello", "hello"},
		{"HTMLParser", "h-t-m-l-parser"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toKebabCase(tt.input)
			if got != tt.want {
				t.Errorf("toKebabCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// --- Additional parameter parsing test types ---

type combinedRequest struct {
	NodeName string `json:"nodeName" path:"nodeName"`
	Status   string `json:"status" query:"status"`
	Page     int    `json:"page" query:"page"`
}

type intIDRequest struct {
	ID int64 `json:"id" path:"id"`
}

type boolQueryRequest struct {
	Verbose bool   `json:"verbose" query:"verbose"`
	Name    string `json:"name" query:"name"`
}

type emptyQueryRequest struct {
	Status string `json:"status" query:"status"`
	Page   int    `json:"page" query:"page"`
}

type updateRequest struct {
	ID   string `json:"id" path:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type threeParamRequest struct {
	Org    string `json:"org" path:"org"`
	Team   string `json:"team" path:"team"`
	Member string `json:"member" path:"member"`
}

type postWithQueryRequest struct {
	Filter string `json:"filter" query:"filter"`
	Body   string `json:"body"`
}

type uintRequest struct {
	Count uint32 `json:"count" path:"count"`
}

type floatQueryRequest struct {
	MinScore float64 `json:"minScore" query:"minScore"`
}

type ssePathRequest struct {
	RoomID string `json:"roomId" path:"roomId"`
}

func TestServer_PathAndQueryCombined(t *testing.T) {
	cfg := x.TypedLazyConfig{Config: json.RawMessage(`{"addr": ":0"}`)}
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	ep := &talk.Endpoint{
		Name:        "GetNodeTasks",
		Path:        "/nodes/{nodeName}/tasks",
		Method:      "GET",
		RequestType: reflect.TypeOf(combinedRequest{}),
		Handler: func(ctx context.Context, req any) (any, error) {
			r := req.(combinedRequest)
			return map[string]any{"nodeName": r.NodeName, "status": r.Status, "page": r.Page}, nil
		},
	}
	server.registerEndpoint(ep)

	ts := httptest.NewServer(server.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/nodes/worker-1/tasks?status=running&page=2")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)

	if result["nodeName"] != "worker-1" {
		t.Errorf("nodeName = %v, want %q", result["nodeName"], "worker-1")
	}
	if result["status"] != "running" {
		t.Errorf("status = %v, want %q", result["status"], "running")
	}
	if result["page"] != float64(2) {
		t.Errorf("page = %v, want 2", result["page"])
	}
}

func TestServer_IntPathParam(t *testing.T) {
	cfg := x.TypedLazyConfig{Config: json.RawMessage(`{"addr": ":0"}`)}
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	ep := &talk.Endpoint{
		Name:        "GetItem",
		Path:        "/items/{id}",
		Method:      "GET",
		RequestType: reflect.TypeOf(intIDRequest{}),
		Handler: func(ctx context.Context, req any) (any, error) {
			r := req.(intIDRequest)
			return map[string]int64{"id": r.ID}, nil
		},
	}
	server.registerEndpoint(ep)

	ts := httptest.NewServer(server.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/items/42")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)

	if result["id"] != float64(42) {
		t.Errorf("id = %v, want 42", result["id"])
	}
}

func TestServer_BoolQueryParam(t *testing.T) {
	cfg := x.TypedLazyConfig{Config: json.RawMessage(`{"addr": ":0"}`)}
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	ep := &talk.Endpoint{
		Name:        "Search",
		Path:        "/search",
		Method:      "GET",
		RequestType: reflect.TypeOf(boolQueryRequest{}),
		Handler: func(ctx context.Context, req any) (any, error) {
			r := req.(boolQueryRequest)
			return map[string]any{"verbose": r.Verbose, "name": r.Name}, nil
		},
	}
	server.registerEndpoint(ep)

	ts := httptest.NewServer(server.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/search?verbose=true&name=test")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)

	if result["verbose"] != true {
		t.Errorf("verbose = %v, want true", result["verbose"])
	}
	if result["name"] != "test" {
		t.Errorf("name = %v, want %q", result["name"], "test")
	}
}

func TestServer_EmptyQueryParams(t *testing.T) {
	cfg := x.TypedLazyConfig{Config: json.RawMessage(`{"addr": ":0"}`)}
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	ep := &talk.Endpoint{
		Name:        "ListItems",
		Path:        "/items",
		Method:      "GET",
		RequestType: reflect.TypeOf(emptyQueryRequest{}),
		Handler: func(ctx context.Context, req any) (any, error) {
			r := req.(emptyQueryRequest)
			return map[string]any{"status": r.Status, "page": r.Page}, nil
		},
	}
	server.registerEndpoint(ep)

	ts := httptest.NewServer(server.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/items")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)

	if result["status"] != "" {
		t.Errorf("status = %v, want empty string", result["status"])
	}
	if result["page"] != float64(0) {
		t.Errorf("page = %v, want 0", result["page"])
	}
}

func TestServer_PostBodyWithPathParams(t *testing.T) {
	cfg := x.TypedLazyConfig{Config: json.RawMessage(`{"addr": ":0"}`)}
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	ep := &talk.Endpoint{
		Name:        "UpdateUser",
		Path:        "/users/{id}",
		Method:      "PUT",
		RequestType: reflect.TypeOf(updateRequest{}),
		Handler: func(ctx context.Context, req any) (any, error) {
			r := req.(updateRequest)
			return map[string]any{"id": r.ID, "name": r.Name, "age": r.Age}, nil
		},
	}
	server.registerEndpoint(ep)

	ts := httptest.NewServer(server.mux)
	defer ts.Close()

	body := strings.NewReader(`{"name":"bob","age":30}`)
	req, _ := http.NewRequest("PUT", ts.URL+"/users/usr-123", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)

	if result["id"] != "usr-123" {
		t.Errorf("id = %v, want %q", result["id"], "usr-123")
	}
	if result["name"] != "bob" {
		t.Errorf("name = %v, want %q", result["name"], "bob")
	}
	if result["age"] != float64(30) {
		t.Errorf("age = %v, want 30", result["age"])
	}
}

func TestServer_MultiplePathParams(t *testing.T) {
	cfg := x.TypedLazyConfig{Config: json.RawMessage(`{"addr": ":0"}`)}
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	ep := &talk.Endpoint{
		Name:        "GetMember",
		Path:        "/orgs/{org}/teams/{team}/members/{member}",
		Method:      "GET",
		RequestType: reflect.TypeOf(threeParamRequest{}),
		Handler: func(ctx context.Context, req any) (any, error) {
			r := req.(threeParamRequest)
			return map[string]string{"org": r.Org, "team": r.Team, "member": r.Member}, nil
		},
	}
	server.registerEndpoint(ep)

	ts := httptest.NewServer(server.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/orgs/acme/teams/backend/members/alice")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)

	if result["org"] != "acme" {
		t.Errorf("org = %q, want %q", result["org"], "acme")
	}
	if result["team"] != "backend" {
		t.Errorf("team = %q, want %q", result["team"], "backend")
	}
	if result["member"] != "alice" {
		t.Errorf("member = %q, want %q", result["member"], "alice")
	}
}

func TestServer_QueryParamWithExplicitQueryTag(t *testing.T) {
	cfg := x.TypedLazyConfig{Config: json.RawMessage(`{"addr": ":0"}`)}
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	ep := &talk.Endpoint{
		Name:        "SearchPosts",
		Path:        "/search",
		Method:      "POST",
		RequestType: reflect.TypeOf(postWithQueryRequest{}),
		Handler: func(ctx context.Context, req any) (any, error) {
			r := req.(postWithQueryRequest)
			return map[string]string{"filter": r.Filter, "body": r.Body}, nil
		},
	}
	server.registerEndpoint(ep)

	ts := httptest.NewServer(server.mux)
	defer ts.Close()

	body := strings.NewReader(`{"body":"content"}`)
	req, _ := http.NewRequest("POST", ts.URL+"/search?filter=active", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)

	if result["filter"] != "active" {
		t.Errorf("filter = %q, want %q", result["filter"], "active")
	}
	if result["body"] != "content" {
		t.Errorf("body = %q, want %q", result["body"], "content")
	}
}

func TestServer_UintPathParam(t *testing.T) {
	cfg := x.TypedLazyConfig{Config: json.RawMessage(`{"addr": ":0"}`)}
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	ep := &talk.Endpoint{
		Name:        "GetStats",
		Path:        "/stats/{count}",
		Method:      "GET",
		RequestType: reflect.TypeOf(uintRequest{}),
		Handler: func(ctx context.Context, req any) (any, error) {
			r := req.(uintRequest)
			return map[string]uint32{"count": r.Count}, nil
		},
	}
	server.registerEndpoint(ep)

	ts := httptest.NewServer(server.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/stats/999")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)

	if result["count"] != float64(999) {
		t.Errorf("count = %v, want 999", result["count"])
	}
}

func TestServer_FloatQueryParam(t *testing.T) {
	cfg := x.TypedLazyConfig{Config: json.RawMessage(`{"addr": ":0"}`)}
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	ep := &talk.Endpoint{
		Name:        "GetResults",
		Path:        "/results",
		Method:      "GET",
		RequestType: reflect.TypeOf(floatQueryRequest{}),
		Handler: func(ctx context.Context, req any) (any, error) {
			r := req.(floatQueryRequest)
			return map[string]float64{"minScore": r.MinScore}, nil
		},
	}
	server.registerEndpoint(ep)

	ts := httptest.NewServer(server.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/results?minScore=0.75")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)

	if result["minScore"] != 0.75 {
		t.Errorf("minScore = %v, want 0.75", result["minScore"])
	}
}

func TestServer_SSEStreamingWithPathParams(t *testing.T) {
	cfg := x.TypedLazyConfig{Config: json.RawMessage(`{"addr": ":0"}`)}
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	var capturedRoomID string
	ep := &talk.Endpoint{
		Name:        "WatchRoom",
		Path:        "/rooms/{roomId}/events",
		Method:      "GET",
		RequestType: reflect.TypeOf(ssePathRequest{}),
		StreamMode:  talk.StreamServerSide,
		StreamHandler: func(ctx context.Context, req any, stream talk.Stream) error {
			r := req.(ssePathRequest)
			capturedRoomID = r.RoomID
			return stream.Send(map[string]string{"roomId": r.RoomID})
		},
	}
	server.registerEndpoint(ep)

	ts := httptest.NewServer(server.mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/rooms/room-42/events")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if capturedRoomID != "room-42" {
		t.Errorf("roomId = %q, want %q", capturedRoomID, "room-42")
	}
}
