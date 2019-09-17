package cli

import (
	"go.zoe.im/x/cli/opts"

	"github.com/spf13/cobra"
)

// Option set the command with value
type Option func(c *Command)

// Version returns option to set set the version
func Version(v string) Option {
	return func(c *Command) {
		c.Command.Version = v
	}
}

// Name returns option to set use
func Name(names ...string) Option {
	return func(c *Command) {
		c.Command.Use = names[0]
		c.Command.Aliases = append(c.Command.Aliases, names[1:]...)
	}
}

// Run returns option to set the main run function
func Run(fn func(cmd *Command, args ...string)) Option {
	return func(c *Command) {
		// try to create func(cmd *cobra.Command, args []string) directly
		c.Command.Run = func(cmd *cobra.Command, args []string) {
			fn(c, args...)
		}
	}
}

// Short returns option to set the short
func Short(desc string) Option {
	return func(c *Command) {
		c.Command.Short = desc
	}
}

// Long returns option to set the long
func Long(desc string) Option {
	return func(c *Command) {
		c.Command.Long = desc
	}
}

// Description returns option to set descrtiption
func Description(desc string) Option {
	return func(c *Command) {
		c.Command.Long = desc
	}
}

// Example returns option to set example
func Example(ex string) Option {
	return func(c *Command) {
		c.Command.Example = ex
	}
}

// Config ...
func Config(v interface{}) Option {
	// set flag from opts
	// new opts from opts
	n := opts.New(v)
	return func(c *Command) {
		c.setflag = func(c *Command) {
			f := c.Flags()
			for _, o := range n.Opts() {
				f.VarP(o.Item(), o.Name(), o.Short(), o.Help())
			}
		}
	}
}

// SetFlags ...
func SetFlags(setflag func(c *Command)) Option {
	return func(c *Command) {
		oldsetflag := c.setflag
		if oldsetflag == nil {
			c.setflag = setflag
		} else {
			c.setflag = func(c *Command) {
				oldsetflag(c)
				setflag(c)
			}
		}
	}
}
