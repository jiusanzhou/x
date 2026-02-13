// Package grpc provides gRPC transport implementations for talk.
package grpc

import (
	"go.zoe.im/x"
	"go.zoe.im/x/factory"
	"go.zoe.im/x/talk/codec"
)

// Config holds gRPC transport configuration.
type Config struct {
	Addr        string `json:"addr" yaml:"addr"`
	Insecure    bool   `json:"insecure,omitempty" yaml:"insecure"`
	TLSCertFile string `json:"tls_cert_file,omitempty" yaml:"tls_cert_file"`
	TLSKeyFile  string `json:"tls_key_file,omitempty" yaml:"tls_key_file"`
}

// ServerConfig extends Config with server-specific settings.
type ServerConfig struct {
	Config            `json:",inline" yaml:",inline"`
	MaxRecvMsgSize    int        `json:"max_recv_msg_size,omitempty" yaml:"max_recv_msg_size"`
	MaxSendMsgSize    int        `json:"max_send_msg_size,omitempty" yaml:"max_send_msg_size"`
	ConnectionTimeout x.Duration `json:"connection_timeout,omitempty" yaml:"connection_timeout"`
}

// ClientConfig extends Config with client-specific settings.
type ClientConfig struct {
	Config       `json:",inline" yaml:",inline"`
	Timeout      x.Duration `json:"timeout,omitempty" yaml:"timeout"`
	MaxRetries   int        `json:"max_retries,omitempty" yaml:"max_retries"`
	WaitForReady bool       `json:"wait_for_ready,omitempty" yaml:"wait_for_ready"`
}

// Option configures gRPC transport creation.
type Option func(any)

// WithCodec sets the codec for the transport.
func WithCodec(c codec.Codec) Option {
	return func(v any) {
		if s, ok := v.(interface{ SetCodec(codec.Codec) }); ok {
			s.SetCodec(c)
		}
	}
}

// ServerFactory creates gRPC server transport implementations.
var ServerFactory = factory.NewFactory[ServerTransport, Option]()

// ClientFactory creates gRPC client transport implementations.
var ClientFactory = factory.NewFactory[ClientTransport, Option]()

// ServerTransport handles gRPC server operations.
type ServerTransport interface {
	SetCodec(codec.Codec)
}

// ClientTransport handles gRPC client operations.
type ClientTransport interface {
	SetCodec(codec.Codec)
}
