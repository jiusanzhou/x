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
	"reflect"
	"strings"
	"testing"
)

func TestLineWriter(t *testing.T) {

	var got []string
	wrt := LineWriter(func(line []byte) error {
		got = append(got, string(line))
		return nil
	}, true)

	cases := []struct{
		input []string
		want []string
	}{
		{[]string{"a\n"}, []string{"a"}},
		{[]string{"aaaa\naaaa\n"}, []string{"aaaa", "aaaa"}},
		{[]string{"aaaa\naaaa"}, []string{"aaaa"}},

		{[]string{"a", "\n"}, []string{"a"}},
		{[]string{"aaaa\naa", "aa\n"}, []string{"aaaa", "aaaa"}},
		{[]string{"aa", "aa\naaaa"}, []string{"aaaa"}},
		{[]string{"aaaa", "\naaaa"}, []string{"aaaa"}},

		{[]string{"aaaa\naa", "aa", "\n"}, []string{"aaaa", "aaaa"}},
	}
	
	for _, c := range cases {
		got = nil
		io.Copy(wrt, bytes.NewReader([]byte(strings.Join(c.input, ""))))
		if !reflect.DeepEqual(c.want, got) {
			t.Errorf("LineWriter = %v, want %v", got, c.want)
		}
	}
}
