package httputil

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
)

func TestNewResponse(t *testing.T) {
	recorder := httptest.NewRecorder()
	resp := NewResponse(recorder)
	if resp == nil {
		t.Fatal("NewResponse() returned nil")
	}
}

func TestNewResponse_WithFields(t *testing.T) {
	recorder := httptest.NewRecorder()
	resp := NewResponse(recorder, Code(200), Data("test"))
	if resp.Code != 200 {
		t.Errorf("Code = %d, want 200", resp.Code)
	}
	if resp.Data != "test" {
		t.Errorf("Data = %v, want 'test'", resp.Data)
	}
}

func TestResponse_WithCode(t *testing.T) {
	recorder := httptest.NewRecorder()
	resp := NewResponse(recorder).WithCode(404)
	if resp.Code != 404 {
		t.Errorf("WithCode() Code = %d, want 404", resp.Code)
	}
}

func TestResponse_WithData(t *testing.T) {
	recorder := httptest.NewRecorder()
	resp := NewResponse(recorder).WithData(map[string]int{"count": 42})
	if resp.Data == nil {
		t.Error("WithData() Data should not be nil")
	}
}

func TestResponse_WithError(t *testing.T) {
	recorder := httptest.NewRecorder()
	err := errors.New("test error")
	resp := NewResponse(recorder).WithError(err)
	if resp.err != err {
		t.Error("WithError() err not set correctly")
	}
}

func TestResponse_WithErrorf(t *testing.T) {
	recorder := httptest.NewRecorder()
	resp := NewResponse(recorder).WithErrorf("error: %s", "test")
	if resp.Error != "error: test" {
		t.Errorf("WithErrorf() Error = %q, want 'error: test'", resp.Error)
	}
}

func TestResponse_WithDataOrErr(t *testing.T) {
	recorder := httptest.NewRecorder()
	resp := NewResponse(recorder).WithDataOrErr("data", nil)
	if resp.Data != "data" {
		t.Errorf("WithDataOrErr() Data = %v, want 'data'", resp.Data)
	}

	err := errors.New("error")
	resp2 := NewResponse(recorder).WithDataOrErr(nil, err)
	if resp2.err != err {
		t.Error("WithDataOrErr() err not set correctly")
	}
}

func TestResponse_Flush_Success(t *testing.T) {
	recorder := httptest.NewRecorder()
	resp := NewResponse(recorder).WithData("success data")
	resp.Flush()

	if recorder.Code != 200 {
		t.Errorf("Flush() status code = %d, want 200", recorder.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &result)
	if result["status"] != StatusSuccess {
		t.Errorf("Flush() status = %v, want %v", result["status"], StatusSuccess)
	}
}

func TestResponse_Flush_WithError(t *testing.T) {
	recorder := httptest.NewRecorder()
	resp := NewResponse(recorder).WithError(errors.New("test error"))
	resp.Flush()

	var result map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &result)
	if result["status"] != StatusFailed {
		t.Errorf("Flush() with error status = %v, want %v", result["status"], StatusFailed)
	}
	if result["error"] != "test error" {
		t.Errorf("Flush() error = %v, want 'test error'", result["error"])
	}
}

func TestResponse_Flush_WithCode(t *testing.T) {
	recorder := httptest.NewRecorder()
	resp := NewResponse(recorder).WithCode(uint32(CodeNotFound))
	resp.Flush()

	if recorder.Code != 404 {
		t.Errorf("Flush() with CodeNotFound status = %d, want 404", recorder.Code)
	}
}

func TestResponse_Flush_WithFields(t *testing.T) {
	recorder := httptest.NewRecorder()
	resp := NewResponse(recorder)
	resp.Flush(Code(uint32(CodeOK)), Data("created"))

	if recorder.Code != 200 {
		t.Errorf("Flush() with fields status = %d, want 200", recorder.Code)
	}
}

func TestCode(t *testing.T) {
	f := Code(500)
	resp := &Response{}
	f(resp)
	if resp.Code != 500 {
		t.Errorf("Code() = %d, want 500", resp.Code)
	}
}

func TestData(t *testing.T) {
	f := Data("test data")
	resp := &Response{}
	f(resp)
	if resp.Data != "test data" {
		t.Errorf("Data() = %v, want 'test data'", resp.Data)
	}
}

func TestErrorf(t *testing.T) {
	f := Errorf("error: %d", 123)
	resp := &Response{}
	f(resp)
	if resp.Error != "error: 123" {
		t.Errorf("Errorf() = %q, want 'error: 123'", resp.Error)
	}
}

func TestError(t *testing.T) {
	err := errors.New("test error")
	f := Error(err)
	resp := &Response{}
	f(resp)
	if resp.err != err {
		t.Error("Error() err not set correctly")
	}
}
