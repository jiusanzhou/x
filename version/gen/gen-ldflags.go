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

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	onlyVersion   = false
	envStoreKey   = "X_LDFLAGS"
	defaultArgs   = []string{}
	incrementType = ""
	createTag     = false

	gitVersion string
	gitCommit  string
	gitState   string
)

func init() {
	flag.BoolVar(&onlyVersion, "v", onlyVersion, "Print only version tag")
	flag.StringVar(&envStoreKey, "env-key", envStoreKey, "Env key to store the LDFLAGS")
	flag.StringVar(&incrementType, "increment", incrementType, "Auto increment version (patch, minor, major, prerelease)")
	flag.BoolVar(&createTag, "tag", createTag, "Create git tag with new version")
}

func genLDFlags() string {

	if gitState == "dirty" {
		gitVersion += "-dirty"
	}

	ldflagsStr := ""
	ldflagsStr += " -X go.zoe.im/x/version.GitVersion=" + gitVersion
	ldflagsStr += " -X go.zoe.im/x/version.GitCommit=" + gitCommit
	ldflagsStr += " -X go.zoe.im/x/version.GitTreeState=" + gitState
	ldflagsStr += " -X go.zoe.im/x/version.BuildDate=" + time.Now().UTC().Format(time.RFC3339)
	return ldflagsStr
}

func releaseTag(gitCommit string) string {
	gv := gitRun("describe", "--tags", "--match", "v*", "--abbrev=14", gitCommit+"^{commit}")

	if gv == "" {
		return "v0.0.0"
	}

	parts := strings.Split(gv, "-")
	last := len(parts) - 1

	if last >= 2 {
		return strings.Join(parts[:last], "-") + "+" + parts[last]
	}

	return gv
}

func treeState() string {
	if len(gitRun("status", "--porcelain")) == 0 {
		return "clean"
	}
	return "dirty"
}

func commitID() string {
	id := gitRun("rev-parse", "HEAD^{commit}")
	if id == "" {
		id = "0000000000000000000000000000000000000000"
	}
	return id
}

func gitRun(args ...string) string {
	var nargs = []string{}
	nargs = append(nargs, defaultArgs...)
	nargs = append(nargs, args...)

	var s []byte
	s, _ = exec.Command("git", nargs...).Output()
	return strings.TrimSpace(string(s))
}

func latestTag() string {
	tag := gitRun("describe", "--tags", "--abbrev=0", "--match", "v*")
	if tag == "" {
		return "v0.0.0"
	}
	return tag
}

func parseVersion(v string) (major, minor, patch int64, prerelease, metadata string) {
	v = strings.TrimPrefix(v, "v")

	if idx := strings.Index(v, "+"); idx >= 0 {
		metadata = v[idx+1:]
		v = v[:idx]
	}

	if idx := strings.Index(v, "-"); idx >= 0 {
		prerelease = v[idx+1:]
		v = v[:idx]
	}

	parts := strings.Split(v, ".")
	if len(parts) >= 1 {
		major, _ = strconv.ParseInt(parts[0], 10, 64)
	}
	if len(parts) >= 2 {
		minor, _ = strconv.ParseInt(parts[1], 10, 64)
	}
	if len(parts) >= 3 {
		patch, _ = strconv.ParseInt(parts[2], 10, 64)
	}
	return
}

func incrementVersion(currentTag, incType string) string {
	major, minor, patch, prerelease, _ := parseVersion(currentTag)

	switch strings.ToLower(incType) {
	case "major":
		major++
		minor = 0
		patch = 0
		prerelease = ""
	case "minor":
		minor++
		patch = 0
		prerelease = ""
	case "patch":
		patch++
		prerelease = ""
	case "prerelease", "pre":
		if prerelease == "" {
			patch++
			prerelease = "0"
		} else {
			parts := strings.Split(prerelease, ".")
			lastIdx := len(parts) - 1
			if num, err := strconv.ParseInt(parts[lastIdx], 10, 64); err == nil {
				parts[lastIdx] = strconv.FormatInt(num+1, 10)
				prerelease = strings.Join(parts, ".")
			} else {
				prerelease = prerelease + ".1"
			}
		}
	default:
		return currentTag
	}

	newVersion := fmt.Sprintf("v%d.%d.%d", major, minor, patch)
	if prerelease != "" {
		newVersion += "-" + prerelease
	}
	return newVersion
}

func createGitTag(tag string) error {
	output := gitRun("tag", "-a", tag, "-m", "Release "+tag)
	if output != "" {
		return fmt.Errorf("failed to create tag: %s", output)
	}
	return nil
}

func main() {
	flag.Parse()

	if len(flag.Args()) > 0 {
		defaultArgs = append(defaultArgs, "--work-tree", flag.Args()[0])
	}

	gitCommit = commitID()
	gitState = treeState()

	if incrementType != "" {
		currentTag := latestTag()
		newVersion := incrementVersion(currentTag, incrementType)

		if createTag && gitState == "clean" {
			if err := createGitTag(newVersion); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating tag: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Created tag: %s\n", newVersion)
		}

		gitVersion = newVersion
	} else {
		gitVersion = releaseTag(gitCommit)
	}

	st := genLDFlags()

	if onlyVersion {
		fmt.Println(gitVersion)
		return
	}

	os.Setenv(envStoreKey, st)
	fmt.Println(st)
}
