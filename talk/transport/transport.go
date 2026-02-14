// Package transport provides transport layer abstractions for talk.
package transport

import (
	"context"
	"fmt"
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

// ServerFactory creates server transport instances from configuration.
// It supports hierarchical type selection like "http/gin" where:
// - "http" selects the HTTP transport family
// - "gin" selects the specific implementation within HTTP
var ServerFactory = &serverTransportFactory{
	families: make(map[string]factory.Factory[ServerTransport, TransportOption]),
}

// ClientFactory creates client transport instances from configuration.
// It supports hierarchical type selection like "http/std".
var ClientFactory = &clientTransportFactory{
	families: make(map[string]factory.Factory[ClientTransport, TransportOption]),
}

// serverTransportFactory manages server transport creation with hierarchical type support.
type serverTransportFactory struct {
	families map[string]factory.Factory[ServerTransport, TransportOption]
}

// RegisterFamily registers a sub-factory for a transport family (e.g., "http", "grpc").
func (f *serverTransportFactory) RegisterFamily(family string, subFactory factory.Factory[ServerTransport, TransportOption]) {
	f.families[family] = subFactory
}

// Register registers a transport creator directly for simple types (e.g., "grpc", "websocket").
func (f *serverTransportFactory) Register(typeName string, creator factory.Creator[ServerTransport, TransportOption], alias ...string) error {
	// Create a simple factory for this type
	simple := factory.NewFactory[ServerTransport, TransportOption]()
	if err := simple.Register("default", creator); err != nil {
		return err
	}
	f.families[typeName] = simple
	for _, a := range alias {
		f.families[a] = simple
	}
	return nil
}

// Create creates a server transport based on configuration.
// Supports formats:
// - "http" -> uses HTTP family with default implementation
// - "http/gin" -> uses HTTP family with "gin" implementation
// - "grpc" -> uses gRPC transport
// - "websocket" -> uses WebSocket transport
func (f *serverTransportFactory) Create(cfg x.TypedLazyConfig, opts ...TransportOption) (ServerTransport, error) {
	family, impl := parseType(cfg.Type)

	subFactory, ok := f.families[family]
	if !ok {
		return nil, fmt.Errorf("unknown transport family: %s", family)
	}

	// Create a new config with the implementation type
	subCfg := x.TypedLazyConfig{
		Name:   cfg.Name,
		Type:   impl,
		Config: cfg.Config,
	}

	return subFactory.Create(subCfg, opts...)
}

// clientTransportFactory manages client transport creation with hierarchical type support.
type clientTransportFactory struct {
	families map[string]factory.Factory[ClientTransport, TransportOption]
}

// RegisterFamily registers a sub-factory for a transport family.
func (f *clientTransportFactory) RegisterFamily(family string, subFactory factory.Factory[ClientTransport, TransportOption]) {
	f.families[family] = subFactory
}

// Register registers a transport creator directly for simple types.
func (f *clientTransportFactory) Register(typeName string, creator factory.Creator[ClientTransport, TransportOption], alias ...string) error {
	simple := factory.NewFactory[ClientTransport, TransportOption]()
	if err := simple.Register("default", creator); err != nil {
		return err
	}
	f.families[typeName] = simple
	for _, a := range alias {
		f.families[a] = simple
	}
	return nil
}

// Create creates a client transport based on configuration.
func (f *clientTransportFactory) Create(cfg x.TypedLazyConfig, opts ...TransportOption) (ClientTransport, error) {
	family, impl := parseType(cfg.Type)

	subFactory, ok := f.families[family]
	if !ok {
		return nil, fmt.Errorf("unknown transport family: %s", family)
	}

	subCfg := x.TypedLazyConfig{
		Name:   cfg.Name,
		Type:   impl,
		Config: cfg.Config,
	}

	return subFactory.Create(subCfg, opts...)
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

// Get returns a transport by name using the provided config.
func Get(name string, cfg x.TypedLazyConfig) (Transport, error) {
	return Factory.Create(cfg)
}
