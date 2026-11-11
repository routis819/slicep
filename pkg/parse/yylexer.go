package parse

import (
	"fmt"
	"io"

	"github.com/routis819/slicep/pkg/lex"
)

// yyLexerImpl is a bridge between the synchronous lexer (lex.Lexer)
// and the interface expected by the goyacc-generated parser (yyLexer).
type yyLexerImpl struct {
	lexer      *lex.Lexer
	lastTok    lex.Token // To store the last token for error reporting
	openParens []lex.Pos // Stack of unclosed '(' positions
	result     []Datum   // To store the parsed top-level datums
	err        error     // To store any lexical or parsing error
}

func newYyLexer(r io.Reader) *yyLexerImpl {
	return &yyLexerImpl{
		lexer: lex.New("input", r),
	}
}

func (l *yyLexerImpl) Lex(lval *yySymType) int {
	tok := l.lexer.Next()
	l.lastTok = tok
	lval.pos = tok.Pos

	switch tok.Type {
	case lex.TokenTypeIdent:
		lval.sval = tok.Value

		return IDENT
	case lex.TokenTypeBoolean:
		lval.bval = tok.Value == "#t" || tok.Value == "#true"

		return BOOLEAN
	case lex.TokenTypeNumber:
		d, err := parseNumber(tok)
		if err != nil {
			l.err = err

			return 0
		}
		if iv, ok := d.(*IntDatum); ok {
			lval.ival = iv.Value

			return INTEGER
		}
		lval.fval = d.(*FloatDatum).Value

		return FLOAT
	case lex.TokenTypeString:
		s, err := unquoteSchemeString(tok.Value)
		if err != nil {
			l.err = fmt.Errorf("%s: invalid string: %w", tok.Pos, err)

			return 0
		}
		lval.sval = s

		return STRING
	case lex.TokenTypeLparen:
		l.openParens = append(l.openParens, tok.Pos)

		return LPAREN
	case lex.TokenTypeRparen:
		if len(l.openParens) > 0 {
			l.openParens = l.openParens[:len(l.openParens)-1]
		}

		return RPAREN
	case lex.TokenError:
		l.err = fmt.Errorf("lexical error: %s", tok.Value)

		return 0
	case lex.TokenEOF:
		return 0
	default:
		return -1
	}
}

func (l *yyLexerImpl) Error(s string) {
	if l.err != nil {
		return
	}

	if l.lastTok.Type == lex.TokenEOF {
		if len(l.openParens) > 0 {
			openPos := l.openParens[len(l.openParens)-1]
			l.err = fmt.Errorf("%s: %s: unexpected EOF (unclosed '(' at %s)", l.lastTok.Pos, s, openPos)

			return
		}

		l.err = fmt.Errorf("%s: %s: unexpected EOF", l.lastTok.Pos, s)

		return
	}

	l.err = fmt.Errorf("%s: %s near %q", l.lastTok.Pos, s, l.lastTok.Value)
}
