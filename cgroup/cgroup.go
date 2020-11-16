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

package cgroup

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

const (
	// _cgroupFSType is the Linux CGroup file system type used in
	// `/proc/$PID/mountinfo`.
	_cgroupFSType = "cgroup"
	// _cgroupSubsysCPU is the CPU CGroup subsystem.
	_cgroupSubsysCPU = "cpu"
	// _cgroupSubsysCPUAcct is the CPU accounting CGroup subsystem.
	_cgroupSubsysCPUAcct = "cpuacct"
	// _cgroupSubsysCPUSet is the CPUSet CGroup subsystem.
	_cgroupSubsysCPUSet = "cpuset"
	// _cgroupSubsysMemory is the Memory CGroup subsystem.
	_cgroupSubsysMemory = "memory"

	// _cgroupCPUCFSQuotaUsParam is the file name for the CGroup CFS quota
	// parameter.
	_cgroupCPUCFSQuotaUsParam = "cpu.cfs_quota_us"
	// _cgroupCPUCFSPeriodUsParam is the file name for the CGroup CFS period
	// parameter.
	_cgroupCPUCFSPeriodUsParam = "cpu.cfs_period_us"
)

const (
	_procSelfPathCgroup    = "/proc/self/cgroup"
	_procSelfPathMountInfo = "/proc/self/mountinfo"

	_procTplPathCgroup    = "/proc/%d/cgroup"
	_procTplPathMountInfo = "/proc/%d/mountinfo"
)

// Cgroup represents the data structure for a Linux control group.
type Cgroup struct {
	path string // the base path
}

// NewCgroup returns a new *Cgroup from a given path.
func NewCgroup(path string) *Cgroup {
	return &Cgroup{path: path}
}

// ParamPath returns the path of the given cgroup param under itself.
func (cg *Cgroup) ParamPath(param string) string {
	return filepath.Join(cg.path, param)
}

// readFirstLine reads the first line from a cgroup param file.
func (cg *Cgroup) readFirstLine(param string) (string, error) {
	paramFile, err := os.Open(cg.ParamPath(param))
	if err != nil {
		return "", err
	}
	defer paramFile.Close()

	scanner := bufio.NewScanner(paramFile)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", io.ErrUnexpectedEOF
}

// readInt parses the first line from a cgroup param file as int.
func (cg *Cgroup) readInt(param string) (int, error) {
	text, err := cg.readFirstLine(param)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(text)
}

// Cgroups is a map that associates each CGroup with its subsystem name.
type Cgroups map[string]*Cgroup

// NewCgroupsFromPath returns a new *CGroups from given `mountinfo` and `cgroup` files
// under for some process under `/proc` file system (see also proc(5) for more
// information).
func NewCgroupsFromPath(procPathMountInfo, procPathCGroup string) (Cgroups, error) {
	cgroupSubsystems, err := NewCgroupSubsys(procPathCGroup)
	if err != nil {
		return nil, err
	}

	cgroups := make(Cgroups)

	// the create mount point func
	newMountPoint := func(mp *MountPoint) error {
		if mp.FSType != _cgroupFSType {
			return nil
		}

		for _, opt := range mp.SuperOptions {
			subsys, exists := cgroupSubsystems[opt]
			if !exists {
				continue
			}

			cgroupPath, err := mp.Translate(subsys.Name)
			if err != nil {
				return err
			}
			cgroups[opt] = NewCgroup(cgroupPath)
		}

		return nil
	}

	if err := parseMountInfo(procPathMountInfo, newMountPoint); err != nil {
		return nil, err
	}
	return cgroups, nil
}

// CPUQuota returns the CPU quota applied with the CPU cgroup controller.
// It is a result of `cpu.cfs_quota_us / cpu.cfs_period_us`. If the value of
// `cpu.cfs_quota_us` was not set (-1), the method returns `(-1, nil)`.
func (cg Cgroups) CPUQuota() (float64, bool, error) {
	cpuCGroup, exists := cg[_cgroupSubsysCPU]
	if !exists {
		return -1, false, nil
	}

	cfsQuotaUs, err := cpuCGroup.readInt(_cgroupCPUCFSQuotaUsParam)
	if defined := cfsQuotaUs > 0; err != nil || !defined {
		return -1, defined, err
	}

	cfsPeriodUs, err := cpuCGroup.readInt(_cgroupCPUCFSPeriodUsParam)
	if err != nil {
		return -1, false, err
	}

	return float64(cfsQuotaUs) / float64(cfsPeriodUs), true, nil
}

// NewCGroupsForPid returns a new *CGroups instance for the current
// process.
func NewCGroupsForPid(pid int) (Cgroups, error) {
	return NewCgroupsFromPath(fmt.Sprintf(_procTplPathMountInfo, pid), fmt.Sprintf(_procTplPathCgroup, pid))
}

// NewCGroupsForSelf returns a new *CGroups instance for the current
// process.
func NewCGroupsForSelf() (Cgroups, error) {
	return NewCgroupsFromPath(_procSelfPathMountInfo, _procSelfPathCgroup)
}
