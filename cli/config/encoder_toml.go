package config

import (
	"bytes"

	toml "github.com/pelletier/go-toml/v2"
)

type tomlEncoder struct{}

func (t tomlEncoder) Encode(v interface{}) ([]byte, error) {
	// toml encode won't work with raw message.
	// so we unmarshal it to json first and then marshal to toml

	var obj any

	jsonEncoder := jsonEncoder{}
	data, err := jsonEncoder.Encode(v)
	if err != nil {
		return nil, err
	}
	err = jsonEncoder.Decode(data, &obj)
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(nil)
	defer b.Reset()
	err = toml.NewEncoder(b).Encode(obj)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (t tomlEncoder) Decode(d []byte, v interface{}) error {
	return toml.Unmarshal(d, v)
}

func (t tomlEncoder) String() string {
	return "toml"
}

// NewTOMLEncoder return a toml encoder
func NewTOMLEncoder() Encoder {
	return tomlEncoder{}
}
