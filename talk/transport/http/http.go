// Package http provides HTTP transport implementations for talk.
package http

import (
	"go.zoe.im/x"
	"go.zoe.im/x/factory"
	"go.zoe.im/x/talk/codec"
)

// Config holds HTTP transport configuration.
type Config struct {
	Addr           string     `json:"addr" yaml:"addr"`
	Implementation string     `json:"implementation,omitempty" yaml:"implementation"`
	ReadTimeout    x.Duration `json:"read_timeout,omitempty" yaml:"read_timeout"`
	WriteTimeout   x.Duration `json:"write_timeout,omitempty" yaml:"write_timeout"`
	IdleTimeout    x.Duration `json:"idle_timeout,omitempty" yaml:"idle_timeout"`
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

// Option configures HTTP transport creation.
type Option func(any)

// WithCodec sets the codec for the transport.
func WithCodec(c codec.Codec) Option {
	return func(v any) {
		if s, ok := v.(interface{ SetCodec(codec.Codec) }); ok {
			s.SetCodec(c)
		}
	}
}

// ServerFactory creates HTTP server transport implementations.
var ServerFactory = factory.NewFactory[ServerTransport, Option]()

// ClientFactory creates HTTP client transport implementations.
var ClientFactory = factory.NewFactory[ClientTransport, Option]()

// ServerTransport handles HTTP server operations.
type ServerTransport interface {
	SetCodec(codec.Codec)
}

// ClientTransport handles HTTP client operations.
type ClientTransport interface {
	SetCodec(codec.Codec)
}
