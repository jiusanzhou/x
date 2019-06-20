package sh

import (
	"time"

	"mvdan.cc/sh/v3/interp"
)

// Option configuration
type Option func(*interp.Runner) error

// Env sets the interpreter's environment. If nil, a copy of the current
// process's environment is used.
var Env = interp.Env


// Dir sets the interpreter's working directory. If empty, the process's current
// directory is used.
var Dir = interp.Dir

// Params populates the shell options and parameters. For example, Params("-e",
// "--", "foo") will set the "-e" option and the parameters ["foo"], and
// Params("+e") will unset the "-e" option and leave the parameters untouched.
//
// This is similar to what the interpreter's "set" builtin does.
var Params = interp.Params


// WithExecModule sets up a runner with a chain of ExecModule middlewares. The
// chain is set up starting at the end, so that the first middleware in the list
// will be the first one to execute as part of the interpreter.
//
// The last or innermost module is always DefaultExec. You can make it
// unreachable by adding a middleware that never calls its next module.
var WithExecModule = interp.WithExecModules


// WithOpenModule sets up a runner with a chain of OpenModule middlewares. The
// chain is set up starting at the end, so that the first middleware in the list
// will be the first one to execute as part of the interpreter.
//
// The last or innermost module is always DefaultOpen. You can make it
// unreachable by adding a middleware that never calls its next module.
var WithOpenModule = interp.WithOpenModules


// StdIO configures an interpreter's standard input, standard output, and
// standard error. If out or err are nil, they default to a writer that discards
// the output.
var StdIO = interp.StdIO


// Timeout configures holds how much time the interpreter will wait for a
// program to stop after being sent an interrupt signal, after
// which a kill signal will be sent. This process will happen when the
// interpreter's context is cancelled.
func Timeout(t time.Duration) func(r *interp.Runner) error {
	return func(r *interp.Runner) error {
		r.KillTimeout = t
		return nil
	}
}