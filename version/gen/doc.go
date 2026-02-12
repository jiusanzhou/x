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
Command gen generates ldflags for injecting version information into Go binaries.

This tool extracts version information from git and outputs ldflags that can
be passed to `go build -ldflags`.

# Usage

Generate ldflags:

	go run go.zoe.im/x/version/gen

Print only the version tag:

	go run go.zoe.im/x/version/gen -v

Specify a custom environment key:

	go run go.zoe.im/x/version/gen -env-key MY_LDFLAGS

# Build Integration

Use in a Makefile:

	VERSION_FLAGS := $(shell go run go.zoe.im/x/version/gen)

	build:
	    go build -ldflags "$(VERSION_FLAGS)" ./cmd/myapp

Or directly:

	go build -ldflags "$(go run go.zoe.im/x/version/gen)" ./cmd/myapp

# Output Format

The tool outputs ldflags in this format:

	-X go.zoe.im/x/version.GitVersion=v1.2.3
	-X go.zoe.im/x/version.GitCommit=abc123...
	-X go.zoe.im/x/version.GitTreeState=clean
	-X go.zoe.im/x/version.BuildDate=2024-01-15T10:30:00Z

# Version Detection

The version is derived from git tags:

  - Uses `git describe --tags --match "v*"`
  - Converts to semver format (e.g., v1.1.0-alpha.0.6+84c76d1)
  - Appends "-dirty" if working tree has uncommitted changes
  - Falls back to "v0.0.0" if no tags exist
*/
package main
