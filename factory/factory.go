package factory

import (
	"fmt"

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
}

// Creator is a function that creates an instance of a type
// using the given configuration and options.
type Creator[Interface any, Option any] func(cfg x.TypedLazyConfig, opts ...Option) (Interface, error)

// creatorFactory is a factory that stores creator functions for a type.
// It is used to register creator functions and create instances of a type
// using the registered creators.
type creatorFactory[Interface any, Option any] struct {
	// creators is a map that stores creator functions for a type.
	// It is used to store the creator functions registered by the user.
	creators x.SyncMap[string, Creator[Interface, Option]]
}

// Register adds a creator function with a given name and optional aliases
// to the factory. It returns an error if the registration fails.
func (c *creatorFactory[Interface, Option]) Register(typeName string, creator Creator[Interface, Option], alias ...string) error {
	if _, exists := c.creators.Load(typeName); exists {
		return fmt.Errorf("creator for %s already exists", typeName)
	}
	c.creators.Store(typeName, creator)
	for _, a := range alias {
		c.creators.Store(a, creator)
	}
	return nil
}

// Create initializes an instance of the specified type using the provided
// configuration and options. It returns the created instance and an error
// if the creation fails.
func (c *creatorFactory[Interface, Option]) Create(cfg x.TypedLazyConfig, opts ...Option) (Interface, error) {
	creator, ok := c.creators.Load(cfg.Type)
	if !ok {
		var null Interface
		return null, fmt.Errorf("no creator for %s", cfg.Type)
	}
	return creator(cfg, opts...)
}

// NewFactory creates a new instance of the default implementation of the Factory interface.
func NewFactory[Interface any, Option any]() Factory[Interface, Option] {
	return &creatorFactory[Interface, Option]{
		creators: x.SyncMap[string, Creator[Interface, Option]]{},
	}
}
