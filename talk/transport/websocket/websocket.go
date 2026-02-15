// Package websocket provides WebSocket transport implementations for talk.
package websocket

import (
	"go.zoe.im/x"
	"go.zoe.im/x/factory"
	"go.zoe.im/x/talk"
	"go.zoe.im/x/talk/codec"
	"go.zoe.im/x/talk/transport"
)

type Config struct {
	Addr            string     `json:"addr" yaml:"addr"`
	Path            string     `json:"path,omitempty" yaml:"path"`
	ReadBufferSize  int        `json:"read_buffer_size,omitempty" yaml:"read_buffer_size"`
	WriteBufferSize int        `json:"write_buffer_size,omitempty" yaml:"write_buffer_size"`
	PingInterval    x.Duration `json:"ping_interval,omitempty" yaml:"ping_interval"`
	PongTimeout     x.Duration `json:"pong_timeout,omitempty" yaml:"pong_timeout"`
}

type ServerConfig struct {
	Config            `json:",inline" yaml:",inline"`
	CheckOrigin       bool `json:"check_origin,omitempty" yaml:"check_origin"`
	EnableCompression bool `json:"enable_compression,omitempty" yaml:"enable_compression"`
}

type ClientConfig struct {
	Config           `json:",inline" yaml:",inline"`
	HandshakeTimeout x.Duration `json:"handshake_timeout,omitempty" yaml:"handshake_timeout"`
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

type MessageType int

const (
	TextMessage   MessageType = 1
	BinaryMessage MessageType = 2
)

type Message struct {
	Type    string `json:"type"`
	Payload []byte `json:"payload"`
	Error   string `json:"error,omitempty"`
}

type wsTransportFamily struct{}

func (f *wsTransportFamily) CreateServer(cfg x.TypedLazyConfig, opts ...transport.TransportOption) (transport.ServerTransport, error) {
	server, err := serverFactory.Create(cfg)
	if err != nil {
		return nil, err
	}

	if full, ok := server.(transport.ServerTransport); ok {
		return full, nil
	}

	return nil, talk.NewError(talk.Internal, "WebSocket server does not implement transport.ServerTransport")
}

func (f *wsTransportFamily) CreateClient(cfg x.TypedLazyConfig, opts ...transport.TransportOption) (transport.ClientTransport, error) {
	client, err := clientFactory.Create(cfg)
	if err != nil {
		return nil, err
	}

	if full, ok := client.(transport.ClientTransport); ok {
		return full, nil
	}

	return nil, talk.NewError(talk.Internal, "WebSocket client does not implement transport.ClientTransport")
}

func init() {
	transport.Factory.RegisterFamily("websocket", &wsTransportFamily{}, "ws")
}
