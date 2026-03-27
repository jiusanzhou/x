package cli

import (
	"os"
	"testing"

	"go.zoe.im/x/cli/opts"
)

type envTestConfig struct {
	Host string `opts:"env=TEST_CLI_HOST"`
	Port int    `opts:"env=TEST_CLI_PORT"`
}

func TestParseGlobalFlags_EnvSet(t *testing.T) {
	os.Setenv("TEST_CLI_HOST", "from-env.example.com")
	defer os.Unsetenv("TEST_CLI_HOST")

	c := &envTestConfig{}
	nn := []opts.Opts{opts.New(c)}

	_parseGlobalFlags(nn)

	if c.Host != "from-env.example.com" {
		t.Errorf("Host = %q, want %q", c.Host, "from-env.example.com")
	}
}

func TestParseGlobalFlags_EnvNotSet_PreservesValue(t *testing.T) {
	os.Unsetenv("TEST_CLI_HOST")
	os.Unsetenv("TEST_CLI_PORT")

	c := &envTestConfig{
		Host: "yaml-loaded.example.com",
		Port: 8080,
	}
	nn := []opts.Opts{opts.New(c)}

	_parseGlobalFlags(nn)

	if c.Host != "yaml-loaded.example.com" {
		t.Errorf("Host = %q, want %q (should be preserved)", c.Host, "yaml-loaded.example.com")
	}
	if c.Port != 8080 {
		t.Errorf("Port = %d, want %d (should be preserved)", c.Port, 8080)
	}
}

func TestParseFlags_EnvSet(t *testing.T) {
	os.Setenv("TEST_CLI_PORT", "9090")
	defer os.Unsetenv("TEST_CLI_PORT")

	c := &envTestConfig{Port: 3000}
	nn := []opts.Opts{opts.New(c)}

	_parseFlags(nn)

	if c.Port != 9090 {
		t.Errorf("Port = %d, want %d", c.Port, 9090)
	}
}

func TestParseFlags_EnvNotSet_PreservesValue(t *testing.T) {
	os.Unsetenv("TEST_CLI_HOST")
	os.Unsetenv("TEST_CLI_PORT")

	c := &envTestConfig{
		Host: "config-value",
		Port: 5432,
	}
	nn := []opts.Opts{opts.New(c)}

	_parseFlags(nn)

	if c.Host != "config-value" {
		t.Errorf("Host = %q, want %q (should be preserved)", c.Host, "config-value")
	}
	if c.Port != 5432 {
		t.Errorf("Port = %d, want %d (should be preserved)", c.Port, 5432)
	}
}
