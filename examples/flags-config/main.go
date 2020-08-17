package main

import (
	"fmt"
	"log"
	"time"

	"go.zoe.im/x/cli"
	"go.zoe.im/x/version"
)

type globallConfig struct {
	Name  string `opts:"env" json:"name"`
	Male  bool   `opts:"env" json:"male"`
	Sleep time.Duration
}

type sleepConfig struct {
	Time   time.Duration `json:"time"`
	Dreams []string      `opts:"env=DREAMS" json:"dreams"`
}

func cliRun() {
	var cfg globallConfig
	// _, err := config.New(&cfg)
	// if err != nil {
	// 	fmt.Println("create a configuration error:", err)
	// }
	cmd := cli.New(
		cli.Name("test"),
		version.NewOption(true),
		cli.Run(func(c *cli.Command, args ...string) {
			c.Help()
		}),
	)

	cmd.Option(cli.GlobalConfig(
		&cfg,
		cli.WithConfigName(),
		cli.WithConfigChanged(func(o, n interface{}) {
			log.Println("config changed:", n)
		}),
	))

	sleepCfg := &sleepConfig{
		// Dreams: []string{"a", "v"},
	}

	cmd.Register(
		cli.New(
			cli.Name("stop"),
			cli.Config(sleepCfg),
			cli.Run(func(c *cli.Command, args ...string) {
				fmt.Printf("Hello my name is %s, and I'm a ", cfg.Name)
				if cfg.Male {
					fmt.Println("man.")
				} else {
					fmt.Println("women.")
				}
				fmt.Println("Sleeping for", cfg.Sleep)
				time.Sleep(cfg.Sleep)
				fmt.Println("Wakeup.\x00")
				fmt.Println("Sleeping for", sleepCfg.Time)
				fmt.Println("Dreams", sleepCfg.Dreams)
			}),
		),
	)
	cmd.Run()
}

func main() {
	cliRun()
}
