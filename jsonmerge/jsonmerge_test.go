package jsonmerge

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestMerge(t *testing.T) {
	type args struct {
		dst string
		src []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		result  string
	}{
		{
			name: "Zero Value",
			args: args{
				dst: `{"a": {"b": 0, "c": true, "d": 1}}`,
				src: []string{
					`{"a": {"b": 1, "c": false}}`,
				},
			},
			result: `{"a":{"b":1,"c":false,"d":1}}`,
		},
		{
			name: "Slice Append",
			args: args{
				dst: `{"a": [0, 1]}`,
				src: []string{
					`{"a": [3, 4]}`,
				},
			},
			result: `{"a":[0,1,3,4]}`,
		},
		{
			name: "Map Fields",
			args: args{
				dst: `{"a": [0, 1]}`,
				src: []string{
					`{"a": [3, 4], "b": {"c": 1}}`,
					`{"b": {"c": 2, "d": 2}}`,
				},
			},
			result: `{"a":[0,1,3,4],"b":{"c":2,"d":2}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var dstv map[string]interface{}
			json.Unmarshal([]byte(tt.args.dst), &dstv)

			var srcvs []interface{}
			for _, s := range tt.args.src {
				var sv map[string]interface{}
				json.Unmarshal([]byte(s), &sv)
				srcvs = append(srcvs, sv)
			}

			if err := Merge(&dstv, srcvs...); (err != nil) != tt.wantErr {
				t.Errorf("Merge() error = %v, wantErr %v", err, tt.wantErr)
			}

			var buf bytes.Buffer
			json.NewEncoder(&buf).Encode(dstv)

			if !bytes.Equal(bytes.TrimSpace(buf.Bytes()), []byte(tt.result)) {
				t.Errorf("Merge() result not corrent, got = %s, except = %s", bytes.TrimSpace(buf.Bytes()), tt.result)
			}
		})
	}
}
