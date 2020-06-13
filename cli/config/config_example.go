package config

import (
	"fmt"
	"time"
)

// ExampleConfig example configuration
type ExampleConfig struct {
	// .env file we need to offer
	Timeout time.Duration // should we need to add all tags?
	Name    string        //
}

func newExample() {
	var c ExampleConfig
	// config.RegistLoader() 注册加载器
	// config.RegistProvider() 注册提供器
	// config.Local() config.Http() config.Type("json", "yaml", "toml")
	// config.WithName() ??? my-app
	// config.WithEnv("MY_APP_ENV", ".") my-app.test
	// config.New(c)

	/**
	// auto detect loader and provider
	config.New(
		&c,
		config.WithName("my-app"),
		config.WithType("json", "yaml"), // asiign to loader, register at first
		config.WithProvider("mysql"),// register a first
		config.WithEnv(), // distinguish different environment like develop, production and test
		// name: test, ci: ..., cd: ...
		config.WithChild("ci", &c1).WithChild("cd", &c2)
	)

	// params for new config from flags
	cli.Config(&c)
	*/

	fmt.Println("parse configuraton =>", c)
}
