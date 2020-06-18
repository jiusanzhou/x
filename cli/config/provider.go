package config

import (
	"crypto/md5"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	providerFactory = NewProviderFactory()

	// register provider creators
	providerCreators = map[string]ProviderCreator{}
)

// Provider defined how to take payload of configuration
// file, http, database or something else
//
// source is the interface for sources
// or we need to use a container for []byte
// like struct{ Data []byte, Checksum string, Timestamp time.Time }
type Provider interface {
	// load configuraton data from name
	Read(name string, typs ...string) ([]byte, error)

	// dumps configuration
	// will we need to update the conguration
	Write(name string, data []byte, typs ...string) error

	// TODO: patch special value with key selector
	// Patch(name string, key string, value interface{}) error

	// watch the config whilte change
	Watch(name string, typs ...string) (Watcher, error)

	// String returns name of provider
	String() string
}

// ProviderCreator function to create provider
type ProviderCreator func(c interface{}) (Provider, error)

// ProviderFactory provider factory
type ProviderFactory struct {
	providers map[string]Provider
}

// Register register a provider
func (f *ProviderFactory) Register(name string, provider Provider) error {
	if _, ok := f.providers[name]; ok {
		// TODO: return aleary register error
	}
	f.providers[name] = provider
	return nil
}

// RegisterProvider register provider
func RegisterProvider(name string, provider Provider) error {
	return providerFactory.Register(name, provider)
}

// NewProviderFactory create a factory
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{
		providers: make(map[string]Provider),
	}
}

// NewProvider create a provider with name and configuration
func NewProvider(name string, config interface{}) (Provider, error) {
	creator, ok := providerCreators[name]
	if !ok {
		return nil, errors.New("provider creator not exits")
	}
	return creator(config)
}

// NewProviderFromURI create a provider with uri string
// at most time, format schema://config
func NewProviderFromURI(url string) (Provider, error) {
	sep := "://"

	s := strings.Index(url, sep)
	if s < 0 {
		return NewProvider("fs", url)
	}

	// split with `://`
	return NewProvider(url[:s], url[s+len(sep):])
}

// RegisterProviderCreator register a provider creator
func RegisterProviderCreator(name string, creator ProviderCreator) error {
	if _, ok := providerCreators[name]; ok {
		// TODO: return a exits error
	}
	providerCreators[name] = creator
	return nil
}

// ChangeSet represents a set of changes from a source
type ChangeSet struct {
	Data      []byte
	Checksum  string
	Format    string
	Source    string
	Timestamp time.Time
}

// Sum returns the md5 checksum of the ChangeSet data
func (c *ChangeSet) Sum() string {
	h := md5.New()
	h.Write(c.Data)
	return fmt.Sprintf("%x", h.Sum(nil))
}
