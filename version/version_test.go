package version

import (
	"runtime"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	info := Get()
	if info == nil {
		t.Fatal("Get() returned nil")
	}
}

func TestInfo_GoVersion(t *testing.T) {
	info := Get()
	if info.GoVersion == "" {
		t.Error("Info.GoVersion is empty")
	}
	if info.GoVersion != runtime.Version() {
		t.Errorf("Info.GoVersion = %q, want %q", info.GoVersion, runtime.Version())
	}
}

func TestInfo_Compiler(t *testing.T) {
	info := Get()
	if info.Compiler == "" {
		t.Error("Info.Compiler is empty")
	}
	if info.Compiler != runtime.Compiler {
		t.Errorf("Info.Compiler = %q, want %q", info.Compiler, runtime.Compiler)
	}
}

func TestInfo_Platform(t *testing.T) {
	info := Get()
	if info.Platform == "" {
		t.Error("Info.Platform is empty")
	}
	if !strings.Contains(info.Platform, "/") {
		t.Errorf("Info.Platform = %q, expected format 'os/arch'", info.Platform)
	}
}

func TestInfo_String(t *testing.T) {
	info := Get()
	str := info.String()
	if str != info.GitVersion {
		t.Errorf("Info.String() = %q, want GitVersion %q", str, info.GitVersion)
	}
}

func TestInfo_Version(t *testing.T) {
	info := Get()
	if info.Version == nil {
		t.Error("Info.Version is nil")
	}
}

func TestSemver_NewSemver(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"1.0.0", false},
		{"v1.2.3", false},
		{"1.2.3-alpha.1", false},
		{"1.2.3+build.123", false},
		{"1.2.3-beta.1+build.456", false},
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := NewSemver(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSemver(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestSemver_IncrementPatch(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.0.0", "1.0.1"},
		{"1.2.3", "1.2.4"},
		{"0.0.0", "0.0.1"},
		{"v1.0.0", "1.0.1"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			s, err := NewSemver(tt.input)
			if err != nil {
				t.Fatalf("NewSemver(%q) error: %v", tt.input, err)
			}
			result := s.IncrementPatch()
			if result.Version.String() != tt.expected {
				t.Errorf("IncrementPatch() = %q, want %q", result.Version.String(), tt.expected)
			}
		})
	}
}

func TestSemver_IncrementMinor(t *testing.T) {
	s := MustSemver("1.2.3")
	result := s.IncrementMinor()
	if result.Version.String() != "1.3.0" {
		t.Errorf("IncrementMinor() = %q, want %q", result.Version.String(), "1.3.0")
	}
}

func TestSemver_IncrementMajor(t *testing.T) {
	s := MustSemver("1.2.3")
	result := s.IncrementMajor()
	if result.Version.String() != "2.0.0" {
		t.Errorf("IncrementMajor() = %q, want %q", result.Version.String(), "2.0.0")
	}
}

func TestSemver_IncrementPrerelease(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.0.0-alpha.0", "alpha.1"},
		{"1.0.0-beta.5", "beta.6"},
		{"1.0.0-rc.1.2", "rc.1.3"},
		{"1.0.0-dev", "dev.1"},
		{"1.0.0", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			s, err := NewSemver(tt.input)
			if err != nil {
				t.Fatalf("NewSemver(%q) error: %v", tt.input, err)
			}
			result, err := s.IncrementPrerelease()
			if err != nil {
				t.Fatalf("IncrementPrerelease() error: %v", err)
			}
			if result.Version.Prerelease() != tt.expected {
				t.Errorf("IncrementPrerelease() prerelease = %q, want %q", result.Version.Prerelease(), tt.expected)
			}
		})
	}
}

func TestSemver_NextVersion(t *testing.T) {
	s := MustSemver("1.2.3")

	tests := []struct {
		incType  string
		expected string
	}{
		{"patch", "1.2.4"},
		{"minor", "1.3.0"},
		{"major", "2.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.incType, func(t *testing.T) {
			result, err := s.NextVersion(tt.incType)
			if err != nil {
				t.Fatalf("NextVersion(%q) error: %v", tt.incType, err)
			}
			if result.Version.String() != tt.expected {
				t.Errorf("NextVersion(%q) = %q, want %q", tt.incType, result.Version.String(), tt.expected)
			}
		})
	}
}

func TestSemver_IsPrerelease(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1.0.0", false},
		{"1.0.0-alpha", true},
		{"1.0.0-beta.1", true},
		{"1.0.0+build", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			s := MustSemver(tt.input)
			if s.IsPrerelease() != tt.expected {
				t.Errorf("IsPrerelease() = %v, want %v", s.IsPrerelease(), tt.expected)
			}
		})
	}
}

func TestSemver_IsStable(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1.0.0", true},
		{"0.1.0", false},
		{"1.0.0-alpha", false},
		{"2.0.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			s := MustSemver(tt.input)
			if s.IsStable() != tt.expected {
				t.Errorf(" %v, want %v", s.IsStable(), tt.expected)
			}
		})
	}
}

func TestSemver_TagString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.0.0", "v1.0.0"},
		{"v1.0.0", "v1.0.0"},
		{"1.2.3-alpha", "v1.2.3-alpha"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			s := MustSemver(tt.input)
			if s.TagString() != tt.expected {
				t.Errorf("TagString() = %q, want %q", s.TagString(), tt.expected)
			}
		})
	}
}
