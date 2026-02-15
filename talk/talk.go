package talk

import (
	"context"

	"go.zoe.im/x/talk/codec"
)

// Transport abstracts the underlying protocol for serving and invoking endpoints.
type Transport interface {
	String() string

	Serve(ctx context.Context, endpoints []*Endpoint) error
	Shutdown(ctx context.Context) error

	Invoke(ctx context.Context, endpoint string, req any, resp any) error
	InvokeStream(ctx context.Context, endpoint string, req any) (Stream, error)
	Close() error
}

// Extractor extracts endpoints from a service implementation.
type Extractor interface {
	Extract(service any) ([]*Endpoint, error)
}

// Server handles incoming requests and routes them to registered endpoints.
type Server struct {
	transport  Transport
	codec      codec.Codec
	extractor  Extractor
	pathPrefix string
	endpoints  []*Endpoint
}

// NewServer creates a new server with the given transport.
func NewServer(t Transport, opts ...ServerOption) *Server {
	s := &Server{transport: t}
	for _, opt := range opts {
		opt(s)
	}

	if s.codec == nil {
		if c, err := codec.Get("json"); err == nil {
			s.codec = c
		}
	}

	return s
}

// RegisterOption configures service registration.
type RegisterOption func(*registerConfig)

type registerConfig struct {
	pathPrefix string
}

// WithPrefix sets a path prefix for all endpoints (e.g., "/api/v1").
func WithPrefix(prefix string) RegisterOption {
	return func(c *registerConfig) {
		c.pathPrefix = prefix
	}
}

// Register extracts endpoints from a service implementation and registers them.
// Use WithPrefix("/api/v1") to override the server's default path prefix.
func (s *Server) Register(service any, opts ...RegisterOption) error {
	if s.extractor == nil {
		return NewError(FailedPrecondition, "no extractor configured")
	}

	cfg := &registerConfig{
		pathPrefix: s.pathPrefix,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	endpoints, err := s.extractor.Extract(service)
	if err != nil {
		return err
	}

	if cfg.pathPrefix != "" {
		for _, ep := range endpoints {
			ep.Path = cfg.pathPrefix + ep.Path
		}
	}

	s.endpoints = append(s.endpoints, endpoints...)
	return nil
}

// RegisterEndpoints adds pre-defined endpoints.
func (s *Server) RegisterEndpoints(endpoints ...*Endpoint) {
	s.endpoints = append(s.endpoints, endpoints...)
}

// RegisterEndpointsWithPrefix adds pre-defined endpoints with a path prefix.
func (s *Server) RegisterEndpointsWithPrefix(prefix string, endpoints ...*Endpoint) {
	for _, ep := range endpoints {
		ep.Path = prefix + ep.Path
	}
	s.endpoints = append(s.endpoints, endpoints...)
}

// RegisterWithPrefix is a convenience method for Register(service, WithPrefix(prefix)).
func (s *Server) RegisterWithPrefix(service any, prefix string) error {
	return s.Register(service, WithPrefix(prefix))
}

// Endpoints returns all registered endpoints.
func (s *Server) Endpoints() []*Endpoint {
	return s.endpoints
}

// Serve starts the server and blocks until context is cancelled.
func (s *Server) Serve(ctx context.Context) error {
	return s.transport.Serve(ctx, s.endpoints)
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.transport.Shutdown(ctx)
}

// Client invokes remote endpoints.
type Client struct {
	transport Transport
	codec     codec.Codec
}

// NewClient creates a new client with the given transport.
func NewClient(t Transport, opts ...ClientOption) *Client {
	c := &Client{transport: t}
	for _, opt := range opts {
		opt(c)
	}

	if c.codec == nil {
		if cdc, err := codec.Get("json"); err == nil {
			c.codec = cdc
		}
	}

	return c
}

// Call invokes an endpoint and decodes the response.
func (c *Client) Call(ctx context.Context, endpoint string, req any, resp any) error {
	return c.transport.Invoke(ctx, endpoint, req, resp)
}

// Stream opens a streaming connection to an endpoint.
func (c *Client) Stream(ctx context.Context, endpoint string, req any) (Stream, error) {
	return c.transport.InvokeStream(ctx, endpoint, req)
}

// Close closes the client connection.
func (c *Client) Close() error {
	return c.transport.Close()
}

// ServerOption configures a Server.
type ServerOption func(*Server)

// WithServerCodec sets the server's codec.
func WithServerCodec(c codec.Codec) ServerOption {
	return func(s *Server) {
		s.codec = c
	}
}

// WithServerCodecName sets the server's codec by name.
func WithServerCodecName(name string) ServerOption {
	return func(s *Server) {
		if c, err := codec.Get(name); err == nil {
			s.codec = c
		}
	}
}

// WithExtractor sets the endpoint extractor for service registration.
func WithExtractor(e Extractor) ServerOption {
	return func(s *Server) {
		s.extractor = e
	}
}

// WithPathPrefix sets a default path prefix for all registered endpoints.
func WithPathPrefix(prefix string) ServerOption {
	return func(s *Server) {
		s.pathPrefix = prefix
	}
}

// ClientOption configures a Client.
type ClientOption func(*Client)

// WithClientCodec sets the client's codec.
func WithClientCodec(c codec.Codec) ClientOption {
	return func(cl *Client) {
		cl.codec = c
	}
}

// WithClientCodecName sets the client's codec by name.
func WithClientCodecName(name string) ClientOption {
	return func(cl *Client) {
		if c, err := codec.Get(name); err == nil {
			cl.codec = c
		}
	}
}
