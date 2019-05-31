package cli

// New returns a command with options
func New(opts ...Option) *Command {
	c := newFromCobra()
	for _, o := range opts {
		o(c)
	}
	return c
}

// Register create a sub command
func (c *Command) Register(scs ...*Command) error {
	for _, sc := range scs {
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
