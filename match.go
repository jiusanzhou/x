package x

import (
	"github.com/gobwas/glob"
)

// TODO: should we need to clean the cache?
var globCache SyncMap[string, glob.Glob]

func Glob(pattern string) glob.Glob {
	g, ok := globCache.Load(pattern)
	if ok {
		return g
	}
	g = glob.MustCompile(pattern)
	globCache.Store(pattern, g)
	return g
}
