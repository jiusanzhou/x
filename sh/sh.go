// Package sh to run command programing
package sh

import (
	"context"
	"io"
	"os"
	"strings"

	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

var (
	parser = syntax.NewParser()

	// DefaultRunner use to execute with package
	DefaultRunner, _ = New()
)

// Runner container runner configuration to execute command
type Runner struct {
	runner *interp.Runner
	opts   []Option

	running bool
}

func newRunner() (*interp.Runner,  error) {
	return interp.New(interp.StdIO(os.Stdin, os.Stdout, os.Stderr))
}

// New create a runner
func New(opts ...Option) (*Runner,  error) {

	var r = &Runner{
		opts: opts,
	}

	var err error

	r.runner, err = newRunner()
	if err != nil {
		return nil, err
	}

	for _, o := range opts {
		err = o(r.runner)
		if err != nil {
			return nil, err
		}
	}

	return r, err
}

func (r *Runner) copy() (*Runner, error) {
	return New(r.opts...)
}

// Run to execute with a runner
// we wont to handle sh file, so need to read them all at first
// but we can use a special char to mark this content is a file
// like: @, let's do it.
func (r *Runner) Run(ctx context.Context, s string, opts ...Option) error {

	// if s startswith a `@`
	// that means this is a file
	var (
		prg *syntax.File
		rdr io.Reader
		err error
	)
	
	var rn = r

	// must copy for overwrite options
	rn, err = r.copy()
	if err != nil {
		return err
	}

	// reload options
	for _, o := range opts {
		o(rn.runner)
	}

	if s[0] == '@' {
		rdr, err = os.Open(s[1:])
		if err != nil {
			return err
		}
	} else {
		rdr = strings.NewReader(s)
	}

	prg, err = parser.Parse(rdr, "__todo__")
	if err != nil {
		return err
	}

	// TODO: add resource limit

	rn.running = true
	return rn.runner.Run(ctx, prg)
}

// Run use default runner to run command
func Run(s string, opts ...Option) error {
	return DefaultRunner.Run(context.Background(), s, opts...)
}