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

package x

import (
	"bytes"
	"io"
)

// lineWriter return a writer with a line callback
type lineWriter struct {
	callback func(line []byte) error
	tmp      []byte
	// ignore empty line default true
	ignoreEmpty bool
	// TODO: end with timeout as a line
}

func (t *lineWriter) Write(p []byte) (n int, err error) {
	// TODO: do we real need this?
	if len(p) == 0 {
		return 0, nil
	}

	// TODO:
	bts := bytes.Split(p, []byte("\n"))

	// maybe the previous line is ending
	if len(bts[0]) == 0 && (!t.ignoreEmpty || len(t.tmp) > 0) {
		t.callback(t.tmp)
	}

	// if last item is not empty, we are not ending
	if len(bts[len(bts)-1]) > 0 {
		t.tmp = bts[len(bts)-1]
	}

	if len(bts) > 1 {
		for _, bs := range bts[0 : len(bts)-1] {
			t.callback(bs)
		}
	}

	return len(p), nil
}

// LineWriter return a writer with a line callback
func LineWriter(fn func(line []byte) error, ignoreEmpty bool) io.Writer {
	return &lineWriter{
		callback:    fn,
		ignoreEmpty: ignoreEmpty,
	}
}
