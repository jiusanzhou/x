package talk

import (
	"fmt"
	"sync"

	"go.zoe.im/x"
)

type TransportCreators struct {
	Server func(cfg x.TypedLazyConfig) (Transport, error)
	Client func(cfg x.TypedLazyConfig) (Transport, error)
}

var (
	transportCreators   = make(map[string]*TransportCreators)
	transportCreatorsMu sync.RWMutex
)

func RegisterTransport(name string, creators *TransportCreators, aliases ...string) {
	transportCreatorsMu.Lock()
	defer transportCreatorsMu.Unlock()
	transportCreators[name] = creators
	for _, alias := range aliases {
		transportCreators[alias] = creators
	}
}

func NewServerFromConfig(cfg x.TypedLazyConfig, opts ...ServerOption) (*Server, error) {
	transportCreatorsMu.RLock()
	creators, ok := transportCreators[cfg.Type]
	transportCreatorsMu.RUnlock()

	if !ok || creators.Server == nil {
		return nil, fmt.Errorf("unknown server transport type: %s", cfg.Type)
	}

	transport, err := creators.Server(cfg)
	if err != nil {
		return nil, err
	}

	return NewServer(transport, opts...), nil
}

func NewClientFromConfig(cfg x.TypedLazyConfig, opts ...ClientOption) (*Client, error) {
	transportCreatorsMu.RLock()
	creators, ok := transportCreators[cfg.Type]
	transportCreatorsMu.RUnlock()

	if !ok || creators.Client == nil {
		return nil, fmt.Errorf("unknown client transport type: %s", cfg.Type)
	}

	transport, err := creators.Client(cfg)
	if err != nil {
		return nil, err
	}

	return NewClient(transport, opts...), nil
}
