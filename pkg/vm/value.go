// Package vm provides the slicep bytecode compiler and the stack-based
// virtual machine that executes the compiled bytecode.
//
// The compiler statically guarantees termination of every compiled program:
// names are bound strictly in declaration order, so self recursion and mutual
// recursion are rejected at compile time, and functions are not first-class
// values (they can neither be referenced as values nor reassigned).
package vm

import (
	"math"
	"strconv"
	"strings"
)

// Value is a runtime value of the slicep VM.
//
//goaugtype:sum Int | Float | String | Bool | Void
type Value any

// Int is an integer runtime value.
type Int int64

// Float is a floating-point runtime value.
type Float float64

// String is a string runtime value.
type String string

// Bool is a boolean runtime value.
type Bool bool

// Void is the unspecified value, produced by forms such as set! and by a
// one-armed if whose test is false.
type Void struct{}

// IsTruthy reports whether v counts as true in a conditional test.
// Following R7RS-small, only #f is false; every other value (including 0,
// "" and Void) is true.
func IsTruthy(v Value) bool {
	b, ok := v.(Bool)

	return !ok || bool(b)
}

// FormatValue returns the external representation of v.
func FormatValue(v Value) string {
	switch x := v.(type) {
	case nil:
		return "<nil>"
	case Int:
		return strconv.FormatInt(int64(x), 10)
	case Float:
		return formatFloat(float64(x))
	case String:
		return strconv.Quote(string(x))
	case Bool:
		if x {
			return "#t"
		}

		return "#f"
	case Void:
		return "#<void>"
	}

	return "<unknown>"
}

func formatFloat(f float64) string {
	switch {
	case math.IsNaN(f):
		return "+nan.0"
	case math.IsInf(f, 1):
		return "+inf.0"
	case math.IsInf(f, -1):
		return "-inf.0"
	}

	s := strconv.FormatFloat(f, 'g', -1, 64)
	if !strings.ContainsAny(s, ".eEn") {
		s += ".0"
	}

	return s
}
