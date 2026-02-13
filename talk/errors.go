// Package talk provides a transport abstraction layer for building
// protocol-agnostic services. It allows defining service methods once
// and exposing them over HTTP, gRPC, WebSocket, or other transports.
package talk

import (
	"fmt"
	"net/http"
)

// ErrorCode represents a canonical error code that can be mapped
// to various protocol-specific error codes (HTTP status, gRPC code, WebSocket close code).
// The codes are modeled after gRPC status codes for consistency.
type ErrorCode int

const (
	// OK indicates the operation completed successfully.
	OK ErrorCode = 0

	// Cancelled indicates the operation was cancelled (typically by the caller).
	Cancelled ErrorCode = 1

	// Unknown indicates an unknown error occurred.
	Unknown ErrorCode = 2

	// InvalidArgument indicates the client specified an invalid argument.
	InvalidArgument ErrorCode = 3

	// DeadlineExceeded indicates the deadline expired before the operation could complete.
	DeadlineExceeded ErrorCode = 4

	// NotFound indicates the requested resource was not found.
	NotFound ErrorCode = 5

	// AlreadyExists indicates the resource already exists.
	AlreadyExists ErrorCode = 6

	// PermissionDenied indicates the caller does not have permission.
	PermissionDenied ErrorCode = 7

	// ResourceExhausted indicates some resource has been exhausted (e.g., quota).
	ResourceExhausted ErrorCode = 8

	// FailedPrecondition indicates the operation was rejected because the system
	// is not in a state required for the operation's execution.
	FailedPrecondition ErrorCode = 9

	// Aborted indicates the operation was aborted.
	Aborted ErrorCode = 10

	// OutOfRange indicates the operation was attempted past the valid range.
	OutOfRange ErrorCode = 11

	// Unimplemented indicates the operation is not implemented.
	Unimplemented ErrorCode = 12

	// Internal indicates an internal error occurred.
	Internal ErrorCode = 13

	// Unavailable indicates the service is currently unavailable.
	Unavailable ErrorCode = 14

	// DataLoss indicates unrecoverable data loss or corruption.
	DataLoss ErrorCode = 15

	// Unauthenticated indicates the request does not have valid authentication credentials.
	Unauthenticated ErrorCode = 16
)

// String returns the string representation of the error code.
func (c ErrorCode) String() string {
	switch c {
	case OK:
		return "OK"
	case Cancelled:
		return "CANCELLED"
	case Unknown:
		return "UNKNOWN"
	case InvalidArgument:
		return "INVALID_ARGUMENT"
	case DeadlineExceeded:
		return "DEADLINE_EXCEEDED"
	case NotFound:
		return "NOT_FOUND"
	case AlreadyExists:
		return "ALREADY_EXISTS"
	case PermissionDenied:
		return "PERMISSION_DENIED"
	case ResourceExhausted:
		return "RESOURCE_EXHAUSTED"
	case FailedPrecondition:
		return "FAILED_PRECONDITION"
	case Aborted:
		return "ABORTED"
	case OutOfRange:
		return "OUT_OF_RANGE"
	case Unimplemented:
		return "UNIMPLEMENTED"
	case Internal:
		return "INTERNAL"
	case Unavailable:
		return "UNAVAILABLE"
	case DataLoss:
		return "DATA_LOSS"
	case Unauthenticated:
		return "UNAUTHENTICATED"
	default:
		return fmt.Sprintf("CODE(%d)", c)
	}
}

