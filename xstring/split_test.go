package xstring

import (
	"reflect"
	"testing"
)

func TestSplitToInt64(t *testing.T) {
	type args struct {
		s   string
		sep string
	}
	tests := []struct {
		name    string
		args    args
		want    []int64
		wantErr bool
	}{
		{
			name:    "测试分隔符",
			args:    args{s: "11111_1", sep: "_"},
			want:    []int64{11111, 1},
			wantErr: false,
		},
		{
			name:    "测试包含特殊符号",
			args:    args{s: "111s1_1", sep: "_"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SplitToInt64(tt.args.s, tt.args.sep)
			if (err != nil) != tt.wantErr {
				t.Errorf("SplitToInt64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitToInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}
