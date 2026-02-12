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
Package opts provides struct-based command-line option parsing.

This package uses reflection to automatically generate command-line flags
from struct field tags, providing a declarative approach to CLI argument
parsing.

# Basic Usage

Define options using struct tags:

	type Options struct {
	    Port    int    `flag:"port,p" usage:"Server port" default:"8080"`
	    Verbose bool   `flag:"verbose,v" usage:"Enable verbose mode"`
	    Config  string `flag:"config,c" usage:"Config file path"`
	}

	opts := &Options{}
	o := opts.New(opts)
	o.Parse()

# Struct Tags

Supported struct tags:

  - flag: Flag name and optional short form (e.g., "port,p")
  - usage: Help text for the flag
  - default: Default value
  - env: Environment variable name

# Types

The package supports these field types:

  - string, int, int64, float64, bool
  - time.Duration
  - []string (for repeated flags)
  - Custom types implementing [flag.Value]

# Environment Variables

Bind flags to environment variables:

	type Options struct {
	    APIKey string `flag:"api-key" env:"API_KEY"`
	}

# Interfaces

  - Opt: Represents a single option/flag
  - Opts: Collection of options with parsing capability
  - ParsedOpts: Result of parsing with Run/Help methods
  - Value: Interface for custom flag types (compatible with flag.Value)
*/
package opts
