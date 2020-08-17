// Package config provider load configuration from file
package config

import (
	"bytes"
	"encoding/json"
	"log"
	"reflect"

	"go.zoe.im/x"
	"go.zoe.im/x/jsonmerge"
)

// Config present runtime
type Config struct {
	// parsed and merged result
	valtyp reflect.Type
	v      interface{} // store the value
	obj    map[string]interface{}
	data   []byte // use obj to marshal

	names []string // store all names from options

	opts *Options // options

	// store the latest error state
	errs x.Errors

	// all sources loaded
	sources map[string]*source
}

// souce is the real data of a config source
type source struct {
	// store all type of data from srouce
	v    interface{}            // store the value
	obj  map[string]interface{} // with map
	data []byte                 // raw source data

	// options of source
	typ      string   // encoder type
	name     string   // config name
	provider Provider // active provider

	// store the latest error state
	err error
}

func (s *source) load(data []byte) error {
	// TODO: decode to v object directlly
	// decode data to obj
	s.err = encoderFactory.Decode(s.typ, data, &s.obj)
	return s.err
}

// Options return the options
func (c *Config) Options() *Options {
	return c.opts
}

// CopyValue copy the raw data of config to object
func (c *Config) CopyValue(v interface{}) error {
	// check if v is a pointer type
	// FIXME: ugly way
	return json.NewDecoder(bytes.NewBuffer(c.data)).Decode(v)
}

// Get take a value out from
func (c *Config) Get(key string) (interface{}, bool) {
	v, ok := c.obj[key]
	return v, ok
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
	for _, n := range c.opts.names {
		c.names = append(c.names, n)
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
					obj:      make(map[string]interface{}),
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
	c.obj = map[string]interface{}{}

	// merge obejct and data?
	for _, n := range c.names {
		if s, ok := c.sources[n]; ok {
			c.errs.Add(jsonmerge.Merge(&c.obj, s.obj))
		}
	}

	c.data, err = json.Marshal(c.obj)
	c.errs.Add(err)
	c.errs.Add(json.Unmarshal(c.data, &c.v))

	// NOTE: re parse flags from os.Args in onChanged
	return c.errs
}

func (c *Config) watch() error {

	if c.opts.onChanged == nil {
		// log.Println("[DEBUG] without onchanged listener")
		return nil
	}

	for _, s := range c.sources {
		w, err := s.provider.Watch(s.name, s.typ)
		if err != nil {
			log.Println("[ERROR] start watcher for", s.name, "error:", err)
			continue
		}

		log.Println("[DEBUG] start watcher for", s.name)

		go func(s *source) {
			// TODO: watch the source, auto load while change and re-mount

			for {
				cs, err := w.Next()
				if err != nil {
					continue
				}

				if err != nil {
					log.Println("[ERROR] reload error", err)
				}

				// load one source
				s.load(cs.Data)

				// mount config again
				c.mount()

				// NOTE: notify with onChanged who we can get the obj of old one
				// TODO: add a old one
				c.opts.onChanged(nil, c.v)
			}
		}(s)
	}

	return nil
}

// New create a new config
func New(v interface{}, opts ...Option) (*Config, error) {
	c := &Config{
		v:       v,
		opts:    NewOptions(opts...),
		sources: make(map[string]*source),
	}

	// generate the type of v
	c.valtyp = reflect.TypeOf(v)

	var err error

	// marsal to data and unmarshal to object
	c.data, err = json.Marshal(c.v)
	c.errs.Add(err)

	c.errs.Add(json.Unmarshal(c.data, &c.obj))

	// new a value from valtyp

	// we are trying to load configuration
	// simple way is just load data
	c.load() // NOTE: main process

	// mount loads all sources to config
	// load all
	c.mount() // NOTE: main process

	// if we have a listener function just start a new wwatcher
	// create a new watcher
	c.watch() // NOTE: main process

	// TODO: make sure errs can be a nil
	if c.errs.IsNil() {
		err = nil
	} else {
		err = c.errs
	}
	return c, err
}
