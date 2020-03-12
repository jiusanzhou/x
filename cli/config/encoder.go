package config

import (
	"errors"
	"sync"
)

var (
	// default package level Encoder factory
	encoderFactory = NewEncoderFactory()

	// ErrUnsupportedEncoder is error
	ErrUnsupportedEncoder = errors.New("unsupported encoder")
)

// Encoder handles source encoding formats
//
// Unmarshal []byte to object
// Marshal object to []byte
type Encoder interface {
	Decode([]byte, interface{}) error
	Encode(interface{}) ([]byte, error)
	String() string
}

// EncoderFactory multi encoders
type EncoderFactory struct {
	sync.RWMutex
	encoders map[string]Encoder
}

// Register register a loader
func (t *EncoderFactory) Register(Encoder Encoder, names ...string) error {
	for _, name := range names {
		if _, ok := t.encoders[name]; ok {
			// TODO: need to return error?
		}
		t.encoders[name] = Encoder
	}

	return nil
}

// Decode data to out with special type name
func (t *EncoderFactory) Decode(name string, data []byte, out interface{}) error {
	er, ok := t.encoders[name]
	if !ok {
		return ErrUnsupportedEncoder
	}
	return er.Decode(data, out)
}

// Encode out to data with special type name
func (t *EncoderFactory) Encode(name string, out interface{}) ([]byte, error) {
	er, ok := t.encoders[name]
	if !ok {
		return nil, ErrUnsupportedEncoder
	}
	return er.Encode(out)
}

// NewEncoderFactory create a new Encoder factry
func NewEncoderFactory() *EncoderFactory {
	return &EncoderFactory{
		encoders: make(map[string]Encoder),
	}
}

// RegisterEncoder retister a type loader to package level
func RegisterEncoder(Encoder Encoder, names ...string) error {
	return encoderFactory.Register(Encoder, names...)
}

func init() {
	// register all encoder
	RegisterEncoder(NewJSONEncoder(), "json")
	RegisterEncoder(NewYAMLEncoder(), "yaml", "yml")
	RegisterEncoder(NewTOMLEncoder(), "toml")
	RegisterEncoder(NewHCLEncoder(), "hcl")
	RegisterEncoder(NewXMLEncoder(), "xml")
}
