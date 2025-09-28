package cli

import (
	"log"
	"os"

	"go.zoe.im/x/cli/config"
)

func NewConfigCommand(cfg *config.Config) *Command {
	var options = &struct {
		Type string `opts:"short=t,help=config output type"`
	}{
		Type: "yaml",
	}

	// rootCmd is the root command
	cfgCmd := New(
		Name("config"),
		Short("Print the application configuration information"),
		Description("Print the application configuration information"),
	)

	// register sub command for cmd config
	cfgCmd.Register(New(
		Name("default"),
		Short("Print the application default configuration information"),
		Description("Print the application default configuration information"),
		Config(options),
		Run(func(cmd *Command, args ...string) {
			// print the default config
			data, err := config.Encode(options.Type, cfg.Default())
			if err != nil {
				log.Fatalln(err)
			}
			os.Stdout.Write(data)
		}),
	))

	cfgCmd.Register(New(
		Name("current", "dump"),
		Short("Print the application current configuration information"),
		Description("Print the application current configuration information"),
		Config(options),
		Run(func(cmd *Command, args ...string) {
			// print the current config
			data, err := config.Encode(options.Type, cfg.Current())
			if err != nil {
				log.Fatalln(err)
			}
			os.Stdout.Write(data)
		}),
	))

	return cfgCmd
}

func init() {
}
