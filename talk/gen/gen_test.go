package gen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerator_Generate(t *testing.T) {
	// Create a temp directory for test files
	tmpDir, err := os.MkdirTemp("", "talk-gen-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test source file
	sourceFile := filepath.Join(tmpDir, "service.go")
	sourceContent := `package testservice

import "context"

// UserService handles user operations.
type UserService interface {
	// @talk path=/users/{id} method=GET
	GetUser(ctx context.Context, id string) (*User, error)

	// @talk path=/users method=POST
	CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error)

	// ListUsers returns all users
	ListUsers(ctx context.Context) ([]*User, error)

	// DeleteUser removes a user
	DeleteUser(ctx context.Context, id string) error
}

type User struct {
	ID   string
	Name string
}

type CreateUserRequest struct {
	Name string
}
`
	if err := os.WriteFile(sourceFile, []byte(sourceContent), 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	// Run generator
	g := &Generator{
		TypeName: "UserService",
	}

	if err := g.Generate(sourceFile); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check output file exists
	outputFile := filepath.Join(tmpDir, "service_talk.go")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatal("output file not created")
	}

	// Read and verify output
	output, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	outputStr := string(output)

	// Verify key elements
	checks := []string{
		"package testservice",
		"UserServiceEndpoints",
		`Name:       "GetUser"`,
		`Path:       "/users/{id}"`,
		`Method:     "GET"`,
		`Name:       "CreateUser"`,
		`Path:       "/users"`,
		`Method:     "POST"`,
		"DO NOT EDIT",
	}

	for _, check := range checks {
		if !strings.Contains(outputStr, check) {
			t.Errorf("output missing: %q", check)
		}
	}
}

func TestDeriveMethodAndPath(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"GetUser", "GET", "/user"},
		{"ListUsers", "GET", "/users"},
		{"CreateUser", "POST", "/user"},
		{"UpdateUser", "PUT", "/user/{id}"},
		{"DeleteUser", "DELETE", "/user/{id}"},
		{"WatchUsers", "GET", "/users/watch"},
		{"DoSomething", "POST", "/do-something"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use nil funcType since we're not checking params here
			method, path := deriveMethodAndPath(tt.name, nil)

			if method != tt.method {
				t.Errorf("method = %q, want %q", method, tt.method)
			}

			if path != tt.path {
				t.Errorf("path = %q, want %q", path, tt.path)
			}
		})
	}
}

func TestExtractResource(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"GetUser", "user"},
		{"ListUsers", "users"},
		{"CreateProduct", "product"},
		{"DeleteItem", "item"},
		{"WatchEvents", "events"},
		{"DoSomething", "dosomething"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := extractResource(tt.input); got != tt.expected {
				t.Errorf("extractResource(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"GetUser", "get-user"},
		{"DoSomething", "do-something"},
		{"HTTPServer", "h-t-t-p-server"},
		{"simple", "simple"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := toKebabCase(tt.input); got != tt.expected {
				t.Errorf("toKebabCase(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
