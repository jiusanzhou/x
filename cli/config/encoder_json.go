package config

import (
	"encoding/json"
)

type jsonEncoder struct{}

func (j jsonEncoder) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (j jsonEncoder) Decode(d []byte, v interface{}) error {
	return json.Unmarshal(d, v)
}

func (j jsonEncoder) String() string {
	return "json"
}

// NewJSONEncoder return a json encoder
func NewJSONEncoder() Encoder {
	return jsonEncoder{}
}
