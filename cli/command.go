package cli

import (
	"encoding/json"
	"os"
)

// New returns a command with options
func New(opts ...Option) *Command {
	c := newFromCobra()
	for _, o := range opts {
		o(c)
	}

	// if c.config is not nil we can load value
	// from parent or root
	PreRun(func(cmd *Command, args ...string) {
		// if we are not a root cmd, and set the config and with root config
		// we try to load config
		// but what about the ...
		if !cmd.IsRoot() && cmd.configv != nil && cmd.root.configobj != nil {
			v, ok := cmd.root.configobj.Get(c.Name())
			if ok {
				// FIXME:
				b, err := json.Marshal(v)

				// ignore error
				if err != nil {
					return
				}

				json.Unmarshal(b, cmd.configv)
				// re parse flag
				cmd.ParseFlags(os.Args)
			}
		}
	})(c)

	// we must do it at init time.
	// load set flag
	c.initFlags()

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
