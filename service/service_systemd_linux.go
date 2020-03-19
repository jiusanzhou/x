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
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

type sysd struct {
	typ    string
	detect bool

	// configuration
	*Config

	svcname          string
	configPath       string
	execPath         string
	templ            *template.Template
	outfileSupported bool
}

// Type returns type of creator of current system
func (s sysd) Type() string {
	return s.typ
}

// Detect is we are detected
func (s sysd) Detect() bool {
	return s.detect
}

// New create a new service
func (s sysd) New(c *Config) (Service, error) {
	// create a new service at here
	if !s.Detect() {
		return nil, ErrNoServiceSystemDetected
	}

	// new a sysd
	ss := &sysd{Config: c}

	var err error

	// init config path
	ss.svcname = ss.Name + ".service"
	ss.configPath = "/etc/systemd/system/" + ss.Name + ".service"
	ss.templ, err = ss.getTemplate()
	if err != nil {
		return nil, err
	}
	ss.execPath, err = ss.getExecPath()
	if err != nil {
		return nil, err
	}
	ss.outfileSupported = ss.hasOutputFileSupport()

	return ss, nil
}

func (s sysd) run(cmd string, args ...string) error {
	var margs = append(
		[]string{"start", s.svcname},
		args...,
	)
	return run("systemctl", margs...)
}

// String return serivce's name
func (s sysd) String() string {
	if len(s.DisplayName) > 0 {
		return s.DisplayName
	}
	return s.Name
}

func (s sysd) Run() error {
	// TODO: wait implement
	// run the service
	return nil
}

func (s sysd) Start() error {
	return s.run("start")
}

func (s sysd) Stop() error {
	return s.run("stop")
}

func (s sysd) Restart() error {
	return s.run("restart")
}

func (s sysd) Install() error {

	// check if have installed?
	_, err := os.Stat(s.configPath)
	if err == nil {
		return fmt.Errorf("Init already exists: %s", s.configPath)
	}

	// create config path
	f, err := os.Create(s.configPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// create service file
	var data = &struct {
		*Config
		Path                 string
		HasOutputFileSupport bool
		ReloadSignal         string
		PIDFile              string
		Restart              string
		SuccessExitStatus    string
		LogOutput            bool
	}{
		s.Config,
		s.execPath,
		s.outfileSupported,

		// TODO: make those field confurable
		"",
		"",
		"always",
		"",
		true,
		// TODO: make those field confurable
	}

	// execute template to generate file
	err = s.templ.Execute(f, data)
	if err != nil {
		return err
	}

	// enable service
	err = run("systemctl", "enable", s.svcname)
	if err != nil {
		return err
	}

	return run("systemctl", "daemon-reload")
}

func (s sysd) Uninstall() error {
	err := run("systemctl", "disable", s.svcname)
	if err != nil {
		return err
	}

	return os.Remove(s.configPath)
}

func (s sysd) Status() (Status, error) {
	exitCode, out, err := runWithOutput("systemctl", "is-active", s.Name)
	if exitCode == 0 && err != nil {
		return StatusUnknown, err
	}

	switch {
	case strings.HasPrefix(out, "active"):
		return StatusRunning, nil
	case strings.HasPrefix(out, "inactive"):
		return StatusStopped, nil
	case strings.HasPrefix(out, "failed"):
		return StatusUnknown, errors.New("service in failed state")
	default:
		return StatusUnknown, ErrNotInstalled
	}
}

func (s sysd) getTemplate() (*template.Template, error) {
	// TODO: custom  template
	return template.New("").Parse(systemdScript)
}

func (s sysd) getSystemdVersion() int64 {
	_, out, err := runWithOutput("systemctl", "--version")
	if err != nil {
		return -1
	}

	re := regexp.MustCompile(`systemd ([0-9]+)`)
	matches := re.FindStringSubmatch(out)
	if len(matches) != 2 {
		return -1
	}

	v, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return -1
	}

	return v
}

func (s sysd) hasOutputFileSupport() bool {
	defaultValue := true
	version := s.getSystemdVersion()
	if version == -1 {
		return defaultValue
	}

	if version < 236 {
		return false
	}

	return defaultValue
}

func init() {
	// register sysd creator
	Register(
		sysd{typ: "systemd", detect: isSystemd()},
	)
}

// detect is systemd
func isSystemd() bool {
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		return true
	}
	if _, err := os.Stat("/proc/1/comm"); err == nil {
		filerc, err := os.Open("/proc/1/comm")
		if err != nil {
			return false
		}
		defer filerc.Close()

		buf := new(bytes.Buffer)
		buf.ReadFrom(filerc)
		contents := buf.String()

		if strings.Trim(contents, " \r\n") == "systemd" {
			return true
		}
	}
	return false
}

const systemdScript = `[Unit]
Description={{.Description}}
ConditionFileIsExecutable={{.Path|cmdEscape}}
{{range $i, $dep := .Dependencies}} 
{{$dep}} {{end}}
[Service]
StartLimitInterval=5
StartLimitBurst=10
ExecStart={{.Path|cmdEscape}}{{range .Arguments}} {{.|cmd}}{{end}}
{{if .ChRoot}}RootDirectory={{.ChRoot|cmd}}{{end}}
{{if .WorkingDirectory}}WorkingDirectory={{.WorkingDirectory|cmdEscape}}{{end}}
{{if .UserName}}User={{.UserName}}{{end}}
{{if .ReloadSignal}}ExecReload=/bin/kill -{{.ReloadSignal}} "$MAINPID"{{end}}
{{if .PIDFile}}PIDFile={{.PIDFile|cmd}}{{end}}
{{if and .LogOutput .HasOutputFileSupport -}}
StandardOutput=file:/var/log/{{.Name}}.out
StandardError=file:/var/log/{{.Name}}.err
{{- end}}
{{if .Restart}}Restart={{.Restart}}{{end}}
{{if .SuccessExitStatus}}SuccessExitStatus={{.SuccessExitStatus}}{{end}}
RestartSec=120
EnvironmentFile=-/etc/sysconfig/{{.Name}}
[Install]
WantedBy=multi-user.target
`
