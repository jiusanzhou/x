package config

// Options present current params
type Options struct {
	// file's names we are trhing to load, merge data with the order
	names []string

	// Deprecated
	name string

	// file types we are supported
	isTypSet bool
	typs     []string

	// provider multi provider
	isProvSet bool
	providers []Provider

	// change lisner
	onChanged func(o, n interface{})

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

// WithNames set names
func WithNames(names ...string) Option {
	return func(c *Options) {
		c.names = append(c.names, names...)
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

// WithConfigChanged set on config chaned
func WithConfigChanged(f func(o, n interface{})) Option {
	return func(co *Options) {
		oldfn := co.onChanged
		if oldfn != nil {
			co.onChanged = func(o, n interface{}) {
				oldfn(o, n)
				f(o, n)
			}
		} else {
			co.onChanged = f
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
	o := &Options{
		name:      "config",
		typs:      []string{"yaml"},              // encoder
		providers: []Provider{DefaultFSProvider}, // provider
	}

	for _, op := range opts {
		op(o)
	}
	return o
}
