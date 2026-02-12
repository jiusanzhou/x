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

package x

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// GraceSignalChan grace signal chan
type GraceSignalChan chan struct{}

// GraceStart start a grace function
func GraceStart(f func(ch GraceSignalChan) error) error {

	// handler signal to exit
	stopCh := make(GraceSignalChan, 1)
	defer close(stopCh)

	errCh := make(chan error, 1)
	go func() {
		errCh <- f(stopCh)
	}()

	// register sigint
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM, syscall.SIGINT)

	defer func() {
		stopCh <- struct{}{}
	}()

	select {
	case <-sigterm:
		return nil
	case err := <-errCh:
		return err
	}
}

// GraceRun grace run a function
func GraceRun(f func() error) error {

	// make a signal handler

	errCh := make(chan error, 1)
	go func() {
		errCh <- f()
	}()

	// register sigint
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-sigterm:
		return nil
	case err := <-errCh:
		return err
	}
}

// CleanupFunc is a function that performs cleanup operations.
type CleanupFunc func()

// GraceRunner provides graceful shutdown with cleanup function registration.
type GraceRunner struct {
	cleanups []CleanupFunc
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewGraceRunner creates a new GraceRunner with context support.
func NewGraceRunner() *GraceRunner {
	ctx, cancel := context.WithCancel(context.Background())
	return &GraceRunner{
		cleanups: make([]CleanupFunc, 0),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// RegisterCleanup adds a cleanup function to be called on shutdown.
// Cleanup functions are called in reverse order (LIFO).
func (g *GraceRunner) RegisterCleanup(f CleanupFunc) {
	g.cleanups = append(g.cleanups, f)
}

// Context returns the runner's context, which is cancelled on shutdown.
func (g *GraceRunner) Context() context.Context {
	return g.ctx
}

// runCleanups executes all registered cleanup functions in reverse order.
func (g *GraceRunner) runCleanups() {
	for i := len(g.cleanups) - 1; i >= 0; i-- {
		g.cleanups[i]()
	}
}

// Run executes the given function and handles graceful shutdown.
// On signal (SIGTERM/SIGINT) or function return, cleanup functions are called.
func (g *GraceRunner) Run(f func() error) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- f()
	}()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(sigterm)

	var retErr error
	select {
	case <-sigterm:
		g.cancel()
	case retErr = <-errCh:
		g.cancel()
	}

	g.runCleanups()
	return retErr
}

// GraceRunWithCleanup runs f with cleanup support via the setup function.
// The setup function receives a register function to add cleanup handlers.
func GraceRunWithCleanup(setup func(register func(CleanupFunc)), f func() error) error {
	runner := NewGraceRunner()
	setup(runner.RegisterCleanup)
	return runner.Run(f)
}
