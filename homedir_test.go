package x

import (
	"path/filepath"
	"testing"
)

func TestWithHomeDir(t *testing.T) {
	home, err := HomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Only Wavy",
			args: args{
				path: "~",
			},
			want:    home,
			wantErr: false,
		},
		{
			name: "Start With Wavy",
			args: args{
				path: "~/tmp",
			},
			want:    filepath.Join(home, "tmp"),
			wantErr: false,
		},
		{
			name: "Contains Dots",
			args: args{
				path: "~/../tmp",
			},
			want:    filepath.Join(filepath.Dir(home), "tmp"),
			wantErr: false,
		},
		{
			name: "Empty String",
			args: args{
				path: "",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Start With Slash",
			args: args{
				path: "/tmp",
			},
			want:    "/tmp",
			wantErr: false,
		},
		{
			name: "Pure Directory",
			args: args{
				path: "tmp",
			},
			want:    "tmp",
			wantErr: false,
		},
		{
			name: "With Dot",
			args: args{
				path: "../tmp",
			},
			want:    "../tmp",
			wantErr: false,
		},
		{
			name: "Invalid User Expansion",
			args: args{
				path: "~user/tmp",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := WithHomeDir(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("WithHomeDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("WithHomeDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHomeDir(t *testing.T) {
	home, err := HomeDir()
	if err != nil {
		t.Fatalf("HomeDir() error = %v", err)
	}
	if home == "" {
		t.Error("HomeDir() returned empty string")
	}

	home2, err := HomeDir()
	if err != nil {
		t.Fatalf("HomeDir() second call error = %v", err)
	}
	if home != home2 {
		t.Errorf("HomeDir() caching failed: got %v, then %v", home, home2)
	}
}

func TestReset(t *testing.T) {
	home1, err := HomeDir()
	if err != nil {
		t.Fatalf("HomeDir() error = %v", err)
	}

	Reset()

	home2, err := HomeDir()
	if err != nil {
		t.Fatalf("HomeDir() after Reset() error = %v", err)
	}

	if home1 != home2 {
		t.Errorf("HomeDir() returned different values before and after Reset(): %v vs %v", home1, home2)
	}
}
