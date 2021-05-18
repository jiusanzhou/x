package main

import (
	"fmt"
	"log"
	"time"

	"go.zoe.im/x/cli"
	"go.zoe.im/x/version"
)

type globallConfig struct {
	Name  string `opts:"env, group=demo" json:"name"`
	Male  bool   `opts:"env, group=demo" json:"male"`
	Sleep time.Duration
	SleepConfig sleepConfig `opts:"-"` // global, or command sub flags
}

type sleepConfig struct {
	Time   time.Duration `json:"time"`
	Dreams []string      `opts:"env=DREAMS" json:"dreams"`
}

func cliRun() {
	cfg := globallConfig{
		Sleep: time.Second * 1,
	}
	cmd := cli.New(
		cli.Name("example-flags-config"),
		version.NewOption(true),
		cli.Run(func(c *cli.Command, args ...string) {
			c.Help()
		}),
	)

	cmd.Option(cli.GlobalConfig(
		&cfg,
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
