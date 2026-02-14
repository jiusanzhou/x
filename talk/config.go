package talk

import (
	"fmt"
	"strings"
	"sync"

	"go.zoe.im/x"
)

type ServerTransportCreator func(cfg x.TypedLazyConfig) (Transport, error)

type ClientTransportCreator func(cfg x.TypedLazyConfig) (Transport, error)

var (
	serverCreators   = make(map[string]ServerTransportCreator)
	serverCreatorsMu sync.RWMutex
	clientCreators   = make(map[string]ClientTransportCreator)
	clientCreatorsMu sync.RWMutex
)

func RegisterServerTransport(name string, creator ServerTransportCreator, aliases ...string) {
	serverCreatorsMu.Lock()
	defer serverCreatorsMu.Unlock()
	serverCreators[name] = creator
	for _, alias := range aliases {
		serverCreators[alias] = creator
	}
}

func RegisterClientTransport(name string, creator ClientTransportCreator, aliases ...string) {
	clientCreatorsMu.Lock()
	defer clientCreatorsMu.Unlock()
	clientCreators[name] = creator
	for _, alias := range aliases {
		clientCreators[alias] = creator
	}
}

func NewServerFromConfig(cfg x.TypedLazyConfig, opts ...ServerOption) (*Server, error) {
	serverCreatorsMu.RLock()
	creator, ok := serverCreators[cfg.Type]
	serverCreatorsMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown server transport type: %s", cfg.Type)
	}

	transport, err := creator(cfg)
	if err != nil {
		return nil, err
	}

	return NewServer(transport, opts...), nil
}

func NewClientFromConfig(cfg x.TypedLazyConfig, opts ...ClientOption) (*Client, error) {
	clientCreatorsMu.RLock()
	creator, ok := clientCreators[cfg.Type]
	clientCreatorsMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown client transport type: %s", cfg.Type)
	}

	transport, err := creator(cfg)
	if err != nil {
		return nil, err
	}

	return NewClient(transport, opts...), nil
}

func parseTransportType(t string) (family, impl string) {
	parts := strings.SplitN(t, "/", 2)
	if len(parts) == 1 {
		return parts[0], "default"
	}
	return parts[0], parts[1]
}

var _ = parseTransportType
