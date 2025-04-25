package main

import (
	"fmt"
	"time"

	"go.zoe.im/x"
	"go.zoe.im/x/cli"
	"go.zoe.im/x/cli/config"
	"go.zoe.im/x/version"
)

type globallConfig struct {
	Name  string     `opts:"env" json:"name"`
	Male  bool       `opts:"env" json:"male"`
	Sleep x.Duration `opts:"name=sleep" json:"sleep"`
}

func cliRun() {

	var cfg globallConfig

	fsprovider, _ := config.NewFSProvider("./examples")

	cmd := cli.New(
		cli.Name("test"),
		version.NewOption(true),
		cli.GlobalConfig(
			&cfg,
			cli.WithConfigName("config", "config2"),
			cli.WithConfigOptions(config.WithProvider(fsprovider)),
		), // this should change the default config name
		cli.Run(func(c *cli.Command, args ...string) {
			fmt.Println("=====> Name:", cfg.Name)
			fmt.Println("=====> Male:", cfg.Male)
			fmt.Println("=====> Sleep:", cfg.Sleep)
			time.Sleep(cfg.Sleep.Duration())
		}),
	)

	cmd.Run()
}

func main() {
	cliRun()
}
