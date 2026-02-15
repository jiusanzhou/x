package codec

import (
	"encoding/json"

	"go.zoe.im/x"
)

const (
	jsonName        = "json"
	jsonContentType = "application/json"
)

type jsonCodec struct{}

func (c *jsonCodec) Name() string {
	return jsonName
}

func (c *jsonCodec) ContentType() string {
	return jsonContentType
}

func (c *jsonCodec) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (c *jsonCodec) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func init() {
	Factory.Register(jsonName, func(cfg x.TypedLazyConfig, opts ...CodecOption) (Codec, error) {
		return &jsonCodec{}, nil
	})
}
