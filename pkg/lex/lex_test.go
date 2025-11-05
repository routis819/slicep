package lex_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/routis819/slicep/pkg/lex"
)

func TestLex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []lex.Token
	}{
		{
			name:  "Parens",
			input: "()",
			want: []lex.Token{
				{Type: lex.TokenTypeLparen, Value: "("},
				{Type: lex.TokenTypeRparen, Value: ")"},
				{Type: lex.TokenEOF, Value: ""},
			},
		},
		{
			name:  "Numbers",
			input: "123 45",
			want: []lex.Token{
				{Type: lex.TokenTypeNumber, Value: "123"},
				{Type: lex.TokenTypeNumber, Value: "45"},
				{Type: lex.TokenEOF, Value: ""},
			},
		},
		{
			name:  "Parens and Numbers",
			input: "(123 45)",
			want: []lex.Token{
				{Type: lex.TokenTypeLparen, Value: "("},
				{Type: lex.TokenTypeNumber, Value: "123"},
				{Type: lex.TokenTypeNumber, Value: "45"},
				{Type: lex.TokenTypeRparen, Value: ")"},
				{Type: lex.TokenEOF, Value: ""},
			},
		},
		{
			name:  "Signed Numbers",
			input: "+123 -45",
			want: []lex.Token{
				{Type: lex.TokenTypeNumber, Value: "+123"},
				{Type: lex.TokenTypeNumber, Value: "-45"},
				{Type: lex.TokenEOF, Value: ""},
			},
		},
		{
			name:  "Parens and Signed Numbers",
			input: "(+123 -45)",
			want: []lex.Token{
				{Type: lex.TokenTypeLparen, Value: "("},
				{Type: lex.TokenTypeNumber, Value: "+123"},
				{Type: lex.TokenTypeNumber, Value: "-45"},
				{Type: lex.TokenTypeRparen, Value: ")"},
				{Type: lex.TokenEOF, Value: ""},
			},
		},
		{
			name:  "Infnan values",
			input: "+inf.0 -inf.0 +nan.0 -nan.0",
			want: []lex.Token{
				{Type: lex.TokenTypeNumber, Value: "+inf.0"},
				{Type: lex.TokenTypeNumber, Value: "-inf.0"},
				{Type: lex.TokenTypeNumber, Value: "+nan.0"},
				{Type: lex.TokenTypeNumber, Value: "-nan.0"},
				{Type: lex.TokenEOF, Value: ""},
			},
		},
		{
			name:  "Parens and Infnan values",
			input: "(+inf.0 -nan.0)",
			want: []lex.Token{
				{Type: lex.TokenTypeLparen, Value: "("},
				{Type: lex.TokenTypeNumber, Value: "+inf.0"},
				{Type: lex.TokenTypeNumber, Value: "-nan.0"},
				{Type: lex.TokenTypeRparen, Value: ")"},
				{Type: lex.TokenEOF, Value: ""},
			},
		},
		{
			name:  "Decimal Numbers",
			input: "1.23 .45 1.e2 1.2e+3 1e-4",
			want: []lex.Token{
				{Type: lex.TokenTypeNumber, Value: "1.23"},
				{Type: lex.TokenTypeNumber, Value: ".45"},
				{Type: lex.TokenTypeNumber, Value: "1.e2"},
				{Type: lex.TokenTypeNumber, Value: "1.2e+3"},
				{Type: lex.TokenTypeNumber, Value: "1e-4"},
				{Type: lex.TokenEOF, Value: ""},
			},
		},
		{
			name:  "Signed Decimal Numbers",
			input: "+1.23 -0.5",
			want: []lex.Token{
				{Type: lex.TokenTypeNumber, Value: "+1.23"},
				{Type: lex.TokenTypeNumber, Value: "-0.5"},
				{Type: lex.TokenEOF, Value: ""},
			},
		},
		{
			name:  "Parens and Decimal Numbers",
			input: "(-1.23 +4.5e-2)",
			want: []lex.Token{
				{Type: lex.TokenTypeLparen, Value: "("},
				{Type: lex.TokenTypeNumber, Value: "-1.23"},
				{Type: lex.TokenTypeNumber, Value: "+4.5e-2"},
				{Type: lex.TokenTypeRparen, Value: ")"},
				{Type: lex.TokenEOF, Value: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokchan := lex.Lex(tt.name, strings.NewReader(tt.input))

			tokens := []lex.Token{}
			for tok := range tokchan {
				tokens = append(tokens, tok)
			}

			if !reflect.DeepEqual(tt.want, tokens) {
				t.Errorf("Expected %#v, got %#v", tt.want, tokens)
			}
		})
	}
}
