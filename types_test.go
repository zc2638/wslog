// Created by zc on 2023/8/18.

package wslog

import (
	"reflect"
	"testing"
)

func Test_convertToColorKey(t *testing.T) {
	prefix := "\x1b[31m"
	suffix := "\x1b[0m"
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "1",
			args: args{
				b: []byte(`a=1 b="1 2" c=1=2 d="1\n\"2"`),
			},
			want: []byte(prefix + `a` + suffix + `=1 ` + prefix + `b` + suffix + `="1 2" ` + prefix + `c` + suffix + `=1=2 ` + prefix + `d` + suffix + `="1\n\"2"`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToColorKey(tt.args.b, []byte(prefix), []byte(suffix)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertToColorKey() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}
