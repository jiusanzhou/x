package cli

import (
	"github.com/spf13/cobra"
)

// TODO: implement my own cli package

// Command is the main struct which comes from cobar
type Command struct {
	*cobra.Command
	setflag func(c *Command)
}

// New returns a wraper of cobra
func newFromCobra() *Command {
	return &Command{
		// Init the cobar command
		Command: &cobra.Command{},
	}
}
