package x

import (
	"testing"
)

func TestGlob(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		want    bool
	}{
		{"*.txt", "readme.txt", true},
		{"*.txt", "readme.md", false},
		{"*.go", "main.go", true},
		{"test_*", "test_file", true},
		{"test_*", "file_test", false},
		{"**/file.go", "src/pkg/file.go", true},
		{"*.{txt,md}", "readme.txt", true},
		{"*.{txt,md}", "readme.md", true},
		{"*.{txt,md}", "readme.go", false},
	}
	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.input, func(t *testing.T) {
			g := Glob(tt.pattern)
			if got := g.Match(tt.input); got != tt.want {
				t.Errorf("Glob(%q).Match(%q) = %v, want %v", tt.pattern, tt.input, got, tt.want)
			}
		})
	}
}

func TestGlob_Caching(t *testing.T) {
	pattern := "cache_test_*.txt"

	g1 := Glob(pattern)
	g2 := Glob(pattern)

	if g1 != g2 {
		t.Error("Glob() should return cached glob for same pattern")
	}
}

func TestGlob_DifferentPatterns(t *testing.T) {
	g1 := Glob("pattern1_*")
	g2 := Glob("pattern2_*")

	if g1 == g2 {
		t.Error("Glob() should return different globs for different patterns")
	}
}
