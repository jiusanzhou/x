package clientlb

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewEndpoint(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{"valid http", "http://localhost:8080", false},
		{"valid https", "https://api.example.com", false},
		{"with path", "http://localhost/api/v1", false},
		{"invalid url", "://invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep, err := NewEndpoint(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && ep == nil {
				t.Error("NewEndpoint() returned nil without error")
			}
		})
	}
}

func TestEndpoint_Apply(t *testing.T) {
	ep, _ := NewEndpoint("http://newhost:9090")
	req, _ := http.NewRequest("GET", "http://originalhost:8080/path", nil)

	ep.Apply(req)

	if req.URL.Host != "newhost:9090" {
		t.Errorf("Apply() req.URL.Host = %q, want %q", req.URL.Host, "newhost:9090")
	}
}

func TestNewSimpleHealthCheck_Prepare(t *testing.T) {
	checker := NewSimpleHealthCheck("POST", "/health", "")
	ep, _ := NewEndpoint("http://localhost:8080")
	req, _ := http.NewRequest("GET", "http://localhost:8080/other", nil)

	result := checker(ep, req, nil)

	if !result {
		t.Error("HealthChecker preparation should return true")
	}
	if req.URL.Path != "/health" {
		t.Errorf("HealthChecker should modify path to %q, got %q", "/health", req.URL.Path)
	}
	if req.Method != "POST" {
		t.Errorf("HealthChecker should modify method to %q, got %q", "POST", req.Method)
	}
}

func TestNewSimpleHealthCheck_SuccessStatus(t *testing.T) {
	checker := NewSimpleHealthCheck("GET", "/health", "")
	ep, _ := NewEndpoint("http://localhost:8080")
	req, _ := http.NewRequest("GET", "http://localhost:8080/health", nil)

	recorder := httptest.NewRecorder()
	recorder.WriteHeader(http.StatusOK)
	resp := recorder.Result()

	result := checker(ep, req, resp)

	if !result {
		t.Error("HealthChecker should return true for 200 status")
	}
}

func TestNewSimpleHealthCheck_FailureStatus(t *testing.T) {
	checker := NewSimpleHealthCheck("GET", "/health", "")
	ep, _ := NewEndpoint("http://localhost:8080")
	req, _ := http.NewRequest("GET", "http://localhost:8080/health", nil)

	recorder := httptest.NewRecorder()
	recorder.WriteHeader(http.StatusInternalServerError)
	resp := recorder.Result()

	result := checker(ep, req, resp)

	if result {
		t.Error("HealthChecker should return false for 500 status")
	}
}

func TestNewSimpleHealthCheck_ContentMatch(t *testing.T) {
	checker := NewSimpleHealthCheck("GET", "/health", "ok")
	ep, _ := NewEndpoint("http://localhost:8080")
	req, _ := http.NewRequest("GET", "http://localhost:8080/health", nil)

	recorder := httptest.NewRecorder()
	recorder.WriteHeader(http.StatusOK)
	recorder.WriteString("ok")
	resp := recorder.Result()

	result := checker(ep, req, resp)

	if !result {
		t.Error("HealthChecker should return true when content matches")
	}
}

func TestNewSimpleHealthCheck_ContentMismatch(t *testing.T) {
	checker := NewSimpleHealthCheck("GET", "/health", "ok")
	ep, _ := NewEndpoint("http://localhost:8080")
	req, _ := http.NewRequest("GET", "http://localhost:8080/health", nil)

	recorder := httptest.NewRecorder()
	recorder.WriteHeader(http.StatusOK)
	recorder.WriteString("not ok")
	resp := recorder.Result()

	result := checker(ep, req, resp)

	if result {
		t.Error("HealthChecker should return false when content doesn't match")
	}
}
