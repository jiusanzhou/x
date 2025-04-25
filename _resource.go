package x

// tmp file, need to be removed
// and implement dynamic object

import (
	"encoding/json"
	"errors"
	"reflect"
)

var (
	// auto build the maker function
	_registry = make(map[string]func(data json.RawMessage) (SourceType, error))
)

// Register the resource type
func Register(res SourceType, typs ...string) error {

	vtype := reflect.TypeOf(res)
	if vtype.Kind() == reflect.Ptr {
		vtype = vtype.Elem()
	}

	// build the generator function
	fn := func(data json.RawMessage) (SourceType, error) {
		src := reflect.New(vtype).Interface().(SourceType)
		err := json.Unmarshal(data, &src)
		return src, err
	}

	// typ := src.Type()
	for _, typ := range typs {
		if _, ok := _registry[typ]; ok {
			return errors.New("type exits: " + typ)
		}
		_registry[typ] = fn
	}

	return nil
}

// SourceType definie the resource interfaace
type SourceType interface {
	// TODO: load resource
	Init() error
}

// Object is the main resource wraper
type Object struct {
	Type string `json:"type,omitempty" yaml:"type"`

	// resource cannot handle a name !!!

	// the real sub resource
	res SourceType

	// ==================================
	// TODO: auto have this data
	// TODO: auto have this data
	_raw       json.RawMessage
	_rawfields map[string]json.RawMessage
	// ==================================

	// use object to caculate the ...
	// store the stat of all nodes???
}

// Init ...
func (s *Object) Init() error {
	// TODO: do we need do something?
	return s.res.Init()
}

// MarshalJSON ...
func (s Object) MarshalJSON() ([]byte, error) {
	// TODO: marshal from the real one
	return json.Marshal(s._rawfields)
}

// UnmarshalJSON ...
func (s *Object) UnmarshalJSON(data []byte) error {
	s._raw = data

	err := json.Unmarshal(data, &s._rawfields)
	if err != nil {
		return err
	}

	// TODO: auto have this fields unmarshalling
	// TODO: auto have this fields unmarshalling
	json.Unmarshal(s._rawfields["type"], &s.Type)

	// auto create the wrapper
	if fn, ok := _registry[s.Type]; ok {
		s.res, err = fn(s._raw)
		return err
	}

	return errors.New("source type not supported: " + s.Type)
}
