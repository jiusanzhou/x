/*
 * Copyright (c) 2019 wellwell.work, LLC by Zoe
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
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
	"testing"
)

func TestStr2Bytes(t *testing.T) {
	var cases = []string{
		"a",
		"ab",
		"abc",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\nsaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}

	for _, c := range cases {
		var (
			got = Str2Bytes(c)
			won = []byte(c)
		)
		if !bytes.Equal(got, won) {
			t.Errorf("covert error. we get %#v, but we want %#v", got, won)
		}
	}
}

func TestBytes2Str(t *testing.T) {
	var cases = []string{
		"a",
		"ab",
		"abc",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\nsaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}

	for _, c := range cases {
		var (
			got = Bytes2Str([]byte(c))
		)
		if got != c {
			t.Errorf("covert error. we get %#v, but we want %#v", got, c)
		}
	}
}

func BenchmarkStr2Bytes(b *testing.B) {
	var str = "aaaaaaaaaaaaaaaaa"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Str2Bytes(str)
	}
}

func BenchmarkStr2BytesOrigin(b *testing.B) {
	var str = "aaaaaaaaaaaaaaaaa"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = []byte(str)
	}
}

func BenchmarkBytes2Str(b *testing.B) {
	var data = []byte("aaaaaaaaaaaaaaaaaad")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Bytes2Str(data)
	}
}

func BenchmarkBytes2StrOrigin(b *testing.B) {
	var data = []byte("aaaaaaaaaaaaaaaaaad")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = string(data)
	}
}
