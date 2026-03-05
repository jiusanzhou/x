package x

import (
	"encoding/json"
	"fmt"
	"reflect"
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
	return fmt.Sprintf("{%s@%s %s}", e.Name, e.Type, string(e.Config))
}

// Unmarshal parses the Config field into the provided object.
func (e *TypedLazyConfig) Unmarshal(obj any) error {
	return json.Unmarshal(e.Config, obj)
}

// Validate checks that Config can be unmarshalled into the given type.
// It returns a descriptive error with the field name when validation fails.
func (e *TypedLazyConfig) Validate(target any) error {
	if e.Type == "" {
		return fmt.Errorf("TypedLazyConfig: type is required")
	}
	if len(e.Config) == 0 {
		return nil // no config to validate
	}

	// Create a fresh instance of the target type
	t := reflect.TypeOf(target)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	v := reflect.New(t).Interface()

	if err := json.Unmarshal(e.Config, v); err != nil {
		return fmt.Errorf("TypedLazyConfig[%s/%s]: config validation failed: %w", e.Type, e.Name, err)
	}
	return nil
}

// MustUnmarshal parses Config into the provided object, panicking on error.
// Useful during startup / init() where failure is fatal.
func (e *TypedLazyConfig) MustUnmarshal(obj any) {
	if err := e.Unmarshal(obj); err != nil {
		panic(fmt.Sprintf("TypedLazyConfig[%s/%s].MustUnmarshal: %v", e.Type, e.Name, err))
	}
}
