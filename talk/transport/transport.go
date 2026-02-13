// Package transport provides transport layer abstractions for talk.
package transport

import (
	"context"

	"go.zoe.im/x"
	"go.zoe.im/x/factory"
	"go.zoe.im/x/talk"
)

// Transport abstracts the underlying protocol for serving and invoking endpoints.
type Transport interface {
	String() string

	// Server operations
	Serve(ctx context.Context, endpoints []*talk.Endpoint) error
	Shutdown(ctx context.Context) error

	// Client operations
	Invoke(ctx context.Context, endpoint string, req any, resp any) error
	InvokeStream(ctx context.Context, endpoint string, req any) (talk.Stream, error)
	Close() error
}

// TransportOption configures transport creation.
type TransportOption func(any)

// Factory creates Transport instances from configuration.
var Factory = factory.NewFactory[Transport, TransportOption]()

// ServerTransport is a transport that only handles server operations.
type ServerTransport interface {
	String() string
	Serve(ctx context.Context, endpoints []*talk.Endpoint) error
	Shutdown(ctx context.Context) error
}

// ClientTransport is a transport that only handles client operations.
type ClientTransport interface {
	String() string
	Invoke(ctx context.Context, endpoint string, req any, resp any) error
	InvokeStream(ctx context.Context, endpoint string, req any) (talk.Stream, error)
	Close() error
}

// Config holds common transport configuration.
type Config struct {
	Addr string `json:"addr" yaml:"addr"`
}

// ServerConfig extends Config with server-specific settings.
type ServerConfig struct {
	Config `json:",inline" yaml:",inline"`
}

// ClientConfig extends Config with client-specific settings.
type ClientConfig struct {
	Config  `json:",inline" yaml:",inline"`
	Timeout x.Duration `json:"timeout,omitempty" yaml:"timeout"`
}

// Get returns a transport by name using the provided config.
func Get(name string, cfg x.TypedLazyConfig) (Transport, error) {
	return Factory.Create(cfg)
}
