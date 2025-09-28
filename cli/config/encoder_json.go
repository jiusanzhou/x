package config

import (
	"encoding/json"
)

type jsonEncoder struct{}

func (j jsonEncoder) Encode(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

func (j jsonEncoder) Decode(d []byte, v any) error {
	return json.Unmarshal(d, v)
}

func (j jsonEncoder) String() string {
	return "json"
}

// NewJSONEncoder return a json encoder
func NewJSONEncoder() Encoder {
	return jsonEncoder{}
}
