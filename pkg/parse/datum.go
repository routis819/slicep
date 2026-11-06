package parse

// Datum represents an R7RS §7.1.2 external representation (S-expression).
//
//goaugtype:sum *BoolDatum | *IntDatum | *FloatDatum | *StringDatum | *SymbolDatum | *ListDatum
type Datum any

// BoolDatum represents a <boolean> datum (#t, #f, #true, #false).
type BoolDatum struct {
	Value bool
}

// IntDatum represents an integer <number> datum (<sign> <uinteger 10>).
type IntDatum struct {
	Value int64
}

// FloatDatum represents a floating-point <number> datum (<sign> <decimal 10> or <infnan>).
type FloatDatum struct {
	Value float64
}

// StringDatum represents a <string> datum.
type StringDatum struct {
	Value string
}

// SymbolDatum represents a <symbol> (<identifier>) datum.
type SymbolDatum struct {
	Value string
}

// ListDatum represents a <list> ((<datum>*)) compound datum.
type ListDatum struct {
	Elements []Datum
}
