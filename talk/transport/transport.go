// Package transport provides transport layer abstractions for talk.
package transport

import (
	"context"
	"strings"

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

// TransportFamily provides both server and client transport creation for a transport type.
type TransportFamily interface {
	CreateServer(cfg x.TypedLazyConfig, opts ...TransportOption) (ServerTransport, error)
	CreateClient(cfg x.TypedLazyConfig, opts ...TransportOption) (ClientTransport, error)
}

type familyOption struct{}

var familyFactory = factory.NewFactory[TransportFamily, familyOption]()

// Factory provides unified access to transport creation.
// Use Factory.RegisterFamily to register a transport family (e.g., "http", "grpc", "websocket").
// Use Factory.CreateServer/CreateClient to create transports.
var Factory = struct {
	RegisterFamily func(name string, family TransportFamily, aliases ...string) error
	CreateServer   func(cfg x.TypedLazyConfig, opts ...TransportOption) (ServerTransport, error)
	CreateClient   func(cfg x.TypedLazyConfig, opts ...TransportOption) (ClientTransport, error)
}{
	RegisterFamily: func(name string, family TransportFamily, aliases ...string) error {
		return familyFactory.Register(name, func(cfg x.TypedLazyConfig, opts ...familyOption) (TransportFamily, error) {
			return family, nil
		}, aliases...)
	},
	CreateServer: func(cfg x.TypedLazyConfig, opts ...TransportOption) (ServerTransport, error) {
		familyName, impl := parseType(cfg.Type)

		familyCfg := x.TypedLazyConfig{Type: familyName}
		family, err := familyFactory.Create(familyCfg)
		if err != nil {
			return nil, err
		}

		subCfg := x.TypedLazyConfig{
			Name:   cfg.Name,
			Type:   impl,
			Config: cfg.Config,
		}
		return family.CreateServer(subCfg, opts...)
	},
	CreateClient: func(cfg x.TypedLazyConfig, opts ...TransportOption) (ClientTransport, error) {
		familyName, impl := parseType(cfg.Type)

		familyCfg := x.TypedLazyConfig{Type: familyName}
		family, err := familyFactory.Create(familyCfg)
		if err != nil {
			return nil, err
		}

		subCfg := x.TypedLazyConfig{
			Name:   cfg.Name,
			Type:   impl,
			Config: cfg.Config,
		}
		return family.CreateClient(subCfg, opts...)
	},
}

// parseType parses a type string like "http/gin" into family and implementation.
// Returns (family, implementation). If no "/" is present, returns (type, "default").
func parseType(t string) (family, impl string) {
	parts := strings.SplitN(t, "/", 2)
	if len(parts) == 1 {
		return parts[0], "default"
	}
	return parts[0], parts[1]
}
