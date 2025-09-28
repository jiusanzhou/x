package config

import (
	"testing"
)

type testConfig struct {
	Foo string
}

func TestNew(t *testing.T) {
	type args struct {
		v    *testConfig
		opts []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"base",
			args{
				v: &testConfig{},
			},
			&Config{},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.v, tt.args.opts...)
			err := got.Init()
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
