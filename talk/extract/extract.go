// Package extract provides endpoint extraction from Go types.
package extract

import (
	"go.zoe.im/x/talk"
)

// Extractor extracts endpoints from a service implementation.
type Extractor interface {
	Extract(service any) ([]*talk.Endpoint, error)
}

// ExtractorFunc is a function adapter for Extractor.
type ExtractorFunc func(service any) ([]*talk.Endpoint, error)

func (f ExtractorFunc) Extract(service any) ([]*talk.Endpoint, error) {
	return f(service)
}

// Options configures endpoint extraction.
type Options struct {
	PathPrefix    string
	MethodMapping map[string]MethodInfo
}

// MethodInfo describes how to expose a method as an endpoint.
type MethodInfo struct {
	Path       string
	HTTPMethod string
	StreamMode talk.StreamMode
}

// Option configures extraction.
type Option func(*Options)

// WithPathPrefix sets a prefix for all extracted endpoint paths.
func WithPathPrefix(prefix string) Option {
	return func(o *Options) {
		o.PathPrefix = prefix
	}
}

// WithMethodMapping provides explicit mapping for specific methods.
func WithMethodMapping(mapping map[string]MethodInfo) Option {
	return func(o *Options) {
		o.MethodMapping = mapping
	}
}
