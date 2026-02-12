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

package version

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Masterminds/semver"
)

type Semver struct {
	*semver.Version
}

func NewSemver(v string) (*Semver, error) {
	ver, err := semver.NewVersion(v)
	if err != nil {
		return nil, err
	}
	return &Semver{Version: ver}, nil
}

func MustSemver(v string) *Semver {
	s, err := NewSemver(v)
	if err != nil {
		panic(err)
	}
	return s
}

func (s *Semver) IncrementMajor() *Semver {
	v := s.Version.IncMajor()
	return &Semver{Version: &v}
}

func (s *Semver) IncrementMinor() *Semver {
	v := s.Version.IncMinor()
	return &Semver{Version: &v}
}

func (s *Semver) IncrementPatch() *Semver {
	v := s.Version.IncPatch()
	return &Semver{Version: &v}
}

func (s *Semver) WithPrerelease(prerelease string) (*Semver, error) {
	v, err := s.Version.SetPrerelease(prerelease)
	if err != nil {
		return nil, err
	}
	return &Semver{Version: &v}, nil
}

func (s *Semver) WithMetadata(metadata string) (*Semver, error) {
	v, err := s.Version.SetMetadata(metadata)
	if err != nil {
		return nil, err
	}
	return &Semver{Version: &v}, nil
}

func (s *Semver) IncrementPrerelease() (*Semver, error) {
	pre := s.Version.Prerelease()
	if pre == "" {
		return s.WithPrerelease("0")
	}

	parts := strings.Split(pre, ".")
	lastIdx := len(parts) - 1

	if num, err := strconv.ParseInt(parts[lastIdx], 10, 64); err == nil {
		parts[lastIdx] = strconv.FormatInt(num+1, 10)
		return s.WithPrerelease(strings.Join(parts, "."))
	}

	return s.WithPrerelease(pre + ".1")
}

func (s *Semver) IsPrerelease() bool {
	return s.Version.Prerelease() != ""
}

func (s *Semver) IsStable() bool {
	return !s.IsPrerelease() && s.Version.Major() > 0
}

func (s *Semver) NextVersion(incrementType string) (*Semver, error) {
	switch strings.ToLower(incrementType) {
	case "major":
		return s.IncrementMajor(), nil
	case "minor":
		return s.IncrementMinor(), nil
	case "patch":
		return s.IncrementPatch(), nil
	case "prerelease", "pre":
		return s.IncrementPrerelease()
	default:
		return nil, fmt.Errorf("unknown increment type: %s", incrementType)
	}
}

func (s *Semver) String() string {
	return s.Version.Original()
}

func (s *Semver) TagString() string {
	orig := s.Version.Original()
	if strings.HasPrefix(orig, "v") {
		return orig
	}
	return "v" + s.Version.String()
}
