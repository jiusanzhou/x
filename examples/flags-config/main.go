package main

import (
	"fmt"
	"log"
	"time"

	"go.zoe.im/x/cli"
	"go.zoe.im/x/version"
)

type globallConfig struct {
	Name string `opts:"env, group=demo" json:"name"`
	Male bool   `opts:"env, group=demo" json:"male"`
	// command=sleep prefix=sleep group=sleep
	// default global flag, if we add command, move to command
	Sleep *sleepConfig `opts:"prefix=sleep"` // flat to global
	Wait  *sleepConfig `opts:"command=wait"` // register for sub command
}

type sleepConfig struct {
	Time   time.Duration `json:"time"`
	Dreams []string      `opts:"env=DREAMS" json:"dreams"`
}

func cliRun() {
	cfg := globallConfig{
		Sleep: &sleepConfig{
			Time: time.Second * 2,
		},
		Wait: &sleepConfig{
			Time: time.Second * 1,
		},
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

	waitCfg := &sleepConfig{}

	cmd.Register(
		cli.New(
			cli.Name("wait"),
			cli.Run(func(c *cli.Command, args ...string) {
				fmt.Printf("Hello my name is %s, and I'm a ", cfg.Name)
				if cfg.Male {
					fmt.Println("man.")
				} else {
					fmt.Println("women.")
				}
				fmt.Println("Sleeping for", cfg.Sleep.Time)
				time.Sleep(cfg.Sleep.Time)
				fmt.Println("Wakeup.\x00")
				fmt.Println("Waiting for", waitCfg.Time)
				fmt.Println("Dreams", waitCfg.Dreams)
			}),
		),
	)
	cmd.Run()
}

func main() {
	cliRun()
}
