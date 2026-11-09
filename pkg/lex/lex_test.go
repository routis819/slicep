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
				{Type: lex.TokenTypeLparen, Value: "(", Pos: lex.Pos{Line: 1, Column: 1, Offset: 0}},
				{Type: lex.TokenTypeRparen, Value: ")", Pos: lex.Pos{Line: 1, Column: 2, Offset: 1}},
				{Type: lex.TokenEOF, Value: "", Pos: lex.Pos{Line: 1, Column: 3, Offset: 2}},
			},
		},
		{
			name:  "Numbers",
			input: "123 45",
			want: []lex.Token{
				{Type: lex.TokenTypeNumber, Value: "123", Pos: lex.Pos{Line: 1, Column: 1, Offset: 0}},
				{Type: lex.TokenTypeNumber, Value: "45", Pos: lex.Pos{Line: 1, Column: 5, Offset: 4}},
				{Type: lex.TokenEOF, Value: "", Pos: lex.Pos{Line: 1, Column: 7, Offset: 6}},
			},
		},
		{
			name:  "Parens and Numbers",
			input: "(123 45)",
			want: []lex.Token{
				{Type: lex.TokenTypeLparen, Value: "(", Pos: lex.Pos{Line: 1, Column: 1, Offset: 0}},
				{Type: lex.TokenTypeNumber, Value: "123", Pos: lex.Pos{Line: 1, Column: 2, Offset: 1}},
				{Type: lex.TokenTypeNumber, Value: "45", Pos: lex.Pos{Line: 1, Column: 6, Offset: 5}},
				{Type: lex.TokenTypeRparen, Value: ")", Pos: lex.Pos{Line: 1, Column: 8, Offset: 7}},
				{Type: lex.TokenEOF, Value: "", Pos: lex.Pos{Line: 1, Column: 9, Offset: 8}},
			},
		},
		{
			name:  "Signed Numbers",
			input: "+123 -45",
			want: []lex.Token{
				{Type: lex.TokenTypeNumber, Value: "+123", Pos: lex.Pos{Line: 1, Column: 1, Offset: 0}},
				{Type: lex.TokenTypeNumber, Value: "-45", Pos: lex.Pos{Line: 1, Column: 6, Offset: 5}},
				{Type: lex.TokenEOF, Value: "", Pos: lex.Pos{Line: 1, Column: 9, Offset: 8}},
			},
		},
		{
			name:  "Parens and Signed Numbers",
			input: "(+123 -45)",
			want: []lex.Token{
				{Type: lex.TokenTypeLparen, Value: "(", Pos: lex.Pos{Line: 1, Column: 1, Offset: 0}},
				{Type: lex.TokenTypeNumber, Value: "+123", Pos: lex.Pos{Line: 1, Column: 2, Offset: 1}},
				{Type: lex.TokenTypeNumber, Value: "-45", Pos: lex.Pos{Line: 1, Column: 7, Offset: 6}},
				{Type: lex.TokenTypeRparen, Value: ")", Pos: lex.Pos{Line: 1, Column: 10, Offset: 9}},
				{Type: lex.TokenEOF, Value: "", Pos: lex.Pos{Line: 1, Column: 11, Offset: 10}},
			},
		},
		{
			name:  "Infnan values",
			input: "+inf.0 -inf.0 +nan.0 -nan.0",
			want: []lex.Token{
				{Type: lex.TokenTypeNumber, Value: "+inf.0", Pos: lex.Pos{Line: 1, Column: 1, Offset: 0}},
				{Type: lex.TokenTypeNumber, Value: "-inf.0", Pos: lex.Pos{Line: 1, Column: 8, Offset: 7}},
				{Type: lex.TokenTypeNumber, Value: "+nan.0", Pos: lex.Pos{Line: 1, Column: 15, Offset: 14}},
				{Type: lex.TokenTypeNumber, Value: "-nan.0", Pos: lex.Pos{Line: 1, Column: 22, Offset: 21}},
				{Type: lex.TokenEOF, Value: "", Pos: lex.Pos{Line: 1, Column: 28, Offset: 27}},
			},
		},
		{
			name:  "Parens and Infnan values",
			input: "(+inf.0 -nan.0)",
			want: []lex.Token{
				{Type: lex.TokenTypeLparen, Value: "(", Pos: lex.Pos{Line: 1, Column: 1, Offset: 0}},
				{Type: lex.TokenTypeNumber, Value: "+inf.0", Pos: lex.Pos{Line: 1, Column: 2, Offset: 1}},
				{Type: lex.TokenTypeNumber, Value: "-nan.0", Pos: lex.Pos{Line: 1, Column: 9, Offset: 8}},
				{Type: lex.TokenTypeRparen, Value: ")", Pos: lex.Pos{Line: 1, Column: 15, Offset: 14}},
				{Type: lex.TokenEOF, Value: "", Pos: lex.Pos{Line: 1, Column: 16, Offset: 15}},
			},
		},
		{
			name:  "Decimal Numbers",
			input: "1.23 .45 1.e2 1.2e+3 1e-4",
			want: []lex.Token{
				{Type: lex.TokenTypeNumber, Value: "1.23", Pos: lex.Pos{Line: 1, Column: 1, Offset: 0}},
				{Type: lex.TokenTypeNumber, Value: ".45", Pos: lex.Pos{Line: 1, Column: 6, Offset: 5}},
				{Type: lex.TokenTypeNumber, Value: "1.e2", Pos: lex.Pos{Line: 1, Column: 10, Offset: 9}},
				{Type: lex.TokenTypeNumber, Value: "1.2e+3", Pos: lex.Pos{Line: 1, Column: 15, Offset: 14}},
				{Type: lex.TokenTypeNumber, Value: "1e-4", Pos: lex.Pos{Line: 1, Column: 22, Offset: 21}},
				{Type: lex.TokenEOF, Value: "", Pos: lex.Pos{Line: 1, Column: 26, Offset: 25}},
			},
		},
		{
			name:  "Signed Decimal Numbers",
			input: "+1.23 -0.5",
			want: []lex.Token{
				{Type: lex.TokenTypeNumber, Value: "+1.23", Pos: lex.Pos{Line: 1, Column: 1, Offset: 0}},
				{Type: lex.TokenTypeNumber, Value: "-0.5", Pos: lex.Pos{Line: 1, Column: 7, Offset: 6}},
				{Type: lex.TokenEOF, Value: "", Pos: lex.Pos{Line: 1, Column: 11, Offset: 10}},
			},
		},
		{
			name:  "Parens and Decimal Numbers",
			input: "(-1.23 +4.5e-2)",
			want: []lex.Token{
				{Type: lex.TokenTypeLparen, Value: "(", Pos: lex.Pos{Line: 1, Column: 1, Offset: 0}},
				{Type: lex.TokenTypeNumber, Value: "-1.23", Pos: lex.Pos{Line: 1, Column: 2, Offset: 1}},
				{Type: lex.TokenTypeNumber, Value: "+4.5e-2", Pos: lex.Pos{Line: 1, Column: 8, Offset: 7}},
				{Type: lex.TokenTypeRparen, Value: ")", Pos: lex.Pos{Line: 1, Column: 15, Offset: 14}},
				{Type: lex.TokenEOF, Value: "", Pos: lex.Pos{Line: 1, Column: 16, Offset: 15}},
			},
		},
		{
			name:  "Booleans",
			input: "(#t #f #true #false)",
			want: []lex.Token{
				{Type: lex.TokenTypeLparen, Value: "(", Pos: lex.Pos{Line: 1, Column: 1, Offset: 0}},
				{Type: lex.TokenTypeBoolean, Value: "#t", Pos: lex.Pos{Line: 1, Column: 2, Offset: 1}},
				{Type: lex.TokenTypeBoolean, Value: "#f", Pos: lex.Pos{Line: 1, Column: 5, Offset: 4}},
				{Type: lex.TokenTypeBoolean, Value: "#true", Pos: lex.Pos{Line: 1, Column: 8, Offset: 7}},
				{Type: lex.TokenTypeBoolean, Value: "#false", Pos: lex.Pos{Line: 1, Column: 14, Offset: 13}},
				{Type: lex.TokenTypeRparen, Value: ")", Pos: lex.Pos{Line: 1, Column: 20, Offset: 19}},
				{Type: lex.TokenEOF, Value: "", Pos: lex.Pos{Line: 1, Column: 21, Offset: 20}},
			},
		},
		{
			name:  "Strings and Escapes",
			input: `"hello" "world\n\t\"\\" "line1\   ` + "\n" + `   line2" "데미지 공식: 🔥"`,
			want: []lex.Token{
				{Type: lex.TokenTypeString, Value: `"hello"`, Pos: lex.Pos{Line: 1, Column: 1, Offset: 0}},
				{Type: lex.TokenTypeString, Value: `"world\n\t\"\\"`, Pos: lex.Pos{Line: 1, Column: 9, Offset: 8}},
				{Type: lex.TokenTypeString, Value: "\"line1\\   \n   line2\"", Pos: lex.Pos{Line: 1, Column: 25, Offset: 24}},
				{Type: lex.TokenTypeString, Value: `"데미지 공식: 🔥"`, Pos: lex.Pos{Line: 2, Column: 11, Offset: 45}},
				{Type: lex.TokenEOF, Value: "", Pos: lex.Pos{Line: 2, Column: 22, Offset: 69}},
			},
		},
		{
			name:  "Identifiers and Peculiar Identifiers",
			input: "(define set! if + - * / <= >= -> ... +.. .foo foo-bar_1?)",
			want: []lex.Token{
				{Type: lex.TokenTypeLparen, Value: "(", Pos: lex.Pos{Line: 1, Column: 1, Offset: 0}},
				{Type: lex.TokenTypeIdent, Value: "define", Pos: lex.Pos{Line: 1, Column: 2, Offset: 1}},
				{Type: lex.TokenTypeIdent, Value: "set!", Pos: lex.Pos{Line: 1, Column: 9, Offset: 8}},
				{Type: lex.TokenTypeIdent, Value: "if", Pos: lex.Pos{Line: 1, Column: 14, Offset: 13}},
				{Type: lex.TokenTypeIdent, Value: "+", Pos: lex.Pos{Line: 1, Column: 17, Offset: 16}},
				{Type: lex.TokenTypeIdent, Value: "-", Pos: lex.Pos{Line: 1, Column: 19, Offset: 18}},
				{Type: lex.TokenTypeIdent, Value: "*", Pos: lex.Pos{Line: 1, Column: 21, Offset: 20}},
				{Type: lex.TokenTypeIdent, Value: "/", Pos: lex.Pos{Line: 1, Column: 23, Offset: 22}},
				{Type: lex.TokenTypeIdent, Value: "<=", Pos: lex.Pos{Line: 1, Column: 25, Offset: 24}},
				{Type: lex.TokenTypeIdent, Value: ">=", Pos: lex.Pos{Line: 1, Column: 28, Offset: 27}},
				{Type: lex.TokenTypeIdent, Value: "->", Pos: lex.Pos{Line: 1, Column: 31, Offset: 30}},
				{Type: lex.TokenTypeIdent, Value: "...", Pos: lex.Pos{Line: 1, Column: 34, Offset: 33}},
				{Type: lex.TokenTypeIdent, Value: "+..", Pos: lex.Pos{Line: 1, Column: 38, Offset: 37}},
				{Type: lex.TokenTypeIdent, Value: ".foo", Pos: lex.Pos{Line: 1, Column: 42, Offset: 41}},
				{Type: lex.TokenTypeIdent, Value: "foo-bar_1?", Pos: lex.Pos{Line: 1, Column: 47, Offset: 46}},
				{Type: lex.TokenTypeRparen, Value: ")", Pos: lex.Pos{Line: 1, Column: 57, Offset: 56}},
				{Type: lex.TokenEOF, Value: "", Pos: lex.Pos{Line: 1, Column: 58, Offset: 57}},
			},
		},
		{
			name:  "Comments and Nested Comments",
			input: "; line comment\n(+ 1 #| outer #| inner |# outer |# 2)",
			want: []lex.Token{
				{Type: lex.TokenTypeLparen, Value: "(", Pos: lex.Pos{Line: 2, Column: 1, Offset: 15}},
				{Type: lex.TokenTypeIdent, Value: "+", Pos: lex.Pos{Line: 2, Column: 2, Offset: 16}},
				{Type: lex.TokenTypeNumber, Value: "1", Pos: lex.Pos{Line: 2, Column: 4, Offset: 18}},
				{Type: lex.TokenTypeNumber, Value: "2", Pos: lex.Pos{Line: 2, Column: 36, Offset: 50}},
				{Type: lex.TokenTypeRparen, Value: ")", Pos: lex.Pos{Line: 2, Column: 37, Offset: 51}},
				{Type: lex.TokenEOF, Value: "", Pos: lex.Pos{Line: 2, Column: 38, Offset: 52}},
			},
		},
		{
			name:  "CRLF and CR Line Endings",
			input: "1\r\n2\r3",
			want: []lex.Token{
				{Type: lex.TokenTypeNumber, Value: "1", Pos: lex.Pos{Line: 1, Column: 1, Offset: 0}},
				{Type: lex.TokenTypeNumber, Value: "2", Pos: lex.Pos{Line: 2, Column: 1, Offset: 3}},
				{Type: lex.TokenTypeNumber, Value: "3", Pos: lex.Pos{Line: 3, Column: 1, Offset: 5}},
				{Type: lex.TokenEOF, Value: "", Pos: lex.Pos{Line: 3, Column: 2, Offset: 6}},
			},
		},
		{
			name:  "Closed Channel Zero Value is TokenEOF",
			input: "",
			want: []lex.Token{
				{Type: lex.TokenEOF, Value: "", Pos: lex.Pos{Line: 1, Column: 1, Offset: 0}},
			},
		},
		{
			name:  "Invalid Token Emits Full Token Text",
			input: "123abc",
			want: []lex.Token{
				{
					Type:  lex.TokenError,
					Value: `Invalid Token Emits Full Token Text:1:1: invalid token "123abc"`,
					Pos:   lex.Pos{Line: 1, Column: 1, Offset: 0},
				},
			},
		},
		{
			name:  "Invalid Sharp Token",
			input: "(#invalid)",
			want: []lex.Token{
				{Type: lex.TokenTypeLparen, Value: "(", Pos: lex.Pos{Line: 1, Column: 1, Offset: 0}},
				{
					Type:  lex.TokenError,
					Value: `Invalid Sharp Token:1:2: invalid token "#invalid"`,
					Pos:   lex.Pos{Line: 1, Column: 2, Offset: 1},
				},
			},
		},
		{
			name:  "Unexpected Single Character Rune",
			input: "\n  🔥",
			want: []lex.Token{
				{
					Type:  lex.TokenError,
					Value: "Unexpected Single Character Rune:2:3: unexpected character '🔥'",
					Pos:   lex.Pos{Line: 2, Column: 3, Offset: 3},
				},
			},
		},
		{
			name:  "Unclosed String Literal",
			input: `"hello`,
			want: []lex.Token{
				{
					Type:  lex.TokenError,
					Value: "Unclosed String Literal:1:1: unclosed string literal",
					Pos:   lex.Pos{Line: 1, Column: 1, Offset: 0},
				},
			},
		},
		{
			name:  "Invalid Escape Sequence in String Literal",
			input: `"hello\x"`,
			want: []lex.Token{
				{
					Type:  lex.TokenError,
					Value: "Invalid Escape Sequence in String Literal:1:1: invalid escape sequence in string literal",
					Pos:   lex.Pos{Line: 1, Column: 1, Offset: 0},
				},
			},
		},
		{
			name:  "Unclosed Nested Comment Emits TokenError",
			input: "#| unclosed",
			want: []lex.Token{
				{
					Type:  lex.TokenError,
					Value: "Unclosed Nested Comment Emits TokenError:1:1: unclosed nested comment",
					Pos:   lex.Pos{Line: 1, Column: 1, Offset: 0},
				},
			},
		},
	}

	if (lex.Token{}).Type != lex.TokenEOF {
		t.Fatalf("Expected zero-value Token.Type to be TokenEOF (0), got %v", (lex.Token{}).Type)
	}

	if got := (lex.Pos{}).String(); got != "" {
		t.Fatalf("Expected zero-value Pos.String() to be empty, got %q", got)
	}

	if got := (lex.Pos{Line: 3, Column: 7}).String(); got != "3:7" {
		t.Fatalf("Expected Pos{3, 7}.String() to be \"3:7\", got %q", got)
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
