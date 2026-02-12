package cli

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var defaultEnvFiles = []string{".env", ".env.local"}

type envOptions struct {
	enabled  bool
	paths    []string
	files    []string
	override bool
}

type EnvOption func(*envOptions)

func WithEnvEnabled(enabled bool) EnvOption {
	return func(o *envOptions) {
		o.enabled = enabled
	}
}

func WithEnvPaths(paths ...string) EnvOption {
	return func(o *envOptions) {
		o.paths = append(o.paths, paths...)
	}
}

func WithEnvFiles(files ...string) EnvOption {
	return func(o *envOptions) {
		o.files = append(o.files, files...)
	}
}

func WithEnvOverride(override bool) EnvOption {
	return func(o *envOptions) {
		o.override = override
	}
}

func newEnvOptions(opts ...EnvOption) *envOptions {
	eo := &envOptions{
		enabled:  true,
		paths:    []string{"."},
		files:    defaultEnvFiles,
		override: false,
	}
	for _, o := range opts {
		o(eo)
	}
	return eo
}

func loadEnvFile(path string, override bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}

		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])

		if key == "" {
			continue
		}

		value = strings.Trim(value, `"'`)

		if !override {
			if _, exists := os.LookupEnv(key); exists {
				continue
			}
		}

		os.Setenv(key, value)
	}

	return scanner.Err()
}

func loadEnvFiles(opts *envOptions) error {
	if !opts.enabled {
		return nil
	}

	for _, dir := range opts.paths {
		for _, file := range opts.files {
			path := filepath.Join(dir, file)
			if _, err := os.Stat(path); err == nil {
				loadEnvFile(path, opts.override)
			}
		}
	}

	return nil
}

func LoadEnv(opts ...EnvOption) Option {
	return func(c *Command) {
		eo := newEnvOptions(opts...)

		oldPreRun := c.Command.PersistentPreRun
		c.Command.PersistentPreRun = func(cmd *cobra.Command, args []string) {
			loadEnvFiles(eo)
			if oldPreRun != nil {
				oldPreRun(cmd, args)
			}
		}
	}
}

func AutoLoadEnv(paths ...string) Option {
	opts := []EnvOption{WithEnvEnabled(true)}
	if len(paths) > 0 {
		opts = append(opts, WithEnvPaths(paths...))
	}
	return LoadEnv(opts...)
}
