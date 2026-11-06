package parse

import (
	"reflect"
	"strings"
	"testing"
)

func parseSingleDatum(t *testing.T, input string) Datum {
	t.Helper()

	datums, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("failed to parse %q: %v", input, err)
	}

	if len(datums) != 1 {
		t.Fatalf("expected 1 datum from %q, got %d", input, len(datums))
	}

	return datums[0]
}

func TestAnalyzeExpression_Success(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Node
	}{
		{
			name:  "Boolean true",
			input: "#t",
			want:  &BoolNode{Value: true},
		},
		{
			name:  "Boolean false",
			input: "#f",
			want:  &BoolNode{Value: false},
		},
		{
			name:  "Integer literal",
			input: "123",
			want:  &IntNode{Value: 123},
		},
		{
			name:  "Float literal",
			input: "3.14",
			want:  &FloatNode{Value: 3.14},
		},
		{
			name:  "String literal",
			input: `"hello world"`,
			want:  &StringNode{Value: "hello world"},
		},
		{
			name:  "Identifier",
			input: "foo",
			want:  &IdentifierNode{Name: "foo"},
		},
		{
			name:  "Identifier symbol operator",
			input: "+",
			want:  &IdentifierNode{Name: "+"},
		},
		{
			name:  "Procedure call no args",
			input: "(do-something)",
			want: &ProcedureCallNode{
				Operator: "do-something",
				Operands: []Node{},
			},
		},
		{
			name:  "Procedure call with args",
			input: "(+ 1 2 3)",
			want: &ProcedureCallNode{
				Operator: "+",
				Operands: []Node{
					&IntNode{Value: 1},
					&IntNode{Value: 2},
					&IntNode{Value: 3},
				},
			},
		},
		{
			name:  "Nested procedure call",
			input: "(+ (* 2 3) 4)",
			want: &ProcedureCallNode{
				Operator: "+",
				Operands: []Node{
					&ProcedureCallNode{
						Operator: "*",
						Operands: []Node{
							&IntNode{Value: 2},
							&IntNode{Value: 3},
						},
					},
					&IntNode{Value: 4},
				},
			},
		},
		{
			name:  "If two arguments",
			input: "(if #t 1)",
			want: &IfNode{
				Test:       &BoolNode{Value: true},
				Consequent: &IntNode{Value: 1},
				Alternate:  nil,
			},
		},
		{
			name:  "If three arguments",
			input: "(if #f 1 2)",
			want: &IfNode{
				Test:       &BoolNode{Value: false},
				Consequent: &IntNode{Value: 1},
				Alternate:  &IntNode{Value: 2},
			},
		},
		{
			name:  "Set assignment",
			input: "(set! x (+ 1 2))",
			want: &SetNode{
				Name: "x",
				Value: &ProcedureCallNode{
					Operator: "+",
					Operands: []Node{
						&IntNode{Value: 1},
						&IntNode{Value: 2},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			datum := parseSingleDatum(t, tt.input)

			got, err := analyzeExpression(datum)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("analyzeExpression(%s) = %#v, want %#v", tt.input, got, tt.want)
			}
		})
	}
}

func TestAnalyzeExpression_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "Empty list",
			input:   "()",
			wantErr: "empty list () is not a valid expression",
		},
		{
			name:    "Operator is number",
			input:   "(1 2 3)",
			wantErr: "operator must be an identifier",
		},
		{
			name:    "Operator is list",
			input:   "((+ 1 2) 3)",
			wantErr: "operator must be an identifier",
		},
		{
			name:    "Keyword define as expression",
			input:   "define",
			wantErr: `unexpected syntax keyword "define" as expression`,
		},
		{
			name:    "Keyword if as expression",
			input:   "if",
			wantErr: `unexpected syntax keyword "if" as expression`,
		},
		{
			name:    "Define in expression context",
			input:   "(define x 1)",
			wantErr: `"define" is not allowed in expression context`,
		},
		{
			name:    "Begin in expression context",
			input:   "(begin 1 2)",
			wantErr: `"begin" is not allowed in expression context`,
		},
		{
			name:    "If with 0 args",
			input:   "(if)",
			wantErr: "if: expected 2 or 3 arguments, got 0",
		},
		{
			name:    "If with 1 arg",
			input:   "(if #t)",
			wantErr: "if: expected 2 or 3 arguments, got 1",
		},
		{
			name:    "If with 4 args",
			input:   "(if #t 1 2 3)",
			wantErr: "if: expected 2 or 3 arguments, got 4",
		},
		{
			name:    "Set with 0 args",
			input:   "(set!)",
			wantErr: "set!: expected 2 arguments, got 0",
		},
		{
			name:    "Set with 1 arg",
			input:   "(set! x)",
			wantErr: "set!: expected 2 arguments, got 1",
		},
		{
			name:    "Set with 3 args",
			input:   "(set! x 1 2)",
			wantErr: "set!: expected 2 arguments, got 3",
		},
		{
			name:    "Set with non-identifier target",
			input:   "(set! 123 456)",
			wantErr: "set!: first argument must be an identifier",
		},
		{
			name:    "Set with keyword target",
			input:   "(set! if 1)",
			wantErr: `set!: cannot assign to syntax keyword "if"`,
		},
		{
			name:    "Procedure call invalid operand",
			input:   "(+ 1 ())",
			wantErr: "empty list () is not a valid expression",
		},
		{
			name:    "If invalid condition",
			input:   "(if () 1 2)",
			wantErr: "empty list () is not a valid expression",
		},
		{
			name:    "Set invalid value",
			input:   "(set! x ())",
			wantErr: "empty list () is not a valid expression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			datum := parseSingleDatum(t, tt.input)

			_, err := analyzeExpression(datum)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestAnalyzeExpression_NilDatum(t *testing.T) {
	_, err := analyzeExpression(nil)
	if err == nil {
		t.Fatal("expected error for nil datum, got nil")
	}
}
