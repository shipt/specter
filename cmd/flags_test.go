package cmd

import (
	"testing"

	"github.com/namsral/flag"
)

var test2 = flag.String("test2", "", "test flag")

func init() {
	flag.Set("test2", "this has been set")
	flag.Parse()
}

func TestIsFlagPassed(t *testing.T) {

	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"test", args{"test"}, false},
		{"test2", args{"test2"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsFlagPassed(tt.args.name); got != tt.want {
				t.Errorf("IsFlagPassed() = %v, want %v", got, tt.want)
			}
		})
	}
}
