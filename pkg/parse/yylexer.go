package parse

import (
	"fmt"
	"io"
	"strconv"

	"github.com/routis819/slicep/pkg/lex"
)

// currentLexer holds the state of the lexer being used by the parser.
// It's a package-level variable to allow semantic actions in the .y file
// to communicate back to the lexer (e.g., to store the final result).
var currentLexer *yyLexerImpl

// yyLexerImpl is a bridge between the channel-based lexer (lex.Lex)
// and the interface expected by the goyacc-generated parser (yyLexer).
type yyLexerImpl struct {
	tokchan <-chan lex.Token
	lastTok lex.Token // To store the last token for error reporting
	result  Node      // To store the final parsed AST
	err     error     // To store any parsing error
}

// NewYyLexer creates a new lexer that is compliant with the yyLexer interface.
func NewYyLexer(r io.Reader) *yyLexerImpl {
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
	case lex.TokenTypeNumber:
		// The grammar currently expects UINTEGER10.
		// This is a simplification; a real implementation would handle various number types.
		u, err := strconv.ParseUint(tok.Value, 10, 64)
		if err != nil {
			l.Error(fmt.Sprintf("invalid number: %s", tok.Value))
			return 0 // Return EOF on error
		}
		lval.uintval = uint(u)
		return UINTEGER10
	case lex.TokenTypeLparen:
		return LPAREN
	case lex.TokenTypeRparen:
		return RPAREN
	case lex.TokenEOF:
		return 0 // 0 is EOF for goyacc
	default:
		return -1 // Invalid token
	}
}

// Error satisfies the yyLexer interface.
// It's called by the parser when it encounters a syntax error.
func (l *yyLexerImpl) Error(s string) {
	l.err = fmt.Errorf("syntax error near '%s': %s", l.lastTok.Value, s)
}

// Parse is the main entry point for the parser.
// It initializes the lexer, runs the goyacc-generated parser,
// and returns the resulting AST or an error.
func Parse(r io.Reader) (Node, error) {
	lexer := NewYyLexer(r)
	currentLexer = lexer

	if yyParse(lexer) != 0 {
		return nil, lexer.err
	}

	return lexer.result, nil
}
