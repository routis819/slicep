package parse_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/routis819/slicep/pkg/parse"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    parse.Node
		wantErr bool
	}{
		{
			name:  "Number literal",
			input: "123",
			want:  &parse.NumberNode{Value: 123},
		},
		{
			name:  "Identifier literal",
			input: "foo",
			want:  &parse.IdentifierNode{Value: "foo"},
		},
		{
			name:  "Simple procedure call",
			input: "(foo bar)",
			want: &parse.ProcedureCallNode{
				Operator: &parse.IdentifierNode{Value: "foo"},
				Operands: []parse.Node{
					&parse.IdentifierNode{Value: "bar"},
				},
			},
		},
		{
			name:  "Procedure call with numbers",
			input: "(+ 1 2)",
			want: &parse.ProcedureCallNode{
				Operator: &parse.IdentifierNode{Value: "+"},
				Operands: []parse.Node{
					&parse.NumberNode{Value: 1},
					&parse.NumberNode{Value: 2},
				},
			},
		},
		{
			name:  "Nested procedure call",
			input: "(+ 1 (* 2 3))",
			want: &parse.ProcedureCallNode{
				Operator: &parse.IdentifierNode{Value: "+"},
				Operands: []parse.Node{
					&parse.NumberNode{Value: 1},
					&parse.ProcedureCallNode{
						Operator: &parse.IdentifierNode{Value: "*"},
						Operands: []parse.Node{
							&parse.NumberNode{Value: 2},
							&parse.NumberNode{Value: 3},
						},
					},
				},
			},
		},
		{
			name:    "Error: Unclosed parenthesis",
			input:   "(+ 1 2",
			want:    nil,
			wantErr: true,
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}
