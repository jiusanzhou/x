package grpc

import (
	"encoding/json"
	"testing"

	"go.zoe.im/x"
	"go.zoe.im/x/talk"
	"go.zoe.im/x/talk/codec"
)

func TestNewServer(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":50051"}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	if server.String() != "grpc" {
		t.Errorf("String() = %q, want %q", server.String(), "grpc")
	}
}

func TestNewClient(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": "localhost:50051", "insecure": true}`),
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	if client.String() != "grpc/client" {
		t.Errorf("String() = %q, want %q", client.String(), "grpc/client")
	}
}

func TestServerConfig(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{
			"addr": ":50051",
			"max_recv_msg_size": 4194304,
			"max_send_msg_size": 4194304
		}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	if server.config.MaxRecvMsgSize != 4194304 {
		t.Errorf("MaxRecvMsgSize = %d, want %d", server.config.MaxRecvMsgSize, 4194304)
	}
}

func TestClientConfig(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{
			"addr": "localhost:50051",
			"insecure": true,
			"timeout": "5s",
			"wait_for_ready": true
		}`),
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	if !client.config.WaitForReady {
		t.Error("WaitForReady should be true")
	}
}

func TestSetCodec(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":50051"}`),
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

func TestServerBuildMethods(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Config: json.RawMessage(`{"addr": ":50051"}`),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	endpoints := []*talk.Endpoint{
		{
			Name:       "GetUser",
			Path:       "/users/{id}",
			Method:     "GET",
			StreamMode: talk.StreamNone,
		},
		{
			Name:       "WatchUsers",
			Path:       "/users/watch",
			Method:     "GET",
			StreamMode: talk.StreamServerSide,
		},
	}

	for _, ep := range endpoints {
		server.endpoints[ep.Name] = ep
	}

	unaryMethods := server.buildUnaryMethods()
	if len(unaryMethods) != 1 {
		t.Errorf("expected 1 unary method, got %d", len(unaryMethods))
	}

	streamMethods := server.buildStreamMethods()
	if len(streamMethods) != 1 {
		t.Errorf("expected 1 stream method, got %d", len(streamMethods))
	}
}

func TestErrorConversion(t *testing.T) {
	talkErr := talk.NewError(talk.NotFound, "user not found")
	grpcCode := talkErr.GRPCCode()

	if grpcCode != 5 {
		t.Errorf("GRPCCode() = %d, want 5 (NotFound)", grpcCode)
	}

	converted := ErrorCodeFromGRPC(5)
	if converted != talk.NotFound {
		t.Errorf("ErrorCodeFromGRPC(5) = %v, want NotFound", converted)
	}
}

func TestFactoryRegistration(t *testing.T) {
	cfg := x.TypedLazyConfig{
		Type:   "default",
		Config: json.RawMessage(`{"addr": ":50051"}`),
	}

	server, err := ServerFactory.Create(cfg)
	if err != nil {
		t.Fatalf("ServerFactory.Create failed: %v", err)
	}
	if server == nil {
		t.Error("ServerFactory.Create returned nil")
	}

	clientCfg := x.TypedLazyConfig{
		Type:   "default",
		Config: json.RawMessage(`{"addr": "localhost:50051", "insecure": true}`),
	}

	client, err := ClientFactory.Create(clientCfg)
	if err != nil {
		t.Fatalf("ClientFactory.Create failed: %v", err)
	}
	if client == nil {
		t.Error("ClientFactory.Create returned nil")
	}
}
