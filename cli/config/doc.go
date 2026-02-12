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
Package config provides configuration loading with multiple sources and hot-reload.

This package supports loading configuration from multiple file formats,
merging configurations from multiple sources, and watching for changes
with automatic reload.

# Basic Usage

Create and load a configuration:

	type AppConfig struct {
	    Port     int    `json:"port"`
	    LogLevel string `json:"logLevel"`
	}

	cfg := &AppConfig{Port: 8080} // defaults
	c := config.New(cfg,
	    config.Name("config"),
	    config.WithPaths(".", "/etc/myapp"),
	)

	err := c.Init()
	// cfg is now populated from config files

# Supported Formats

  - JSON (.json)
  - YAML (.yaml, .yml)
  - TOML (.toml)
  - HCL (.hcl)
  - XML (.xml)

# Multiple Sources

Load and merge from multiple config files:

	c := config.New(cfg,
	    config.Names("defaults", "config", "local"),
	    config.WithPaths(".", "/etc/myapp"),
	)

Files are merged in order, with later files taking precedence.

# Hot Reload

Watch for configuration changes:

	c := config.New(cfg,
	    config.Name("config"),
	    config.OnChanged(func(old, new any) {
	        fmt.Println("Config updated!")
	    }),
	)

# Configuration Access

Access configuration values:

	c.Current()   // Get current config value
	c.Default()   // Get default config value
	c.Get("key")  // Get specific key from merged config
	c.CopyValue(&v) // Deep copy current config to v
*/
package config
