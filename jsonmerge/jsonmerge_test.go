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

func TestNew(t *testing.T) {
	f := New()
	if f == nil {
		t.Fatal("New() returned nil")
	}
}

func TestNew_WithOptions(t *testing.T) {
	f := New(
		Overwrite(true),
		AppendSlice(false),
		TypeCheck(true),
		MaxMergeDepth(5),
	)
	if f == nil {
		t.Fatal("New() with options returned nil")
	}
}

func TestFactory_With(t *testing.T) {
	f := New()
	f2 := f.With(Overwrite(true))
	if f2 == nil {
		t.Fatal("With() returned nil")
	}
}

func TestMerge_NilDst(t *testing.T) {
	src := map[string]interface{}{"a": 1}
	err := Merge(nil, src)
	if err != ErrNilArguments {
		t.Errorf("Merge(nil, ...) error = %v, want ErrNilArguments", err)
	}
}

func TestMerge_NilSrc(t *testing.T) {
	dst := map[string]interface{}{"a": 1}
	err := Merge(&dst, nil)
	if err != ErrNilArguments {
		t.Errorf("Merge(..., nil) error = %v, want ErrNilArguments", err)
	}
}

func TestMerge_NonPointer(t *testing.T) {
	dst := map[string]interface{}{"a": 1}
	src := map[string]interface{}{"b": 2}
	err := New().Merge(dst, src)
	if err != ErrNonPointerArgument {
		t.Errorf("Merge(non-pointer) error = %v, want ErrNonPointerArgument", err)
	}
}

func TestMerge_DifferentTypes(t *testing.T) {
	dst := map[string]interface{}{"a": 1}
	src := "string"
	err := Merge(&dst, src)
	if err != ErrDifferentArgumentsTypes {
		t.Errorf("Merge(different types) error = %v, want ErrDifferentArgumentsTypes", err)
	}
}

func TestMerge_WithOverwrite(t *testing.T) {
	dst := map[string]interface{}{"a": 1, "b": 2}
	src := map[string]interface{}{"a": 10}

	f := New(Overwrite(true))
	err := f.Merge(&dst, src)
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	if dst["a"] != 10 {
		t.Errorf("Merge with Overwrite: dst[a] = %v, want 10", dst["a"])
	}
}

func TestMerge_NestedMaps(t *testing.T) {
	dst := map[string]interface{}{
		"level1": map[string]interface{}{
			"a": 1,
			"b": 2,
		},
	}
	src := map[string]interface{}{
		"level1": map[string]interface{}{
			"b": 20,
			"c": 3,
		},
	}

	err := Merge(&dst, src)
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	level1 := dst["level1"].(map[string]interface{})
	if level1["a"] != 1 {
		t.Errorf("Nested merge: level1[a] = %v, want 1", level1["a"])
	}
	if level1["b"] != 20 {
		t.Errorf("Nested merge: level1[b] = %v, want 20", level1["b"])
	}
	if level1["c"] != 3 {
		t.Errorf("Nested merge: level1[c] = %v, want 3", level1["c"])
	}
}

func TestMerge_MultipleSources(t *testing.T) {
	dst := map[string]interface{}{"a": 1}
	src1 := map[string]interface{}{"b": 2}
	src2 := map[string]interface{}{"c": 3}

	err := Merge(&dst, src1, src2)
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	if len(dst) != 3 {
		t.Errorf("Merge multiple sources: len = %d, want 3", len(dst))
	}
}

func TestOverwriteWithEmptySrc(t *testing.T) {
	f := New(OverwriteWithEmptySrc(false))
	if f == nil {
		t.Fatal("New() returned nil")
	}
}

func TestOverwriteSliceWithEmptySrc(t *testing.T) {
	f := New(OverwriteSliceWithEmptySrc(true))
	if f == nil {
		t.Fatal("New() returned nil")
	}
}

func TestDefault(t *testing.T) {
	if Default == nil {
		t.Error("Default should not be nil")
	}
}
