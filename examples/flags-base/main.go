package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"go.zoe.im/x/cli/config"
)

type globallConfig struct {
	Name  string `opts:"env" json:"name"`
	Male  bool   `opts:"env" json:"male"`
	Sleep time.Duration
}

func flagParse() {
	fset := pflag.NewFlagSet("x", pflag.PanicOnError)

	var cfg globallConfig

	fset.StringVar(&cfg.Name, "name", "", "Just for test")

	fset.Parse(os.Args)

	fmt.Println("===>", cfg.Name)

	err := config.New(&cfg, config.WithType("json"), config.WithName("./testdata/config")).Init()
	if err != nil {
		fmt.Println("[ERROR]", err)
	}

	fmt.Println("===>", cfg.Name)

	fset.Parse(strings.Split("xx  --name aaax", " "))

	fmt.Println("===>", cfg.Name)
}

func main() {
	flagParse()
}
