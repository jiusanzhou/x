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

// Errors contains multi error
type Errors []error

// Error implements the error interface
func (es Errors) Error() string {
	var err string
	for _, e := range es {
		err += "; " + e.Error()
	}
	return err
}

// IsNil return if contains no error
func (es Errors) IsNil() bool {
	return len(es) == 0
}

// Add ad a new error to errors
func (es Errors) Add(ex ...error) {
	for _, e := range ex {
		if e != nil {
			es = append(es, e)
		}
	}
}

// NewErrors create errors
func NewErrors(errs ...error) Errors {
	es := Errors{}
	for _, e := range errs {
		es = append(es, e)
	}
	return es
}
