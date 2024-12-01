package x

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Example struct {
	A string
	B int
	C bool
	D []string
	E map[string]string
}

var example = Example{
	A: "a",
	B: 1,
	C: true,
	D: []string{"d1", "d2"},
	E: map[string]string{
		"Ek1": "Ev1",
	},
}

var node_mock = &v1.Node{
	ObjectMeta: metav1.ObjectMeta{
		Name: "im_node_name",
	},
}

func TestSelectorField_Match(t *testing.T) {

	type fields struct {
		Key    string
		Values []string
	}

	tests := []struct {
		name   string
		fields fields
		input  interface{}
		want   bool
	}{
		{"Match 1", fields{Key: ".A", Values: []string{"a"}}, example, true},
		{"Not Match 1", fields{Key: ".A", Values: []string{""}}, example, false},
		{"Map Match 1", fields{Key: ".E.Ek1", Values: []string{"Ev1"}}, example, true},
		{"Auto prefix Match", fields{Key: "A", Values: []string{"a"}}, example, true},
		{"Glob prefix Match", fields{Key: "A", Values: []string{"*"}}, example, true},
		{"Node Match", fields{Key: ".Name", Values: []string{"im_node_name"}}, node_mock, true},
		{"Node Match Failed", fields{Key: ".Namex", Values: []string{"im_node_name"}}, node_mock, false},
		{"Node Match Failed 1", fields{Key: ".Name", Values: []string{"im_node_name1"}}, node_mock, false},
		{"Raw Template", fields{Key: "`test_name`", Values: []string{"test_name"}}, node_mock, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SelectorField{
				Key:    tt.fields.Key,
				Values: tt.fields.Values,
			}
			s.Init()
			if got := s.Match(tt.input); got != tt.want {
				t.Errorf("SelectorField.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetValueByPath(t *testing.T) {
	type args struct {
		obj interface{}
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{"Simple", args{example, ".A"}, "a", false},
		{"Add Dot", args{example, "B"}, "1", false},
		{"Node Name", args{node_mock, ".Name"}, "im_node_name", false},
		{"Full Template", args{example, `{{ if eq .A "a" }}true{{ else }}false{{ end }}`}, "true", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetValueByPath(tt.args.obj, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetValueByPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetValueByPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
