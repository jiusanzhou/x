package x

import (
	"testing"
)

func Test_Match(t *testing.T) {
	type args struct {
		op     OperatorType
		values []string
		val    string
	}
	tests := []struct {
		name      string
		args      args
		validator error
		want      bool
	}{
		{"In", args{OperatorIn, []string{"1", "2"}, "1"}, nil, true},
		{"In", args{OperatorIn, []string{"1", "*"}, "aaa"}, nil, true},
		{"In", args{OperatorIn, []string{"1", "2"}, "3"}, nil, false},
		{"NotIn", args{OperatorNotIn, []string{"1", "2"}, "3"}, nil, true},
		{"NotIn", args{OperatorNotIn, []string{"1", "2"}, "2"}, nil, false},
		{"Exists", args{OperatorExists, []string{}, "1"}, nil, true},
		{"NotExists", args{OperatorNotExists, []string{}, "<nil>"}, nil, true},
		{"Lt", args{OperatorLt, []string{"3"}, "1"}, nil, true},
		{"Gt", args{OperatorGt, []string{"3"}, "4"}, nil, true},
		{"Range", args{OperatorRange, []string{"3", "6"}, "4"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.op.Validate(tt.args.values); tt.validator != got {
				t.Errorf("Validate() = %v, want %v", got, tt.validator)
			}
			if got := tt.args.op.Match(tt.args.val, tt.args.values); got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}
