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
	Config      []string `opts:"env,short=c,help=configuration's name(config); without extension name(toml|yaml|json)"`
	ConfigTypes []string `opts:"-"`

	onChanged func(o, n interface{})
}

// ConfigOption defined config option for cli
type ConfigOption func(co *configOptions)

// WithConfigName set config name
func WithConfigName(names ...string) ConfigOption {
	return func(co *configOptions) {
		co.Config = append(co.Config, names...)
	}
}

// WithConfigChanged set a watcher to watch the config changed
// params's type can be auto set with interface{}
func WithConfigChanged(f func(o, n interface{})) ConfigOption {
	return func(co *configOptions) {
		oldfn := co.onChanged
		if oldfn != nil {
			co.onChanged = func(o, n interface{}) {
				oldfn(o, n)
				f(o, n)
			}
		} else {
			co.onChanged = f
		}
	}
}

// WithConfigType set config

func newConfigOptions() *configOptions {
	return &configOptions{
		Config: []string{},

		// TODO: opts can't supported slice with default values
		ConfigTypes: []string{"toml", "yaml", "json"},
	}
}

// create a new Config options from flags
func (c *configOptions) build() []config.Option {
	opts := []config.Option{}

	if len(c.Config) > 0 {
		opts = append(opts, config.WithNames(c.Config...))
	}

	if len(c.ConfigTypes) > 0 {
		opts = append(opts, config.WithType(c.ConfigTypes...))
	}

	// with default provider is current path

	// create the watcher
	if c.onChanged != nil {
		opts = append(opts, config.WithConfigChanged(c.onChanged))
	}

	return opts
}
