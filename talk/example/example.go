// Package example demonstrates how to use the talk package.
//
// Talk is a transport abstraction layer that allows you to define service methods
// without caring about the underlying connection implementation (HTTP, gRPC, WebSocket).
// The protocol can be switched via configuration.
//
// This file contains example code that demonstrates:
//   - Defining a service interface
//   - Using reflection-based endpoint extraction
//   - Creating servers and clients with NewServerFromConfig/NewClientFromConfig
//   - Switching between HTTP and WebSocket transports via configuration
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

	_ "go.zoe.im/x/talk/transport/http/std"
	_ "go.zoe.im/x/talk/transport/websocket"
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

// ExampleHTTPServer demonstrates creating an HTTP server using NewServerFromConfig.
func ExampleHTTPServer() {
	userSvc := NewUserService()

	cfg := x.TypedLazyConfig{
		Type:   "http",
		Config: json.RawMessage(`{"addr": ":8080"}`),
	}

	server, err := talk.NewServerFromConfig(cfg, talk.WithExtractor(extract.NewReflectExtractor()))
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Register(userSvc, talk.WithPrefix("/api/v1")); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Registered %d endpoints:\n", len(server.Endpoints()))
	for _, ep := range server.Endpoints() {
		fmt.Printf("  %s %s -> %s\n", ep.Method, ep.Path, ep.Name)
	}

	ctx := context.Background()
	fmt.Println("Server listening on :8080")
	if err := server.Serve(ctx); err != nil {
		log.Fatal(err)
	}
}

// =============================================================================
// HTTP Client Example
// =============================================================================

// ExampleHTTPClient demonstrates using an HTTP client with NewClientFromConfig.
func ExampleHTTPClient() {
	cfg := x.TypedLazyConfig{
		Type:   "http",
		Config: json.RawMessage(`{"addr": "http://localhost:8080"}`),
	}

	client, err := talk.NewClientFromConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	var user User
	err = client.Call(context.Background(), "/api/v1/user/user-1", nil, &user)
	if err != nil {
		if talkErr, ok := talk.IsError(err); ok {
			fmt.Printf("Talk error: %s (code: %s)\n", talkErr.Message, talkErr.Code)
		} else {
			log.Fatal(err)
		}
	}

	fmt.Printf("Got user: %+v\n", user)
}

// =============================================================================
// WebSocket Server Example
// =============================================================================

// ExampleWebSocketServer demonstrates creating a WebSocket server using NewServerFromConfig.
func ExampleWebSocketServer() {
	userSvc := NewUserService()

	cfg := x.TypedLazyConfig{
		Type:   "websocket",
		Config: json.RawMessage(`{"addr": ":8081", "path": "/ws"}`),
	}

	server, err := talk.NewServerFromConfig(cfg, talk.WithExtractor(extract.NewReflectExtractor()))
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Register(userSvc); err != nil {
		log.Fatal(err)
	}

	fmt.Println("WebSocket server listening on :8081/ws")
	ctx := context.Background()
	if err := server.Serve(ctx); err != nil {
		log.Fatal(err)
	}
}

// =============================================================================
// Switching Transports via Configuration
// =============================================================================

// TransportConfig demonstrates how to switch transports via configuration.
type TransportConfig struct {
	Type   string          `json:"type" yaml:"type"`     // "http" or "websocket"
	Config json.RawMessage `json:"config" yaml:"config"` // Transport-specific config
}

// CreateServerFromConfig creates a server based on configuration using NewServerFromConfig.
// This allows switching transports without code changes.
func CreateServerFromConfig(cfg TransportConfig, service any) (*talk.Server, error) {
	extractor := extract.NewReflectExtractor()

	talkCfg := x.TypedLazyConfig{
		Type:   cfg.Type,
		Config: cfg.Config,
	}

	server, err := talk.NewServerFromConfig(talkCfg, talk.WithExtractor(extractor))
	if err != nil {
		return nil, err
	}

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
		Type:   "http",
		Config: json.RawMessage(`{"addr": ":8080"}`),
	}

	server, err := talk.NewServerFromConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	server.RegisterEndpoints(
		talk.NewEndpoint("healthz", func(ctx context.Context, req any) (any, error) {
			return map[string]string{"status": "ok"}, nil
		}, talk.WithPath("/healthz"), talk.WithMethod("GET")),

		talk.NewEndpoint("echo", func(ctx context.Context, req any) (any, error) {
			return req, nil
		}, talk.WithPath("/echo"), talk.WithMethod("POST")),
	)

	fmt.Printf("Registered %d endpoints manually\n", len(server.Endpoints()))
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

