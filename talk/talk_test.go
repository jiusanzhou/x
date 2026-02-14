package talk

import (
	"context"
	"io"
	"testing"
)

// mockTransport implements Transport for testing
type mockTransport struct {
	serveFunc        func(ctx context.Context, endpoints []*Endpoint) error
	shutdownFunc     func(ctx context.Context) error
	invokeFunc       func(ctx context.Context, endpoint string, req any, resp any) error
	invokeStreamFunc func(ctx context.Context, endpoint string, req any) (Stream, error)
}

func (m *mockTransport) String() string { return "mock" }

func (m *mockTransport) Serve(ctx context.Context, endpoints []*Endpoint) error {
	if m.serveFunc != nil {
		return m.serveFunc(ctx, endpoints)
	}
	<-ctx.Done()
	return ctx.Err()
}

func (m *mockTransport) Shutdown(ctx context.Context) error {
	if m.shutdownFunc != nil {
		return m.shutdownFunc(ctx)
	}
	return nil
}

func (m *mockTransport) Invoke(ctx context.Context, endpoint string, req any, resp any) error {
	if m.invokeFunc != nil {
		return m.invokeFunc(ctx, endpoint, req, resp)
	}
	return nil
}

func (m *mockTransport) InvokeStream(ctx context.Context, endpoint string, req any) (Stream, error) {
	if m.invokeStreamFunc != nil {
		return m.invokeStreamFunc(ctx, endpoint, req)
	}
	return nil, NewError(Unimplemented, "not implemented")
}

func (m *mockTransport) Close() error { return nil }

// mockExtractor implements Extractor for testing
type mockExtractor struct {
	endpoints []*Endpoint
	err       error
}

func (m *mockExtractor) Extract(service any) ([]*Endpoint, error) {
	return m.endpoints, m.err
}

func TestNewServer(t *testing.T) {
	transport := &mockTransport{}
	server := NewServer(transport)

	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	if server.transport != transport {
		t.Error("transport not set correctly")
	}

	if server.codec == nil {
		t.Error("default codec should be set")
	}
}

func TestNewServer_WithOptions(t *testing.T) {
	transport := &mockTransport{}
	extractor := &mockExtractor{}

	server := NewServer(transport,
		WithExtractor(extractor),
		WithServerCodecName("json"),
	)

	if server.extractor != extractor {
		t.Error("extractor not set correctly")
	}

	if server.codec == nil {
		t.Error("codec should be set")
	}
}

func TestServer_RegisterEndpoints(t *testing.T) {
	server := NewServer(&mockTransport{})

	ep1 := &Endpoint{Name: "GetUser", Path: "/users/{id}", Method: "GET"}
	ep2 := &Endpoint{Name: "CreateUser", Path: "/users", Method: "POST"}

	server.RegisterEndpoints(ep1, ep2)

	endpoints := server.Endpoints()
	if len(endpoints) != 2 {
		t.Errorf("expected 2 endpoints, got %d", len(endpoints))
	}
}

func TestServer_Register_WithExtractor(t *testing.T) {
	extractor := &mockExtractor{
		endpoints: []*Endpoint{
			{Name: "GetUser", Path: "/users/{id}", Method: "GET"},
		},
	}

	server := NewServer(&mockTransport{}, WithExtractor(extractor))

	err := server.Register(struct{}{})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if len(server.Endpoints()) != 1 {
		t.Errorf("expected 1 endpoint, got %d", len(server.Endpoints()))
	}
}

func TestServer_Register_WithoutExtractor(t *testing.T) {
	server := NewServer(&mockTransport{})

	err := server.Register(struct{}{})
	if err == nil {
		t.Error("expected error when no extractor configured")
	}

	talkErr, ok := err.(*Error)
	if !ok {
		t.Errorf("expected *Error, got %T", err)
	}

	if talkErr.Code != FailedPrecondition {
		t.Errorf("expected FailedPrecondition, got %v", talkErr.Code)
	}
}

func TestServer_Serve(t *testing.T) {
	served := false
	transport := &mockTransport{
		serveFunc: func(ctx context.Context, endpoints []*Endpoint) error {
			served = true
			return nil
		},
	}

	server := NewServer(transport)
	server.RegisterEndpoints(&Endpoint{Name: "Test"})

	err := server.Serve(context.Background())
	if err != nil {
		t.Errorf("Serve failed: %v", err)
	}

	if !served {
		t.Error("transport.Serve was not called")
	}
}