// HTTPStatus returns the HTTP status code corresponding to this error code.
func (c ErrorCode) HTTPStatus() int {
	switch c {
	case OK:
		return http.StatusOK
	case Cancelled:
		return http.StatusRequestTimeout // 408, client cancelled
	case Unknown:
		return http.StatusInternalServerError
	case InvalidArgument:
		return http.StatusBadRequest
	case DeadlineExceeded:
		return http.StatusGatewayTimeout
	case NotFound:
		return http.StatusNotFound
	case AlreadyExists:
		return http.StatusConflict
	case PermissionDenied:
		return http.StatusForbidden
	case ResourceExhausted:
		return http.StatusTooManyRequests
	case FailedPrecondition:
		return http.StatusPreconditionFailed
	case Aborted:
		return http.StatusConflict
	case OutOfRange:
		return http.StatusBadRequest
	case Unimplemented:
		return http.StatusNotImplemented
	case Internal:
		return http.StatusInternalServerError
	case Unavailable:
		return http.StatusServiceUnavailable
	case DataLoss:
		return http.StatusInternalServerError
	case Unauthenticated:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

// GRPCCode returns the gRPC status code corresponding to this error code.
// The values match google.golang.org/grpc/codes.Code directly.
func (c ErrorCode) GRPCCode() uint32 {
	// ErrorCode values are designed to match gRPC codes directly
	return uint32(c)
}

// WSCloseCode returns the WebSocket close code corresponding to this error code.
// Uses standard WebSocket close codes (RFC 6455) where applicable.
func (c ErrorCode) WSCloseCode() int {
	switch c {
	case OK:
		return 1000 // Normal Closure
	case InvalidArgument, OutOfRange:
		return 1003 // Unsupported Data
	case PermissionDenied, Unauthenticated:
		return 1008 // Policy Violation
	case Internal, Unknown, DataLoss:
		return 1011 // Internal Error
	case Unavailable:
		return 1013 // Try Again Later (not standard, but commonly used)
	case Cancelled:
		return 1001 // Going Away
	default:
		return 1011 // Internal Error
	}
}

// Error represents a Talk error with a code, message, and optional details.
type Error struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details any       `json:"details,omitempty"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s", e.Code.String(), e.Message)
	}
	return e.Code.String()
}

// HTTPStatus returns the HTTP status code for this error.
func (e *Error) HTTPStatus() int {
	return e.Code.HTTPStatus()
}

// GRPCCode returns the gRPC code for this error.
func (e *Error) GRPCCode() uint32 {
	return e.Code.GRPCCode()
}

// WSCloseCode returns the WebSocket close code for this error.
func (e *Error) WSCloseCode() int {
	return e.Code.WSCloseCode()
}

// NewError creates a new Error with the given code and message.
func NewError(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// NewErrorf creates a new Error with the given code and formatted message.
func NewErrorf(code ErrorCode, format string, args ...any) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// NewErrorWithDetails creates a new Error with the given code, message, and details.
func NewErrorWithDetails(code ErrorCode, message string, details any) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// FromHTTPStatus converts an HTTP status code to an ErrorCode.
func FromHTTPStatus(status int) ErrorCode {
	switch status {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent:
		return OK
	case http.StatusBadRequest:
		return InvalidArgument
	case http.StatusUnauthorized:
		return Unauthenticated
	case http.StatusForbidden:
		return PermissionDenied
	case http.StatusNotFound:
		return NotFound
	case http.StatusConflict:
		return AlreadyExists
	case http.StatusPreconditionFailed:
		return FailedPrecondition
	case http.StatusTooManyRequests:
		return ResourceExhausted
	case http.StatusRequestTimeout:
		return Cancelled
	case http.StatusGatewayTimeout:
		return DeadlineExceeded
	case http.StatusNotImplemented:
		return Unimplemented
	case http.StatusServiceUnavailable:
		return Unavailable
	default:
		if status >= 400 && status < 500 {
			return InvalidArgument
		}
		return Internal
	}
}

// FromGRPCCode converts a gRPC status code to an ErrorCode.
func FromGRPCCode(code uint32) ErrorCode {
	if code <= 16 {
		return ErrorCode(code)
	}
	return Unknown
}

// FromWSCloseCode converts a WebSocket close code to an ErrorCode.
func FromWSCloseCode(code int) ErrorCode {
	switch code {
	case 1000:
		return OK
	case 1001:
		return Cancelled
	case 1003:
		return InvalidArgument
	case 1008:
		return PermissionDenied
	case 1011:
		return Internal
	case 1013:
		return Unavailable
	default:
		return Unknown
	}
}

// IsError checks if an error is a Talk Error and returns it.
func IsError(err error) (*Error, bool) {
	if err == nil {
		return nil, false
	}
	if e, ok := err.(*Error); ok {
		return e, true
	}
	return nil, false
}

// ToError converts any error to a Talk Error.
// If the error is already a Talk Error, it returns it directly.
// Otherwise, it wraps the error with the Unknown code.
func ToError(err error) *Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		return e
	}
	return NewError(Unknown, err.Error())
}
