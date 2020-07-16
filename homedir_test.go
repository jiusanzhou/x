package x

import "testing"

func TestWithHomeDir(t *testing.T) {
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
			want:    "/home/zoe",
			wantErr: false,
		},
		{
			name: "Start With Wavy",
			args: args{
				path: "~/tmp",
			},
			want:    "/home/zoe/tmp",
			wantErr: false,
		},
		{
			name: "Contains Dots",
			args: args{
				path: "~/../tmp",
			},
			want:    "/home/tmp",
			wantErr: false,
		},
		// {
		// 	name: "Empty String",
		// 	args: args{
		// 		path: "",
		// 	},
		// 	want:    "/home/zoe",
		// 	wantErr: false,
		// },
		// {
		// 	name: "Start With Slash",
		// 	args: args{
		// 		path: "/tmp",
		// 	},
		// 	want:    "/tmp",
		// 	wantErr: false,
		// },
		// {
		// 	name: "Pure Directory",
		// 	args: args{
		// 		path: "tmp",
		// 	},
		// 	want:    "/home/zoe/tmp",
		// 	wantErr: false,
		// },
		// {
		// 	name: "With Dot",
		// 	args: args{
		// 		path: "../tmp",
		// 	},
		// 	want:    "/home",
		// 	wantErr: false,
		// },
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
