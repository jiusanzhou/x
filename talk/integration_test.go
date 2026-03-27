package talk_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"go.zoe.im/x"
	"go.zoe.im/x/talk"
	"go.zoe.im/x/talk/extract"

	// Import transports so init() registers them
	stdhttp "go.zoe.im/x/talk/transport/http/std"
)

// ============================================================
// Test service types
// ============================================================

type Task struct {
	ID     string `json:"id" path:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type CreateTaskRequest struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type GetTaskRequest struct {
	ID string `json:"id" path:"id"`
}

type ListTasksRequest struct {
	Status string `json:"status" query:"status"`
	Page   int    `json:"page" query:"page"`
	Limit  int    `json:"limit" query:"limit"`
}

type UpdateTaskRequest struct {
	ID     string `json:"id" path:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type DeleteTaskRequest struct {
	ID string `json:"id" path:"id"`
}

// taskService captures requests for verification
type taskService struct {
	lastCreate *CreateTaskRequest
	lastGet    *GetTaskRequest
	lastList   *ListTasksRequest
	lastUpdate *UpdateTaskRequest
	lastDelete *DeleteTaskRequest
}

func (s *taskService) CreateTask(ctx context.Context, req CreateTaskRequest) (*Task, error) {
	s.lastCreate = &req
	return &Task{ID: "new-1", Name: req.Name, Status: req.Status}, nil
}

func (s *taskService) GetTask(ctx context.Context, req GetTaskRequest) (*Task, error) {
	s.lastGet = &req
	return &Task{ID: req.ID, Name: "fetched", Status: "open"}, nil
}

func (s *taskService) ListTasks(ctx context.Context, req ListTasksRequest) ([]*Task, error) {
	s.lastList = &req
	return []*Task{{ID: "1", Name: "task1", Status: req.Status}}, nil
}

func (s *taskService) UpdateTask(ctx context.Context, req UpdateTaskRequest) (*Task, error) {
	s.lastUpdate = &req
	return &Task{ID: req.ID, Name: req.Name, Status: req.Status}, nil
}

func (s *taskService) DeleteTask(ctx context.Context, req DeleteTaskRequest) error {
	s.lastDelete = &req
	return nil
}

func (s *taskService) TalkAnnotations() map[string]string {
	return map[string]string{
		"GetTask":    "@talk path=/tasks/{id} method=GET",
		"UpdateTask": "@talk path=/tasks/{id} method=PUT",
		"DeleteTask": "@talk path=/tasks/{id} method=DELETE",
	}
}

// ============================================================
// Nested resource service
// ============================================================

type Comment struct {
	ID     string `json:"id" path:"id"`
	TaskID string `json:"taskId" path:"taskId"`
	Body   string `json:"body"`
}

type ListCommentsRequest struct {
	TaskID string `json:"taskId" path:"taskId"`
	Sort   string `json:"sort" query:"sort"`
}

type CreateCommentRequest struct {
	TaskID string `json:"taskId" path:"taskId"`
	Body   string `json:"body"`
}

type GetCommentRequest struct {
	TaskID string `json:"taskId" path:"taskId"`
	ID     string `json:"id" path:"id"`
}

type DeleteCommentRequest struct {
	TaskID string `json:"taskId" path:"taskId"`
	ID     string `json:"id" path:"id"`
}

type commentService struct {
	lastList   *ListCommentsRequest
	lastCreate *CreateCommentRequest
	lastGet    *GetCommentRequest
	lastDelete *DeleteCommentRequest
}

func (s *commentService) ListComments(ctx context.Context, req ListCommentsRequest) ([]*Comment, error) {
	s.lastList = &req
	return []*Comment{{ID: "c1", TaskID: req.TaskID, Body: "hello"}}, nil
}

func (s *commentService) CreateComment(ctx context.Context, req CreateCommentRequest) (*Comment, error) {
	s.lastCreate = &req
	return &Comment{ID: "c-new", TaskID: req.TaskID, Body: req.Body}, nil
}

func (s *commentService) GetComment(ctx context.Context, req GetCommentRequest) (*Comment, error) {
	s.lastGet = &req
	return &Comment{ID: req.ID, TaskID: req.TaskID, Body: "found"}, nil
}

func (s *commentService) DeleteComment(ctx context.Context, req DeleteCommentRequest) error {
	s.lastDelete = &req
	return nil
}

func (s *commentService) TalkAnnotations() map[string]string {
	return map[string]string{
		"ListComments":   "@talk path=/tasks/{taskId}/comments method=GET",
		"CreateComment":  "@talk path=/tasks/{taskId}/comments method=POST",
		"GetComment":     "@talk path=/tasks/{taskId}/comments/{id} method=GET",
		"DeleteComment":  "@talk path=/tasks/{taskId}/comments/{id} method=DELETE",
	}
}

// ============================================================
// Helpers
// ============================================================

func setupStdServer(t *testing.T, svc any, prefix string) *httptest.Server {
	t.Helper()

	cfg := x.TypedLazyConfig{Config: json.RawMessage(`{"addr":":0"}`)}
	transport, err := stdhttp.NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	extractor := extract.NewReflectExtractor()
	endpoints, err := extractor.Extract(svc)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}

	if prefix != "" {
		for _, ep := range endpoints {
			ep.Path = prefix + ep.Path
		}
	}

	transport.RegisterEndpoints(endpoints)
	return httptest.NewServer(transport.ServeMux())
}

