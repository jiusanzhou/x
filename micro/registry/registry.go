// Package registry is an interface for service discovery
package registry

import (
	"errors"
)

var (
	// ErrNotFound error when GetService is called
	ErrNotFound = errors.New("service not found")
)