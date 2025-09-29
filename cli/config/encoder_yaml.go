package config

import (
	yaml "sigs.k8s.io/yaml"
)

type yamlEncoder struct{}

func (y yamlEncoder) Encode(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

func (y yamlEncoder) Decode(d []byte, v interface{}) error {
	return yaml.Unmarshal(d, v)
}

func (y yamlEncoder) String() string {
	return "yaml"
}

// NewYAMLEncoder return a yaml encoder
func NewYAMLEncoder() Encoder {
	return yamlEncoder{}
}
