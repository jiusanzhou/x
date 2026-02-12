// Copyright (c) 2020 wellwell.work, LLC by Zoe
//
// Licensed under the Apache License 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
Package jsonmerge provides deep merge functionality for Go maps and structs.

This package allows merging multiple source values into a destination,
with configurable behavior for handling conflicts, empty values, and slices.

# Basic Usage

Merge maps or structs using the default merger:

	dst := map[string]any{"a": 1, "b": 2}
	src := map[string]any{"b": 3, "c": 4}

	err := jsonmerge.Merge(&dst, src)
	// dst is now {"a": 1, "b": 3, "c": 4}

# Multiple Sources

Merge multiple sources in order:

	err := jsonmerge.Merge(&dst, src1, src2, src3)

# Custom Merger

Create a merger with custom options:

	m := jsonmerge.New(
	    jsonmerge.WithOverwrite(true),
	    jsonmerge.WithAppendSlice(true),
	)
	err := m.Merge(&dst, src)

# Options

  - WithOverwrite: Overwrite existing values in destination
  - WithOverwriteWithEmptySrc: Allow empty source values to overwrite
  - WithOverwriteSliceWithEmptySrc: Allow empty slices to overwrite
  - WithTypeCheck: Enable type checking between src and dst
  - WithAppendSlice: Append slices instead of replacing
  - WithMaxMergeDepth: Set maximum recursion depth for merging
*/
package jsonmerge
