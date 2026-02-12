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
Package service provides cross-platform system service management.

This package allows Go programs to be installed, started, stopped,
and managed as system services on multiple platforms including:

  - Linux (systemd)
  - macOS (launchd)
  - Windows (Windows Service Manager)

# Basic Usage

Create and manage a service:

	config := &service.Config{
	    Name:        "myservice",
	    DisplayName: "My Service",
	    Description: "A description of my service",
	}

	svc, err := service.New(config)
	if err != nil {
	    log.Fatal(err)
	}

	// Install the service
	err = svc.Install()

	// Start the service
	err = svc.Start()

	// Check status
	status, err := svc.Status()
	if status == service.StatusRunning {
	    fmt.Println("Service is running")
	}

	// Stop and uninstall
	err = svc.Stop()
	err = svc.Uninstall()

# Configuration Options

The Config struct supports various options:

	config := &service.Config{
	    Name:             "myservice",
	    DisplayName:      "My Service",
	    Description:      "Long description",
	    UserName:         "serviceuser",
	    Arguments:        []string{"--config", "/etc/myservice.conf"},
	    Executable:       "/usr/bin/myservice",
	    WorkingDirectory: "/var/lib/myservice",
	    Dependencies:     []string{"After=network.target"},
	}

# Service Interface

The Service interface provides:

  - Run: Block and run as a service
  - Start/Stop/Restart: Control the service
  - Install/Uninstall: Manage service registration
  - Status: Query current service status

# Status Values

	StatusUnknown  // Unable to determine status
	StatusRunning  // Service is running
	StatusStopped  // Service is stopped
*/
package service