func TestServer_Shutdown(t *testing.T) {
	shutdownCalled := false
	transport := &mockTransport{
		shutdownFunc: func(ctx context.Context) error {
			shutdownCalled = true
			return nil
		},
	}

	server := NewServer(transport)

	err := server.Shutdown(context.Background())
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	if !shutdownCalled {
		t.Error("transport.Shutdown was not called")
	}
}

func TestNewClient(t *testing.T) {
	transport := &mockTransport{}
	client := NewClient(transport)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.transport != transport {
		t.Error("transport not set correctly")
	}

	if client.codec == nil {
		t.Error("default codec should be set")
	}
}

func TestClient_Call(t *testing.T) {
	type testReq struct{ ID string }
	type testResp struct{ Name string }

	transport := &mockTransport{
		invokeFunc: func(ctx context.Context, endpoint string, req any, resp any) error {
			if endpoint != "/users/123" {
				t.Errorf("endpoint = %q, want %q", endpoint, "/users/123")
			}
			if r, ok := resp.(*testResp); ok {
				r.Name = "Alice"
			}
			return nil
		},
	}

	client := NewClient(transport)

	var resp testResp
	err := client.Call(context.Background(), "/users/123", &testReq{ID: "123"}, &resp)
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	if resp.Name != "Alice" {
		t.Errorf("resp.Name = %q, want %q", resp.Name, "Alice")
	}
}

func TestClient_Stream(t *testing.T) {
	mockStream := NewChanStream[string](context.Background(), 10)

	transport := &mockTransport{
		invokeStreamFunc: func(ctx context.Context, endpoint string, req any) (Stream, error) {
			return mockStream, nil
		},
	}

	client := NewClient(transport)

	stream, err := client.Stream(context.Background(), "/events", nil)
	if err != nil {
		t.Fatalf("Stream failed: %v", err)
	}

	if stream == nil {
		t.Error("expected stream, got nil")
	}
}

