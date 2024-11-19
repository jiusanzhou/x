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

package x

import (
	"time"
)

// Duration implement JSON marshall for time.Duration
type Duration time.Duration

// MarshalJSON implement MarshalJSON
func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

// UnmarshalJSON implement UnmarshalJSON
func (d *Duration) UnmarshalJSON(b []byte) error {
	v, err := time.ParseDuration(string(b))
	*d = Duration(v)
	return err
}

// Min
func Min[V int32 | int64 | float32 | float64](a, b V) V {
	if a < b {
		return a
	}
	return b
}

// Max
func Max[V int32 | int64 | float32 | float64](a, b V) V {
	if a > b {
		return a
	}
	return b
}
