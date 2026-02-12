//go:build !go1.18

/*
 * Copyright (c) 2022 wellwell.work, LLC by Zoe
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
	"errors"
	"reflect"
)

// Iter is the struct to process items
type Iter struct {
	items []interface{}

	// TODO: array should be
	filters func(interface{}) bool
	maps    func(interface{}) interface{}

	err error
}

// Filter register the filter function
// TODO: filter should call as the map chain
func (i *Iter) Filter(f func(interface{}) bool) *Iter {
	i.filters = f
	return i
}

// Map register the map function to convert item
// TODO: map should call as a chain
func (i *Iter) Map(f func(interface{}) interface{}) *Iter {
	i.maps = f
	return i
}

// Collect to take items out
func (i *Iter) Collect() []interface{} {
	// cache
	res := []interface{}{}
	for _, item := range i.items {
		if i.filters != nil && !i.filters(item) {
			continue
		}
		if i.maps != nil {
			v := i.maps(item)
			res = append(res, v)
		} else {
			res = append(res, item)
		}
	}
	return res
}

// ToIter return a new iter from processed items
func (i *Iter) ToIter() *Iter {
	return NewIter(i.Collect())
}

// Error get the init error
func (i *Iter) Error() error {
	return i.err
}

// NewIter returns a new Iter from a slice.
func NewIter(items interface{}) *Iter {
	iter := &Iter{}

	// ignore is not a slice or array
	v := reflect.ValueOf(items)
	if v.Kind() != reflect.Slice {
		iter.err = errors.New("Iteror value not supported")
		return iter
	}
	for i := 0; i < v.Len(); i++ {
		iter.items = append(iter.items, v.Index(i).Interface())
	}

	return iter
}
