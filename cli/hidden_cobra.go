package cli

import (
	"github.com/spf13/cobra"

	"go.zoe.im/x/cli/opts"
)

// TODO: implement my own cli package

// Command is the main struct which comes from cobar
type Command struct {
	*cobra.Command
	setflag func(c *Command)
	// store opts
	globalOpts opts.Opts
	opts       opts.Opts
}

func (c *Command) initFlags() {

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
	return &Command{
		// Init the cobar command
		Command: &cobra.Command{},
	}
}
