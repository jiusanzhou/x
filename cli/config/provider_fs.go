package config

import (
	"errors"
	"io/ioutil"
	"path/filepath"
)

type fsProvider struct {
	root string // base path to load file
}

// Read implement read content from provider
func (fs *fsProvider) Read(name string, typs ...string) ([]byte, error) {
	// check type we need to read all type?
	// no need
	if len(typs) > 0 {
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
