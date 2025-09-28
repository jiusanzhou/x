package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"time"

	"go.zoe.im/x"
	"go.zoe.im/x/cli"
	"go.zoe.im/x/cli/config"
	"go.zoe.im/x/version"
)

type globallConfig struct {
	Name    string     `opts:"env" json:"name"`
	Male    bool       `opts:"env" json:"male"`
	Sleep   x.Duration `opts:"name=sleep" json:"sleep"`
	Message string     `opts:"env,name=message" json:"message"`
}

func cliRun() {

	cfg := &globallConfig{
		Sleep:   x.Duration(time.Second * 1),
		Message: "Hello default message",
	}

	fsprovider, _ := config.NewFSProvider("./examples")

	cmd := cli.New(
		cli.Name("test"),
		version.NewOption(true),
		cli.GlobalConfig(
			cfg,
			cli.WithConfigName("config", "config2"),
			cli.WithConfigOptions(config.WithProvider(fsprovider)),
			cli.WithConfigChanged(func(o, n interface{}) {
				// print the pointer of o and n
				fmt.Printf("config changed: %p %p\n", o, n)
				fmt.Printf("config changed: %v %v\n", o, n)
			}),
			cli.WithConfigCommand(true),
		), // this should change the default config name
		cli.Run(func(c *cli.Command, args ...string) {
			fmt.Println("=====> Name:", cfg.Name)
			fmt.Println("=====> Male:", cfg.Male)
			fmt.Println("=====> Sleep:", cfg.Sleep)
			fmt.Println("=====> Sleep:", cfg.Message)

			var demo any

			demo = &globallConfig{
				Name: "demo",
				Male: true,
			}

			// create the demo2
			demo2 := reflect.New(reflect.TypeOf(demo).Elem()).Interface()
			// unmarshal to demo2
			json.Unmarshal([]byte(`{"sleep": "1s"}`), demo2)
			// copy to demo
			reflect.ValueOf(demo).Elem().Set(reflect.ValueOf(demo2).Elem())

			json.NewEncoder(os.Stdout).Encode(demo)

			time.Sleep(cfg.Sleep.Duration())
		}),
	)

	cmd.Run()
}

func main() {
	cliRun()
}