// ExampleFactory demonstrates using NewServerFromConfig for unified transport creation.
func ExampleFactory() {
	httpCfg := x.TypedLazyConfig{
		Type:   "http",
		Config: json.RawMessage(`{"addr": ":8080"}`),
	}

	httpServer, err := talk.NewServerFromConfig(httpCfg)
	if err != nil {
		log.Printf("HTTP error: %v\n", err)
	} else {
		fmt.Printf("Created HTTP server from config\n")
		_ = httpServer
	}

	wsCfg := x.TypedLazyConfig{
		Type:   "websocket",
		Config: json.RawMessage(`{"addr": ":8081"}`),
	}

	wsServer, err := talk.NewServerFromConfig(wsCfg)
	if err != nil {
		log.Printf("WebSocket error: %v\n", err)
	} else {
		fmt.Printf("Created WebSocket server from config\n")
		_ = wsServer
	}
}

// ExampleRegisterTransport demonstrates registering a custom transport.
func ExampleRegisterTransport() {
	talk.RegisterTransport("custom", &talk.TransportCreators{
		Server: func(cfg x.TypedLazyConfig) (talk.Transport, error) {
			return nil, fmt.Errorf("custom server not implemented")
		},
		Client: func(cfg x.TypedLazyConfig) (talk.Transport, error) {
			return nil, fmt.Errorf("custom client not implemented")
		},
	}, "custom-alias")

	cfg := x.TypedLazyConfig{
		Type:   "custom",
		Config: json.RawMessage(`{}`),
	}

	_, err := talk.NewServerFromConfig(cfg)
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	}
}

// =============================================================================
// Swagger Documentation Example
// =============================================================================

// ExampleSwagger demonstrates enabling Swagger documentation for HTTP servers.
func ExampleSwagger() {
	userSvc := NewUserService()

	cfg := x.TypedLazyConfig{
		Type: "http",
		Config: json.RawMessage(`{
			"addr": ":8080",
			"swagger": {
				"enabled": true,
				"path": "/swagger",
				"title": "User Service API",
				"description": "API for managing users",
				"version": "1.0.0"
			}
		}`),
	}

	server, err := talk.NewServerFromConfig(cfg, talk.WithExtractor(extract.NewReflectExtractor()))
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Register(userSvc, talk.WithPrefix("/api/v1")); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Registered %d endpoints\n", len(server.Endpoints()))
	fmt.Println("Swagger UI available at: http://localhost:8080/swagger/")
	fmt.Println("OpenAPI spec available at: http://localhost:8080/swagger/openapi.json")

	ctx := context.Background()
	if err := server.Serve(ctx); err != nil {
		log.Fatal(err)
	}
}

// =============================================================================
// Path Parameters Example
// =============================================================================

// GetNodeModelRequest demonstrates using `path` struct tags to extract
// multiple path parameters from the URL.
type GetNodeModelRequest struct {
	NodeName  string `json:"nodeName" path:"nodeName"`
	ModelName string `json:"modelName" path:"modelName"`
}

// GetNodeModelResponse is the response for GetNodeModel.
type GetNodeModelResponse struct {
	Node  string `json:"node"`
	Model string `json:"model"`
	Ready bool   `json:"ready"`
}

// nodeModelService demonstrates a service with path parameters.
type nodeModelService struct{}

// GetNodeModel handles GET /nodes/{nodeName}/models/{modelName}
// The `path` struct tags on GetNodeModelRequest automatically bind URL path segments.
func (s *nodeModelService) GetNodeModel(ctx context.Context, req *GetNodeModelRequest) (*GetNodeModelResponse, error) {
	return &GetNodeModelResponse{
		Node:  req.NodeName,
		Model: req.ModelName,
		Ready: true,
	}, nil
}

// DeleteNodeModel handles DELETE /nodes/{nodeName}/models/{modelName}
// Even DELETE requests with no body can receive path parameters via struct tags.
func (s *nodeModelService) DeleteNodeModel(ctx context.Context, req *GetNodeModelRequest) error {
	fmt.Printf("Deleting model %s from node %s\n", req.ModelName, req.NodeName)
	return nil
}