func TestClient_Close(t *testing.T) {
	transport := &mockTransport{}
	client := NewClient(transport)

	err := client.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

// Test Endpoint

func TestEndpoint_IsStreaming(t *testing.T) {
	tests := []struct {
		mode     StreamMode
		expected bool
	}{
		{StreamNone, false},
		{StreamClientSide, true},
		{StreamServerSide, true},
		{StreamBidirect, true},
	}

	for _, tt := range tests {
		t.Run(tt.mode.String(), func(t *testing.T) {
			ep := &Endpoint{StreamMode: tt.mode}
			if got := ep.IsStreaming(); got != tt.expected {
				t.Errorf("IsStreaming() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEndpoint_Clone(t *testing.T) {
	original := &Endpoint{
		Name:   "GetUser",
		Path:   "/users/{id}",
		Method: "GET",
		Metadata: map[string]any{
			"auth": true,
		},
	}

	clone := original.Clone()

	if clone == original {
		t.Error("Clone should return a new pointer")
	}

	if clone.Name != original.Name {
		t.Errorf("Name = %q, want %q", clone.Name, original.Name)
	}

	// Modifying clone's metadata should not affect original
	clone.Metadata["auth"] = false
	if original.Metadata["auth"] != true {
		t.Error("modifying clone should not affect original")
	}
}

func TestEndpoint_CloneWithNilMetadata(t *testing.T) {
	original := &Endpoint{
		Name:   "GetUser",
		Path:   "/users/{id}",
		Method: "GET",
	}

	clone := original.Clone()

	if clone.Metadata != nil {
		t.Error("cloned nil metadata should remain nil")
	}
}

func TestNewEndpoint(t *testing.T) {
	handler := func(ctx context.Context, req any) (any, error) {
		return "response", nil
	}

	ep := NewEndpoint("Test", handler,
		WithPath("/test"),
		WithMethod("GET"),
		WithMetadata("version", "v1"),
	)

	if ep.Name != "Test" {
		t.Errorf("Name = %q, want %q", ep.Name, "Test")
	}

	if ep.Path != "/test" {
		t.Errorf("Path = %q, want %q", ep.Path, "/test")
	}

	if ep.Method != "GET" {
		t.Errorf("Method = %q, want %q", ep.Method, "GET")
	}

	if ep.Metadata["version"] != "v1" {
		t.Errorf("Metadata[version] = %v, want %q", ep.Metadata["version"], "v1")
	}
}

func TestNewEndpoint_DefaultMethod(t *testing.T) {
	handler := func(ctx context.Context, req any) (any, error) {
		return nil, nil
	}

	ep := NewEndpoint("Test", handler)

	if ep.Method != "POST" {
		t.Errorf("default Method = %q, want %q", ep.Method, "POST")
	}
}

func TestNewStreamEndpoint(t *testing.T) {
	handler := func(ctx context.Context, req any, stream Stream) error {
		return stream.Send("event")
	}

	ep := NewStreamEndpoint("WatchEvents", handler, StreamServerSide,
		WithPath("/events/watch"),
	)

	if ep.Name != "WatchEvents" {
		t.Errorf("Name = %q, want %q", ep.Name, "WatchEvents")
	}

	if ep.StreamMode != StreamServerSide {
		t.Errorf("StreamMode = %v, want %v", ep.StreamMode, StreamServerSide)
	}

	if ep.Method != "GET" {
		t.Errorf("Method = %q, want %q", ep.Method, "GET")
	}
}

func TestStreamMode_String(t *testing.T) {
	tests := []struct {
		mode     StreamMode
		expected string
	}{
		{StreamNone, "none"},
		{StreamClientSide, "client"},
		{StreamServerSide, "server"},
		{StreamBidirect, "bidirectional"},
		{StreamMode(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// Test ChanStream

func TestChanStream_SendRecv(t *testing.T) {
	stream := NewChanStream[string](context.Background(), 1)

	// Wire up send to recv for testing
	go func() {
		stream.recvCh <- "hello"
	}()

	var msg string
	err := stream.Recv(&msg)
	if err != nil {
		t.Fatalf("Recv failed: %v", err)
	}

	if msg != "hello" {
		t.Errorf("msg = %q, want %q", msg, "hello")
	}
}

func TestChanStream_Send(t *testing.T) {
	stream := NewChanStream[string](context.Background(), 1)

	err := stream.Send("hello")
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	select {
	case msg := <-stream.sendCh:
		if msg != "hello" {
			t.Errorf("msg = %q, want %q", msg, "hello")
		}
	default:
		t.Error("message not in send channel")
	}
}

func TestChanStream_Close(t *testing.T) {
	stream := NewChanStream[string](context.Background(), 1)

	err := stream.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Send after close should fail
	err = stream.Send("hello")
	if err != io.ErrClosedPipe {
		t.Errorf("Send after close should return ErrClosedPipe, got %v", err)
	}

	// Close again should not error
	err = stream.Close()
	if err != nil {
		t.Errorf("second Close should not error: %v", err)
	}
}

func TestChanStream_RecvAfterClose(t *testing.T) {
	stream := NewChanStream[string](context.Background(), 1)
	stream.Close()

	var msg string
	err := stream.Recv(&msg)
	if err != io.EOF {
		t.Errorf("Recv after close should return EOF, got %v", err)
	}
}

func TestChanStream_Context(t *testing.T) {
	ctx := context.Background()
	stream := NewChanStream[string](ctx, 1)

	if stream.Context() == nil {
		t.Error("Context() should not return nil")
	}
}

func TestChanStream_SendChan(t *testing.T) {
	stream := NewChanStream[string](context.Background(), 1)

	sendCh := stream.SendChan()
	if sendCh == nil {
		t.Error("SendChan() should not return nil")
	}
}

func TestChanStream_RecvChan(t *testing.T) {
	stream := NewChanStream[string](context.Background(), 1)

	recvCh := stream.RecvChan()
	if recvCh == nil {
		t.Error("RecvChan() should not return nil")
	}
}

func TestChanStream_SetRecvChan(t *testing.T) {
	stream := NewChanStream[string](context.Background(), 1)
	newCh := make(chan string, 1)
	newCh <- "from new channel"

	stream.SetRecvChan(newCh)

	var msg string
	err := stream.Recv(&msg)
	if err != nil {
		t.Fatalf("Recv failed: %v", err)
	}

	if msg != "from new channel" {
		t.Errorf("msg = %q, want %q", msg, "from new channel")
	}
}

func TestChanStream_SendInvalidType(t *testing.T) {
	stream := NewChanStream[string](context.Background(), 1)

	err := stream.Send(123) // int instead of string
	if err == nil {
		t.Error("Send with wrong type should fail")
	}

	talkErr, ok := err.(*Error)
	if !ok {
		t.Errorf("expected *Error, got %T", err)
	}

	if talkErr.Code != InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", talkErr.Code)
	}
}

func TestChanStream_RecvInvalidType(t *testing.T) {
	stream := NewChanStream[string](context.Background(), 1)

	go func() {
		stream.recvCh <- "hello"
	}()

	var msg int // wrong type
	err := stream.Recv(&msg)
	if err == nil {
		t.Error("Recv with wrong type should fail")
	}
}

func TestChanStream_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	stream := NewChanStream[string](ctx, 0) // unbuffered

	cancel()

	err := stream.Send("hello")
	if err != context.Canceled {
		t.Errorf("Send after cancel should return context.Canceled, got %v", err)
	}
}
