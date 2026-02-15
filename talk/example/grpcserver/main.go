// Command grpcserver demonstrates a simple gRPC server using talk.
//
// Run with:
//
//	go run go.zoe.im/x/talk/example/grpcserver
//
// Then test with grpcurl:
//
//	# List services
//	grpcurl -plaintext localhost:9090 list
//
//	# Create a user
//	grpcurl -plaintext -d '{"name": "Alice", "email": "alice@example.com"}' \
//	  localhost:9090 talk.UserService/CreateUser
//
//	# Get the user
//	grpcurl -plaintext -d '{"id": "user-1"}' \
//	  localhost:9090 talk.UserService/GetUser
//
//	# List all users
//	grpcurl -plaintext localhost:9090 talk.UserService/ListUsers
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

	_ "go.zoe.im/x/talk/transport/grpc"
)

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserEvent struct {
	Type      string `json:"type"`
	UserID    string `json:"user_id,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

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
		Type:   "grpc",
		Config: json.RawMessage(`{"addr": ":9090"}`),
	}

	server, err := talk.NewServerFromConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Register(svc); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Registered endpoints:")
	for _, ep := range server.Endpoints() {
		fmt.Printf("  %s -> %s", ep.Path, ep.Name)
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

	log.Println("gRPC server listening on :9090")
	if err := server.Serve(ctx); err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}
