package parse

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/routis819/slicep/pkg/lex"
)

// yyLexerImpl is a bridge between the channel-based lexer (lex.Lex)
// and the interface expected by the goyacc-generated parser (yyLexer).
type yyLexerImpl struct {
	tokchan <-chan lex.Token
	lastTok lex.Token // To store the last token for error reporting
	result  []Datum   // To store the parsed top-level datums
	err     error     // To store any lexical or parsing error
}

// newYyLexer creates a new lexer that is compliant with the yyLexer interface.
func newYyLexer(r io.Reader) *yyLexerImpl {
	return &yyLexerImpl{
		tokchan: lex.Lex("input", r),
	}
}

// Lex satisfies the yyLexer interface.
// It reads the next token from the channel and maps it to the
// integer token types defined in the parser. It also populates the
// lval structure with the token's semantic value.
func (l *yyLexerImpl) Lex(lval *yySymType) int {
	tok := <-l.tokchan
	l.lastTok = tok

	switch tok.Type {
	case lex.TokenTypeIdent:
		lval.sval = tok.Value

		return IDENT
	case lex.TokenTypeBoolean:
		lval.bval = tok.Value == "#t" || tok.Value == "#true"

		return BOOLEAN
	case lex.TokenTypeNumber:
		return l.lexNumber(lval, tok.Value)
	case lex.TokenTypeString:
		s, err := unquoteSchemeString(tok.Value)
		if err != nil {
			l.Error(fmt.Sprintf("invalid string: %v", err))

			return 0
		}
		lval.sval = s

		return STRING
	case lex.TokenTypeLparen:
		return LPAREN
	case lex.TokenTypeRparen:
		return RPAREN
	case lex.TokenError:
		l.err = fmt.Errorf("lexical error: %s", tok.Value)

		return 0
	case lex.TokenEOF:
		return 0 // 0 is EOF for goyacc
	default:
		return -1 // Invalid token
	}
}

func (l *yyLexerImpl) lexNumber(lval *yySymType, raw string) int {
	lower := strings.ToLower(raw)
	switch lower {
	case "+inf.0":
		lval.fval = math.Inf(1)

		return FLOAT
	case "-inf.0":
		lval.fval = math.Inf(-1)

		return FLOAT
	case "+nan.0", "-nan.0":
		lval.fval = math.NaN()

		return FLOAT
	}

	if strings.ContainsAny(lower, ".e") {
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			l.Error(fmt.Sprintf("invalid float: %s", raw))

			return 0
		}
		lval.fval = f

		return FLOAT
	}

	i, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		l.Error(fmt.Sprintf("invalid integer: %s", raw))

		return 0
	}
	lval.ival = i

	return INTEGER
}

func unquoteSchemeString(raw string) (string, error) {
	if len(raw) < 2 || raw[0] != '"' || raw[len(raw)-1] != '"' {
		return "", fmt.Errorf("malformed string literal %s", raw)
	}

	inner := raw[1 : len(raw)-1]
	var b strings.Builder
	b.Grow(len(inner))

	for i := 0; i < len(inner); i++ {
		if inner[i] != '\\' {
			b.WriteByte(inner[i])

			continue
		}

		i++
		if i >= len(inner) {
			return "", fmt.Errorf("trailing backslash in %s", raw)
		}

		switch inner[i] {
		case 'a':
			b.WriteByte('\a')
		case 'b':
			b.WriteByte('\b')
		case 't':
			b.WriteByte('\t')
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		case '"':
			b.WriteByte('"')
		case '\\':
			b.WriteByte('\\')
		case ' ', '\t', '\r', '\n':
			// Line continuation: \<intraline whitespace>* <line ending> <intraline whitespace>*
			for i < len(inner) && (inner[i] == ' ' || inner[i] == '\t') {
				i++
			}
			if i < len(inner) && inner[i] == '\r' {
				i++
				if i < len(inner) && inner[i] == '\n' {
					i++
				}
			} else if i < len(inner) && inner[i] == '\n' {
				i++
			}
			for i < len(inner) && (inner[i] == ' ' || inner[i] == '\t') {
				i++
			}
			i-- // Compensate for loop increment
		default:
			return "", fmt.Errorf("unknown escape \\%c", inner[i])
		}
	}

	return b.String(), nil
}

// Error satisfies the yyLexer interface.
// It's called by the parser when it encounters a syntax error.
func (l *yyLexerImpl) Error(s string) {
	if l.err == nil {
		l.err = fmt.Errorf("syntax error near '%s': %s", l.lastTok.Value, s)
	}
}

// Parse is the main entry point for the Stage 1 <datum> parser.
// It initializes the lexer, runs the goyacc-generated parser,
// and returns the resulting slice of top-level Datums or an error.
func Parse(r io.Reader) ([]Datum, error) {
	lexer := newYyLexer(r)

	if yyParse(lexer) != 0 || lexer.err != nil {
		return nil, lexer.err
	}

	return lexer.result, nil
}
