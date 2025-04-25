package httputil

import "testing"

func TestResponse_Flush(t *testing.T) {
	type args struct {
		fs []Field
	}
	tests := []struct {
		name string
		r    *Response
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tt.r.WithDataOrErr(returnDemo())
		tt.r.Flush(tt.args.fs...)
	}
}

func returnDemo() (int, error) {
	return 0, nil
}
