package parse

// Node represents an AST node in slicep.
//
//goaugtype:sum *ProgramNode | *IntNode | *FloatNode | *StringNode | *BoolNode | *IdentifierNode | *BeginNode | *DefineVarNode | *DefineFuncNode | *SetNode | *IfNode | *ProcedureCallNode
type Node any

// ProgramNode represents <program> -> <command or definition>+.
type ProgramNode struct {
	Body []Node
}

// IntNode represents an integer <literal>.
type IntNode struct {
	Value int64
}

// FloatNode represents a floating-point <literal>.
type FloatNode struct {
	Value float64
}

// StringNode represents a string <literal>.
type StringNode struct {
	Value string
}

// BoolNode represents a boolean <literal>.
type BoolNode struct {
	Value bool
}

// IdentifierNode represents an <identifier> expression.
type IdentifierNode struct {
	Name string
}

// BeginNode represents (begin <command or definition>+).
type BeginNode struct {
	Body []Node
}

// DefineVarNode represents (define <identifier> <expression>).
type DefineVarNode struct {
	Name  string
	Value Node
}

// DefineFuncNode represents (define (<identifier> <def formals>) <body>).
// Body corresponds to <body> -> <definition>* <sequence>, where Definitions
// holds leading <definition>* (*DefineVarNode or *DefineFuncNode) and
// Sequence holds <command>* <expression> (at least 1 expression).
type DefineFuncNode struct {
	Name        string
	Params      []string
	Definitions []Node
	Sequence    []Node
}

// SetNode represents <assignment> -> (set! <identifier> <expression>).
type SetNode struct {
	Name  string
	Value Node
}

// IfNode represents <conditional> -> (if <test> <consequent> <alternate>).
// Alternate is nil when omitted: (if <test> <consequent>).
type IfNode struct {
	Test       Node
	Consequent Node
	Alternate  Node
}

// ProcedureCallNode represents <procedure call> -> (<operator> <operand>*).
// To guarantee static termination in Milestone 1, Operator is restricted to an identifier name.
type ProcedureCallNode struct {
	Operator string
	Operands []Node
}
