package cli

import (
	"github.com/spf13/cobra"

	"go.zoe.im/x/cli/config"
	"go.zoe.im/x/cli/opts"
)

// TODO: implement my own cli package

// Command is the main struct which comes from cobar
type Command struct {
	*cobra.Command
	setflag func(c *Command)
	// store opts
	globalOpts []opts.Opts
	opts       []opts.Opts

	// config value, we only support one config
	configv   interface{}
	configobj *config.Config

	// root command
	root *Command

	// parent command
	parent *Command
}

// InitFlags init flags from config
func (c *Command) InitFlags() {

	// loads flags from config flag
	if c.globalOpts != nil {
		SetFlags(func(c *Command) {
			_setFlagsFromConfig(c.PersistentFlags(), c.globalOpts)
		})(c)
		PersistentPreRun(func(c *Command, args ...string) {
			_parseGlobalFlags(c.globalOpts)
		})(c)
	}

	if c.opts != nil {
		SetFlags(func(c *Command) {
			_setFlagsFromConfig(c.Flags(), c.opts)
		})(c)
		PreRun(func(c *Command, args ...string) {
			_parseFlags(c.opts)
		})(c)
	}

	// init functions
	if c.setflag != nil {
		c.setflag(c)
	}

	// add load prerun
}

// New returns a wraper of cobra
func newFromCobra() *Command {
	c := &Command{
		// Init the cobar command
		Command: &cobra.Command{},
	}

	// set parent and root to self, or nil
	c.parent = c
	c.root = c

	return c
}
