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
	"strings"
	"time"
)

var (
	onlyVersion = false
	envStoreKey = "X_LDFLAGS"
	defaultArgs = []string{}

	// data
	gitVersion string
	gitCommit  string
	gitState   string
)

func init() {
	flag.BoolVar(&onlyVersion, "v", onlyVersion, "Print only version tag")
	flag.StringVar(&envStoreKey, "env-key", envStoreKey, "Env key to store the LDFLAGS")
}

func genLDFlags() string {

	if gitState == "dirty" {
		gitVersion += "-dirty"
	}

	ldflagsStr := "" // "-s -w"
	ldflagsStr += " -X go.zoe.im/x/version.GitVersion=" + gitVersion
	ldflagsStr += " -X go.zoe.im/x/version.GitCommit=" + gitCommit
	ldflagsStr += " -X go.zoe.im/x/version.GitTreeState=" + gitState
	ldflagsStr += " -X go.zoe.im/x/version.BuildDate=" + time.Now().UTC().Format(time.RFC3339)
	return ldflagsStr
}

// genReleaseTag Use git describe to find the version based on tags.
func releaseTag(gitCommit string) string {
	gv := gitRun("describe", "--tags", "--match", "v*", "--abbrev=14", gitCommit+"^{commit}")

	if gv == "" {
		return "v0.0.0"
	}

	// This translates the "git describe" to an actual semver.org
	// compatible semantic version that looks something like this:
	//   v1.1.0-alpha.0.6+84c76d1142ea4d
	// #
	// TODO: We continue calling this "git version" because so many
	// downstream consumers are expecting it there.
	// #
	// These regexes are painful enough in sed...
	// We don't want to do them in pure shell, so disable SC2001
	// shellcheck disable=SC2001

	// We have distance to subversion (v1.1.0-subversion-1-gCommitHash)

	// We have distance to base tag (v1.1.0-1-gCommitHash)

	// replace last - to +

	parts := strings.Split(gv, "-")
	last := len(parts) - 1

	if last >= 2 {
		return strings.Join(parts[:last], "-") + "+" + parts[last]
	}

	return gv
}

// treeState returns Check if the tree is dirty.  default to dirty
func treeState() string {
	if len(gitRun("status", "--porcelain")) == 0 {
		return "clean"
	}
	return "dirty"
}

// commitID returns the abbreviated commit-id hash of the last commit.
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

func main() {
	// parse command
	flag.Parse()

	// add more args for git
	if len(os.Args) > 1 {
		// set root path
		defaultArgs = append(defaultArgs, "--work-tree", os.Args[1])
	}

	gitCommit = commitID()
	gitState = treeState()
	gitVersion = releaseTag(gitCommit)

	if onlyVersion {
		fmt.Println(gitVersion)
		return
	}

	// store ldflags into env
	st := genLDFlags()
	os.Setenv(envStoreKey, st)
	fmt.Println(st)
}
