package config

import (
	"encoding/json"

	"github.com/hashicorp/hcl"
)

type hclEncoder struct{}

func (h hclEncoder) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (h hclEncoder) Decode(d []byte, v interface{}) error {
	return hcl.Unmarshal(d, v)
}

func (h hclEncoder) String() string {
	return "hcl"
}

// NewHCLEncoder return a hcl encoder
func NewHCLEncoder() Encoder {
	return hclEncoder{}
}
