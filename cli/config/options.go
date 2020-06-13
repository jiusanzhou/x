package config

// Options present current params
type Options struct {
	// file's names we are trhing to load, only load one exits
	// TODO: comboine fileds with names order
	names []string

	name string

	// file types we are supported
	isTypSet bool
	typs     []string

	// provider multi provider
	isProvSet bool
	providers []Provider

	// children
	children []*Config
}

// Option present config for C
type Option func(c *Options)

// WithName set name
func WithName(name string) Option {
	return func(c *Options) {
		c.name = name
	}
}

// WithType set file extension type for loader
func WithType(typ ...string) Option {
	return func(c *Options) {
		if c.isTypSet {
			c.typs = append(c.typs, typ...)
		} else {
			c.typs = typ
			c.isTypSet = true
		}
	}
}

// WithProvider set provider
func WithProvider(provides ...Provider) Option {
	return func(c *Options) {
		if c.isProvSet {
			c.providers = append(c.providers, provides...)
		} else {
			c.providers = provides
			c.isProvSet = true
		}
	}
}

// WithEnv set env loader, distinguish different environment
func WithEnv() Option {
	return func(c *Options) {
		// TODO: implement
	}
}

// WithChild add sub config
func WithChild() Option {
	return func(c *Options) {
		// TODO: implement
	}
}

// NewOptions return a new options
func NewOptions(opts ...Option) *Options {
	fs, _ := NewFSProvider("") // default current path
	o := &Options{
		// names:     []string{"config"}, // name
		name:      "config",
		typs:      []string{"yaml"}, // encoder
		providers: []Provider{fs},   // provider
	}

	for _, op := range opts {
		op(o)
	}
	return o
}
