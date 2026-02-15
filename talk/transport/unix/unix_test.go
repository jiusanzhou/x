package unix

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.zoe.im/x"
	"go.zoe.im/x/talk"
)

func TestUnixSocket(t *testing.T) {
	socketPath := filepath.Join(os.TempDir(), "talk_test.sock")
	defer os.Remove(socketPath)

	serverCfg := x.TypedLazyConfig{
		Type:   "unix",
		Config: json.RawMessage(`{"path": "` + socketPath + `"}`),
	}

	server, err := NewServer(serverCfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	endpoints := []*talk.Endpoint{
		talk.NewEndpoint("ping", func(ctx context.Context, req any) (any, error) {
			return map[string]string{"status": "pong"}, nil
		}, talk.WithPath("/ping"), talk.WithMethod("POST")),
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
		Type:   "unix",
		Config: json.RawMessage(`{"path": "` + socketPath + `"}`),
	}

	client, err := NewClient(clientCfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	var pingResp map[string]string
	if err := client.Invoke(ctx, "/ping", nil, &pingResp); err != nil {
		t.Fatalf("Invoke /ping failed: %v", err)
	}

	if pingResp["status"] != "pong" {
		t.Errorf("expected status 'pong', got %q", pingResp["status"])
	}

	cancel()
	time.Sleep(100 * time.Millisecond)
}

func TestUnixSocketStreaming(t *testing.T) {
	socketPath := filepath.Join(os.TempDir(), "talk_stream_test.sock")
	defer os.Remove(socketPath)

	serverCfg := x.TypedLazyConfig{
		Type:   "unix",
		Config: json.RawMessage(`{"path": "` + socketPath + `"}`),
	}

	server, err := NewServer(serverCfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	endpoints := []*talk.Endpoint{
		talk.NewStreamEndpoint("events", func(ctx context.Context, req any, stream talk.Stream) error {
			for i := 0; i < 3; i++ {
				if err := stream.Send(map[string]int{"count": i}); err != nil {
					return err
				}
			}
			return nil
		}, talk.StreamServerSide, talk.WithPath("/events"), talk.WithMethod("GET")),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		server.Serve(ctx, endpoints)
	}()

	time.Sleep(100 * time.Millisecond)

	clientCfg := x.TypedLazyConfig{
		Type:   "unix",
		Config: json.RawMessage(`{"path": "` + socketPath + `"}`),
	}

	client, err := NewClient(clientCfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	stream, err := client.InvokeStream(ctx, "/events", nil)
	if err != nil {
		t.Fatalf("InvokeStream failed: %v", err)
	}
	defer stream.Close()

	for i := 0; i < 3; i++ {
		var event map[string]int
		if err := stream.Recv(&event); err != nil {
			t.Fatalf("Recv failed: %v", err)
		}
		if event["count"] != i {
			t.Errorf("expected count %d, got %d", i, event["count"])
		}
	}
}
