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
	"errors"
	"os"
	"path/filepath"
)

// ControlAction list valid string texts to use in Control.
type ControlAction string

// Status represents service status as an byte value
type Status byte

// Status of service represented as an byte
const (
	StatusUnknown Status = iota // Status is unable to be determined due to an error or it was not installed.
	StatusRunning
	StatusStopped
)

// Config provides the setup for a Service. The Name field is required.
type Config struct {
	Name        string   // Required name of the service. No spaces suggested.
	DisplayName string   // Display name, spaces allowed.
	Description string   // Long description of service.
	UserName    string   // Run as username.
	Arguments   []string // Run with arguments.

	// Optional field to specify the executable for service.
	// If empty the current executable is used.
	Executable string

	// Array of service dependencies.
	// Not yet fully implemented on Linux or OS X:
	//  1. Support linux-systemd dependencies, just put each full line as the
	//     element of the string array, such as
	//     "After=network.target syslog.target"
	//     "Requires=syslog.target"
	//     Note, such lines will be directly appended into the [Unit] of
	//     the generated service config file, will not check their correctness.
	Dependencies []string

	// The following fields are not supported on Windows.
	WorkingDirectory string // Initial working directory.
	ChRoot           string

	// System specific options.
	//  * OS X
	//    - LaunchdConfig string ()      - Use custom launchd config
	//    - KeepAlive     bool   (true)
	//    - RunAtLoad     bool   (false)
	//    - UserService   bool   (false) - Install as a current user service.
	//    - SessionCreate bool   (false) - Create a full user session.
	//  * POSIX
	//    - SystemdScript string ()                 - Use custom systemd script
	//    - UpstartScript string ()                 - Use custom upstart script
	//    - SysvScript    string ()                 - Use custom sysv script
	//    - RunWait       func() (wait for SIGNAL)  - Do not install signal but wait for this function to return.
	//    - ReloadSignal  string () [USR1, ...]     - Signal to send on reaload.
	//    - PIDFile       string () [/run/prog.pid] - Location of the PID file.
	//    - LogOutput     bool   (false)            - Redirect StdErr & StandardOutPath to files.
	//    - Restart       string (always)           - How shall service be restarted.
	//    - SuccessExitStatus string ()             - The list of exit status that shall be considered as successful,
	//                                                in addition to the default ones.
	// Option KeyValue
}

func (c *Config) getExecPath() (string, error) {
	if len(c.Executable) != 0 {
		return filepath.Abs(c.Executable)
	}
	return os.Executable()
}

// Service represents a service that can be run or controlled.
type Service interface {
	// Run should be called shortly after the program entry point.
	// After Interface.Stop has finished running, Run will stop blocking.
	// After Run stops blocking, the program must exit shortly after.
	Run() error

	// Start signals to the OS service manager the given service should start.
	Start() error

	// Stop signals to the OS service manager the given service should stop.
	Stop() error

	// Restart signals to the OS service manager the given service should stop then start.
	Restart() error

	// Install setups up the given service in the OS service manager. This may require
	// greater rights. Will return an error if it is already installed.
	Install() error

	// Uninstall removes the given service from the OS service manager. This may require
	// greater rights. Will return an error if the service is not present.
	Uninstall() error

	// String displays the name of the service. The display name if present,
	// otherwise the name.
	String() string

	// Platform displays the name of the system that manages the service.
	// In most cases this will be the same as service.Platform().
	// Platform() string

	// Status returns the current service status.
	Status() (Status, error)
}

// Creator create a service from config
type Creator interface {
	// Type returns a description of the system.
	Type() string

	// Detect returns true if the system is available to use.
	Detect() bool

	// New creates a new service for this creator.
	New(c *Config) (Service, error)
}

var (
	// registed creators
	creators = []Creator{}

	// ErrNameFieldRequired is returned when Config.Name is empty.
	ErrNameFieldRequired = errors.New("Config.Name field is required")
	// ErrNoServiceSystemDetected is returned when no system was detected.
	ErrNoServiceSystemDetected = errors.New("No service system detected")
	// ErrNotInstalled is returned when the service is not installed
	ErrNotInstalled = errors.New("the service is not installed")
)

// Register register a new ceator
func Register(c Creator) {
	creators = append(creators, c)
}

// New creates a new service based on a service interface and configuration.
func New(c *Config) (Service, error) {
	if len(c.Name) == 0 {
		return nil, ErrNameFieldRequired
	}

	// TODO: can we choose we one we want to use?
	for _, creator := range creators {
		if creator.Detect() {
			return creator.New(c)
		}
	}
	return nil, ErrNoServiceSystemDetected
}
