// Package config provider load configuration from file
package config

// Config present runtime
type Config struct {
	v    interface{} // store the value
	opts *Options    // options

	// store the latest error state
	err error

	typ  string // encoder type
	data []byte // raw source data
	obj  map[string]interface{}
}

// Options return the options
func (c *Config) Options() *Options {
	return c.opts
}

// CopyValue copy the raw data of config to object
func (c *Config) CopyValue(v interface{}) error {
	// check if v is a pointer type
	return encoderFactory.Decode(c.typ, c.data, v)
}

// Get take a value out from
func (c *Config) Get(key string) (interface{}, bool) {
	v, ok := c.obj[key]
	return v, ok
}

// load load data from provider
func (c *Config) load() error {
	// load config from provider to v
	// 1. load source from provider
	// 2. decode source data to v

	var (
		data   []byte
		parsed bool
	)

	// providers are store in optiosns, why we don't use factory?
	for _, provider := range c.Options().providers {
		// with type?
		// filestorage provider should provide extension
		for _, typ := range c.Options().typs {
			// read data from provider
			data, c.err = provider.Read(c.Options().name, typ) // with type
			if c.err != nil {
				continue
			}

			c.obj = make(map[string]interface{})
			encoderFactory.Decode(typ, data, &c.obj)

			// decode data to value
			c.err = encoderFactory.Decode(typ, data, c.v)
			if c.err != nil {
				continue
			}

			// store the type and the data
			c.typ = typ
			c.data = data

			// we need to end the process
			parsed = true
			goto _end
		}
	}

_end:
	if !parsed {
		return c.err
	}

	return c.err
}

// New create a new config
func New(v interface{}, opts ...Option) (*Config, error) {
	c := &Config{
		v:    v,
		opts: NewOptions(opts...),
	}

	// we are trying to load configuration
	// simple way is just load data
	// c.Watch()
	c.err = c.load()
	if c.err != nil {
		return nil, c.err
	}

	return c, nil
}
