package talk

import (
	"context"
	"reflect"
)

// StreamMode indicates the streaming behavior of an endpoint.
type StreamMode int

const (
	StreamNone       StreamMode = iota // Request-response (no streaming)
	StreamClientSide                   // Client streams to server (param contains <-chan)
	StreamServerSide                   // Server streams to client (return contains <-chan)
	StreamBidirect                     // Bidirectional streaming (both contain chan)
)

func (m StreamMode) String() string {
	switch m {
	case StreamNone:
		return "none"
	case StreamClientSide:
		return "client"
	case StreamServerSide:
		return "server"
	case StreamBidirect:
		return "bidirectional"
	default:
		return "unknown"
	}
}

// EndpointFunc is the unified handler signature for all endpoints.
// It receives a context and a request, returning a response and error.
type EndpointFunc func(ctx context.Context, request any) (response any, err error)

// StreamEndpointFunc is the handler signature for streaming endpoints.
// It receives a context, request, and a stream for bidirectional communication.
type StreamEndpointFunc func(ctx context.Context, request any, stream Stream) error

// MiddlewareFunc wraps an EndpointFunc to add pre/post processing logic.
// It receives the next handler and returns a wrapped handler.
type MiddlewareFunc func(next EndpointFunc) EndpointFunc

// Endpoint represents a service endpoint with its routing and handler information.
type Endpoint struct {
	Name          string             // Method name (e.g., "GetUser")
	Path          string             // URL path (e.g., "/users/{id}")
	Method        string             // HTTP method (e.g., "GET", "POST")
	Handler       EndpointFunc       // Handler for non-streaming endpoints
	StreamHandler StreamEndpointFunc // Handler for streaming endpoints
	StreamMode    StreamMode         // Streaming behavior
	Middleware    []MiddlewareFunc   // Middleware chain applied to Handler

	// Type information for request/response
	RequestType  reflect.Type
	ResponseType reflect.Type

	Metadata map[string]any // Additional metadata (e.g., from @talk annotations)
}

// IsStreaming returns true if the endpoint uses any form of streaming.
func (e *Endpoint) IsStreaming() bool {
	return e.StreamMode != StreamNone
}

// WrappedHandler returns the Handler with all middleware applied.
// Middleware is applied in order: first middleware is outermost wrapper.
func (e *Endpoint) WrappedHandler() EndpointFunc {
	if e.Handler == nil || len(e.Middleware) == 0 {
		return e.Handler
	}
	h := e.Handler
	// Apply in reverse so that Middleware[0] is the outermost
	for i := len(e.Middleware) - 1; i >= 0; i-- {
		h = e.Middleware[i](h)
	}
	return h
}

// Clone creates a copy of the endpoint.
func (e *Endpoint) Clone() *Endpoint {
	clone := *e
	if e.Metadata != nil {
		clone.Metadata = make(map[string]any, len(e.Metadata))
		for k, v := range e.Metadata {
			clone.Metadata[k] = v
		}
	}
	return &clone
}

// EndpointOption configures an endpoint.
type EndpointOption func(*Endpoint)

// WithPath sets the endpoint path.
func WithPath(path string) EndpointOption {
	return func(e *Endpoint) {
		e.Path = path
	}
}

// WithMethod sets the HTTP method.
func WithMethod(method string) EndpointOption {
	return func(e *Endpoint) {
		e.Method = method
	}
}

// WithMetadata adds metadata to the endpoint.
func WithMetadata(key string, value any) EndpointOption {
	return func(e *Endpoint) {
		if e.Metadata == nil {
			e.Metadata = make(map[string]any)
		}
		e.Metadata[key] = value
	}
}

// WithMiddleware adds middleware functions to the endpoint.
func WithMiddleware(mw ...MiddlewareFunc) EndpointOption {
	return func(e *Endpoint) {
		e.Middleware = append(e.Middleware, mw...)
	}
}

// NewEndpoint creates a new endpoint with the given name and handler.
func NewEndpoint(name string, handler EndpointFunc, opts ...EndpointOption) *Endpoint {
	e := &Endpoint{
		Name:       name,
		Handler:    handler,
		StreamMode: StreamNone,
		Method:     "POST", // Default to POST for RPC-style calls
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// NewStreamEndpoint creates a new streaming endpoint.
func NewStreamEndpoint(name string, handler StreamEndpointFunc, mode StreamMode, opts ...EndpointOption) *Endpoint {
	e := &Endpoint{
		Name:          name,
		StreamHandler: handler,
		StreamMode:    mode,
		Method:        "GET", // Streaming typically uses GET for SSE/WebSocket
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}
