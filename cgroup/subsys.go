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
	"os"
	"strconv"
	"strings"
)

const (
	_cgroupSep       = ":"
	_cgroupSubsysSep = ","
)

const (
	_csFieldIDID         = 0
	_csFieldIDSubsystems = 1
	_csFieldIDName       = 2
	_csFieldCount        = 3
)

type parseErr struct {
	typ, line string
}

func (e parseErr) Error() string {
	return fmt.Sprintf("parse %s, but \"%s\"", e.typ, e.line)
}

// Subsys represents the data structure for entities in
// `/proc/$PID/cgroup`. See also proc(5) for more information.
type Subsys struct {
	ID         int
	Subsystems []string
	Name       string
}

// NewSubsysFromLine returns a new *Subsys by parsing a string in
// the format of `/proc/$PID/cgroup`
func NewSubsysFromLine(line string) (*Subsys, error) {
	fields := strings.Split(line, _cgroupSep)

	if len(fields) != _csFieldCount {
		return nil, parseErr{"sub system", line}
	}

	id, err := strconv.Atoi(fields[_csFieldIDID])
	if err != nil {
		return nil, err
	}

	return &Subsys{
		ID:         id,
		Subsystems: strings.Split(fields[_csFieldIDSubsystems], _cgroupSubsysSep),
		Name:       fields[_csFieldIDName],
	}, nil
}

// NewCgroupSubsys parses procPathCGroup (usually at `/proc/$PID/cgroup`)
// and returns a new map[string]*CGroupSubsys.
func NewCgroupSubsys(procPathCGroup string) (map[string]*Subsys, error) {
	cgroupFile, err := os.Open(procPathCGroup)
	if err != nil {
		return nil, err
	}
	defer cgroupFile.Close()

	scanner := bufio.NewScanner(cgroupFile)
	subsystems := make(map[string]*Subsys)

	for scanner.Scan() {
		cgroup, err := NewSubsysFromLine(scanner.Text())
		if err != nil {
			return nil, err
		}
		for _, subsys := range cgroup.Subsystems {
			subsystems[subsys] = cgroup
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return subsystems, nil
}
