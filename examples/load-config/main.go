package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"go.zoe.im/x"
	"go.zoe.im/x/cli"
	"go.zoe.im/x/cli/config"
	"go.zoe.im/x/version"
)

type lazyConfig struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

type globallConfig struct {
	Name       string          `opts:"env" json:"name"`
	Male       bool            `opts:"env" json:"male"`
	Sleep      x.Duration      `opts:"name=sleep" json:"sleep"`
	Message    string          `json:"message"`
	NightSleep *sleepConfig    `json:"night_sleep"`
	Lazy       json.RawMessage `json:"lazy"`
}

type sleepConfig struct {
	Place    string     `json:"place"`
	Duration x.Duration `opts:"name=duration" json:"duration"`
}

func cliRun() {

	cfg := &globallConfig{
		Sleep:   x.Duration(time.Second * 1),
		Message: "Hello default message",
		NightSleep: &sleepConfig{
			Place:    "home",
			Duration: x.Duration(time.Second * 2),
		},
	}

	fsprovider, _ := config.NewFSProvider("./examples")

	cmd := cli.New(
		cli.Name("test"),
		version.NewOption(true),
		cli.GlobalConfig(
			cfg,
			cli.WithConfigName("config"),
			cli.WithConfigOptions(config.WithProvider(fsprovider)),
			cli.WithConfigChanged(func(o, n interface{}) {
				// print the pointer of o and n
				oc := o.(*globallConfig)
				nc := n.(*globallConfig)
				fmt.Printf("config changed: %p %p\n", o, n)
				fmt.Printf("config changed: %v %v\n", o, n)
				fmt.Printf("config changed: %v %v\n", oc.NightSleep, nc.NightSleep)
			}),
			cli.WithConfigCommand(true),
		), // this should change the default config name
		cli.Run(func(c *cli.Command, args ...string) {
			fmt.Println("=====> Name:", cfg.Name)
			fmt.Println("=====> Male:", cfg.Male)
			fmt.Println("=====> Sleep:", cfg.Sleep)
			fmt.Println("=====> Sleep:", cfg.Message)
			fmt.Println("=====> Sleep:", cfg.NightSleep)

			json.NewEncoder(os.Stdout).Encode(cfg)

			time.Sleep(cfg.Sleep.Duration())
		}),
	)

	cmd.Run()
}

func main() {
	cliRun()
}
