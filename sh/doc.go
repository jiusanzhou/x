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
Package sh provides shell command execution utilities using mvdan.cc/sh
as the underlying shell interpreter.

This package allows running shell commands and scripts programmatically
without spawning external shell processes. It supports both inline commands
and script files.

# Basic Usage

Run simple commands using the default runner:

	// Run a command
	sh.Run("echo hello")

	// Run a script file (prefix with @)
	sh.Run("@script.sh")

# Custom Runner

Create a custom runner with options:

	runner, err := sh.New(
	    sh.WithStdout(os.Stdout),
	    sh.WithStderr(os.Stderr),
	    sh.WithDir("/path/to/workdir"),
	)
	if err != nil {
	    log.Fatal(err)
	}

	err = runner.Run(context.Background(), "ls -la")

# Script Files

When the command string starts with '@', the package treats the rest
as a file path and executes the contents of that file:

	sh.Run("@deploy.sh")  // Executes the contents of deploy.sh
*/
package sh
