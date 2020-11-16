// +build linux

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

package cgroup

import (
	"fmt"
	"testing"
)

func TestNewCGroupsForSelf(t *testing.T) {
	tests := []struct {
		name    string
		want    Cgroups
		wantErr bool
	}{
		// TODO: Add test cases.
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCGroupsForSelf()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCGroupsForSelf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			fmt.Println(got)
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("NewCGroupsForSelf() = %v, want %v", got, tt.want)
			// }
		})
	}
}
