package vm

import (
	"math"
	"testing"
)

func TestIsTruthy(t *testing.T) {
	tests := []struct {
		v    Value
		want bool
	}{
		{Bool(false), false},
		{Bool(true), true},
		{Int(0), true},
		{Float(0), true},
		{String(""), true},
		{Void{}, true},
	}

	for _, tt := range tests {
		if got := IsTruthy(tt.v); got != tt.want {
			t.Errorf("IsTruthy(%s) = %v, want %v", FormatValue(tt.v), got, tt.want)
		}
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		v    Value
		want string
	}{
		{nil, "<nil>"},
		{Int(-42), "-42"},
		{Float(1.5), "1.5"},
		{Float(2), "2.0"},
		{Float(1e21), "1e+21"},
		{Float(math.Inf(1)), "+inf.0"},
		{Float(math.Inf(-1)), "-inf.0"},
		{Float(math.NaN()), "+nan.0"},
		{String("a\"b"), `"a\"b"`},
		{Bool(true), "#t"},
		{Bool(false), "#f"},
		{Void{}, "#<void>"},
	}

	for _, tt := range tests {
		if got := FormatValue(tt.v); got != tt.want {
			t.Errorf("FormatValue(%#v) = %q, want %q", tt.v, got, tt.want)
		}
	}
}
