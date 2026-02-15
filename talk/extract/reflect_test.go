package extract

import (
	"context"
	"testing"

	"go.zoe.im/x/talk"
)

type testUser struct {
	ID   string
	Name string
}

type testUserService struct{}

func (s *testUserService) GetUser(ctx context.Context, id string) (*testUser, error) {
	return &testUser{ID: id, Name: "Test"}, nil
}

func (s *testUserService) CreateUser(ctx context.Context, req *testUser) (*testUser, error) {
	return req, nil
}

func (s *testUserService) ListUsers(ctx context.Context) ([]*testUser, error) {
	return nil, nil
}

func (s *testUserService) DeleteUser(ctx context.Context, id string) error {
	return nil
}

func (s *testUserService) WatchUsers(ctx context.Context) (<-chan *testUser, error) {
	return nil, nil
}

func TestReflectExtractor_Extract(t *testing.T) {
	extractor := NewReflectExtractor()
	endpoints, err := extractor.Extract(&testUserService{})
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(endpoints) != 5 {
		t.Errorf("expected 5 endpoints, got %d", len(endpoints))
	}

	endpointMap := make(map[string]*talk.Endpoint)
	for _, ep := range endpoints {
		endpointMap[ep.Name] = ep
	}

	tests := []struct {
		name       string
		method     string
		path       string
		streamMode talk.StreamMode
	}{
		{"GetUser", "GET", "/user/{id}", talk.StreamNone},
		{"CreateUser", "POST", "/user", talk.StreamNone},
		{"ListUsers", "GET", "/users", talk.StreamNone},
		{"DeleteUser", "DELETE", "/user/{id}", talk.StreamNone},
		{"WatchUsers", "GET", "/users/watch", talk.StreamServerSide},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep, ok := endpointMap[tt.name]
			if !ok {
				t.Fatalf("endpoint %s not found", tt.name)
			}

			if ep.Method != tt.method {
				t.Errorf("method: got %s, want %s", ep.Method, tt.method)
			}

			if ep.Path != tt.path {
				t.Errorf("path: got %s, want %s", ep.Path, tt.path)
			}

			if ep.StreamMode != tt.streamMode {
				t.Errorf("streamMode: got %v, want %v", ep.StreamMode, tt.streamMode)
			}
		})
	}
}

func TestReflectExtractor_WithPathPrefix(t *testing.T) {
	extractor := NewReflectExtractor(WithPathPrefix("/api/v1"))
	endpoints, err := extractor.Extract(&testUserService{})
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	for _, ep := range endpoints {
		if ep.Path[:7] != "/api/v1" {
			t.Errorf("expected path prefix /api/v1, got %s", ep.Path)
		}
	}
}

func TestReflectExtractor_Handler(t *testing.T) {
	extractor := NewReflectExtractor()
	endpoints, err := extractor.Extract(&testUserService{})
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	var getUserEndpoint *talk.Endpoint
	for _, ep := range endpoints {
		if ep.Name == "GetUser" {
			getUserEndpoint = ep
			break
		}
	}

	if getUserEndpoint == nil {
		t.Fatal("GetUser endpoint not found")
	}

	resp, err := getUserEndpoint.Handler(context.Background(), "123")
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}

	user, ok := resp.(*testUser)
	if !ok {
		t.Fatalf("expected *testUser, got %T", resp)
	}

	if user.ID != "123" {
		t.Errorf("expected ID 123, got %s", user.ID)
	}
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user", "users"},
		{"box", "boxes"},
		{"city", "cities"},
		{"day", "days"},
		{"bus", "buses"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := pluralize(tt.input)
			if result != tt.expected {
				t.Errorf("pluralize(%s) = %s, want %s", tt.input, result, tt.expected)
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
		{"CreateUserProfile", "create-user-profile"},
		{"ID", "i-d"},
		{"simple", "simple"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toKebabCase(tt.input)
			if result != tt.expected {
				t.Errorf("toKebabCase(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

type annotatedService struct{}

func (s *annotatedService) GetUser(ctx context.Context, id string) (*testUser, error) {
	return &testUser{ID: id}, nil
}

func (s *annotatedService) CreateUser(ctx context.Context, req *testUser) (*testUser, error) {
	return req, nil
}

func (s *annotatedService) InternalMethod(ctx context.Context) error {
	return nil
}

func (s *annotatedService) CustomPath(ctx context.Context) (*testUser, error) {
	return &testUser{}, nil
}

func (s *annotatedService) TalkAnnotations() map[string]string {
	return map[string]string{
		"InternalMethod": "@talk skip",
		"CustomPath":     "@talk path=/custom/endpoint method=PUT",
	}
}

func TestReflectExtractor_WithAnnotations(t *testing.T) {
	extractor := NewReflectExtractor()
	endpoints, err := extractor.Extract(&annotatedService{})
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	endpointMap := make(map[string]*talk.Endpoint)
	for _, ep := range endpoints {
		endpointMap[ep.Name] = ep
	}

	if _, ok := endpointMap["InternalMethod"]; ok {
		t.Error("InternalMethod should be skipped")
	}

	if ep, ok := endpointMap["CustomPath"]; ok {
		if ep.Path != "/custom/endpoint" {
			t.Errorf("CustomPath path = %s, want /custom/endpoint", ep.Path)
		}
		if ep.Method != "PUT" {
			t.Errorf("CustomPath method = %s, want PUT", ep.Method)
		}
	} else {
		t.Error("CustomPath endpoint not found")
	}

	if _, ok := endpointMap["GetUser"]; !ok {
		t.Error("GetUser should be extracted normally")
	}
}
