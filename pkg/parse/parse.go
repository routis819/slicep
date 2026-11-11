// Package parse provides the functionality of parsing slicep token
// stream.
package parse

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/routis819/slicep/pkg/lex"
)

//go:generate go run golang.org/x/tools/cmd/goyacc@v0.52.0 slicep.y

// Parse is the main entry point for the Stage 1 <datum> parser.
// It reads tokens from r and returns the resulting slice of top-level
// Datums or an error.
func Parse(r io.Reader) ([]Datum, error) {
	lexer := newYyLexer(r)

	if yyParse(lexer) != 0 || lexer.err != nil {
		return nil, lexer.err
	}

	return lexer.result, nil
}

func parseNumber(tok lex.Token) (Datum, error) {
	raw := tok.Value
	lower := strings.ToLower(raw)
	switch lower {
	case "+inf.0":
		return &FloatDatum{Value: math.Inf(1), Pos: tok.Pos}, nil
	case "-inf.0":
		return &FloatDatum{Value: math.Inf(-1), Pos: tok.Pos}, nil
	case "+nan.0", "-nan.0":
		return &FloatDatum{Value: math.NaN(), Pos: tok.Pos}, nil
	}

	if strings.ContainsAny(lower, ".e") {
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, fmt.Errorf("%s: invalid float %q: %w", tok.Pos, raw, err)
		}

		return &FloatDatum{Value: f, Pos: tok.Pos}, nil
	}

	i, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid integer %q: %w", tok.Pos, raw, err)
	}

	return &IntDatum{Value: i, Pos: tok.Pos}, nil
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
