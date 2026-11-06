package parse

import (
	"errors"
	"fmt"
	"io"
)

// Analyze converts a slice of external representation Datums into an AST ProgramNode.
func Analyze(datums []Datum) (*ProgramNode, error) {
	if len(datums) == 0 {
		return nil, errors.New("empty program: expected at least one command or definition")
	}

	body := make([]Node, 0, len(datums))
	for _, d := range datums {
		node, err := analyzeCommandOrDefinition(d)
		if err != nil {
			return nil, err
		}

		body = append(body, node)
	}

	return &ProgramNode{Body: body}, nil
}

// ParseProgram parses the input token stream from r and transforms it into an AST ProgramNode.
func ParseProgram(r io.Reader) (*ProgramNode, error) {
	datums, err := Parse(r)
	if err != nil {
		return nil, err
	}

	return Analyze(datums)
}

// NOTE: Syntax keywords ("define", "set!", "if", "begin") are kept as string
// literals while the set is small (4 keywords in Milestone 1). Extract them
// into const declarations once the number of syntax keywords exceeds 8
// (e.g., when restoring R7RS-small forms such as lambda, quote, cond, let).
func isSyntaxKeyword(name string) bool {
	switch name {
	case "define", "set!", "if", "begin":
		return true
	default:
		return false
	}
}

func analyzeCommandOrDefinition(d Datum) (Node, error) {
	if list, ok := d.(*ListDatum); ok && len(list.Elements) > 0 {
		if head, ok := list.Elements[0].(*SymbolDatum); ok {
			switch head.Value {
			case "define":
				return analyzeDefinition(list.Elements[1:])
			case "begin":
				return analyzeBegin(list.Elements[1:])
			}
		}
	}

	return analyzeExpression(d)
}

func analyzeBegin(args []Datum) (Node, error) {
	if len(args) == 0 {
		return nil, errors.New("empty begin form")
	}

	body := make([]Node, 0, len(args))
	for _, arg := range args {
		node, err := analyzeCommandOrDefinition(arg)
		if err != nil {
			return nil, err
		}

		body = append(body, node)
	}

	return &BeginNode{Body: body}, nil
}

func isDefinitionDatum(d Datum) bool {
	if list, ok := d.(*ListDatum); ok && len(list.Elements) > 0 {
		if head, ok := list.Elements[0].(*SymbolDatum); ok && head.Value == "define" {
			return true
		}
	}

	return false
}

func analyzeDefinition(args []Datum) (Node, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("define: expected at least 2 arguments, got %d", len(args))
	}

	switch target := args[0].(type) {
	case nil:
		return nil, errors.New("define: nil datum")
	case *SymbolDatum:
		if len(args) != 2 {
			return nil, fmt.Errorf("define: variable definition expects 1 value expression, got %d", len(args)-1)
		}

		if isSyntaxKeyword(target.Value) {
			return nil, fmt.Errorf("define: cannot bind syntax keyword %q", target.Value)
		}

		val, err := analyzeExpression(args[1])
		if err != nil {
			return nil, err
		}

		return &DefineVarNode{
			Name:  target.Value,
			Value: val,
		}, nil

	case *ListDatum:
		if len(target.Elements) == 0 {
			return nil, errors.New("define: function signature cannot be empty")
		}

		nameSym, ok := target.Elements[0].(*SymbolDatum)
		if !ok {
			return nil, errors.New("define: function name must be an identifier")
		}

		if isSyntaxKeyword(nameSym.Value) {
			return nil, fmt.Errorf("define: cannot bind syntax keyword %q as function name", nameSym.Value)
		}

		params := make([]string, 0, len(target.Elements)-1)
		seen := make(map[string]bool, len(target.Elements)-1)
		for _, elem := range target.Elements[1:] {
			paramSym, ok := elem.(*SymbolDatum)
			if !ok {
				return nil, errors.New("define: parameter must be an identifier")
			}

			if isSyntaxKeyword(paramSym.Value) {
				return nil, fmt.Errorf("define: parameter cannot be syntax keyword %q", paramSym.Value)
			}

			if seen[paramSym.Value] {
				return nil, fmt.Errorf("define: duplicate parameter name %q", paramSym.Value)
			}

			seen[paramSym.Value] = true
			params = append(params, paramSym.Value)
		}

		defs, seq, err := analyzeBody(args[1:])
		if err != nil {
			return nil, err
		}

		return &DefineFuncNode{
			Name:        nameSym.Value,
			Params:      params,
			Definitions: defs,
			Sequence:    seq,
		}, nil

	case *BoolDatum, *IntDatum, *FloatDatum, *StringDatum:
		return nil, errors.New("define: first argument must be an identifier or function signature list")
	}

	return nil, errors.New("unreachable")
}

