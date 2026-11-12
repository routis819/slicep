// Package vm implements the slicep bytecode compiler and virtual machine.
package vm

import (
	"fmt"
	"strconv"
)

// Value represents a runtime value in the slicep virtual machine.
//
//goaugtype:sum IntVal | FloatVal | StringVal | BoolVal | VoidVal
type Value any

// IntVal represents a 64-bit signed integer runtime value.
type IntVal int64

func (v IntVal) String() string {
	return strconv.FormatInt(int64(v), 10)
}

// FloatVal represents a 64-bit IEEE-754 floating-point runtime value.
type FloatVal float64

func (v FloatVal) String() string {
	return strconv.FormatFloat(float64(v), 'g', -1, 64)
}

// StringVal represents a UTF-8 string runtime value.
type StringVal string

func (v StringVal) String() string {
	return string(v)
}

// BoolVal represents a boolean runtime value (#t or #f).
type BoolVal bool

func (v BoolVal) String() string {
	if v {
		return "#t"
	}

	return "#f"
}

// VoidVal represents an unspecified/void runtime value (returned by define, set!, etc.).
type VoidVal struct{}

func (VoidVal) String() string {
	return "#<void>"
}

// IsTruthy reports whether v is considered true in a conditional test.
// Per R7RS §6.3, only #f (BoolVal(false)) counts as false; all other values count as true.
func IsTruthy(v Value) bool {
	switch val := v.(type) {
	case nil:
		return false
	case BoolVal:
		return bool(val)
	case IntVal, FloatVal, StringVal, VoidVal:
		return true
	}

	panic(fmt.Sprintf("unexpected Value type %T", v))
}
