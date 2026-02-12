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
Package cgroup provides utilities for reading Linux cgroup information.

This package allows reading CPU quotas and other cgroup parameters
from the Linux /proc filesystem. It supports cgroup v1.

# Basic Usage

Read cgroup information for the current process:

	cgs, err := cgroup.NewCGroupsForSelf()
	if err != nil {
	    log.Fatal(err)
	}

	quota, defined, err := cgs.CPUQuota()
	if defined {
	    fmt.Printf("CPU quota: %.2f cores\n", quota)
	}

# Reading for Other Processes

Read cgroup info for a specific process by PID:

	cgs, err := cgroup.NewCGroupsForPid(1234)

# CPU Quota

The CPUQuota method returns the CPU quota as a float representing
the number of CPU cores allocated:

	quota, defined, err := cgs.CPUQuota()
	// quota = 2.5 means 2.5 CPU cores
	// defined = false means no quota is set (unlimited)

# Supported Subsystems

The package recognizes these cgroup subsystems:

  - cpu: CPU bandwidth control
  - cpuacct: CPU accounting
  - cpuset: CPU and memory node assignment
  - memory: Memory limits

# Platform Support

This package only works on Linux. On other platforms, the
cgroup-related functions are not available.
*/
package cgroup
