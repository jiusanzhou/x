// Command server demonstrates a simple HTTP server using talk.
//
// Run with:
//
//	go run go.zoe.im/x/talk/example/server
//
// Then test with:
//
//	# Create a user
//	curl -X POST http://localhost:8080/api/v1/user \
//	  -H "Content-Type: application/json" \
//	  -d '{"name": "Alice", "email": "alice@example.com"}'
//
//	# Get the user
//	curl http://localhost:8080/api/v1/user/user-1
//
//	# List all users
//	curl http://localhost:8080/api/v1/users
//
//	# Watch for events (SSE)
//	curl http://localhost:8080/api/v1/users/watch
//
//	# Swagger UI
//	open http://localhost:8080/swagger/
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.zoe.im/x"
	"go.zoe.im/x/talk"
	"go.zoe.im/x/talk/extract"

	_ "go.zoe.im/x/talk/transport/http/std"
)

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
	Type      string `json:"type"`
	UserID    string `json:"user_id,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// userService implements the user service.
type userService struct {
	users map[string]*User
}

func newUserService() *userService {
	return &userService{
		users: make(map[string]*User),
	}
}

func (s *userService) GetUser(ctx context.Context, id string) (*User, error) {
	user, ok := s.users[id]
	if !ok {
		return nil, talk.NewError(talk.NotFound, "user not found: "+id)
	}
	return user, nil
}

func (s *userService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
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
	log.Printf("Created user: %s (%s)", user.Name, user.ID)
	return user, nil
}

func (s *userService) ListUsers(ctx context.Context) ([]*User, error) {
	users := make([]*User, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u)
	}
	return users, nil
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	if _, ok := s.users[id]; !ok {
		return talk.NewError(talk.NotFound, "user not found: "+id)
	}
	delete(s.users, id)
	log.Printf("Deleted user: %s", id)
	return nil
}

func (s *userService) WatchUsers(ctx context.Context) (<-chan *UserEvent, error) {
	ch := make(chan *UserEvent)
	go func() {
		defer close(ch)
		log.Println("Client connected to watch stream")
		for i := 0; i < 10; i++ {
			select {
			case <-ctx.Done():
				log.Println("Watch stream closed by client")
				return
			case <-time.After(time.Second):
				event := &UserEvent{
					Type:      "heartbeat",
					Timestamp: time.Now().Unix(),
				}
				ch <- event
			}
		}
		log.Println("Watch stream completed")
	}()
	return ch, nil
}

func main() {
	svc := newUserService()

	cfg := x.TypedLazyConfig{
		Type: "http",
		Config: json.RawMessage(`{
			"addr": ":8080",
			"swagger": {
				"enabled": true,
				"path": "/swagger",
				"title": "User Service API",
				"description": "Example API for managing users",
				"version": "1.0.0"
			}
		}`),
	}

	server, err := talk.NewServerFromConfig(cfg, talk.WithExtractor(extract.NewReflectExtractor()))
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Register(svc, talk.WithPrefix("/api/v1")); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Registered endpoints:")
	for _, ep := range server.Endpoints() {
		fmt.Printf("  %s %s -> %s", ep.Method, ep.Path, ep.Name)
		if ep.IsStreaming() {
			fmt.Printf(" (streaming: %s)", ep.StreamMode.String())
		}
		fmt.Println()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Shutting down...")
		cancel()
	}()

	log.Println("Server listening on :8080")
	log.Println("Swagger UI: http://localhost:8080/swagger/")
	if err := server.Serve(ctx); err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}
