package cli

import (
	"log"
	"os"

	"go.zoe.im/x/cli/config"
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

// PersistentPreRun children of this command will inherit and execute.
func PersistentPreRun(fn func(cmd *Command, args ...string)) Option {
	return func(c *Command) {
		oldfn := c.Command.PersistentPreRun
		if oldfn != nil {
			c.Command.PersistentPreRun = func(cmd *cobra.Command, args []string) {
				// oldfn(c, args...)
				oldfn(cmd, args)
				fn(c, args...)
			}
		} else {
			c.Command.PersistentPreRun = func(cmd *cobra.Command, args []string) {
				fn(c, args...)
			}
		}

	}
}

// PreRun children of this command will not inherit.
func PreRun(fn func(cmd *Command, args ...string)) Option {
	return func(c *Command) {
		oldfn := c.Command.PreRun
		if oldfn != nil {
			c.Command.PreRun = func(cmd *cobra.Command, args []string) {
				oldfn(cmd, args)
				fn(c, args...)
			}
		} else {
			c.Command.PreRun = func(cmd *cobra.Command, args []string) {
				fn(c, args...)
			}
		}
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

// GlobalConfig returns option to set the global config
var _lockGlobalConfigOption bool

func GlobalConfig(v any, cfos ...ConfigOption) Option {
	// create a new config loader, and load content
	// - add a config file flag to flagset, create flags with config
	// - create a config instance

	if _lockGlobalConfigOption {
		panic("GlobalConfig can only be called once")
	}

	_lockGlobalConfigOption = true

	// with out default value at here
	cfopts := newConfigOptions()

	return func(c *Command) {
		for _, o := range cfos {
			o(cfopts)
		}

		// register config flags
		c.globalOpts = append(c.globalOpts, opts.New(cfopts))

		if cfopts.AutoFlags {
			// create a new flags set from config struct
			// generate flags from config
			// load config from source before flags parsed(get flags)
			c.globalOpts = append(c.globalOpts, opts.New(v))
		}

		// parse flags while onchanged, before call custom onchanged
		// at last version, this function called before o, why???
		if cfopts.onChanged != nil {
			WithConfigChanged(func(o, n any) { c.ParseFlags(os.Args) })(cfopts)
		}

		// create config from flags
		c.configobj = config.New(v)

		// register the config command if needs
		if cfopts.enableCommand {
			c.Register(NewConfigCommand(c.configobj))
		}

		// registe PersistentPreRun, but when to get flags
		PersistentPreRun(func(cmd *Command, _ ...string) {
			// check if the config flags is setted

			if err := c.configobj.Init(cfopts.build()...); err != nil {
				log.Println("[WARN] init config error:", err)
				return
			}

			// can reset with flag parse
			// parsed flags again to set to v
			cmd.ParseFlags(os.Args)
		})(c)

	}
}

// Config loads configuration from provider
func Config(v any) Option {
	return func(c *Command) {
		c.opts = append(c.opts, opts.New(v))
		// we won't to register opts
		// but we can load from global config
		// TODO: how to load config with command name
		c.configv = v
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

// RunE returns option to set the main run function with error handling.
// If the function returns an error, it will be printed to stderr.
func RunE(fn func(cmd *Command, args ...string) error) Option {
	return func(c *Command) {
		c.Command.RunE = func(cmd *cobra.Command, args []string) error {
			return fn(c, args...)
		}
	}
}

// HelpTemplate sets a custom help template.
func HelpTemplate(tpl string) Option {
	return func(c *Command) {
		c.Command.SetHelpTemplate(tpl)
	}
}

// UsageTemplate sets a custom usage template.
func UsageTemplate(tpl string) Option {
	return func(c *Command) {
		c.Command.SetUsageTemplate(tpl)
	}
}

// VersionTemplate sets a custom version template.
func VersionTemplate(tpl string) Option {
	return func(c *Command) {
		c.Command.SetVersionTemplate(tpl)
	}
}

// HelpFunc sets a custom help function.
func HelpFunc(fn func(cmd *Command, args []string)) Option {
	return func(c *Command) {
		c.Command.SetHelpFunc(func(cmd *cobra.Command, args []string) {
			fn(c, args)
		})
	}
}

// UsageFunc sets a custom usage function.
func UsageFunc(fn func(cmd *Command) error) Option {
	return func(c *Command) {
		c.Command.SetUsageFunc(func(cmd *cobra.Command) error {
			return fn(c)
		})
	}
}
