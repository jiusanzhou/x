// Copyright (c) 2020 wellwell.work, LLC by Zoe
//
// Licensed under the Apache License 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
Package cli provides a CLI builder with configuration management support.

This package wraps [github.com/spf13/cobra] to provide a more ergonomic
API for building command-line applications with integrated configuration
file support, environment variable binding, and flag parsing.

# Basic Usage

Create a simple CLI application:

	cmd := cli.New(
	    cli.Name("myapp"),
	    cli.Short("A brief description"),
	    cli.Run(func() error {
	        fmt.Println("Hello!")
	        return nil
	    }),
	)
	cmd.Run()

# With Configuration

Bind a configuration struct to flags and config files:

	type Config struct {
	    Port    int    `json:"port" flag:"port,p" usage:"Server port"`
	    Verbose bool   `json:"verbose" flag:"verbose,v" usage:"Enable verbose output"`
	}

	cfg := &Config{Port: 8080}

	cmd := cli.New(
	    cli.Name("server"),
	    cli.WithConfig(cfg),
	    cli.Run(func() error {
	        fmt.Printf("Starting on port %d\n", cfg.Port)
	        return nil
	    }),
	)
	cmd.Run()

# Subcommands

Register subcommands:

	root := cli.New(cli.Name("myapp"))

	start := cli.New(
	    cli.Name("start"),
	    cli.Run(startFunc),
	)

	stop := cli.New(
	    cli.Name("stop"),
	    cli.Run(stopFunc),
	)

	root.Register(start, stop)
	root.Run()

# Configuration Sources

Configuration is loaded from (in order of precedence):

 1. Command-line flags
 2. Environment variables
 3. Configuration files (YAML, JSON, TOML, HCL)
 4. Default values in the struct
*/
package cli
