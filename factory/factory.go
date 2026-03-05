package factory

import (
	"fmt"
	"sort"
	"strings"

	"go.zoe.im/x"
)

// Factory is an interface that defines methods for registering and creating
// instances using a specified configuration and options.
type Factory[Interface any, Option any] interface {
	// Register adds a creator function with a given name and optional aliases
	// to the factory. It returns an error if the registration fails.
	Register(typeName string, creator Creator[Interface, Option], alias ...string) error

	// Create initializes an instance of the specified type using the provided
	// configuration and options. It returns the created instance and an error
	// if the creation fails.
	Create(cfg x.TypedLazyConfig, opts ...Option) (Interface, error)

	// MustCreate is like Create but panics on error. Useful during init().
	MustCreate(cfg x.TypedLazyConfig, opts ...Option) Interface

	// Has reports whether a creator for the given type name is registered.
	Has(typeName string) bool

	// List returns all registered type names (sorted, no aliases).
	List() []string

	// Types returns all registered type names including aliases (sorted).
	Types() []string
}

// Creator is a function that creates an instance of a type
// using the given configuration and options.
type Creator[Interface any, Option any] func(cfg x.TypedLazyConfig, opts ...Option) (Interface, error)

// creatorFactory is a factory that stores creator functions for a type.
type creatorFactory[Interface any, Option any] struct {
	creators x.SyncMap[string, Creator[Interface, Option]]
	// primary tracks non-alias names for List()
	primary x.SyncMap[string, struct{}]
}

// Register adds a creator function with a given name and optional aliases
// to the factory. It returns an error if the registration fails.
func (c *creatorFactory[Interface, Option]) Register(typeName string, creator Creator[Interface, Option], alias ...string) error {
	if _, exists := c.creators.Load(typeName); exists {
		return fmt.Errorf("creator for %s already exists", typeName)
	}
	c.creators.Store(typeName, creator)
	c.primary.Store(typeName, struct{}{})
	for _, a := range alias {
		c.creators.Store(a, creator)
	}
	return nil
}

// Create initializes an instance of the specified type using the provided
// configuration and options. It returns the created instance and an error
// if the creation fails. The error message includes all registered types
// for easier debugging.
func (c *creatorFactory[Interface, Option]) Create(cfg x.TypedLazyConfig, opts ...Option) (Interface, error) {
	creator, ok := c.creators.Load(cfg.Type)
	if !ok {
		var null Interface
		registered := c.Types()
		if len(registered) == 0 {
			return null, fmt.Errorf("no creator for type %q (no types registered)", cfg.Type)
		}
		return null, fmt.Errorf("no creator for type %q (registered: %s)", cfg.Type, strings.Join(registered, ", "))
	}
	return creator(cfg, opts...)
}

// MustCreate is like Create but panics on error.
func (c *creatorFactory[Interface, Option]) MustCreate(cfg x.TypedLazyConfig, opts ...Option) Interface {
	v, err := c.Create(cfg, opts...)
	if err != nil {
		panic(fmt.Sprintf("factory.MustCreate: %v", err))
	}
	return v
}

// Has reports whether a creator for the given type name is registered.
func (c *creatorFactory[Interface, Option]) Has(typeName string) bool {
	_, ok := c.creators.Load(typeName)
	return ok
}

// List returns all registered primary type names (sorted, no aliases).
func (c *creatorFactory[Interface, Option]) List() []string {
	var names []string
	c.primary.Range(func(key string, _ struct{}) bool {
		names = append(names, key)
		return true
	})
	sort.Strings(names)
	return names
}

// Types returns all registered type names including aliases (sorted).
func (c *creatorFactory[Interface, Option]) Types() []string {
	var names []string
	c.creators.Range(func(key string, _ Creator[Interface, Option]) bool {
		names = append(names, key)
		return true
	})
	sort.Strings(names)
	return names
}

// NewFactory creates a new instance of the default implementation of the Factory interface.
func NewFactory[Interface any, Option any]() Factory[Interface, Option] {
	return &creatorFactory[Interface, Option]{
		creators: x.SyncMap[string, Creator[Interface, Option]]{},
		primary:  x.SyncMap[string, struct{}]{},
	}
}