// TalkAnnotations provides custom path patterns with multiple path parameters.
func (s *nodeModelService) TalkAnnotations() map[string]string {
	return map[string]string{
		"GetNodeModel":    "@talk path=/nodes/{nodeName}/models/{modelName} method=GET",
		"DeleteNodeModel": "@talk path=/nodes/{nodeName}/models/{modelName} method=DELETE",
	}
}

// ExamplePathParams demonstrates using path parameters with struct tags.
// Path segments like /nodes/{nodeName}/models/{modelName} are automatically
// extracted and bound to request struct fields tagged with `path:"paramName"`.
func ExamplePathParams() {
	svc := &nodeModelService{}

	cfg := x.TypedLazyConfig{
		Type:   "http",
		Config: json.RawMessage(`{"addr": ":8080"}`),
	}

	server, err := talk.NewServerFromConfig(cfg, talk.WithExtractor(extract.NewReflectExtractor()))
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Register(svc, talk.WithPrefix("/api/v1")); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Path parameter endpoints:")
	for _, ep := range server.Endpoints() {
		fmt.Printf("  %s %s\n", ep.Method, ep.Path)
	}

	// Test with:
	//   curl http://localhost:8080/api/v1/nodes/gpu-01/models/llama-70b
	//   curl -X DELETE http://localhost:8080/api/v1/nodes/gpu-01/models/llama-70b

	ctx := context.Background()
	if err := server.Serve(ctx); err != nil {
		log.Fatal(err)
	}
}

// =============================================================================
// Query Parameters Example
// =============================================================================

// ListTasksRequest demonstrates using `query` struct tags to extract
// query string parameters from the URL (e.g., /tasks?status=running&node=gpu-01).
type ListTasksRequest struct {
	Status string `json:"status" query:"status"`
	Node   string `json:"node" query:"node"`
	Limit  string `json:"limit" query:"limit"`
}

// Task represents a task entity.
type Task struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Node   string `json:"node"`
}

// taskService demonstrates a service with query parameters.
type taskService struct{}

// ListTasks handles GET /tasks?status=running&node=gpu-01
// The `query` struct tags on ListTasksRequest automatically bind URL query parameters.
func (s *taskService) ListTasks(ctx context.Context, req *ListTasksRequest) ([]*Task, error) {
	fmt.Printf("Listing tasks: status=%s, node=%s, limit=%s\n", req.Status, req.Node, req.Limit)
	return []*Task{
		{ID: "task-1", Status: req.Status, Node: req.Node},
	}, nil
}

// ExampleQueryParams demonstrates using query parameters with struct tags.
// Query parameters like ?status=running&node=gpu-01 are automatically
// extracted and bound to request struct fields tagged with `query:"paramName"`.
func ExampleQueryParams() {
	svc := &taskService{}

	cfg := x.TypedLazyConfig{
		Type:   "http",
		Config: json.RawMessage(`{"addr": ":8080"}`),
	}

	server, err := talk.NewServerFromConfig(cfg, talk.WithExtractor(extract.NewReflectExtractor()))
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Register(svc, talk.WithPrefix("/api/v1")); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Query parameter endpoints:")
	for _, ep := range server.Endpoints() {
		fmt.Printf("  %s %s\n", ep.Method, ep.Path)
	}

	// Test with:
	//   curl "http://localhost:8080/api/v1/tasks?status=running&node=gpu-01&limit=10"

	ctx := context.Background()
	if err := server.Serve(ctx); err != nil {
		log.Fatal(err)
	}
}

// =============================================================================
// Middleware Example
// =============================================================================

// LoggingMiddleware logs every request and response.
func LoggingMiddleware() talk.MiddlewareFunc {
	return func(next talk.EndpointFunc) talk.EndpointFunc {
		return func(ctx context.Context, req any) (any, error) {
			log.Printf("[LOG] request: %+v", req)
			resp, err := next(ctx, req)
			if err != nil {
				log.Printf("[LOG] error: %v", err)
			} else {
				log.Printf("[LOG] response: %+v", resp)
			}
			return resp, err
		}
	}
}

// AuthMiddleware checks for an "authorized" key in context.
// In a real application, you would validate JWT tokens, API keys, etc.
func AuthMiddleware() talk.MiddlewareFunc {
	return func(next talk.EndpointFunc) talk.EndpointFunc {
		return func(ctx context.Context, req any) (any, error) {
			// In a real app, extract and validate the token from context/headers.
			// For this example, we just log and pass through.
			log.Println("[AUTH] checking authorization")
			return next(ctx, req)
		}
	}
}

