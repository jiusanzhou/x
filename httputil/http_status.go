package httputil

import (
	"net/http"
)

// A StatusCode is an unsigned 32-bit error code as defined in the gRPC spec.
type StatusCode uint32

const (
	// CodeOK is returned on success.
	CodeOK StatusCode = 0

	// CodeCanceled indicates the operation was canceled (typically by the caller).
	CodeCanceled StatusCode = 1

	// CodeUnknown error. An example of where this error may be returned is
	// if a Status value received from another address space belongs to
	// an error-space that is not known in this address space. Also
	// errors raised by APIs that do not return enough error information
	// may be converted to this error.
	CodeUnknown StatusCode = 2

	// CodeInvalidArgument indicates client specified an invalid argument.
	// Note that this differs from FailedPrecondition. It indicates arguments
	// that are problematic regardless of the state of the system
	// (e.g., a malformed file name).
	CodeInvalidArgument StatusCode = 3

	// CodeDeadlineExceeded means operation expired before completion.
	// For operations that change the state of the system, this error may be
	// returned even if the operation has completed successfully. For
	// example, a successful response from a server could have been delayed
	// long enough for the deadline to expire.
	CodeDeadlineExceeded StatusCode = 4

	// CodeNotFound means some requested entity (e.g., file or directory) was
	// not found.
	CodeNotFound StatusCode = 5

	// CodeAlreadyExists means an attempt to create an entity failed because one
	// already exists.
	CodeAlreadyExists StatusCode = 6

	// CodePermissionDenied indicates the caller does not have permission to
	// execute the specified operation. It must not be used for rejections
	// caused by exhausting some resource (use ResourceExhausted
	// instead for those errors). It must not be
	// used if the caller cannot be identified (use Unauthenticated
	// instead for those errors).
	CodePermissionDenied StatusCode = 7

	// CodeResourceExhausted indicates some resource has been exhausted, perhaps
	// a per-user quota, or perhaps the entire file system is out of space.
	CodeResourceExhausted StatusCode = 8

	// CodeFailedPrecondition indicates operation was rejected because the
	// system is not in a state required for the operation's execution.
	// For example, directory to be deleted may be non-empty, an rmdir
	// operation is applied to a non-directory, etc.
	//
	// A litmus test that may help a service implementor in deciding
	// between FailedPrecondition, Aborted, and Unavailable:
	//  (a) Use Unavailable if the client can retry just the failing call.
	//  (b) Use Aborted if the client should retry at a higher-level
	//      (e.g., restarting a read-modify-write sequence).
	//  (c) Use FailedPrecondition if the client should not retry until
	//      the system state has been explicitly fixed. E.g., if an "rmdir"
	//      fails because the directory is non-empty, FailedPrecondition
	//      should be returned since the client should not retry unless
	//      they have first fixed up the directory by deleting files from it.
	//  (d) Use FailedPrecondition if the client performs conditional
	//      REST Get/Update/Delete on a resource and the resource on the
	//      server does not match the condition. E.g., conflicting
	//      read-modify-write on the same resource.
	CodeFailedPrecondition StatusCode = 9

	// CodeAborted indicates the operation was aborted, typically due to a
	// concurrency issue like sequencer check failures, transaction aborts,
	// etc.
	//
	// See litmus test above for deciding between FailedPrecondition,
	// Aborted, and Unavailable.
	CodeAborted StatusCode = 10

	// CodeOutOfRange means operation was attempted past the valid range.
	// E.g., seeking or reading past end of file.
	//
	// Unlike InvalidArgument, this error indicates a problem that may
	// be fixed if the system state changes. For example, a 32-bit file
	// system will generate InvalidArgument if asked to read at an
	// offset that is not in the range [0,2^32-1], but it will generate
	// OutOfRange if asked to read from an offset past the current
	// file size.
	//
	// There is a fair bit of overlap between FailedPrecondition and
	// OutOfRange. We recommend using OutOfRange (the more specific
	// error) when it applies so that callers who are iterating through
	// a space can easily look for an OutOfRange error to detect when
	// they are done.
	CodeOutOfRange StatusCode = 11

	// CodeUnimplemented indicates operation is not implemented or not
	// supported/enabled in this service.
	CodeUnimplemented StatusCode = 12

	// CodeInternal errors. Means some invariants expected by underlying
	// system has been broken. If you see one of these errors,
	// something is very broken.
	CodeInternal StatusCode = 13

	// CodeUnavailable indicates the service is currently unavailable.
	// This is a most likely a transient condition and may be corrected
	// by retrying with a backoff. Note that it is not always safe to retry
	// non-idempotent operations.
	//
	// See litmus test above for deciding between FailedPrecondition,
	// Aborted, and Unavailable.
	CodeUnavailable StatusCode = 14

	// CodeDataLoss indicates unrecoverable data loss or corruption.
	CodeDataLoss StatusCode = 15

	// CodeUnauthenticated indicates the request does not have valid
	// authentication credentials for the operation.
	CodeUnauthenticated StatusCode = 16

	_maxCode = 17
)

// defaultHTTPStatusFromCode converts a gRPC error code into the corresponding HTTP response status.
// See: https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto
func defaultHTTPStatusFromCode(code StatusCode) int {
	switch code {
	case CodeOK:
		return http.StatusOK
	case CodeCanceled:
		return http.StatusRequestTimeout
	case CodeUnknown:
		return http.StatusInternalServerError
	case CodeInvalidArgument:
		return http.StatusBadRequest
	case CodeDeadlineExceeded:
		return http.StatusGatewayTimeout
	case CodeNotFound:
		return http.StatusNotFound
	case CodeAlreadyExists:
		return http.StatusConflict
	case CodePermissionDenied:
		return http.StatusForbidden
	case CodeUnauthenticated:
		return http.StatusUnauthorized
	case CodeResourceExhausted:
		return http.StatusTooManyRequests
	case CodeFailedPrecondition:
		return http.StatusPreconditionFailed
	case CodeAborted:
		return http.StatusConflict
	case CodeOutOfRange:
		return http.StatusBadRequest
	case CodeUnimplemented:
		return http.StatusNotImplemented
	case CodeInternal:
		return http.StatusInternalServerError
	case CodeUnavailable:
		return http.StatusServiceUnavailable
	case CodeDataLoss:
		return http.StatusInternalServerError
	}

	return http.StatusInternalServerError
}

var httpStatusFromCode = defaultHTTPStatusFromCode

// HTTPStatusFromCode set http status from code function
func HTTPStatusFromCode(fn func(code StatusCode) int) {
	var oldfn = httpStatusFromCode
	httpStatusFromCode = func(code StatusCode) int {
		var v = oldfn(code)
		if v != http.StatusInternalServerError {
			return v
		}
		return fn(code)
	}
}
