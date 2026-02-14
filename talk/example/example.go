// Package example demonstrates how to use the talk package.
//
// Talk is a transport abstraction layer that allows you to define service methods
// without caring about the underlying connection implementation (HTTP, gRPC, WebSocket).
// The protocol can be switched via configuration.
//
// This file contains example code that demonstrates:
//   - Defining a service interface
//   - Using reflection-based endpoint extraction
//   - Creating servers and clients
//   - Switching between HTTP and WebSocket transports
package example

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go.zoe.im/x"
	"go.zoe.im/x/talk"
	"go.zoe.im/x/talk/extract"
	thttp "go.zoe.im/x/talk/transport/http"
	"go.zoe.im/x/talk/transport/http/std"
	"go.zoe.im/x/talk/transport/websocket"
)

// =============================================================================
// Service Definition
// =============================================================================

// User represents a user entity.
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CreateUserRequest is the request for creating a user.
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UserEvent represents an event about a user.
type UserEvent struct {
	Type   string `json:"type"`
	UserID string `json:"user_id"`
	Data   any    `json:"data,omitempty"`
}

// UserService defines the user service interface.
// Method names are automatically mapped to HTTP methods and paths:
//   - GetUser    -> GET    /user/{id}
//   - CreateUser -> POST   /user
//   - ListUsers  -> GET    /users
//   - DeleteUser -> DELETE /user/{id}
//   - WatchUsers -> GET    /users/watch (SSE streaming)
type UserService interface {
	GetUser(ctx context.Context, id string) (*User, error)
	CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error)
	ListUsers(ctx context.Context) ([]*User, error)
	DeleteUser(ctx context.Context, id string) error
	WatchUsers(ctx context.Context) (<-chan *UserEvent, error)
}

// =============================================================================
// Service Implementation
// =============================================================================

// userServiceImpl implements UserService.
type userServiceImpl struct {
	users map[string]*User
}

// NewUserService creates a new user service.
func NewUserService() *userServiceImpl {
	return &userServiceImpl{
		users: make(map[string]*User),
	}
}

func (s *userServiceImpl) GetUser(ctx context.Context, id string) (*User, error) {
	user, ok := s.users[id]
	if !ok {
		return nil, talk.NewError(talk.NotFound, "user not found: "+id)
	}
	return user, nil
}

func (s *userServiceImpl) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	if req.Name == "" {
		return nil, talk.NewError(talk.InvalidArgument, "name is required")
	}

	id := fmt.Sprintf("user-%d", len(s.users)+1)
	user := &User{
		ID:    id,
		Name:  req.Name,
		Email: req.Email,
	}
	s.users[id] = user
	return user, nil
}

func (s *userServiceImpl) ListUsers(ctx context.Context) ([]*User, error) {
	users := make([]*User, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u)
	}
	return users, nil
}

func (s *userServiceImpl) DeleteUser(ctx context.Context, id string) error {
	if _, ok := s.users[id]; !ok {
		return talk.NewError(talk.NotFound, "user not found: "+id)
	}
	delete(s.users, id)
	return nil
}

func (s *userServiceImpl) WatchUsers(ctx context.Context) (<-chan *UserEvent, error) {
	ch := make(chan *UserEvent)
	go func() {
		defer close(ch)
		for i := 0; i < 5; i++ {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
				ch <- &UserEvent{
					Type:   "heartbeat",
					UserID: "",
					Data:   map[string]any{"count": i + 1},
				}
			}
		}
	}()
	return ch, nil
}

// =============================================================================
// HTTP Server Example
// =============================================================================

