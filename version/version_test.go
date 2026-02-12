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
