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

const maxNestingDepth = 256

type parser struct {
	lexer *lex.Lexer
}

// Parse is the main entry point for the Stage 1 <datum> parser.
// It reads tokens from r and returns the resulting slice of top-level
// Datums or an error.
func Parse(r io.Reader) ([]Datum, error) {
	p := &parser{
		lexer: lex.New("input", r),
	}

	return p.parseTop()
}

func (p *parser) parseTop() ([]Datum, error) {
	datums := make([]Datum, 0)

	for {
		tok := p.lexer.Next()
		if tok.Type == lex.TokenEOF {
			return datums, nil
		}

		d, err := p.parseDatum(tok, 0)
		if err != nil {
			return nil, err
		}

		datums = append(datums, d)
	}
}

func (p *parser) parseDatum(tok lex.Token, depth int) (Datum, error) {
	switch tok.Type {
	case lex.TokenTypeBoolean:
		return &BoolDatum{
			Value: tok.Value == "#t" || tok.Value == "#true",
			Pos:   tok.Pos,
		}, nil
	case lex.TokenTypeNumber:
		return parseNumber(tok)
	case lex.TokenTypeString:
		s, err := unquoteSchemeString(tok.Value)
		if err != nil {
			return nil, fmt.Errorf("%s: invalid string: %w", tok.Pos, err)
		}

		return &StringDatum{
			Value: s,
			Pos:   tok.Pos,
		}, nil
	case lex.TokenTypeIdent:
		return &SymbolDatum{
			Value: tok.Value,
			Pos:   tok.Pos,
		}, nil
	case lex.TokenTypeLparen:
		return p.parseList(tok.Pos, depth+1)
	case lex.TokenTypeRparen:
		return nil, fmt.Errorf("%s: syntax error near %q", tok.Pos, tok.Value)
	case lex.TokenError:
		return nil, fmt.Errorf("lexical error: %s", tok.Value)
	case lex.TokenEOF:
		return nil, fmt.Errorf("%s: syntax error: unexpected EOF", tok.Pos)
	default:
		return nil, fmt.Errorf("%s: unexpected token %v", tok.Pos, tok.Type)
	}
}

func (p *parser) parseList(openPos lex.Pos, depth int) (Datum, error) {
	if depth > maxNestingDepth {
		return nil, fmt.Errorf("%s: syntax error: maximum nesting depth (%d) exceeded", openPos, maxNestingDepth)
	}

	elements := make([]Datum, 0)

	for {
		tok := p.lexer.Next()
		switch tok.Type {
		case lex.TokenTypeRparen:
			return &ListDatum{
				Elements: elements,
				Pos:      openPos,
			}, nil
		case lex.TokenEOF:
			return nil, fmt.Errorf("%s: syntax error: unexpected EOF (unclosed '(' at %s)", tok.Pos, openPos)
		default:
			elem, err := p.parseDatum(tok, depth)
			if err != nil {
				return nil, err
			}

			elements = append(elements, elem)
		}
	}
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
