package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	content := `
# Comment line
APP_NAME=testapp
APP_PORT=8080
APP_DEBUG=true
QUOTED_VALUE="hello world"
SINGLE_QUOTED='single quotes'
EMPTY_VALUE=
`

	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test .env file: %v", err)
	}

	os.Unsetenv("APP_NAME")
	os.Unsetenv("APP_PORT")
	os.Unsetenv("APP_DEBUG")
	os.Unsetenv("QUOTED_VALUE")
	os.Unsetenv("SINGLE_QUOTED")
	os.Unsetenv("EMPTY_VALUE")

	if err := loadEnvFile(envFile, false); err != nil {
		t.Fatalf("loadEnvFile failed: %v", err)
	}

	tests := []struct {
		key      string
		expected string
	}{
		{"APP_NAME", "testapp"},
		{"APP_PORT", "8080"},
		{"APP_DEBUG", "true"},
		{"QUOTED_VALUE", "hello world"},
		{"SINGLE_QUOTED", "single quotes"},
		{"EMPTY_VALUE", ""},
	}

	for _, tt := range tests {
		got := os.Getenv(tt.key)
		if got != tt.expected {
			t.Errorf("Getenv(%q) = %q, want %q", tt.key, got, tt.expected)
		}
	}
}

func TestLoadEnvFileNoOverride(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	content := `EXISTING_VAR=fromfile`
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test .env file: %v", err)
	}

	os.Setenv("EXISTING_VAR", "fromenv")
	defer os.Unsetenv("EXISTING_VAR")

	if err := loadEnvFile(envFile, false); err != nil {
		t.Fatalf("loadEnvFile failed: %v", err)
	}

	got := os.Getenv("EXISTING_VAR")
	if got != "fromenv" {
		t.Errorf("expected env var to not be overridden, got %q", got)
	}
}

func TestLoadEnvFileWithOverride(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	content := `OVERRIDE_VAR=fromfile`
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test .env file: %v", err)
	}

	os.Setenv("OVERRIDE_VAR", "fromenv")
	defer os.Unsetenv("OVERRIDE_VAR")

	if err := loadEnvFile(envFile, true); err != nil {
		t.Fatalf("loadEnvFile failed: %v", err)
	}

	got := os.Getenv("OVERRIDE_VAR")
	if got != "fromfile" {
		t.Errorf("expected env var to be overridden to 'fromfile', got %q", got)
	}
}

func TestLoadEnvFiles(t *testing.T) {
	tmpDir := t.TempDir()

	envFile := filepath.Join(tmpDir, ".env")
	content := `MULTI_FILE_VAR=value1`
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create .env file: %v", err)
	}

	envLocalFile := filepath.Join(tmpDir, ".env.local")
	localContent := `MULTI_FILE_LOCAL=value2`
	if err := os.WriteFile(envLocalFile, []byte(localContent), 0644); err != nil {
		t.Fatalf("failed to create .env.local file: %v", err)
	}

	os.Unsetenv("MULTI_FILE_VAR")
	os.Unsetenv("MULTI_FILE_LOCAL")

	opts := newEnvOptions(
		WithEnvPaths(tmpDir),
		WithEnvFiles(".env", ".env.local"),
	)

	if err := loadEnvFiles(opts); err != nil {
		t.Fatalf("loadEnvFiles failed: %v", err)
	}

	if got := os.Getenv("MULTI_FILE_VAR"); got != "value1" {
		t.Errorf("MULTI_FILE_VAR = %q, want %q", got, "value1")
	}

	if got := os.Getenv("MULTI_FILE_LOCAL"); got != "value2" {
		t.Errorf("MULTI_FILE_LOCAL = %q, want %q", got, "value2")
	}
}

func TestLoadEnvFilesDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	content := `DISABLED_VAR=shouldnotload`
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test .env file: %v", err)
	}

	os.Unsetenv("DISABLED_VAR")

	opts := newEnvOptions(
		WithEnvEnabled(false),
		WithEnvPaths(tmpDir),
	)

	if err := loadEnvFiles(opts); err != nil {
		t.Fatalf("loadEnvFiles failed: %v", err)
	}

	if got := os.Getenv("LED_VAR"); got != "" {
		t.Errorf("DISABLED_VAR should be empty when disabled, got %q", got)
	}
}