// RecoveryMiddleware catches panics and converts them to errors.
func RecoveryMiddleware() talk.MiddlewareFunc {
	return func(next talk.EndpointFunc) talk.EndpointFunc {
		return func(ctx context.Context, req any) (resp any, err error) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[RECOVERY] caught panic: %v", r)
					err = talk.NewError(talk.Internal, fmt.Sprintf("internal error: %v", r))
				}
			}()
			return next(ctx, req)
		}
	}
}

// ExampleMiddleware demonstrates using middleware at both server and endpoint levels.
func ExampleMiddleware() {
	userSvc := NewUserService()

	cfg := x.TypedLazyConfig{
		Type:   "http",
		Config: json.RawMessage(`{"addr": ":8080"}`),
	}

	// Server-level middleware applies to ALL registered endpoints.
	server, err := talk.NewServerFromConfig(cfg,
		talk.WithExtractor(extract.NewReflectExtractor()),
		talk.WithServerMiddleware(
			RecoveryMiddleware(), // outermost: catch panics
			LoggingMiddleware(),  // log all requests
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Register(userSvc, talk.WithPrefix("/api/v1")); err != nil {
		log.Fatal(err)
	}

	// Endpoint-level middleware applies only to specific endpoints.
	server.RegisterEndpoints(
		talk.NewEndpoint("admin-stats", func(ctx context.Context, req any) (any, error) {
			return map[string]any{"users": 42, "requests": 1000}, nil
		},
			talk.WithPath("/admin/stats"),
			talk.WithMethod("GET"),
			talk.WithMiddleware(
				AuthMiddleware(), // only this endpoint requires auth
			),
		),
	)

	fmt.Println("Middleware example endpoints:")
	for _, ep := range server.Endpoints() {
		mwCount := len(ep.Middleware)
		fmt.Printf("  %s %s (%d middleware)\n", ep.Method, ep.Path, mwCount)
	}

	// Test with:
	//   curl http://localhost:8080/api/v1/users          # recovery + logging
	//   curl http://localhost:8080/admin/stats            # auth middleware
	//   curl -X POST http://localhost:8080/api/v1/user \\
	//     -H "Content-Type: application/json" \\
	//     -d '{"name": "Alice", "email": "alice@example.com"}'

	ctx := context.Background()
	if err := server.Serve(ctx); err != nil {
		log.Fatal(err)
	}
}

// =============================================================================
// Annotation-Based Middleware Example
// =============================================================================

// protectedService demonstrates middleware configured via @talk annotations.
type protectedService struct{}

// PublicEndpoint is accessible without auth.
func (s *protectedService) GetStatus(ctx context.Context) (map[string]string, error) {
	return map[string]string{"status": "ok"}, nil
}

// AdminEndpoint requires auth and admin middleware (configured via annotation).
func (s *protectedService) DeleteAllData(ctx context.Context) (map[string]string, error) {
	return map[string]string{"deleted": "true"}, nil
}

// TalkAnnotations configures middleware via annotations.
// The `middleware=auth,admin` annotation attaches named middleware to the endpoint.
// The actual middleware functions must be registered separately.
func (s *protectedService) TalkAnnotations() map[string]string {
	return map[string]string{
		"GetStatus":     "@talk path=/status method=GET",
		"DeleteAllData": "@talk path=/admin/delete-all method=DELETE middleware=auth,admin",
	}
}

// ExampleAnnotationMiddleware demonstrates configuring middleware via @talk annotations.
func ExampleAnnotationMiddleware() {
	svc := &protectedService{}

	cfg := x.TypedLazyConfig{
		Type:   "http",
		Config: json.RawMessage(`{"addr": ":8080"}`),
	}

	server, err := talk.NewServerFromConfig(cfg, talk.WithExtractor(extract.NewReflectExtractor()))
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Register(svc); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Annotation middleware endpoints:")
	for _, ep := range server.Endpoints() {
		mwNames := ""
		if names, ok := ep.Metadata["middleware"]; ok {
			mwNames = fmt.Sprintf(" [middleware: %v]", names)
		}
		fmt.Printf("  %s %s%s\n", ep.Method, ep.Path, mwNames)
	}
}
