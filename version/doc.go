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
Package version provides version information management for Go applications.

This package stores and provides access to build-time version information
that is typically injected via ldflags during compilation. It uses
semantic versioning and includes git metadata.

# Version Information

Get version information at runtime:

	info := version.Get()
	fmt.Println(info.GitVersion)  // e.g., "v1.2.3"
	fmt.Println(info.GitCommit)   // e.g., "abc123..."
	fmt.Println(info.BuildDate)   // e.g., "2024-01-15T10:30:00Z"
	fmt.Println(info.Platform)    // e.g., "linux/amd64"

# Build with ldflags

Use the version/gen tool to generate ldflags:

	go run go.zoe.im/x/version/gen -v  # Print version only
	go run go.zoe.im/x/version/gen     # Print ldflags

Build with version info:

	LDFLAGS=$(go run go.zoe.im/x/version/gen)
	go build -ldflags "$LDFLAGS" ./cmd/myapp

# Semantic Versioning

The package embeds [github.com/Masterminds/semver] for semantic version
parsing and comparison:

	info := version.Get()
	if info.Version.Major() >= 2 {
	    // Use v2 features
	}

# Version Variables

The following variables are set via ldflags:

  - GitVersion: Semantic version string (e.g., "v1.2.3")
  - GitCommit: Full git commit hash
  - GitTreeState: "clean" or "dirty"
  - BuildDate: RFC3339 formatted build timestamp
*/
package version
