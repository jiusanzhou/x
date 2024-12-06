package x

import (
	"encoding/json"
	"fmt"
)

// TypedLazyConfig represents a configuration with type information that can be lazily processed.
type TypedLazyConfig struct {
	Name   string          `json:"name,omitempty" yaml:"name"`     // Name is the name of the configuration.
	Type   string          `json:"type,omitempty" yaml:"type"`     // Type indicates the configuration type.
	Config json.RawMessage `json:"config,omitempty" yaml:"config"` // Config holds the raw JSON configuration data.
}

// TypedLazyConfigs is a slice of TypedLazyConfig pointers.
type TypedLazyConfigs []*TypedLazyConfig

// TypeLazyConfigWithSelector extends TypedLazyConfig with selectors for additional configuration.
type TypeLazyConfigWithSelector struct {
	TypedLazyConfig
	Selectors Selectors `json:"selectors,omitempty" yaml:"selectors"` // Selectors specify criteria for selecting configurations.
}

// TypeLazyConfigWithSelectors is a slice of TypeLazyConfigWithSelector pointers.
type TypeLazyConfigWithSelectors []*TypeLazyConfigWithSelector

// String returns a string representation of the TypedLazyConfig.
func (e *TypedLazyConfig) String() string {
	// Format the TypedLazyConfig fields into a string.
	return fmt.Sprintf("{%s@%s %s}", e.Name, e.Type, string(e.Config))
}
