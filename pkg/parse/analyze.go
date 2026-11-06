package parse

import (
	"errors"
	"fmt"
)

func isSyntaxKeyword(name string) bool {
	switch name {
	case "define", "set!", "if", "begin":
		return true
	default:
		return false
	}
}

func analyzeExpression(d Datum) (Node, error) {
	if d == nil {
		return nil, errors.New("nil datum")
	}

	switch v := d.(type) {
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
	default:
		return nil, fmt.Errorf("unknown datum type %T", d)
	}
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
