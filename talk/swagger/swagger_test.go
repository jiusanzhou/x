package swagger

import (
	"encoding/json"
	"reflect"
	"testing"

	"go.zoe.im/x/talk"
)

func TestGenerator_Generate(t *testing.T) {
	cfg := Config{
		Title:   "Test API",
		Version: "1.0.0",
	}
	gen := NewGenerator(cfg)

	type TestRequest struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	type TestResponse struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	endpoints := []*talk.Endpoint{
		{
			Name:         "GetUser",
			Path:         "/users/{id}",
			Method:       "GET",
			RequestType:  nil,
			ResponseType: reflect.TypeOf(TestResponse{}),
		},
		{
			Name:         "CreateUser",
			Path:         "/users",
			Method:       "POST",
			RequestType:  reflect.TypeOf(TestRequest{}),
			ResponseType: reflect.TypeOf(TestResponse{}),
		},
	}

	spec := gen.Generate(endpoints)

	if spec.OpenAPI != "3.0.3" {
		t.Errorf("expected OpenAPI 3.0.3, got %s", spec.OpenAPI)
	}
	if spec.Info.Title != "Test API" {
		t.Errorf("expected title 'Test API', got %s", spec.Info.Title)
	}
	if len(spec.Paths) != 2 {
		t.Errorf("expected 2 paths, got %d", len(spec.Paths))
	}

	userPath, ok := spec.Paths["/users/{id}"]
	if !ok {
		t.Error("expected /users/{id} path")
	}
	if userPath.Get == nil {
		t.Error("expected GET operation for /users/{id}")
	}
	if len(userPath.Get.Parameters) != 1 {
		t.Errorf("expected 1 parameter, got %d", len(userPath.Get.Parameters))
	}
	if userPath.Get.Parameters[0].Name != "id" {
		t.Errorf("expected parameter 'id', got %s", userPath.Get.Parameters[0].Name)
	}

	usersPath, ok := spec.Paths["/users"]
	if !ok {
		t.Error("expected /users path")
	}
	if usersPath.Post == nil {
		t.Error("expected POST operation for /users")
	}
	if usersPath.Post.RequestBody == nil {
		t.Error("expected request body for POST /users")
	}
}

func TestGenerator_GenerateJSON(t *testing.T) {
	cfg := Config{
		Title:   "Test API",
		Version: "1.0.0",
	}
	gen := NewGenerator(cfg)

	endpoints := []*talk.Endpoint{
		{
			Name:   "Ping",
			Path:   "/ping",
			Method: "GET",
		},
	}

	data, err := gen.GenerateJSON(endpoints)
	if err != nil {
		t.Fatalf("GenerateJSON failed: %v", err)
	}

	var spec OpenAPI
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if spec.Info.Title != "Test API" {
		t.Errorf("expected title 'Test API', got %s", spec.Info.Title)
	}
}

func TestGenerator_StreamingEndpoint(t *testing.T) {
	cfg := Config{
		Title:   "Test API",
		Version: "1.0.0",
	}
	gen := NewGenerator(cfg)

	endpoints := []*talk.Endpoint{
		{
			Name:       "WatchEvents",
			Path:       "/events/watch",
			Method:     "GET",
			StreamMode: talk.StreamServerSide,
		},
	}

	spec := gen.Generate(endpoints)

	eventsPath, ok := spec.Paths["/events/watch"]
	if !ok {
		t.Error("expected /events/watch path")
	}
	if eventsPath.Get == nil {
		t.Error("expected GET operation")
	}

	resp, ok := eventsPath.Get.Responses["200"]
	if !ok {
		t.Error("expected 200 response")
	}

	_, hasSSE := resp.Content["text/event-stream"]
	if !hasSSE {
		t.Error("expected text/event-stream content type for streaming endpoint")
	}
}
