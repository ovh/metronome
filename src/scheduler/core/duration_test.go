package core

import "testing"

func TestParseInt64(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{{
		name: "Easy",
		args: args{
			value: "100D",
		},
		want: 100,
	}, {
		name: "Real zero",
		args: args{
			value: "0Y",
		},
		want: 0,
	}, {
		name: "zero",
		args: args{
			value: "M",
		},
		want: 0,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseInt64(tt.args.value); got != tt.want {
				t.Errorf("ParseInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}
