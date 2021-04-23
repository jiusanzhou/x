/*
 * Copyright (c) 2021 wellwell.work, LLC by Zoe
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

package notify

import (
	"encoding/json"
	"errors"
	"sync"
)

// Notifier define the interface for notify
type Notifier interface {
	// We have no need to having those method
	// Type() string
	// Name() string

	// Init the notifier before using it
	// Maybe we can apply Option
	Init() error


	Send(Message, ...Subject) error

	// TODO: batch send
}

// Message define the interface for sending message
type Message interface {

}

// Subject define the interface for sending target
type Subject interface {

}

var (
	// ErrUnimplement common error
	ErrUnknownType = errors.New("unknown notifier type:")

	// factory registry store Notifier creator
	registry = make(map[string]func(json.RawMessage) (Notifier, error))
	rlock sync.RWMutex
)

// Object is a struct wraper for nofifier config
type Object struct {
	Type string
}

// New create a notify from config
// 
// Should we need to accept Option function-sytle args?
func New() (Notifier, error) {
	rlock.RLock()
	defer rlock.RUnlock()

	// take the creator out
	if 

	return nil, ErrUnimplement
}