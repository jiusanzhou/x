// Command wsclient demonstrates a WebSocket client using talk.
//
// Run the server first:
//
//	go run go.zoe.im/x/talk/example/wsserver
//
// Then run this client:
//
//	go run go.zoe.im/x/talk/example/wsclient
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go.zoe.im/x"
	"go.zoe.im/x/talk"
	"go.zoe.im/x/talk/transport/websocket"
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

func main() {
	// Create WebSocket client
	cfg := x.TypedLazyConfig{
		Type:   "default",
		Config: json.RawMessage(`{"addr": "localhost:8081", "path": "/ws"}`),
	}

	client, err := websocket.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	talkClient := talk.NewClient(client)

	fmt.Println("Connected to WebSocket server at localhost:8081/ws")

	// Create a user
	fmt.Println("\n1. Creating a user...")
	createReq := &CreateUserRequest{
		Name:  "Bob",
		Email: "bob@example.com",
	}

	var createdUser User
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = talkClient.Call(ctx, "CreateUser", createReq, &createdUser)
	cancel()

	if err != nil {
		if talkErr, ok := talk.IsError(err); ok {
			log.Printf("Create failed: %s (code: %s)", talkErr.Message, talkErr.Code)
		} else {
			log.Fatalf("Create failed: %v", err)
		}
	} else {
		fmt.Printf("   Created: %+v\n", createdUser)
	}

	// Get the user
	fmt.Println("\n2. Getting the user...")
	var fetchedUser User
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	err = talkClient.Call(ctx, "GetUser", createdUser.ID, &fetchedUser)
	cancel()

	if err != nil {
		if talkErr, ok := talk.IsError(err); ok {
			log.Printf("Get failed: %s (code: %s)", talkErr.Message, talkErr.Code)
		} else {
			log.Fatalf("Get failed: %v", err)
		}
	} else {
		fmt.Printf("   Fetched: %+v\n", fetchedUser)
	}

	// List all users
	fmt.Println("\n3. Listing all users...")
	var users []*User
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	err = talkClient.Call(ctx, "ListUsers", nil, &users)
	cancel()

	if err != nil {
		log.Printf("List failed: %v", err)
	} else {
		fmt.Printf("   Found %d user(s):\n", len(users))
		for _, u := range users {
			fmt.Printf("   - %s: %s <%s>\n", u.ID, u.Name, u.Email)
		}
	}

	// Try to get a non-existent user (to demonstrate error handling)
	fmt.Println("\n4. Getting non-existent user...")
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	err = talkClient.Call(ctx, "GetUser", "user-999", &fetchedUser)
	cancel()

	if err != nil {
		if talkErr, ok := talk.IsError(err); ok {
			fmt.Printf("   Expected error: %s (code: %s)\n", talkErr.Message, talkErr.Code)
		} else {
			log.Printf("   Unexpected error type: %v", err)
		}
	}

	fmt.Println("\nDone!")
}
