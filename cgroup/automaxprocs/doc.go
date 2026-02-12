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
Package automaxprocs automatically sets GOMAXPROCS based on cgroup CPU quota.

This package is designed for containerized Go applications where the
container has CPU limits. Import this package for its side effects
to automatically configure GOMAXPROCS at startup.

# Usage

Simply import the package:

	import _ "go.zoe.im/x/cgroup/automaxprocs"

	func main() {
	    // GOMAXPROCS is automatically set based on cgroup limits
	}

# Behavior

On init, the package:

 1. Checks the GOMAXPROCS environment variable
 2. If set, uses that value
 3. Otherwise, reads the CPU quota from cgroup
 4. Sets GOMAXPROCS to the quota (minimum 1)

# Environment Variable

To override automatic detection, set GOMAXPROCS:

	GOMAXPROCS=4 ./myapp

# Unset

To restore the previous GOMAXPROCS value:

	automaxprocs.Unset()

# Platform Support

This package only works on Linux. On other platforms, the
import has no effect.
*/
package automaxprocs
