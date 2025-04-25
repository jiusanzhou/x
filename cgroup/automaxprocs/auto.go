//go:build linux
// +build linux

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

package automaxprocs

import (
	"log"
	"os"
	"runtime"
	"strconv"

	"go.zoe.im/x"
	"go.zoe.im/x/cgroup"
)

const _maxProcsKey = "GOMAXPROCS"

var (
	// keep the prev count
	prevCount = 0
)

// Unset unset the max procs
func Unset() {
	runtime.GOMAXPROCS(prevCount)
}

func init() {

	// first load the prev count at first
	prevCount = runtime.GOMAXPROCS(0)

	maxProc := 0

	// if with some value from env
	if max, ok := os.LookupEnv(_maxProcsKey); ok {
		maxProc, _ = strconv.Atoi(max)
		log.Printf("auto set max proc from env: %v", maxProc)
	} else {
		maxProc, _ = x.V(maxProc).Unwrap(quotaToProcs(1)).Int()
		log.Printf("auto set max proc from cgroup: %v", maxProc)
	}

	// and set to runtime
	runtime.GOMAXPROCS(maxProc)
}

// quotaToProcs load max cpu from cgroup
func quotaToProcs(min int) (int, error) {
	// TODO: load from cgroup
	cgs, err := cgroup.NewCGroupsForSelf()
	if err != nil {
		return min, err
	}

	v, ok, err := cgs.CPUQuota()
	if !ok {
		return min, err
	}

	xv, _ := x.V(int(v)).If(int(v) > min).Or(min).Int()
	return xv, nil
}
