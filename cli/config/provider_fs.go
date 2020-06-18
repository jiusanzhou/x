package config

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// TODO(zoe): use file to wrap the root path file name and decoder/encoder
type file struct{}

type fsProvider struct {
	root string // base path to load file
}

// Read implement read content from provider
func (fs *fsProvider) Read(name string, typs ...string) ([]byte, error) {
	// check type we need to read all type?
	// no need, without the extension make sure
	if len(typs) > 0 && !strings.HasSuffix(name, "."+typs[0]) {
		// suffix with extension
		name = name + "." + typs[0]
	}

	// read file with name from root directory
	return ioutil.ReadFile(filepath.Join(fs.root, name))
}

// Write implement write content to provider
func (fs *fsProvider) Write(name string, data []byte, typs ...string) error {
	// check type we need to read all type?
	// no need
	if len(typs) > 0 {
		// suffix with extension
		name = name + "." + typs[0]
	}
	return ioutil.WriteFile(filepath.Join(fs.root, name), data, 0644)
}

// Watch return a new wather for file source
func (fs *fsProvider) Watch(name string, typs ...string) (Watcher, error) {
	// storage the original name
	originalName := name
	// check type we need to read all type?
	// no need
	if len(typs) > 0 {
		// suffix with extension
		name = name + "." + typs[0]
	}

	// check if the file exits
	fpath := filepath.Join(fs.root, name)
	if _, err := os.Stat(fpath); err != nil {
		return nil, err
	}

	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	fw.Add(fpath)

	// create a new watcher
	return &fsWatcher{
		provider: fs,
		name:     originalName,
		types:    typs,
		fullpath: fpath,

		fw:   fw,
		exit: make(chan bool),
	}, nil
}

func (fs *fsProvider) String() string {
	return "fs" + "://" + fs.root
}

// NewFSProvider return a fs provider
func NewFSProvider(path string) (Provider, error) {
	fs := &fsProvider{
		root: path,
	}
	return fs, nil
}

func init() {
	// register fs creator
	RegisterProviderCreator("fs", func(c interface{}) (Provider, error) {
		f, ok := c.(string)
		if !ok {
			return nil, errors.New("fs provider config must be a path")
		}
		return NewFSProvider(f)
	})
}
