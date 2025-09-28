// Package config provider load configuration from file
package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"reflect"

	"go.zoe.im/x"
	"go.zoe.im/x/jsonmerge"
)

// souce is the real data of a config source
type source struct {
	// store all type of data from srouce
	obj  map[string]any // with map
	data []byte         // raw source data

	// options of source
	typ      string   // encoder type
	name     string   // config name
	provider Provider // active provider

	// store the latest error state
	err error
}

func (s *source) load(data []byte) error {
	// TODO: decode to v object directlly
	// clean the old data first
	s.obj = map[string]any{}
	// decode data to obj
	s.err = encoderFactory.Decode(s.typ, data, &s.obj)
	return s.err
}

// Config present runtime
type Config struct {
	valtyp  reflect.Type
	v       any // store the real one value, current value
	v0      any // store the old value, for notify, null if is the first time laod config
	vdeault any // store the default value

	// parsed and merged result
	obj  map[string]any
	data []byte // use obj to marshal

	names []string // store all names from options
	// all sources loaded
	sources map[string]*source

	opts *Options // options

	// store the latest error state
	errs x.Errors
}

// Options return the options
func (c *Config) Options() *Options {
	return c.opts
}

// CopyValue copy the raw data of config to object
func (c *Config) CopyValue(v any) error {
	// check if v is a pointer type
	// FIXME: ugly way
	return json.NewDecoder(bytes.NewBuffer(c.data)).Decode(v)
}

// Get take a value out from
func (c *Config) Get(key string) (any, bool) {
	v, ok := c.obj[key]
	return v, ok
}

// Default return the default value of config
func (c *Config) Default() any {
	return c.vdeault
}

// Current return the current value of config
func (c *Config) Current() any {
	return c.v
}

// load load data from provider
func (c *Config) load() error {
	// load config from provider to v
	// 1. load source from provider to sources
	// 2. merge sourcs and decode data to v

	// try to merge all names
	if c.opts.name != "" {
		c.names = append(c.names, c.opts.name)
	}

	// distinct the names
	for _, n := range c.opts.names {
		if !x.Contains(c.names, n) {
			c.names = append(c.names, n)
		}
	}

	// providers are store in optiosns, why we don't use factory?
	for _, provider := range c.Options().providers {
		// filestorage provider should provide extension
		for _, typ := range c.Options().typs {
			// add name one by one
			for _, n := range c.names {
				// read data from provider
				data, err := provider.Read(n, typ)
				if err != nil {
					continue
				}

				// build source object
				s := source{
					// TODO: new a v to merge directlly
					// v: reflect.New()
					data:     data,
					obj:      make(map[string]any),
					typ:      typ,
					provider: provider,
					name:     n,
				}

				s.err = s.load(data)
				if s.err != nil {
					c.errs.Add(s.err)
					// NOTE: maby we should add this s to sources map
					continue
				}

				// NOTE: config.yaml and config.json are the same one?
				c.sources[n] = &s
			}
		}
	}

	return c.errs
}

// mount loads  a data from sources to v
func (c *Config) mount() error {
	var err error

	// make sure c.obj is empty
	c.obj = map[string]any{}

	// merge all data from sources
	for _, n := range c.names {
		if s, ok := c.sources[n]; ok {
			c.errs.Add(jsonmerge.Merge(&c.obj, s.obj))
		}
	}

	// marshal obj
	c.data, err = json.Marshal(c.obj)
	c.errs.Add(err)

	// copy and store the old value
	if c.v0 == nil {
		c.v0 = reflect.New(c.valtyp.Elem()).Interface()
	}
	reflect.ValueOf(c.v0).Elem().Set(reflect.ValueOf(c.v).Elem())

	// create a new value from data
	vv := reflect.New(c.valtyp.Elem()).Interface()
	// set the default value to vv
	reflect.ValueOf(vv).Elem().Set(reflect.ValueOf(c.vdeault).Elem())
	// unmarshal data to vv
	c.errs.Add(json.Unmarshal(c.data, vv))

	// TODO:atomic to update the vv to c.v
	// update the vv to c.v
	reflect.ValueOf(c.v).Elem().Set(reflect.ValueOf(vv).Elem())

	return c.errs
}

func (c *Config) watch() error {
	if c.opts.onChanged == nil {
		return nil
	}

	for _, s := range c.sources {
		w, err := s.provider.Watch(s.name, s.typ)
		if err != nil {
			log.Println("[ERROR] start watcher for config:", s.name, "error:", err)
			continue
		}

		go func(s *source) {
			for {
				cs, err := w.Next()
				if err != nil {
					continue
				}

				// load the changed data to the source
				s.load(cs.Data)

				// let's remount all the data again
				c.mount()

				// NOTE: notify with onChanged who we can get the obj of old one
				// TODO: add a old one
				c.opts.onChanged(c.v0, c.v)
			}
		}(s)
	}

	return nil
}

func (c *Config) Init() error {
	if c.valtyp.Kind() != reflect.Ptr {
		return errors.New("value must be a pointer")
	}

	// we are trying to load configuration
	// simple way is just load data
	c.load() // NOTE: main process

	// mount loads all sources to config
	// load all
	c.mount() // NOTE: main process

	// if we have a listener function just start a new wwatcher
	// create a new watcher
	c.watch() // NOTE: main process

	var err error
	// TODO: make sure errs can be a nil
	if c.errs.IsNil() {
		err = nil
	} else {
		err = c.errs
	}
	return err
}

// New create a new config
func New(v any, opts ...Option) *Config {
	c := &Config{
		v:       v,
		opts:    NewOptions(opts...),
		sources: make(map[string]*source),
	}

	// generate the type of v
	c.valtyp = reflect.TypeOf(v)

	// marsal to data and unmarshal to object
	// c.data, err = json.Marshal(c.v)
	// c.errs.Add(err)
	// c.errs.Add(json.Unmarshal(c.data, &c.obj))

	// new a value from valtyp
	// restore the default from the value type
	c.vdeault = reflect.New(c.valtyp.Elem()).Interface()
	reflect.ValueOf(c.vdeault).Elem().Set(reflect.ValueOf(v).Elem())

	return c
}
