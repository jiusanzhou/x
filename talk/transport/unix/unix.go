// Package unix provides Unix socket transport implementation for talk.
// It uses HTTP over Unix domain sockets for communication.
package unix

import (
	"go.zoe.im/x"
	"go.zoe.im/x/factory"
	"go.zoe.im/x/talk"
	"go.zoe.im/x/talk/codec"
	"go.zoe.im/x/talk/transport"
)

type Config struct {
	Path         string     `json:"path" yaml:"path"`
	ReadTimeout  x.Duration `json:"read_timeout,omitempty" yaml:"read_timeout"`
	WriteTimeout x.Duration `json:"write_timeout,omitempty" yaml:"write_timeout"`
	IdleTimeout  x.Duration `json:"idle_timeout,omitempty" yaml:"idle_timeout"`
}

type ServerConfig struct {
	Config `json:",inline" yaml:",inline"`
}

type ClientConfig struct {
	Config  `json:",inline" yaml:",inline"`
	Timeout x.Duration `json:"timeout,omitempty" yaml:"timeout"`
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

type unixTransportFamily struct{}

func (f *unixTransportFamily) CreateServer(cfg x.TypedLazyConfig, opts ...transport.TransportOption) (transport.ServerTransport, error) {
	server, err := serverFactory.Create(cfg)
	if err != nil {
		return nil, err
	}

	if full, ok := server.(transport.ServerTransport); ok {
		return full, nil
	}

	return nil, talk.NewError(talk.Internal, "Unix server does not implement transport.ServerTransport")
}

func (f *unixTransportFamily) CreateClient(cfg x.TypedLazyConfig, opts ...transport.TransportOption) (transport.ClientTransport, error) {
	client, err := clientFactory.Create(cfg)
	if err != nil {
		return nil, err
	}

	if full, ok := client.(transport.ClientTransport); ok {
		return full, nil
	}

	return nil, talk.NewError(talk.Internal, "Unix client does not implement transport.ClientTransport")
}

func init() {
	transport.Factory.RegisterFamily("unix", &unixTransportFamily{})
}
