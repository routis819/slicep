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