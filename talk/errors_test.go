package talk

import (
	"net/http"
	"testing"
)

func TestErrorCode_HTTPStatus(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected int
	}{
		{OK, http.StatusOK},
		{InvalidArgument, http.StatusBadRequest},
		{NotFound, http.StatusNotFound},
		{PermissionDenied, http.StatusForbidden},
		{Unauthenticated, http.StatusUnauthorized},
		{Internal, http.StatusInternalServerError},
		{Unavailable, http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.code.String(), func(t *testing.T) {
			if got := tt.code.HTTPStatus(); got != tt.expected {
				t.Errorf("HTTPStatus() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestErrorCode_GRPCCode(t *testing.T) {
	for code := OK; code <= Unauthenticated; code++ {
		if got := code.GRPCCode(); got != uint32(code) {
			t.Errorf("GRPCCode(%d) = %d, want %d", code, got, code)
		}
	}
}

func TestFromHTTPStatus(t *testing.T) {
	tests := []struct {
		status   int
		expected ErrorCode
	}{
		{http.StatusOK, OK},
		{http.StatusCreated, OK},
		{http.StatusBadRequest, InvalidArgument},
		{http.StatusUnauthorized, Unauthenticated},
		{http.StatusForbidden, PermissionDenied},
		{http.StatusNotFound, NotFound},
		{http.StatusConflict, AlreadyExists},
		{http.StatusInternalServerError, Internal},
		{http.StatusServiceUnavailable, Unavailable},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			if got := FromHTTPStatus(tt.status); got != tt.expected {
				t.Errorf("FromHTTPStatus(%d) = %v, want %v", tt.status, got, tt.expected)
			}
		})
	}
}

func TestError_Error(t *testing.T) {
	err := NewError(NotFound, "user not found")
	expected := "NOT_FOUND: user not found"
	if got := err.Error(); got != expected {
		t.Errorf("Error() = %q, want %q", got, expected)
	}
}

func TestError_WithDetails(t *testing.T) {
	details := map[string]string{"field": "email", "reason": "invalid format"}
	err := NewErrorWithDetails(InvalidArgument, "validation failed", details)

	if err.Code != InvalidArgument {
		t.Errorf("Code = %v, want %v", err.Code, InvalidArgument)
	}
	if err.Details == nil {
		t.Error("Details should not be nil")
	}
}

func TestToError(t *testing.T) {
	talkErr := NewError(NotFound, "not found")
	if got := ToError(talkErr); got != talkErr {
		t.Error("ToError should return the same talk.Error")
	}

	stdErr := http.ErrBodyNotAllowed
	converted := ToError(stdErr)
	if converted.Code != Unknown {
		t.Errorf("Code = %v, want Unknown", converted.Code)
	}
	if converted.Message != stdErr.Error() {
		t.Errorf("Message = %q, want %q", converted.Message, stdErr.Error())
	}
}

func TestIsError(t *testing.T) {
	talkErr := NewError(NotFound, "not found")
	if e, ok := IsError(talkErr); !ok || e != talkErr {
		t.Error("IsError should return true for talk.Error")
	}

	if _, ok := IsError(http.ErrBodyNotAllowed); ok {
		t.Error("IsError should return false for non-talk.Error")
	}

	if _, ok := IsError(nil); ok {
		t.Error("IsError should return false for nil")
	}
}