func analyzeBody(bodyDatums []Datum) ([]Node, []Node, error) {
	if len(bodyDatums) == 0 {
		return nil, nil, errors.New("empty function body: expected at least one expression")
	}

	defs := make([]Node, 0)
	i := 0
	for i < len(bodyDatums) {
		if !isDefinitionDatum(bodyDatums[i]) {
			break
		}

		defNode, err := analyzeCommandOrDefinition(bodyDatums[i])
		if err != nil {
			return nil, nil, err
		}

		defs = append(defs, defNode)
		i++
	}

	if i == len(bodyDatums) {
		return nil, nil, errors.New("function body must end with at least one expression")
	}

	seq := make([]Node, 0, len(bodyDatums)-i)
	for ; i < len(bodyDatums); i++ {
		exprNode, err := analyzeExpression(bodyDatums[i])
		if err != nil {
			return nil, nil, err
		}

		seq = append(seq, exprNode)
	}

	return defs, seq, nil
}

func analyzeExpression(d Datum) (Node, error) {
	switch v := d.(type) {
	case nil:
		return nil, errors.New("nil datum")
	case *BoolDatum:
		return &BoolNode{Value: v.Value}, nil
	case *IntDatum:
		return &IntNode{Value: v.Value}, nil
	case *FloatDatum:
		return &FloatNode{Value: v.Value}, nil
	case *StringDatum:
		return &StringNode{Value: v.Value}, nil
	case *SymbolDatum:
		if isSyntaxKeyword(v.Value) {
			return nil, fmt.Errorf("unexpected syntax keyword %q as expression", v.Value)
		}

		return &IdentifierNode{Name: v.Value}, nil
	case *ListDatum:
		return analyzeListExpression(v)
	}

	return nil, errors.New("unreachable")
}

func analyzeListExpression(list *ListDatum) (Node, error) {
	if len(list.Elements) == 0 {
		return nil, errors.New("empty list () is not a valid expression")
	}

	head, ok := list.Elements[0].(*SymbolDatum)
	if !ok {
		return nil, errors.New("operator must be an identifier")
	}

	switch head.Value {
	case "if":
		return analyzeIf(list.Elements[1:])
	case "set!":
		return analyzeSet(list.Elements[1:])
	case "define", "begin":
		return nil, fmt.Errorf("%q is not allowed in expression context", head.Value)
	default:
		return analyzeProcedureCall(head.Value, list.Elements[1:])
	}
}

func analyzeIf(args []Datum) (Node, error) {
	if len(args) != 2 && len(args) != 3 {
		return nil, fmt.Errorf("if: expected 2 or 3 arguments, got %d", len(args))
	}

	test, err := analyzeExpression(args[0])
	if err != nil {
		return nil, err
	}

	consequent, err := analyzeExpression(args[1])
	if err != nil {
		return nil, err
	}

	var alternate Node
	if len(args) == 3 {
		alt, err := analyzeExpression(args[2])
		if err != nil {
			return nil, err
		}

		alternate = alt
	}

	return &IfNode{
		Test:       test,
		Consequent: consequent,
		Alternate:  alternate,
	}, nil
}

func analyzeSet(args []Datum) (Node, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("set!: expected 2 arguments, got %d", len(args))
	}

	sym, ok := args[0].(*SymbolDatum)
	if !ok {
		return nil, errors.New("set!: first argument must be an identifier")
	}

	if isSyntaxKeyword(sym.Value) {
		return nil, fmt.Errorf("set!: cannot assign to syntax keyword %q", sym.Value)
	}

	val, err := analyzeExpression(args[1])
	if err != nil {
		return nil, err
	}

	return &SetNode{
		Name:  sym.Value,
		Value: val,
	}, nil
}

func analyzeProcedureCall(op string, args []Datum) (Node, error) {
	operands := make([]Node, 0, len(args))
	for _, arg := range args {
		node, err := analyzeExpression(arg)
		if err != nil {
			return nil, err
		}

		operands = append(operands, node)
	}

	return &ProcedureCallNode{
		Operator: op,
		Operands: operands,
	}, nil
}
