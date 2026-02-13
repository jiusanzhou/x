// Package websocket provides WebSocket transport implementations for talk.
package websocket

import (
	"go.zoe.im/x"
	"go.zoe.im/x/factory"
	"go.zoe.im/x/talk/codec"
)

// Config holds WebSocket transport configuration.
type Config struct {
	Addr            string     `json:"addr" yaml:"addr"`
	Path            string     `json:"path,omitempty" yaml:"path"`
	ReadBufferSize  int        `json:"read_buffer_size,omitempty" yaml:"read_buffer_size"`
	WriteBufferSize int        `json:"write_buffer_size,omitempty" yaml:"write_buffer_size"`
	PingInterval    x.Duration `json:"ping_interval,omitempty" yaml:"ping_interval"`
	PongTimeout     x.Duration `json:"pong_timeout,omitempty" yaml:"pong_timeout"`
}

// ServerConfig extends Config with server-specific settings.
type ServerConfig struct {
	Config            `json:",inline" yaml:",inline"`
	CheckOrigin       bool `json:"check_origin,omitempty" yaml:"check_origin"`
	EnableCompression bool `json:"enable_compression,omitempty" yaml:"enable_compression"`
}

// ClientConfig extends Config with client-specific settings.
type ClientConfig struct {
	Config           `json:",inline" yaml:",inline"`
	HandshakeTimeout x.Duration `json:"handshake_timeout,omitempty" yaml:"handshake_timeout"`
}

// Option configures WebSocket transport creation.
type Option func(any)

// WithCodec sets the codec for the transport.
func WithCodec(c codec.Codec) Option {
	return func(v any) {
		if s, ok := v.(interface{ SetCodec(codec.Codec) }); ok {
			s.SetCodec(c)
		}
	}
}

// ServerFactory creates WebSocket server transport implementations.
var ServerFactory = factory.NewFactory[ServerTransport, Option]()

// ClientFactory creates WebSocket client transport implementations.
var ClientFactory = factory.NewFactory[ClientTransport, Option]()

// ServerTransport handles WebSocket server operations.
type ServerTransport interface {
	SetCodec(codec.Codec)
}

// ClientTransport handles WebSocket client operations.
type ClientTransport interface {
	SetCodec(codec.Codec)
}

// MessageType represents WebSocket message types.
type MessageType int

const (
	TextMessage   MessageType = 1
	BinaryMessage MessageType = 2
)

// Message represents a WebSocket message with type information.
type Message struct {
	Type    string `json:"type"`
	Payload []byte `json:"payload"`
	Error   string `json:"error,omitempty"`
}
