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

	defer func(){
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
