package cli

import (
	"os"

	"go.zoe.im/x/cli/opts"

	"github.com/spf13/pflag"
)

func _setFlagsFromConfig(f *pflag.FlagSet, n opts.Opts) {
	for _, o := range n.Opts() {
		flag := f.VarPF(o.Item(), o.Name(), o.Short(), o.Help())
		// add default for flag
		// fmt.Println("type:", o.Item().Type(), "value:", o.Item().String())
		if o.Item().Type() == "bool" {
			flag.NoOptDefVal = "true"
		}
		// TODO: set value from env,  load from envs at opts.New
	}
}

func _parseGlobalFlags(n opts.Opts) {
	for _, o := range n.Opts() {
		if !o.HasSet() && o.FromEnv() {
			o.Item().Set(os.Getenv(o.EnvName()))
		}
	}
}

func _parseFlags(n opts.Opts) {
	for _, o := range n.Opts() {
		if !o.HasSet() && o.FromEnv() {
			o.Item().Set(os.Getenv(o.EnvName()))
		}
	}
}
