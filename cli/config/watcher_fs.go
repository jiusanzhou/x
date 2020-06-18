/*
 * Copyright (c) 2020 wellwell.work, LLC by Zoe
 *
 * Licensed under the Apache License 2.0 (the "License");
 * You may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/fsnotify/fsnotify"
)

type fsWatcher struct {
	fullpath string // full path of the file
	name     string // orignal name
	types    []string

	provider Provider

	fw   *fsnotify.Watcher
	exit chan bool
}

func (w *fsWatcher) Next() (*ChangeSet, error) {

	// is it closed?
	select {
	case <-w.exit:
		return nil, ErrWatcherStopped
	default:
	}

	// try get the event
	select {
	case event, _ := <-w.fw.Events:
		// if event.Op == fsnotify.Rename {
		// 	// check existence of file, and add watch again
		// 	_, err := os.Stat(event.Name)
		// 	if err == nil || os.IsExist(err) {
		// 		w.fw.Add(event.Name)
		// 	}
		// }

		// TODO: vim likes editor: Rename -> Chmod -> Remove
		if event.Op != fsnotify.Write {
			return nil, errors.New("unhandle")
		}

		fh, err := os.Open(w.fullpath)
		if err != nil {
			return nil, err
		}
		defer fh.Close()

		// read file content out
		b, err := ioutil.ReadAll(fh)
		if err != nil {
			return nil, err
		}

		info, err := fh.Stat()
		if err != nil {
			return nil, err
		}

		cs := &ChangeSet{
			Data:      b,
			Timestamp: info.ModTime(),
			Source:    w.fullpath,
		}

		cs.Checksum = cs.Sum()

		// add path again for the event bug of fsnotify
		// w.fw.Add(w.fullpath)

		return cs, nil
	case err := <-w.fw.Errors:
		return nil, err
	case <-w.exit:
		return nil, ErrWatcherStopped
	}
}

func (w *fsWatcher) Stop() error {
	return w.fw.Close()
}
