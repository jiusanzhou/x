// Package store is an interface for distribute data storage.
// The first version comes from eagle/storage
package store

import (
	"context"
	"errors"
)

var (
	// ErrNotFound returns when query record doesn't exit
	ErrNotFound = errors.New("not found")
	// ErrExits ...
	ErrExits = errors.New("aleary exits")
)

// Store is a data stroage interface
// TODO: add more like spawn a new store from bucket?
type Store interface {
	// TODO: batch CRUD
	List(ctx context.Context, ops ...ListOption) ([]Record, error) // what about stream way ??
	Get(ctx context.Context, id string, ops ...GetOption) (Record, error)
	Create(ctx context.Context, r Record, ops ...CreateOption) (Record, error)
	Update(ctx context.Context, r Record, ops ...UpdateOption) (Record, error)
	Delete(ctx context.Context, id string, ops ...DeleteOption) error
}