func doJSON(t *testing.T, method, url string, body any) *http.Response {
	t.Helper()
	var reqBody *bytes.Buffer
	if body != nil {
		data, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(data)
	} else {
		reqBody = &bytes.Buffer{}
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s failed: %v", method, url, err)
	}
	return resp
}

func decodeJSON[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	defer resp.Body.Close()
	var v T
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return v
}

// ============================================================
// Integration tests: std transport CRUD
// ============================================================

func TestIntegration_StdTransport_CRUD(t *testing.T) {
	svc := &taskService{}
	ts := setupStdServer(t, svc, "")
	defer ts.Close()

	t.Run("CreateTask", func(t *testing.T) {
		resp := doJSON(t, "POST", ts.URL+"/task", CreateTaskRequest{Name: "buy milk", Status: "open"})
		result := decodeJSON[Task](t, resp)

		if resp.StatusCode != 200 {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		if result.Name != "buy milk" {
			t.Errorf("Name = %q, want %q", result.Name, "buy milk")
		}
		if svc.lastCreate == nil {
			t.Fatal("handler not called")
		}
		if svc.lastCreate.Name != "buy milk" {
			t.Errorf("captured Name = %q, want %q", svc.lastCreate.Name, "buy milk")
		}
		if svc.lastCreate.Status != "open" {
			t.Errorf("captured Status = %q, want %q", svc.lastCreate.Status, "open")
		}
	})

	t.Run("GetTask_PathParam", func(t *testing.T) {
		resp := doJSON(t, "GET", ts.URL+"/tasks/task-42", nil)
		result := decodeJSON[Task](t, resp)

		if resp.StatusCode != 200 {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		if result.ID != "task-42" {
			t.Errorf("ID = %q, want %q", result.ID, "task-42")
		}
		if svc.lastGet == nil {
			t.Fatal("handler not called")
		}
		if svc.lastGet.ID != "task-42" {
			t.Errorf("captured ID = %q, want %q", svc.lastGet.ID, "task-42")
		}
	})

	t.Run("ListTasks_QueryParams", func(t *testing.T) {
		resp := doJSON(t, "GET", ts.URL+"/tasks?status=done&page=3&limit=25", nil)
		_ = decodeJSON[[]Task](t, resp)

		if resp.StatusCode != 200 {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		if svc.lastList == nil {
			t.Fatal("handler not called")
		}
		if svc.lastList.Status != "done" {
			t.Errorf("Status = %q, want %q", svc.lastList.Status, "done")
		}
		if svc.lastList.Page != 3 {
			t.Errorf("Page = %d, want 3", svc.lastList.Page)
		}
		if svc.lastList.Limit != 25 {
			t.Errorf("Limit = %d, want 25", svc.lastList.Limit)
		}
	})

	t.Run("UpdateTask_PathAndBody", func(t *testing.T) {
		body := map[string]string{"name": "updated", "status": "closed"}
		resp := doJSON(t, "PUT", ts.URL+"/tasks/task-99", body)
		result := decodeJSON[Task](t, resp)

		if resp.StatusCode != 200 {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		if result.ID != "task-99" {
			t.Errorf("ID = %q, want %q", result.ID, "task-99")
		}
		if result.Name != "updated" {
			t.Errorf("Name = %q, want %q", result.Name, "updated")
		}
		if svc.lastUpdate == nil {
			t.Fatal("handler not called")
		}
		if svc.lastUpdate.ID != "task-99" {
			t.Errorf("captured ID = %q, want %q", svc.lastUpdate.ID, "task-99")
		}
		if svc.lastUpdate.Name != "updated" {
			t.Errorf("captured Name = %q, want %q", svc.lastUpdate.Name, "updated")
		}
		if svc.lastUpdate.Status != "closed" {
			t.Errorf("captured Status = %q, want %q", svc.lastUpdate.Status, "closed")
		}
	})

	t.Run("DeleteTask_PathParam", func(t *testing.T) {
		resp := doJSON(t, "DELETE", ts.URL+"/tasks/task-7", nil)
		resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		if svc.lastDelete == nil {
			t.Fatal("handler not called")
		}
		if svc.lastDelete.ID != "task-7" {
			t.Errorf("captured ID = %q, want %q", svc.lastDelete.ID, "task-7")
		}
	})
}

// ============================================================
// Integration tests: nested resources with annotations
// ============================================================

func TestIntegration_StdTransport_NestedResources(t *testing.T) {
	svc := &commentService{}
	ts := setupStdServer(t, svc, "")
	defer ts.Close()

	t.Run("ListComments_PathAndQuery", func(t *testing.T) {
		resp := doJSON(t, "GET", ts.URL+"/tasks/task-1/comments?sort=newest", nil)
		_ = decodeJSON[[]Comment](t, resp)

		if svc.lastList == nil {
			t.Fatal("handler not called")
		}
		if svc.lastList.TaskID != "task-1" {
			t.Errorf("TaskID = %q, want %q", svc.lastList.TaskID, "task-1")
		}
		if svc.lastList.Sort != "newest" {
			t.Errorf("Sort = %q, want %q", svc.lastList.Sort, "newest")
		}
	})

	t.Run("CreateComment_PathAndBody", func(t *testing.T) {
		body := map[string]string{"body": "great work!"}
		resp := doJSON(t, "POST", ts.URL+"/tasks/task-2/comments", body)
		result := decodeJSON[Comment](t, resp)

		if svc.lastCreate == nil {
			t.Fatal("handler not called")
		}
		if svc.lastCreate.TaskID != "task-2" {
			t.Errorf("TaskID = %q, want %q", svc.lastCreate.TaskID, "task-2")
		}
		if svc.lastCreate.Body != "great work!" {
			t.Errorf("Body = %q, want %q", svc.lastCreate.Body, "great work!")
		}
		if result.TaskID != "task-2" {
			t.Errorf("response TaskID = %q, want %q", result.TaskID, "task-2")
		}
	})

	t.Run("GetComment_TwoPathParams", func(t *testing.T) {
		resp := doJSON(t, "GET", ts.URL+"/tasks/task-3/comments/comment-5", nil)
		result := decodeJSON[Comment](t, resp)

		if svc.lastGet == nil {
			t.Fatal("handler not called")
		}
		if svc.lastGet.TaskID != "task-3" {
			t.Errorf("TaskID = %q, want %q", svc.lastGet.TaskID, "task-3")
		}
		if svc.lastGet.ID != "comment-5" {
			t.Errorf("ID = %q, want %q", svc.lastGet.ID, "comment-5")
		}
		if result.TaskID != "task-3" {
			t.Errorf("response TaskID = %q, want %q", result.TaskID, "task-3")
		}
	})

	t.Run("DeleteComment_TwoPathParams", func(t *testing.T) {
		resp := doJSON(t, "DELETE", ts.URL+"/tasks/task-4/comments/comment-9", nil)
		resp.Body.Close()

		if svc.lastDelete == nil {
			t.Fatal("handler not called")
		}
		if svc.lastDelete.TaskID != "task-4" {
			t.Errorf("TaskID = %q, want %q", svc.lastDelete.TaskID, "task-4")
		}
		if svc.lastDelete.ID != "comment-9" {
			t.Errorf("ID = %q, want %q", svc.lastDelete.ID, "comment-9")
		}
	})
}

// ============================================================
// Integration tests: with path prefix
// ============================================================

func TestIntegration_StdTransport_WithPrefix(t *testing.T) {
	svc := &taskService{}
	ts := setupStdServer(t, svc, "/api/v1")
	defer ts.Close()

	t.Run("CreateTask_WithPrefix", func(t *testing.T) {
		resp := doJSON(t, "POST", ts.URL+"/api/v1/task", CreateTaskRequest{Name: "prefixed"})
		result := decodeJSON[Task](t, resp)

		if resp.StatusCode != 200 {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		if result.Name != "prefixed" {
			t.Errorf("Name = %q, want %q", result.Name, "prefixed")
		}
	})

	t.Run("GetTask_WithPrefix", func(t *testing.T) {
		resp := doJSON(t, "GET", ts.URL+"/api/v1/tasks/abc", nil)
		result := decodeJSON[Task](t, resp)

		if result.ID != "abc" {
			t.Errorf("ID = %q, want %q", result.ID, "abc")
		}
	})

	t.Run("ListTasks_WithPrefix_QueryParams", func(t *testing.T) {
		resp := doJSON(t, "GET", ts.URL+"/api/v1/tasks?status=pending&page=1&limit=5", nil)
		_ = decodeJSON[[]Task](t, resp)

		if svc.lastList.Status != "pending" {
			t.Errorf("Status = %q, want %q", svc.lastList.Status, "pending")
		}
		if svc.lastList.Page != 1 {
			t.Errorf("Page = %d, want 1", svc.lastList.Page)
		}
	})
}

// ============================================================
// Integration tests: middleware + params
// ============================================================

func TestIntegration_StdTransport_MiddlewareAndParams(t *testing.T) {
	svc := &taskService{}

	cfg := x.TypedLazyConfig{Config: json.RawMessage(`{"addr":":0"}`)}
	transport, err := stdhttp.NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	var middlewareSawEndpoint string
	mw := talk.MiddlewareFunc(func(next talk.EndpointFunc) talk.EndpointFunc {
		return func(ctx context.Context, req any) (any, error) {
			ep := talk.EndpointFromContext(ctx)
			if ep != nil {
				middlewareSawEndpoint = ep.Name
			}
			return next(ctx, req)
		}
	})

	extractor := extract.NewReflectExtractor()
	endpoints, err := extractor.Extract(svc)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}

	// Add middleware to all endpoints
	for _, ep := range endpoints {
		ep.Middleware = append(ep.Middleware, mw)
	}

	transport.RegisterEndpoints(endpoints)
	ts := httptest.NewServer(transport.ServeMux())
	defer ts.Close()

	resp := doJSON(t, "GET", ts.URL+"/tasks/mw-test", nil)
	resp.Body.Close()

	if middlewareSawEndpoint != "GetTask" {
		t.Errorf("middleware saw endpoint %q, want %q", middlewareSawEndpoint, "GetTask")
	}
	if svc.lastGet == nil {
		t.Fatal("handler not called")
	}
	if svc.lastGet.ID != "mw-test" {
		t.Errorf("ID = %q, want %q", svc.lastGet.ID, "mw-test")
	}
}

// ============================================================
// Integration tests: extractor derived paths (no annotations)
// ============================================================

// plainService has NO TalkAnnotations — tests pure reflection-based path derivation
type plainService struct {
	lastCreate *CreateTaskRequest
	lastList   *ListTasksRequest
}

func (s *plainService) CreateTask(ctx context.Context, req CreateTaskRequest) (*Task, error) {
	s.lastCreate = &req
	return &Task{ID: "p1", Name: req.Name}, nil
}

func (s *plainService) ListTasks(ctx context.Context, req ListTasksRequest) ([]*Task, error) {
	s.lastList = &req
	return []*Task{}, nil
}

func TestIntegration_StdTransport_DerivedPaths(t *testing.T) {
	svc := &plainService{}

	cfg := x.TypedLazyConfig{Config: json.RawMessage(`{"addr":":0"}`)}
	transport, err := stdhttp.NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	extractor := extract.NewReflectExtractor()
	endpoints, err := extractor.Extract(svc)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}

	// Verify derived paths
	epMap := make(map[string]*talk.Endpoint)
	for _, ep := range endpoints {
		epMap[ep.Name] = ep
	}

	if ep, ok := epMap["CreateTask"]; ok {
		if ep.Method != "POST" || ep.Path != "/task" {
			t.Errorf("CreateTask: %s %s, want POST /task", ep.Method, ep.Path)
		}
	} else {
		t.Fatal("CreateTask not extracted")
	}

	if ep, ok := epMap["ListTasks"]; ok {
		if ep.Method != "GET" || ep.Path != "/tasks" {
			t.Errorf("ListTasks: %s %s, want GET /tasks", ep.Method, ep.Path)
		}
	} else {
		t.Fatal("ListTasks not extracted")
	}

	transport.RegisterEndpoints(endpoints)
	ts := httptest.NewServer(transport.ServeMux())
	defer ts.Close()

	t.Run("CreateTask_DerivedPath", func(t *testing.T) {
		resp := doJSON(t, "POST", ts.URL+"/task", CreateTaskRequest{Name: "derived"})
		result := decodeJSON[Task](t, resp)

		if resp.StatusCode != 200 {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		if result.Name != "derived" {
			t.Errorf("Name = %q, want %q", result.Name, "derived")
		}
	})

	t.Run("ListTasks_DerivedPath_WithQuery", func(t *testing.T) {
		resp := doJSON(t, "GET", ts.URL+"/tasks?status=active&page=2&limit=10", nil)
		_ = decodeJSON[[]Task](t, resp)

		if svc.lastList == nil {
			t.Fatal("handler not called")
		}
		if svc.lastList.Status != "active" {
			t.Errorf("Status = %q, want %q", svc.lastList.Status, "active")
		}
		if svc.lastList.Page != 2 {
			t.Errorf("Page = %d, want 2", svc.lastList.Page)
		}
	})
}

// ============================================================
// Integration tests: client-server round trip
// ============================================================

func TestIntegration_ClientServer_RoundTrip(t *testing.T) {
	svc := &taskService{}
	ts := setupStdServer(t, svc, "")
	defer ts.Close()

	cfg := x.TypedLazyConfig{Config: json.RawMessage(fmt.Sprintf(`{"addr":%q}`, ts.URL))}
	clientTransport, err := stdhttp.NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	client := talk.NewClient(clientTransport)
	defer client.Close()

	t.Run("Call_CreateTask", func(t *testing.T) {
		var resp Task
		err := client.Call(context.Background(), "CreateTask", &CreateTaskRequest{Name: "roundtrip", Status: "open"}, &resp)
		if err != nil {
			t.Fatalf("Call: %v", err)
		}
		if resp.Name != "roundtrip" {
			t.Errorf("Name = %q, want %q", resp.Name, "roundtrip")
		}
		if svc.lastCreate.Name != "roundtrip" {
			t.Errorf("captured Name = %q, want %q", svc.lastCreate.Name, "roundtrip")
		}
	})

	t.Run("Call_DirectPath", func(t *testing.T) {
		var resp Task
		err := client.Call(context.Background(), "/task", &CreateTaskRequest{Name: "direct"}, &resp)
		if err != nil {
			t.Fatalf("Call: %v", err)
		}
		if resp.Name != "direct" {
			t.Errorf("Name = %q, want %q", resp.Name, "direct")
		}
	})
}

// ============================================================
// Integration: talk.NewServer + Register() full chain
// ============================================================

func TestIntegration_TalkServer_FullChain(t *testing.T) {
	svc := &taskService{}

	cfg := x.TypedLazyConfig{Config: json.RawMessage(`{"addr":":0"}`)}
	transport, err := stdhttp.NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	server := talk.NewServer(transport,
		talk.WithExtractor(extract.NewReflectExtractor()),
	)

	if err := server.Register(svc); err != nil {
		t.Fatalf("Register: %v", err)
	}

	// Verify endpoints extracted
	endpoints := server.Endpoints()
	if len(endpoints) != 5 {
		t.Fatalf("expected 5 endpoints, got %d", len(endpoints))
	}

	// Manually register to transport and test
	transport.RegisterEndpoints(endpoints)
	ts := httptest.NewServer(transport.ServeMux())
	defer ts.Close()

	// POST /task
	resp := doJSON(t, "POST", ts.URL+"/task", CreateTaskRequest{Name: "fullchain"})
	result := decodeJSON[Task](t, resp)
	if result.Name != "fullchain" {
		t.Errorf("Name = %q, want %q", result.Name, "fullchain")
	}

	// GET /tasks/fc-1 (annotated path)
	resp = doJSON(t, "GET", ts.URL+"/tasks/fc-1", nil)
	result = decodeJSON[Task](t, resp)
	if result.ID != "fc-1" {
		t.Errorf("ID = %q, want %q", result.ID, "fc-1")
	}
}

// ============================================================
// Integration: talk.NewServer with prefix via Register
// ============================================================

func TestIntegration_TalkServer_RegisterWithPrefix(t *testing.T) {
	svc := &taskService{}

	cfg := x.TypedLazyConfig{Config: json.RawMessage(`{"addr":":0"}`)}
	transport, err := stdhttp.NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	server := talk.NewServer(transport,
		talk.WithExtractor(extract.NewReflectExtractor()),
		talk.WithPathPrefix("/api/v2"),
	)

	if err := server.Register(svc); err != nil {
		t.Fatalf("Register: %v", err)
	}

	transport.RegisterEndpoints(server.Endpoints())
	ts := httptest.NewServer(transport.ServeMux())
	defer ts.Close()

	// Verify prefixed path works
	resp := doJSON(t, "POST", ts.URL+"/api/v2/task", CreateTaskRequest{Name: "v2"})
	result := decodeJSON[Task](t, resp)
	if result.Name != "v2" {
		t.Errorf("Name = %q, want %q", result.Name, "v2")
	}

	resp = doJSON(t, "GET", ts.URL+"/api/v2/tasks?status=new&page=1&limit=50", nil)
	_ = decodeJSON[[]Task](t, resp)
	if svc.lastList.Status != "new" {
		t.Errorf("Status = %q, want %q", svc.lastList.Status, "new")
	}
}

// ============================================================
// Integration: RequestType verification in extracted endpoints
// ============================================================

func TestIntegration_Extractor_RequestTypes(t *testing.T) {
	svc := &taskService{}
	extractor := extract.NewReflectExtractor()
	endpoints, err := extractor.Extract(svc)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}

	epMap := make(map[string]*talk.Endpoint)
	for _, ep := range endpoints {
		epMap[ep.Name] = ep
	}

	tests := []struct {
		name        string
		wantReqType reflect.Type
	}{
		{"CreateTask", reflect.TypeOf(CreateTaskRequest{})},
		{"GetTask", reflect.TypeOf(GetTaskRequest{})},
		{"ListTasks", reflect.TypeOf(ListTasksRequest{})},
		{"UpdateTask", reflect.TypeOf(UpdateTaskRequest{})},
		{"DeleteTask", reflect.TypeOf(DeleteTaskRequest{})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep, ok := epMap[tt.name]
			if !ok {
				t.Fatalf("endpoint %s not found", tt.name)
			}
			if ep.RequestType != tt.wantReqType {
				t.Errorf("RequestType = %v, want %v", ep.RequestType, tt.wantReqType)
			}
		})
	}
}
