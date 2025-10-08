package lex_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/routis819/slicep/pkg/lex"
)

func TestLex(t *testing.T) {
	tokchan := lex.Lex("test", strings.NewReader("()"))

	expectedTokens := []lex.Token{
		{Type: lex.TokenTypeLparen, Value: "("},
		{Type: lex.TokenTypeRparen, Value: ")"},
		{Type: lex.TokenEOF, Value: ""},
	}

	tokens := []lex.Token{}
	for tok := range tokchan {
		tokens = append(tokens, tok)
	}

	if !reflect.DeepEqual(expectedTokens, tokens) {
		t.Errorf("Expected %#v, got %#v", expectedTokens, tokens)
	}
}
