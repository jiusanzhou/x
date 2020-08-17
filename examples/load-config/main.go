package main

import (
	"fmt"
	"time"

	"go.zoe.im/x/cli"
	"go.zoe.im/x/version"
)

type globallConfig struct {
	Name  string `opts:"env" json:"name"`
	Male  bool   `opts:"env" json:"male"`
	Sleep time.Duration
}

func cliRun() {

	var cfg globallConfig

	cmd := cli.New(
		cli.Name("test"),
		version.NewOption(true),
		cli.GlobalConfig(
			&cfg,
			cli.WithConfigName("config"),
		), // this should change the default config name
		cli.Run(func(c *cli.Command, args ...string) {
			fmt.Println("=====> Name:", cfg.Name)
			fmt.Println("=====> Male:", cfg.Male)
			fmt.Println("=====> Sleep:", cfg.Sleep)
			time.Sleep(cfg.Sleep)
		}),
	)

	cmd.Run()
}

func main() {
	cliRun()
}
