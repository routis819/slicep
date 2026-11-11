package parse_test

import (
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/routis819/slicep/pkg/lex"
	"github.com/routis819/slicep/pkg/parse"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		want       []parse.Datum
		wantErr    bool
		wantErrSub string
	}{
		{
			name:  "Empty input",
			input: "   ; only comment\n",
			want:  []parse.Datum{},
		},
		{
			name:  "Integer literals",
			input: "123 -45 +67",
			want: []parse.Datum{
				&parse.IntDatum{Value: 123, Pos: lex.Pos{Line: 1, Column: 1, Offset: 0}},
				&parse.IntDatum{Value: -45, Pos: lex.Pos{Line: 1, Column: 5, Offset: 4}},
				&parse.IntDatum{Value: 67, Pos: lex.Pos{Line: 1, Column: 9, Offset: 8}},
			},
		},
		{
			name:  "Float literals",
			input: "1.25 -0.5 1e3 +inf.0 -inf.0",
			want: []parse.Datum{
				&parse.FloatDatum{Value: 1.25, Pos: lex.Pos{Line: 1, Column: 1, Offset: 0}},
				&parse.FloatDatum{Value: -0.5, Pos: lex.Pos{Line: 1, Column: 6, Offset: 5}},
				&parse.FloatDatum{Value: 1000, Pos: lex.Pos{Line: 1, Column: 11, Offset: 10}},
				&parse.FloatDatum{Value: math.Inf(1), Pos: lex.Pos{Line: 1, Column: 15, Offset: 14}},
				&parse.FloatDatum{Value: math.Inf(-1), Pos: lex.Pos{Line: 1, Column: 22, Offset: 21}},
			},
		},
		{
			name:  "Boolean literals",
			input: "#t #f #true #false",
			want: []parse.Datum{
				&parse.BoolDatum{Value: true, Pos: lex.Pos{Line: 1, Column: 1, Offset: 0}},
				&parse.BoolDatum{Value: false, Pos: lex.Pos{Line: 1, Column: 4, Offset: 3}},
				&parse.BoolDatum{Value: true, Pos: lex.Pos{Line: 1, Column: 7, Offset: 6}},
				&parse.BoolDatum{Value: false, Pos: lex.Pos{Line: 1, Column: 13, Offset: 12}},
			},
		},
		{
			name:  "String literals",
			input: `"hello" "a\nb\t\"\\" "line1\  ` + "\n" + `  line2"`,
			want: []parse.Datum{
				&parse.StringDatum{Value: "hello", Pos: lex.Pos{Line: 1, Column: 1, Offset: 0}},
				&parse.StringDatum{Value: "a\nb\t\"\\", Pos: lex.Pos{Line: 1, Column: 9, Offset: 8}},
				&parse.StringDatum{Value: "line1line2", Pos: lex.Pos{Line: 1, Column: 22, Offset: 21}},
			},
		},
		{
			name:  "Symbol literal",
			input: "foo",
			want: []parse.Datum{
				&parse.SymbolDatum{Value: "foo", Pos: lex.Pos{Line: 1, Column: 1, Offset: 0}},
			},
		},
		{
			name:  "Simple list",
			input: "(foo bar)",
			want: []parse.Datum{
				&parse.ListDatum{
					Pos: lex.Pos{Line: 1, Column: 1, Offset: 0},
					Elements: []parse.Datum{
						&parse.SymbolDatum{Value: "foo", Pos: lex.Pos{Line: 1, Column: 2, Offset: 1}},
						&parse.SymbolDatum{Value: "bar", Pos: lex.Pos{Line: 1, Column: 6, Offset: 5}},
					},
				},
			},
		},
		{
			name:  "Nested list and multiple top-level datums",
			input: "(define x (+ 1 2.5)) (set! x (* x 3))",
			want: []parse.Datum{
				&parse.ListDatum{
					Pos: lex.Pos{Line: 1, Column: 1, Offset: 0},
					Elements: []parse.Datum{
						&parse.SymbolDatum{Value: "define", Pos: lex.Pos{Line: 1, Column: 2, Offset: 1}},
						&parse.SymbolDatum{Value: "x", Pos: lex.Pos{Line: 1, Column: 9, Offset: 8}},
						&parse.ListDatum{
							Pos: lex.Pos{Line: 1, Column: 11, Offset: 10},
							Elements: []parse.Datum{
								&parse.SymbolDatum{Value: "+", Pos: lex.Pos{Line: 1, Column: 12, Offset: 11}},
								&parse.IntDatum{Value: 1, Pos: lex.Pos{Line: 1, Column: 14, Offset: 13}},
								&parse.FloatDatum{Value: 2.5, Pos: lex.Pos{Line: 1, Column: 16, Offset: 15}},
							},
						},
					},
				},
				&parse.ListDatum{
					Pos: lex.Pos{Line: 1, Column: 22, Offset: 21},
					Elements: []parse.Datum{
						&parse.SymbolDatum{Value: "set!", Pos: lex.Pos{Line: 1, Column: 23, Offset: 22}},
						&parse.SymbolDatum{Value: "x", Pos: lex.Pos{Line: 1, Column: 28, Offset: 27}},
						&parse.ListDatum{
							Pos: lex.Pos{Line: 1, Column: 30, Offset: 29},
							Elements: []parse.Datum{
								&parse.SymbolDatum{Value: "*", Pos: lex.Pos{Line: 1, Column: 31, Offset: 30}},
								&parse.SymbolDatum{Value: "x", Pos: lex.Pos{Line: 1, Column: 33, Offset: 32}},
								&parse.IntDatum{Value: 3, Pos: lex.Pos{Line: 1, Column: 35, Offset: 34}},
							},
						},
					},
				},
			},
		},
		{
			name:       "Error: Unclosed parenthesis",
			input:      "(+ 1 2",
			want:       nil,
			wantErr:    true,
			wantErrSub: "1:7: syntax error: unexpected EOF (unclosed '(' at 1:1)",
		},
		{
			name:       "Error: Unexpected closing parenthesis",
			input:      ")",
			want:       nil,
			wantErr:    true,
			wantErrSub: `1:1: syntax error near ")"`,
		},
		{
			name:       "Error: Lexical error propagated",
			input:      "123abc",
			want:       nil,
			wantErr:    true,
			wantErrSub: `lexical error: input:1:1: invalid token "123abc"`,
		},
		{
			name:       "Error: Integer overflow without duplicated token",
			input:      "99999999999999999999999999",
			want:       nil,
			wantErr:    true,
			wantErrSub: `1:1: invalid integer "99999999999999999999999999"`,
		},
		{
			name:       "Error: Float overflow",
			input:      "1e9999",
			want:       nil,
			wantErr:    true,
			wantErrSub: `1:1: invalid float "1e9999"`,
		},
		{
			name:       "Error: Nested unclosed parenthesis reports innermost open position",
			input:      "(+ (* 1 2",
			want:       nil,
			wantErr:    true,
			wantErrSub: "1:10: syntax error: unexpected EOF (unclosed '(' at 1:4)",
		},
		{
			name:       "Error: Deeply nested parentheses exceed maxNestingDepth",
			input:      strings.Repeat("(", 257) + strings.Repeat(")", 257),
			want:       nil,
			wantErr:    true,
			wantErrSub: "1:257: syntax error: maximum nesting depth (256) exceeded",
		},
		{
			name:       "Error: Malicious unclosed deep nesting fails safely without panic",
			input:      strings.Repeat("(", 10000),
			want:       nil,
			wantErr:    true,
			wantErrSub: "1:257: syntax error: maximum nesting depth (256) exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			got, err := parse.Parse(reader)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if tt.wantErrSub != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErrSub)) {
				t.Errorf("Parse() error = %v, want substring %q", err, tt.wantErrSub)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() got = %#v, want %#v", got, tt.want)
			}
		})
	}

	t.Run("Max nesting depth boundary (256) succeeds", func(t *testing.T) {
		input := strings.Repeat("(", 256) + "1" + strings.Repeat(")", 256)
		got, err := parse.Parse(strings.NewReader(input))
		if err != nil {
			t.Fatalf("Parse() unexpected error at maxNestingDepth: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("Parse() got %d top-level datums, want 1", len(got))
		}
	})
}
