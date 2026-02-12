package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"go.zoe.im/x/cli/config"
)

func NewConfigCommand(cfg *config.Config) *Command {
	var options = &struct {
		Type string `opts:"short=t,help=config output type"`
	}{
		Type: "yaml",
	}

	cfgCmd := New(
		Name("config"),
		Short("Print the application configuration information"),
		Description("Print the application configuration information"),
	)

	cfgCmd.Register(New(
		Name("default"),
		Short("Print the application default configuration information"),
		Description("Print the application default configuration information"),
		Config(options),
		Run(func(cmd *Command, args ...string) {
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
			data, err := config.Encode(options.Type, cfg.Current())
			if err != nil {
				log.Fatalln(err)
			}
			os.Stdout.Write(data)
		}),
	))

	var newOptions = &struct {
		Type   string `opts:"short=t,help=config file type (yaml/json/toml/hcl/xml)"`
		Output string `opts:"short=o,help=output file path (default: stdout)"`
		Force  bool   `opts:"short=f,help=overwrite existing file"`
	}{
		Type:   "yaml",
		Output: "",
		Force:  false,
	}

	cfgCmd.Register(New(
		Name("new", "init", "create"),
		Short("Create a new example configuration file"),
		Description("Generate a new configuration file with default values. Use -o to write to a file, or leave empty to print to stdout."),
		Config(newOptions),
		Run(func(cmd *Command, args ...string) {
			data, err := config.Encode(newOptions.Type, cfg.Default())
			if err != nil {
				log.Fatalln("failed to encode config:", err)
			}

			if newOptions.Output == "" {
				os.Stdout.Write(data)
				return
			}

			outputPath := newOptions.Output
			if filepath.Ext(outputPath) == "" {
				outputPath = outputPath + "." + newOptions.Type
			}

			if !newOptions.Force {
				if _, err := os.Stat(outputPath); err == nil {
					log.Fatalf("file %s already exists, use -f to overwrite", outputPath)
				}
			}

			if err := os.WriteFile(outputPath, data, 0644); err != nil {
				log.Fatalln("failed to write config file:", err)
			}

			fmt.Printf("config file created: %s\n", outputPath)
		}),
	))

	return cfgCmd
}

func init() {
}
