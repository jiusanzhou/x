package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type dirProvider struct {
	root     string
	patterns []string
}

func (d *dirProvider) Read(name string, typs ...string) ([]byte, error) {
	if len(typs) > 0 && !strings.HasSuffix(name, "."+typs[0]) {
		name = name + "." + typs[0]
	}
	return os.ReadFile(filepath.Join(d.root, name))
}

func (d *dirProvider) Write(name string, data []byte, typs ...string) error {
	if len(typs) > 0 && !strings.HasSuffix(name, "."+typs[0]) {
		name = name + "." + typs[0]
	}
	return os.WriteFile(filepath.Join(d.root, name), data, 0644)
}

func (d *dirProvider) Watch(name string, typs ...string) (Watcher, error) {
	if len(typs) > 0 && !strings.HasSuffix(name, "."+typs[0]) {
		name = name + "." + typs[0]
	}

	fpath := filepath.Join(d.root, name)
	if _, err := os.Stat(fpath); err != nil {
		return nil, err
	}

	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	fw.Add(fpath)

	return &fsWatcher{
		provider: &fsProvider{root: d.root},
		name:     strings.TrimSuffix(name, "."+typs[0]),
		types:    typs,
		fullpath: fpath,
		fw:       fw,
		exit:     make(chan bool),
	}, nil
}

func (d *dirProvider) String() string {
	return "dir://" + d.root
}

func (d *dirProvider) ListConfigs(typs ...string) ([]string, error) {
	var configs []string

	entries, err := os.ReadDir(d.root)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := strings.TrimPrefix(filepath.Ext(name), ".")

		if len(typs) > 0 {
			matched := false
			for _, t := range typs {
				if ext == t {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		if len(d.patterns) > 0 {
			matched := false
			for _, pattern := range d.patterns {
				if m, _ := filepath.Match(pattern, name); m {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		baseName := strings.TrimSuffix(name, filepath.Ext(name))
		configs = append(configs, baseName)
	}

	return configs, nil
}

func (d *dirProvider) WatchDirectory(typs ...string) (DirectoryWatcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if err := fw.Add(d.root); err != nil {
		fw.Close()
		return nil, err
	}

	return &dirWatcher{
		provider: d,
		types:    typs,
		fw:       fw,
		exit:     make(chan bool),
	}, nil
}

func NewDirProvider(path string, patterns ...string) *dirProvider {
	return &dirProvider{
		root:     path,
		patterns: patterns,
	}
}

type DirectoryWatcher interface {
	Next() (*DirChangeSet, error)
	Stop()
}

type DirChangeSet struct {
	Name    string
	Type    string
	Data    []byte
	Deleted bool
}

type dirWatcher struct {
	provider *dirProvider
	types    []string
	fw       *fsnotify.Watcher
	exit     chan bool
}

func (w *dirWatcher) Next() (*DirChangeSet, error) {
	for {
		select {
		case <-w.exit:
			return nil, nil
		case event := <-w.fw.Events:
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) == 0 {
				continue
			}

			name := filepath.Base(event.Name)
			ext := strings.TrimPrefix(filepath.Ext(name), ".")

			if len(w.types) > 0 {
				matched := false
				for _, t := range w.types {
					if ext == t {
						matched = true
						break
					}
				}
				if !matched {
					continue
				}
			}

			if len(w.provider.patterns) > 0 {
				matched := false
				for _, pattern := range w.provider.patterns {
					if m, _ := filepath.Match(pattern, name); m {
						matched = true
						break
					}
				}
				if !matched {
					continue
				}
			}

			baseName := strings.TrimSuffix(name, filepath.Ext(name))

			cs := &DirChangeSet{
				Name:    baseName,
				Type:    ext,
				Deleted: event.Op&fsnotify.Remove != 0,
			}

			if !cs.Deleted {
				data, err := os.ReadFile(event.Name)
				if err != nil {
					continue
				}
				cs.Data = data
			}

			return cs, nil
		case err := <-w.fw.Errors:
			return nil, err
		}
	}
}

func (w *dirWatcher) Stop() {
	select {
	case w.exit <- true:
	default:
	}
	w.fw.Close()
}

func WithDirectory(path string, patterns ...string) Option {
	return func(c *Options) {
		provider := NewDirProvider(path, patterns...)
		if c.isProvSet {
			c.providers = append(c.providers, provider)
		} else {
			c.providers = []Provider{provider}
			c.isProvSet = true
		}
	}
}

func WithDirectories(paths ...string) Option {
	return func(c *Options) {
		for _, path := range paths {
			provider := NewDirProvider(path)
			if c.isProvSet {
				c.providers = append(c.providers, provider)
			} else {
				c.providers = []Provider{provider}
				c.isProvSet = true
			}
		}
	}
}

func init() {
	RegisterProviderCreator("dir", func(c interface{}) (Provider, error) {
		path, ok := c.(string)
		if !ok {
			return nil, ErrUnsupportedEncoder
		}
		return NewDirProvider(path), nil
	})
}
