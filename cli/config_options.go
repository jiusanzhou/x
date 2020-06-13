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
	Config         string `opts:"env,short=c,help=configuration's name"`
	ConfigType     string `opts:"env,help=configuration encoding format"`
	ConfigProvider string `opts:"env,help=configuration provider"`
	// ConfigTypes     []string `opts:"env,help=configuration encoding format"`
	// ConfigProviders []string `opts:"env,help=configuration provider"`
}

func newConfigOptions() configOptions {
	return configOptions{
		Config:         "config",
		ConfigType:     "yaml",
		ConfigProvider: "",
		// TODO: opts can't supported slice with default values
		// ConfigTypes: []string{"yaml", "toml"},
		// ConfigProviders: []string{"./"},
	}
}

// create a new Config options from flags
func (c *configOptions) build() []config.Option {
	opts := []config.Option{}

	if c.Config != "" {
		opts = append(opts, config.WithName(c.Config))
	}

	if len(c.ConfigType) > 0 {
		opts = append(opts, config.WithType(c.ConfigType))
	}

	if len(c.ConfigProvider) > 0 {
		p, err := config.NewProviderFromURI(c.ConfigProvider)
		if err == nil {
			opts = append(opts, config.WithProvider(p))
		}
	}

	return opts
}
