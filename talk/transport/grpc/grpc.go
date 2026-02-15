// Package grpc provides gRPC transport implementations for talk.
package grpc

import (
	"go.zoe.im/x"
	"go.zoe.im/x/factory"
	"go.zoe.im/x/talk"
	"go.zoe.im/x/talk/codec"
	"go.zoe.im/x/talk/transport"
)

type Config struct {
	Addr        string `json:"addr" yaml:"addr"`
	Insecure    bool   `json:"insecure,omitempty" yaml:"insecure"`
	TLSCertFile string `json:"tls_cert_file,omitempty" yaml:"tls_cert_file"`
	TLSKeyFile  string `json:"tls_key_file,omitempty" yaml:"tls_key_file"`
}

type ServerConfig struct {
	Config            `json:",inline" yaml:",inline"`
	MaxRecvMsgSize    int        `json:"max_recv_msg_size,omitempty" yaml:"max_recv_msg_size"`
	MaxSendMsgSize    int        `json:"max_send_msg_size,omitempty" yaml:"max_send_msg_size"`
	ConnectionTimeout x.Duration `json:"connection_timeout,omitempty" yaml:"connection_timeout"`
}

type ClientConfig struct {
	Config       `json:",inline" yaml:",inline"`
	Timeout      x.Duration `json:"timeout,omitempty" yaml:"timeout"`
	MaxRetries   int        `json:"max_retries,omitempty" yaml:"max_retries"`
	WaitForReady bool       `json:"wait_for_ready,omitempty" yaml:"wait_for_ready"`
}

type Option func(any)

func WithCodec(c codec.Codec) Option {
	return func(v any) {
		if s, ok := v.(interface{ SetCodec(codec.Codec) }); ok {
			s.SetCodec(c)
		}
	}
}

var serverFactory = factory.NewFactory[ServerTransport, Option]()

var ServerFactory = struct {
	Create   func(cfg x.TypedLazyConfig, opts ...Option) (ServerTransport, error)
	Register func(typeName string, creator factory.Creator[ServerTransport, Option], alias ...string) error
}{
	Create:   serverFactory.Create,
	Register: serverFactory.Register,
}

var clientFactory = factory.NewFactory[ClientTransport, Option]()

var ClientFactory = struct {
	Create   func(cfg x.TypedLazyConfig, opts ...Option) (ClientTransport, error)
	Register func(typeName string, creator factory.Creator[ClientTransport, Option], alias ...string) error
}{
	Create:   clientFactory.Create,
	Register: clientFactory.Register,
}

type ServerTransport interface {
	SetCodec(codec.Codec)
}

type ClientTransport interface {
	SetCodec(codec.Codec)
}

type adaptedServerFactory struct{}

func (f *adaptedServerFactory) Register(typeName string, creator factory.Creator[transport.ServerTransport, transport.TransportOption], alias ...string) error {
	return nil
}

func (f *adaptedServerFactory) Create(cfg x.TypedLazyConfig, opts ...transport.TransportOption) (transport.ServerTransport, error) {
	server, err := serverFactory.Create(cfg)
	if err != nil {
		return nil, err
	}

	if full, ok := server.(transport.ServerTransport); ok {
		return full, nil
	}

	return nil, talk.NewError(talk.Internal, "gRPC server does not implement transport.ServerTransport")
}

type adaptedClientFactory struct{}

func (f *adaptedClientFactory) Register(typeName string, creator factory.Creator[transport.ClientTransport, transport.TransportOption], alias ...string) error {
	return nil
}

func (f *adaptedClientFactory) Create(cfg x.TypedLazyConfig, opts ...transport.TransportOption) (transport.ClientTransport, error) {
	client, err := clientFactory.Create(cfg)
	if err != nil {
		return nil, err
	}

	if full, ok := client.(transport.ClientTransport); ok {
		return full, nil
	}

	return nil, talk.NewError(talk.Internal, "gRPC client does not implement transport.ClientTransport")
}

func init() {
	transport.ServerFactory.RegisterFamily("grpc", &adaptedServerFactory{})
	transport.ClientFactory.RegisterFamily("grpc", &adaptedClientFactory{})
}
