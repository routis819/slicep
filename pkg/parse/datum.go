package parse

import "github.com/routis819/slicep/pkg/lex"

// Datum represents an R7RS §7.1.2 external representation (S-expression).
//
//goaugtype:sum *BoolDatum | *IntDatum | *FloatDatum | *StringDatum | *SymbolDatum | *ListDatum
type Datum any

// BoolDatum represents a <boolean> datum (#t, #f, #true, #false).
type BoolDatum struct {
	Value bool
	Pos   lex.Pos
}

// IntDatum represents an integer <number> datum (<sign> <uinteger 10>).
type IntDatum struct {
	Value int64
	Pos   lex.Pos
}

// FloatDatum represents a floating-point <number> datum (<sign> <decimal 10> or <infnan>).
type FloatDatum struct {
	Value float64
	Pos   lex.Pos
}

// StringDatum represents a <string> datum.
type StringDatum struct {
	Value string
	Pos   lex.Pos
}

// SymbolDatum represents a <symbol> (<identifier>) datum.
type SymbolDatum struct {
	Value string
	Pos   lex.Pos
}

// ListDatum represents a <list> ((<datum>*)) compound datum.
type ListDatum struct {
	Elements []Datum
	Pos      lex.Pos
}
