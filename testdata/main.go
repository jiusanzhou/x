package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"go.zoe.im/x/cli"
	"go.zoe.im/x/cli/config"
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

func main() {

	// flagParse()
	cliRun()
}

func flagParse() {
	fset := pflag.NewFlagSet("x", pflag.PanicOnError)

	var cfg globallConfig

	fset.StringVar(&cfg.Name, "name", "", "Just for test")

	fset.Parse(os.Args)

	fmt.Println("===>", cfg.Name)

	_, err := config.New(&cfg, config.WithType("json"), config.WithName("./testdata/config"))
	if err != nil {
		fmt.Println("[ERROR]", err)
	}

	fmt.Println("===>", cfg.Name)

	fset.Parse(strings.Split("xx  --name aaax", " "))

	fmt.Println("===>", cfg.Name)
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
		cli.GlobalConfig(&cfg, cli.WithConfigName("config")),
		cli.Run(func(c *cli.Command, args ...string) {
			c.Help()
		}),
	)
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
