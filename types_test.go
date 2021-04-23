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
	"reflect"
	"testing"
	"time"
)

func TestDuration_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		d       Duration
		want    string
		wantErr bool
	}{
		{
			"Simple", Duration(time.Second), "1s", false,
		},
		{
			"Simple 2", Duration(time.Second + 10 * time.Minute + time.Hour), "1h10m1s", false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotbs, err := tt.d.MarshalJSON()
			got := string(gotbs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Duration.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Duration.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDuration_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		want  Duration
		data    string
		wantErr bool
	}{
		{
			"Simple", Duration(time.Second), "1s", false,
		},
		{
			"Simple 2", Duration(time.Second + 10 * time.Minute + time.Hour), "1h10m1s", false,
		},
	}
	var got Duration
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := (&got).UnmarshalJSON([]byte(tt.data))
			if (err != nil) != tt.wantErr {
				t.Errorf("Duration.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Duration.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}
