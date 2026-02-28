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
