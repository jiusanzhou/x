package x

import (
	"github.com/gobwas/glob"
)

// globCache stores compiled glob patterns for reuse.
var globCache SyncMap[string, glob.Glob]

// Glob compiles and caches the glob pattern, returning a reusable matcher.
func Glob(pattern string) glob.Glob {
	g, ok := globCache.Load(pattern)
	if ok {
		return g
	}
	g = glob.MustCompile(pattern)
	globCache.Store(pattern, g)
	return g
}