// ExampleHTTPServer demonstrates creating an HTTP server.
func ExampleHTTPServer() {
	// Create the service implementation
	userSvc := NewUserService()

	// Extract endpoints using reflection
	extractor := extract.NewReflectExtractor(extract.WithPathPrefix("/api/v1"))
	endpoints, err := extractor.Extract(userSvc)
	if err != nil {
		log.Fatal(err)
	}

	// Create HTTP transport
	cfg := x.TypedLazyConfig{
		Type:   "std",
		Config: json.RawMessage(`{"addr": ":8080"}`),
	}

	transport, err := std.NewServer(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Create talk server
	server := talk.NewServer(transport,
		talk.WithExtractor(extractor),
	)
	server.RegisterEndpoints(endpoints...)

	fmt.Printf("Registered %d endpoints:\n", len(endpoints))
	for _, ep := range endpoints {
		fmt.Printf("  %s %s -> %s\n", ep.Method, ep.Path, ep.Name)
	}

	// Start serving
	ctx := context.Background()
	fmt.Println("Server listening on :8080")
	if err := server.Serve(ctx); err != nil {
		log.Fatal(err)
	}

	// Output:
	// Registered 5 endpoints:
	//   GET /api/v1/user/{id} -> GetUser
	//   POST /api/v1/user -> CreateUser
	//   GET /api/v1/users -> ListUsers
	//   DELETE /api/v1/user/{id} -> DeleteUser
	//   GET /api/v1/users/watch -> WatchUsers
	// Server listening on :8080
}

// =============================================================================
// HTTP Client Example
// =============================================================================

// ExampleHTTPClient demonstrates using an HTTP client.
func ExampleHTTPClient() {
	// Create HTTP client transport
	cfg := x.TypedLazyConfig{
		Type:   "std",
		Config: json.RawMessage(`{"addr": "http://localhost:8080"}`),
	}

	transport, err := std.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Create talk client
	client := talk.NewClient(transport)
	defer client.Close()

	// Make a request
	var user User
	err = client.Call(context.Background(), "/api/v1/user/user-1", nil, &user)
	if err != nil {
		// Handle error - check if it's a talk error
		if talkErr, ok := talk.IsError(err); ok {
			fmt.Printf("Talk error: %s (code: %s)\n", talkErr.Message, talkErr.Code)
		} else {
			log.Fatal(err)
		}
	}

	fmt.Printf("Got user: %+v\n", user)

	// Output:
	// Got user: {ID:user-1 Name:Alice Email:alice@example.com}
}

// =============================================================================
// WebSocket Server Example
// =============================================================================

// ExampleWebSocketServer demonstrates creating a WebSocket server.
func ExampleWebSocketServer() {
	// Create the service implementation
	userSvc := NewUserService()

	// Extract endpoints using reflection
	extractor := extract.NewReflectExtractor()
	endpoints, err := extractor.Extract(userSvc)
	if err != nil {
		log.Fatal(err)
	}

	// Create WebSocket transport
	cfg := x.TypedLazyConfig{
		Type:   "default",
		Config: json.RawMessage(`{"addr": ":8081", "path": "/ws"}`),
	}

	transport, err := websocket.NewServer(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Create talk server
	server := talk.NewServer(transport)
	server.RegisterEndpoints(endpoints...)

	fmt.Println("WebSocket server listening on :8081/ws")
	ctx := context.Background()
	if err := server.Serve(ctx); err != nil {
		log.Fatal(err)
	}

	// Output:
	// WebSocket server listening on :8081/ws
}

// =============================================================================
// Switching Transports via Configuration
// =============================================================================

// TransportConfig demonstrates how to switch transports via configuration.
type TransportConfig struct {
	Type   string          `json:"type" yaml:"type"`     // "http" or "websocket"
	Config json.RawMessage `json:"config" yaml:"config"` // Transport-specific config
}

// CreateServerFromConfig creates a server based on configuration.
// This allows switching transports without code changes.
func CreateServerFromConfig(cfg TransportConfig, service any) (*talk.Server, error) {
	extractor := extract.NewReflectExtractor()

	var transport talk.Transport
	var err error

	switch cfg.Type {
	case "http":
		httpCfg := x.TypedLazyConfig{
			Type:   "std",
			Config: cfg.Config,
		}
		transport, err = std.NewServer(httpCfg)
	case "websocket":
		wsCfg := x.TypedLazyConfig{
			Type:   "default",
			Config: cfg.Config,
		}
		transport, err = websocket.NewServer(wsCfg)
	default:
		return nil, talk.NewError(talk.InvalidArgument, "unsupported transport: "+cfg.Type)
	}

	if err != nil {
		return nil, err
	}

	server := talk.NewServer(transport, talk.WithExtractor(extractor))
	if err := server.Register(service); err != nil {
		return nil, err
	}

	return server, nil
}

// ExampleConfigDriven demonstrates configuration-driven transport selection.
func ExampleConfigDriven() {
	userSvc := NewUserService()

	// Configuration can come from file, env vars, etc.
	// Switch between HTTP and WebSocket by changing "type"
	configs := []TransportConfig{
		{
			Type:   "http",
			Config: json.RawMessage(`{"addr": ":8080"}`),
		},
		{
			Type:   "websocket",
			Config: json.RawMessage(`{"addr": ":8081", "path": "/ws"}`),
		},
	}

	for _, cfg := range configs {
		server, err := CreateServerFromConfig(cfg, userSvc)
		if err != nil {
			log.Printf("Failed to create %s server: %v\n", cfg.Type, err)
			continue
		}

		fmt.Printf("Created %s server with %d endpoints\n", cfg.Type, len(server.Endpoints()))
	}

	// Output:
	// Created http server with 5 endpoints
	// Created websocket server with 5 endpoints
}

// =============================================================================
// Manual Endpoint Registration
// =============================================================================

// ExampleManualEndpoints demonstrates registering endpoints manually.
func ExampleManualEndpoints() {
	cfg := x.TypedLazyConfig{
		Type:   "std",
		Config: json.RawMessage(`{"addr": ":8080"}`),
	}

	transport, err := std.NewServer(cfg)
	if err != nil {
		log.Fatal(err)
	}

	server := talk.NewServer(transport)

	// Register endpoints manually with full control
	server.RegisterEndpoints(
		talk.NewEndpoint("healthz", func(ctx context.Context, req any) (any, error) {
			return map[string]string{"status": "ok"}, nil
		}, talk.WithPath("/healthz"), talk.WithMethod("GET")),

		talk.NewEndpoint("echo", func(ctx context.Context, req any) (any, error) {
			return req, nil
		}, talk.WithPath("/echo"), talk.WithMethod("POST")),
	)

	fmt.Printf("Registered %d endpoints manually\n", len(server.Endpoints()))

	// Output:
	// Registered 2 endpoints manually
}

// =============================================================================
// Error Handling
// =============================================================================

// ExampleErrorHandling demonstrates talk error handling.
func ExampleErrorHandling() {
	// Create a talk error
	err := talk.NewError(talk.NotFound, "resource not found")

	// Error implements error interface
	fmt.Printf("Error: %s\n", err.Error())

	// Get HTTP status code
	fmt.Printf("HTTP Status: %d\n", err.HTTPStatus())

	// Get gRPC code
	fmt.Printf("gRPC Code: %d\n", err.GRPCCode())

	// Create error with details
	validationErr := talk.NewErrorWithDetails(
		talk.InvalidArgument,
		"validation failed",
		map[string]string{
			"field":  "email",
			"reason": "invalid format",
		},
	)

	fmt.Printf("Validation error: %s\n", validationErr.Error())

	// Check if an error is a talk error
	if talkErr, ok := talk.IsError(err); ok {
		fmt.Printf("Is talk error with code: %s\n", talkErr.Code.String())
	}

	// Convert standard error to talk error
	stdErr := fmt.Errorf("some error")
	converted := talk.ToError(stdErr)
	fmt.Printf("Converted error code: %s\n", converted.Code.String())

	// Output:
	// Error: NOT_FOUND: resource not found
	// HTTP Status: 404
	// gRPC Code: 5
	// Validation error: INVALID_ARGUMENT: validation failed
	// Is talk error with code: NOT_FOUND
	// Converted error code: UNKNOWN
}

// =============================================================================
// Streaming Example
// =============================================================================

// ExampleStreaming demonstrates server-side streaming.
func ExampleStreaming() {
	// Create a streaming endpoint
	ep := talk.NewStreamEndpoint(
		"WatchEvents",
		func(ctx context.Context, req any, stream talk.Stream) error {
			// Send multiple events
			for i := 0; i < 3; i++ {
				event := map[string]any{
					"id":        i + 1,
					"type":      "update",
					"timestamp": time.Now().Unix(),
				}
				if err := stream.Send(event); err != nil {
					return err
				}
			}
			return nil
		},
		talk.StreamServerSide,
		talk.WithPath("/events/watch"),
	)

	fmt.Printf("Created streaming endpoint: %s\n", ep.Name)
	fmt.Printf("  Path: %s\n", ep.Path)
	fmt.Printf("  Stream Mode: %s\n", ep.StreamMode.String())
	fmt.Printf("  Is Streaming: %v\n", ep.IsStreaming())

	// Output:
	// Created streaming endpoint: WatchEvents
	//   Path: /events/watch
	//   Stream Mode: server
	//   Is Streaming: true
}

// =============================================================================
// Using Factory Pattern
// =============================================================================

// ExampleFactory demonstrates using the transport factory.
func ExampleFactory() {
	// HTTP server factory
	httpCfg := x.TypedLazyConfig{
		Type:   "std",
		Config: json.RawMessage(`{"addr": ":8080"}`),
	}

	httpServer, err := thttp.ServerFactory.Create(httpCfg)
	if err != nil {
		log.Printf("HTTP factory error: %v\n", err)
	} else {
		fmt.Printf("Created HTTP server from factory\n")
		_ = httpServer
	}

	// WebSocket server factory
	wsCfg := x.TypedLazyConfig{
		Type:   "default",
		Config: json.RawMessage(`{"addr": ":8081"}`),
	}

	wsServer, err := websocket.ServerFactory.Create(wsCfg)
	if err != nil {
		log.Printf("WebSocket factory error: %v\n", err)
	} else {
		fmt.Printf("Created WebSocket server from factory\n")
		_ = wsServer
	}

	// Output:
	// Created HTTP server from factory
	// Created WebSocket server from factory
}
