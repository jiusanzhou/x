/*
 * Copyright (c) 2020 wellwell.work, LLC by Zoe
 *
 * Licensed under the Apache License 2.0 (the "License");
 * You may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cli

import (
	"go.zoe.im/x/cli/config"
)

// configConfig contains options for config parse
type configOptions struct {
	// config file name
	Config      string   `opts:"env,short=c,help=configuration's name; without extension name(toml|yaml|json)"`
	ConfigTypes []string `opts:"-"`
}

// ConfigOption defined config option for cli
type ConfigOption func(co *configOptions)

// WithConfigName set config name
func WithConfigName(name string) ConfigOption {
	return func(co *configOptions) {
		co.Config = name
	}
}

// WithConfigType set config

func newConfigOptions() *configOptions {
	return &configOptions{
		Config: "config",
		// TODO: opts can't supported slice with default values
		ConfigTypes: []string{"toml", "yaml", "json"},
	}
}

// create a new Config options from flags
func (c *configOptions) build() []config.Option {
	opts := []config.Option{}

	if c.Config != "" {
		opts = append(opts, config.WithName(c.Config))
	}

	if len(c.ConfigTypes) > 0 {
		opts = append(opts, config.WithType(c.ConfigTypes...))
	}

	// with default provider is current path

	return opts
}
