// Package codec provides encoding/decoding interfaces for talk messages.
package codec

import (
	"go.zoe.im/x"
	"go.zoe.im/x/factory"
)

// Codec encodes and decodes messages between binary and Go types.
type Codec interface {
	Name() string
	ContentType() string
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}

// CodecOption configures codec creation.
type CodecOption func(any)

// Factory creates Codec instances from configuration.
var Factory = factory.NewFactory[Codec, CodecOption]()

// Get returns a codec by name using an empty config.
func Get(name string) (Codec, error) {
	return Factory.Create(x.TypedLazyConfig{Type: name})
}

// MustGet returns a codec by name, panicking on error.
func MustGet(name string) Codec {
	c, err := Get(name)
	if err != nil {
		panic(err)
	}
	return c
}
