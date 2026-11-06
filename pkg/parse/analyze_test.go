package parse

import (
	"reflect"
	"strings"
	"testing"
)

func parseDatums(t *testing.T, input string) []Datum {
	t.Helper()

	datums, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("failed to parse %q: %v", input, err)
	}

	return datums
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
			datums := parseDatums(t, tt.input)

			got, err := Analyze(datums)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got.Body) != 1 || !reflect.DeepEqual(got.Body[0], tt.want) {
				t.Errorf("Analyze(%s) body = %#v, want [%#v]", tt.input, got.Body, tt.want)
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
			datums := parseDatums(t, tt.input)

			_, err := Analyze(datums)
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

func TestAnalyze_Success(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  *ProgramNode
	}{
		{
			name:  "Variable definition",
			input: "(define x 42)",
			want: &ProgramNode{
				Body: []Node{
					&DefineVarNode{
						Name:  "x",
						Value: &IntNode{Value: 42},
					},
				},
			},
		},
		{
			name:  "Simple function definition",
			input: "(define (square x) (* x x))",
			want: &ProgramNode{
				Body: []Node{
					&DefineFuncNode{
						Name:        "square",
						Params:      []string{"x"},
						Definitions: []Node{},
						Sequence: []Node{
							&ProcedureCallNode{
								Operator: "*",
								Operands: []Node{
									&IdentifierNode{Name: "x"},
									&IdentifierNode{Name: "x"},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "Function definition no parameters",
			input: "(define (answer) 42)",
			want: &ProgramNode{
				Body: []Node{
					&DefineFuncNode{
						Name:        "answer",
						Params:      []string{},
						Definitions: []Node{},
						Sequence: []Node{
							&IntNode{Value: 42},
						},
					},
				},
			},
		},
		{
			name:  "Function definition multiple parameters",
			input: "(define (add a b c) (+ a b c))",
			want: &ProgramNode{
				Body: []Node{
					&DefineFuncNode{
						Name:        "add",
						Params:      []string{"a", "b", "c"},
						Definitions: []Node{},
						Sequence: []Node{
							&ProcedureCallNode{
								Operator: "+",
								Operands: []Node{
									&IdentifierNode{Name: "a"},
									&IdentifierNode{Name: "b"},
									&IdentifierNode{Name: "c"},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Function definition with internal definitions and sequence",
			input: `
(define (foo a b)
  (define x (+ a 1))
  (define (bar y) (* y 2))
  (set! a (+ a x))
  (bar b))
`,
			want: &ProgramNode{
				Body: []Node{
					&DefineFuncNode{
						Name:   "foo",
						Params: []string{"a", "b"},
						Definitions: []Node{
							&DefineVarNode{
								Name: "x",
								Value: &ProcedureCallNode{
									Operator: "+",
									Operands: []Node{
										&IdentifierNode{Name: "a"},
										&IntNode{Value: 1},
									},
								},
							},
							&DefineFuncNode{
								Name:        "bar",
								Params:      []string{"y"},
								Definitions: []Node{},
								Sequence: []Node{
									&ProcedureCallNode{
										Operator: "*",
										Operands: []Node{
											&IdentifierNode{Name: "y"},
											&IntNode{Value: 2},
										},
									},
								},
							},
						},
						Sequence: []Node{
							&SetNode{
								Name: "a",
								Value: &ProcedureCallNode{
									Operator: "+",
									Operands: []Node{
										&IdentifierNode{Name: "a"},
										&IdentifierNode{Name: "x"},
									},
								},
							},
							&ProcedureCallNode{
								Operator: "bar",
								Operands: []Node{
									&IdentifierNode{Name: "b"},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Top-level begin with nested begin and definitions",
			input: `
(begin
  (define x 1)
  (begin
    (set! x 2)
    x))
`,
			want: &ProgramNode{
				Body: []Node{
					&BeginNode{
						Body: []Node{
							&DefineVarNode{
								Name:  "x",
								Value: &IntNode{Value: 1},
							},
							&BeginNode{
								Body: []Node{
									&SetNode{
										Name:  "x",
										Value: &IntNode{Value: 2},
									},
									&IdentifierNode{Name: "x"},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Mixed definitions and expressions at top level",
			input: `
(define x 10)
(+ x 5)
(set! x 20)
`,
			want: &ProgramNode{
				Body: []Node{
					&DefineVarNode{
						Name:  "x",
						Value: &IntNode{Value: 10},
					},
					&ProcedureCallNode{
						Operator: "+",
						Operands: []Node{
							&IdentifierNode{Name: "x"},
							&IntNode{Value: 5},
						},
					},
					&SetNode{
						Name:  "x",
						Value: &IntNode{Value: 20},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			datums := parseDatums(t, tt.input)

			got, err := Analyze(datums)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Analyze() =\n%#v\nwant:\n%#v", got, tt.want)
			}
		})
	}
}

func TestAnalyze_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "Empty begin form",
			input:   "(begin)",
			wantErr: "empty begin form",
		},
		{
			name:    "Define with 0 arguments",
			input:   "(define)",
			wantErr: "define: expected at least 2 arguments, got 0",
		},
		{
			name:    "Define with 1 argument",
			input:   "(define x)",
			wantErr: "define: expected at least 2 arguments, got 1",
		},
		{
			name:    "Define variable with 2 value expressions",
			input:   "(define x 1 2)",
			wantErr: "define: variable definition expects 1 value expression, got 2",
		},
		{
			name:    "Define syntax keyword if",
			input:   "(define if 1)",
			wantErr: `define: cannot bind syntax keyword "if"`,
		},
		{
			name:    "Define syntax keyword define",
			input:   "(define define 1)",
			wantErr: `define: cannot bind syntax keyword "define"`,
		},
		{
			name:    "Define syntax keyword set!",
			input:   "(define set! 1)",
			wantErr: `define: cannot bind syntax keyword "set!"`,
		},
		{
			name:    "Define syntax keyword begin",
			input:   "(define begin 1)",
			wantErr: `define: cannot bind syntax keyword "begin"`,
		},
		{
			name:    "Define variable with invalid value",
			input:   "(define x ())",
			wantErr: "empty list () is not a valid expression",
		},
		{
			name:    "Define function empty signature",
			input:   "(define () 1)",
			wantErr: "define: function signature cannot be empty",
		},
		{
			name:    "Define function non-identifier name",
			input:   "(define (123) 1)",
			wantErr: "define: function name must be an identifier",
		},
		{
			name:    "Define function keyword as name",
			input:   "(define (if x) x)",
			wantErr: `define: cannot bind syntax keyword "if" as function name`,
		},
		{
			name:    "Define function non-identifier parameter",
			input:   "(define (f 123) 1)",
			wantErr: "define: parameter must be an identifier",
		},
		{
			name:    "Define function keyword parameter",
			input:   "(define (f set!) 1)",
			wantErr: `define: parameter cannot be syntax keyword "set!"`,
		},
		{
			name:    "Define function duplicate parameter",
			input:   "(define (f x x) 1)",
			wantErr: `define: duplicate parameter name "x"`,
		},
		{
			name:    "Define function duplicate parameter among multiple",
			input:   "(define (f a b a) 1)",
			wantErr: `define: duplicate parameter name "a"`,
		},
		{
			name:    "Define function only definitions no sequence",
			input:   "(define (f) (define x 1))",
			wantErr: "function body must end with at least one expression",
		},
		{
			name:    "Define function multiple definitions no sequence",
			input:   "(define (f) (define x 1) (define y 2))",
			wantErr: "function body must end with at least one expression",
		},
		{
			name:    "Define function definition after expression",
			input:   "(define (f) (+ 1 2) (define x 1))",
			wantErr: `"define" is not allowed in expression context`,
		},
		{
			name:    "Define function begin after expression",
			input:   "(define (f) (+ 1 2) (begin 3))",
			wantErr: `"begin" is not allowed in expression context`,
		},
		{
			name:    "Define function begin in definition position",
			input:   "(define (f) (begin 1) 2)",
			wantErr: `"begin" is not allowed in expression context`,
		},
		{
			name:    "Define invalid target integer literal",
			input:   "(define 123 456)",
			wantErr: "define: first argument must be an identifier or function signature list",
		},
		{
			name:    "Define invalid target boolean literal",
			input:   "(define #t 456)",
			wantErr: "define: first argument must be an identifier or function signature list",
		},
		{
			name:    "Define in procedure call operand",
			input:   "(+ 1 (define x 2))",
			wantErr: `"define" is not allowed in expression context`,
		},
		{
			name:    "Begin in procedure call operand",
			input:   "(+ 1 (begin 2 3))",
			wantErr: `"begin" is not allowed in expression context`,
		},
		{
			name:    "Begin in if test expression",
			input:   "(if (begin #t) 1 2)",
			wantErr: `"begin" is not allowed in expression context`,
		},
		{
			name:    "Begin in set value expression",
			input:   "(set! x (begin 1 2))",
			wantErr: `"begin" is not allowed in expression context`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			datums := parseDatums(t, tt.input)

			_, err := Analyze(datums)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestAnalyze_EmptyProgram(t *testing.T) {
	_, err := Analyze([]Datum{})
	if err == nil {
		t.Fatal("expected error for empty datums, got nil")
	}

	want := "empty program: expected at least one command or definition"
	if !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestAnalyzeDefinition_NilTarget(t *testing.T) {
	_, err := analyzeDefinition([]Datum{nil, &IntDatum{Value: 1}})
	if err == nil {
		t.Fatal("expected error for nil target in define, got nil")
	}
}

func TestParseProgram(t *testing.T) {
	t.Run("Success full program", func(t *testing.T) {
		src := `
;; Full scheme program
(define PI 3.14159)

(define (circle-area r)
  (* PI (* r r)))

(circle-area 10)
`
		prog, err := ParseProgram(strings.NewReader(src))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(prog.Body) != 3 {
			t.Fatalf("expected 3 body nodes, got %d", len(prog.Body))
		}

		defVar, ok := prog.Body[0].(*DefineVarNode)
		if !ok || defVar.Name != "PI" {
			t.Errorf("expected DefineVarNode with Name PI, got %#v", prog.Body[0])
		}

		defFunc, ok := prog.Body[1].(*DefineFuncNode)
		if !ok || defFunc.Name != "circle-area" || len(defFunc.Params) != 1 || defFunc.Params[0] != "r" {
			t.Errorf("expected DefineFuncNode for circle-area, got %#v", prog.Body[1])
		}

		call, ok := prog.Body[2].(*ProcedureCallNode)
		if !ok || call.Operator != "circle-area" || len(call.Operands) != 1 {
			t.Errorf("expected ProcedureCallNode for circle-area, got %#v", prog.Body[2])
		}
	})

	t.Run("Empty program error", func(t *testing.T) {
		_, err := ParseProgram(strings.NewReader(""))
		if err == nil {
			t.Fatal("expected error for empty reader, got nil")
		}

		if !strings.Contains(err.Error(), "empty program") {
			t.Errorf("error = %q, want containing 'empty program'", err.Error())
		}
	})

	t.Run("Comments only error", func(t *testing.T) {
		_, err := ParseProgram(strings.NewReader("; only comment\n"))
		if err == nil {
			t.Fatal("expected error for comments only, got nil")
		}

		if !strings.Contains(err.Error(), "empty program") {
			t.Errorf("error = %q, want containing 'empty program'", err.Error())
		}
	})

	t.Run("Parser syntax error", func(t *testing.T) {
		_, err := ParseProgram(strings.NewReader("(+ 1 2"))
		if err == nil {
			t.Fatal("expected parse error for unclosed paren, got nil")
		}
	})
}
