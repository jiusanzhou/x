package websocket

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"go.zoe.im/x"
	"go.zoe.im/x/talk"
	"go.zoe.im/x/talk/codec"
)

func TestNewServer(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":8090"}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	if server.String() != "websocket" {
		t.Errorf("String() = %q, want %q", server.String(), "websocket")
	}
}

func TestServerDefaultPath(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":8090"}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	if server.config.Path != "/ws" {
		t.Errorf("default Path = %q, want %q", server.config.Path, "/ws")
	}
}

func TestServerCustomPath(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":8090", "path": "/socket"}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	if server.config.Path != "/socket" {
		t.Errorf("custom Path = %q, want %q", server.config.Path, "/socket")
	}
}

func TestServerConfig(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{
			"addr": ":8090",
			"path": "/ws",
			"read_buffer_size": 4096,
			"write_buffer_size": 4096,
			"check_origin": true,
			"enable_compression": true
		}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	if server.config.ReadBufferSize != 4096 {
		t.Errorf("ReadBufferSize = %d, want %d", server.config.ReadBufferSize, 4096)
	}

	if server.config.WriteBufferSize != 4096 {
		t.Errorf("WriteBufferSize = %d, want %d", server.config.WriteBufferSize, 4096)
	}

	if !server.config.CheckOrigin {
		t.Error("CheckOrigin should be true")
	}

	if !server.config.EnableCompression {
		t.Error("EnableCompression should be true")
	}
}

func TestClientConfig(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{
			"addr": "localhost:8090",
			"path": "/ws",
			"handshake_timeout": "5s"
		}`),
	}

	client := &Client{}
	if err := cfg.Unmarshal(&client.config); err != nil {
		t.Fatalf("Config unmarshal failed: %v", err)
	}

	if client.config.Addr != "localhost:8090" {
		t.Errorf("Addr = %q, want %q", client.config.Addr, "localhost:8090")
	}

	if client.config.Path != "/ws" {
		t.Errorf("Path = %q, want %q", client.config.Path, "/ws")
	}

	expectedTimeout := 5 * time.Second
	if time.Duration(client.config.HandshakeTimeout) != expectedTimeout {
		t.Errorf("HandshakeTimeout = %v, want %v", time.Duration(client.config.HandshakeTimeout), expectedTimeout)
	}
}

func TestSetCodec(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":8090"}`),
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

func TestWithCodecOption(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":8090"}`),
	}

	customCodec := codec.MustGet("json")
	server, err := NewServer(cfg, WithCodec(customCodec))
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	if server.codec != customCodec {
		t.Error("WithCodec option did not set codec correctly")
	}
}

func TestServerNotSupportInvoke(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":8090"}`),
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

func TestServerNotSupportInvokeStream(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":8090"}`),
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

func TestFactoryRegistration(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Type:   "default",
		Config: json.RawMessage(`{"addr": ":8090"}`),
	}

	server, err := ServerFactory.Create(cfg)
	if err != nil {
		t.Fatalf("ServerFactory.Create failed: %v", err)
	}
	if server == nil {
		t.Error("ServerFactory.Create returned nil")
	}
}

func TestMessageTypes(t *testing.T) {
	if TextMessage != 1 {
		t.Errorf("TextMessage = %d, want 1", TextMessage)
	}
	if BinaryMessage != 2 {
		t.Errorf("BinaryMessage = %d, want 2", BinaryMessage)
	}
}

func TestMessage(t *testing.T) {
	msg := Message{
		Type:    "request",
		Payload: []byte(`{"method": "test"}`),
		Error:   "",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Type != "request" {
		t.Errorf("Type = %q, want %q", decoded.Type, "request")
	}
}

func TestWSMessageFormat(t *testing.T) {
	msg := wsMessage{
		ID:     "test-123",
		Method: "GetUser",
		Params: json.RawMessage(`{"id": "user-1"}`),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded wsMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.ID != "test-123" {
		t.Errorf("ID = %q, want %q", decoded.ID, "test-123")
	}

	if decoded.Method != "GetUser" {
		t.Errorf("Method = %q, want %q", decoded.Method, "GetUser")
	}
}

func TestWSResponseFormat(t *testing.T) {
	tests := []struct {
		name     string
		response wsResponse
	}{
		{
			name: "success response",
			response: wsResponse{
				ID:     "test-123",
				Result: map[string]string{"name": "Alice"},
			},
		},
		{
			name: "error response",
			response: wsResponse{
				ID: "test-456",
				Error: &wsError{
					Code:    int(talk.NotFound),
					Message: "user not found",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			var decoded wsResponse
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			if decoded.ID != tt.response.ID {
				t.Errorf("ID = %q, want %q", decoded.ID, tt.response.ID)
			}
		})
	}
}

func TestServerShutdownWithoutServe(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":8090"}`),
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

func TestServerClose(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":8090"}`),
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

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	serverCfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":18090", "path": "/ws"}`),
	}

	server, err := NewServer(serverCfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	endpoints := []*talk.Endpoint{
		{
			Name:       "Echo",
			Path:       "/echo",
			Method:     "POST",
			StreamMode: talk.StreamNone,
			Handler: func(ctx context.Context, req any) (any, error) {
				return req, nil
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverReady := make(chan struct{})
	serverErr := make(chan error, 1)

	go func() {
		close(serverReady)
		if err := server.Serve(ctx, endpoints); err != nil {
			serverErr <- err
		}
	}()

	<-serverReady
	time.Sleep(100 * time.Millisecond)

	clientCfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": "localhost:18090", "path": "/ws"}`),
	}

	client, err := NewClient(clientCfg)
	if err != nil {
		cancel()
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	var result map[string]any
	err = client.Invoke(context.Background(), "Echo", map[string]string{"message": "hello"}, &result)
	if err != nil {
		t.Errorf("Invoke failed: %v", err)
	}

	cancel()
}
