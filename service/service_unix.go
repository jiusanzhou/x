//go:build linux || darwin || solaris || aix
// +build linux darwin solaris aix

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

package service

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"syscall"
)

func run(command string, arguments ...string) error {
	_, _, err := runCommand(command, false, arguments...)
	return err
}

func runWithOutput(command string, arguments ...string) (int, string, error) {
	return runCommand(command, true, arguments...)
}

func runCommand(command string, readStdout bool, arguments ...string) (int, string, error) {
	cmd := exec.Command(command, arguments...)

	var output string
	var stdout io.ReadCloser
	var err error

	if readStdout {
		// Connect pipe to read Stdout
		stdout, err = cmd.StdoutPipe()

		if err != nil {
			// Failed to connect pipe
			return 0, "", fmt.Errorf("%q failed to connect stdout pipe: %v", command, err)
		}
	}

	// Connect pipe to read Stderr
	stderr, err := cmd.StderrPipe()

	if err != nil {
		// Failed to connect pipe
		return 0, "", fmt.Errorf("%q failed to connect stderr pipe: %v", command, err)
	}

	// Do not use cmd.Run()
	if err := cmd.Start(); err != nil {
		// Problem while copying stdin, stdout, or stderr
		return 0, "", fmt.Errorf("%q failed: %v", command, err)
	}

	// Zero exit status
	// Darwin: launchctl can fail with a zero exit status,
	// so check for emtpy stderr
	if command == "launchctl" {
		slurp, _ := ioutil.ReadAll(stderr)
		if len(slurp) > 0 && !bytes.HasSuffix(slurp, []byte("Operation now in progress\n")) {
			return 0, "", fmt.Errorf("%q failed with stderr: %s", command, slurp)
		}
	}

	if readStdout {
		out, err := ioutil.ReadAll(stdout)
		if err != nil {
			return 0, "", fmt.Errorf("%q failed while attempting to read stdout: %v", command, err)
		} else if len(out) > 0 {
			output = string(out)
		}
	}

	if err := cmd.Wait(); err != nil {
		exitStatus, ok := isExitError(err)
		if ok {
			// Command didn't exit with a zero exit status.
			return exitStatus, output, err
		}

		// An error occurred and there is no exit status.
		return 0, output, fmt.Errorf("%q failed: %v", command, err)
	}

	return 0, output, nil
}

func isExitError(err error) (int, bool) {
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus(), true
		}
	}

	return 0, false
}
