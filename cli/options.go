package cli

import "github.com/spf13/cobra"

// Option set the command with value
type Option func(c *Command)

// Version returns option to set set the version
func Version(v string) Option {
	return func(c *Command) {
		c.Command.Version = v
	}
}

// Name returns option to set use
func Name(name string) Option {
	return func(c *Command) {
		c.Command.Use = name
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
