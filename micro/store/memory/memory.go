// Package memory is a in-memory store implemented Store
package memory

import (
	"context"
	"sync"

	"go.zoe.im/x/micro/store"
)

type memoryStore struct {
	sync.RWMutex
	values map[string]store.Record
}

func (m *memoryStore) List(ctx context.Context, ops ...store.ListOption) ([]store.Record, error) {
	m.RLock()
	defer m.RUnlock()

	var values []store.Record

	opts := store.NewListOptions(ops...)

	var count int64
	var index int64

	// TODO: implement stream with option watch

	for _, v := range m.values {
		if store.IsDeleted(v) {
			continue
		}

		if store.IsExpired(v) {
			// expired
			m.Delete(ctx, v.GetID())
			continue
		}

		index++
		if index-1 < opts.Skip {
			continue
		}

		count++
		values = append(values, v)

		if count == opts.Limit {
			break
		}

		// TODO: implement continue
	}

	return values, nil
}

func (m *memoryStore) Get(ctx context.Context, id string, ops ...store.GetOption) (store.Record, error) {
	m.RLock()
	defer m.RUnlock()

	_ = store.NewGetOptions(ops...)

	v, ok := m.values[id]
	if !ok {
		return nil, store.ErrNotFound
	}

	return v, nil
}

func (m *memoryStore) Create(ctx context.Context, r store.Record, ops ...store.CreateOption) (store.Record, error) {
	m.RLock()
	defer m.RUnlock()

	opts := store.NewCreateOptions(ops...)

	_, ok := m.values[r.GetID()]
	if ok && !opts.Force {
		return nil, store.ErrExits
	}

	if opts.DryRun {
		return r, nil
	}

	// TODO: set create time and update time
	m.values[r.GetID()] = r
	return r, nil
}

func (m *memoryStore) Update(ctx context.Context, r store.Record, ops ...store.UpdateOption) (store.Record, error) {
	m.RLock()
	defer m.RUnlock()

	opts := store.NewUpdateOptions(ops...)

	if _, ok := m.values[r.GetID()]; !ok {
		return nil, store.ErrNotFound
	}

	if opts.DryRun {
		return r, nil
	}

	// TODO: set create time and update time
	m.values[r.GetID()] = r
	return r, nil
}

func (m *memoryStore) Delete(ctx context.Context, id string, ops ...store.DeleteOption) error {
	m.RLock()
	defer m.RUnlock()

	opts := store.NewDeleteOptions()

	if _, ok := m.values[id]; !ok {
		return store.ErrNotFound
	}

	if opts.DryRun {
		return nil
	}

	if opts.GracePeriod > 0 {
		// TODO: set timer to delete
		delete(m.values, id)
	} else {
		delete(m.values, id)
	}

	return nil
}

// NewStore returns a new store.Store
// TODO: implement more options
func NewStore() store.Store {
	return &memoryStore{
		values: make(map[string]store.Record),
	}
}
