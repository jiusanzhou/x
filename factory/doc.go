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
Package factory provides a generic factory pattern implementation for creating
instances of types based on configuration.

The factory pattern allows registering creator functions with a name and optional
aliases, then creating instances by specifying the type name in configuration.
This is useful for plugin systems where different implementations can be
registered and instantiated dynamically.

# Basic Usage

Create a factory, register creators, and create instances:

	// Define your interface and option types
	type Plugin interface {
	    Execute() error
	}
	type PluginOption struct {
	    Debug bool
	}

	// Create a factory
	f := factory.NewFactory[Plugin, PluginOption]()

	// Register a creator
	f.Register("example", func(cfg x.TypedLazyConfig, opts ...PluginOption) (Plugin, error) {
	    var config ExampleConfig
	    cfg.Unmarshal(&config)
	    return NewExamplePlugin(config), nil
	}, "ex", "sample") // with aliases

	// Create an instance
	cfg := x.TypedLazyConfig{Type: "example", Config: json.RawMessage(`{}`)}
	plugin, err := f.Create(cfg)

# Thread Safety

The factory uses [x.SyncMap] internally, making it safe for concurrent
registration and creation operations.
*/
package factory
