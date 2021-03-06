package cli

// New returns a command with options
func New(opts ...Option) *Command {
	c := newFromCobra()
	for _, o := range opts {
		o(c)
	}

	// if c.config is not nil we can load value
	// from parent or root
	PreParseConfig()(c)
	// we must do it at init time.
	// load set flag
	c.InitFlags()

	return c
}

// Register create a sub command
func (c *Command) Register(scs ...*Command) error {
	for _, sc := range scs {
		// TODO: check if sc.cfg is not a nil, we need to load from global config
		// or auto set from global config
		// NOTE: make sure sub command has no options from config
		sc.parent = c
		sc.root = c.root
		c.AddCommand(sc.Command)
	}
	return nil
}

// Run execute the whole commands
func (c *Command) Run(opts ...Option) error {
	for _, o := range opts {
		o(c)
	}

	return c.Command.Execute()
}

// IsRoot check if the command is the root
func (c *Command) IsRoot() bool {
	return c.root == c
}

// Option merge options after new
func (c *Command) Option(opts ...Option) *Command {
	for _, o := range opts {
		o(c)
	}

	// while we have add config by Option we need to re parse flags and config

	// if c.config is not nil we can load value
	// from parent or root
	PreParseConfig()(c)
	// we must do it at init time.
	// load set flag
	c.InitFlags()

	return c
}
