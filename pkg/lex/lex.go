// Package lex provides lexer for slicep language.
package lex

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

//go:generate stringer -type=TokenType
type TokenType uint

const (
	TokenTypeIdent = TokenType(iota)
	TokenTypeNumber
	TokenTypeLparen
	TokenTypeRparen
	TokenEOF
)

const eof = rune(0)

type Token struct {
	Type  TokenType
	Value string
}

type lexer struct {
	name    string
	reader  *bufio.Reader
	tokchan chan Token
	builder strings.Builder
}

// Lex create new lexer
func Lex(name string, input io.Reader) chan Token {
	l := &lexer{
		name:    name,
		reader:  bufio.NewReader(input),
		tokchan: make(chan Token),
	}
	go l.run()

	return l.tokchan
}

type stateFn func(*lexer) stateFn

func (l *lexer) run() {
	for state := lexDispatch; state != nil; {
		state = state(l)
	}
	close(l.tokchan)
}

func (l *lexer) emit(t TokenType) {
	l.tokchan <- Token{t, l.builder.String()}
	l.builder.Reset()
}

func (l *lexer) next() rune {
	r, _, err := l.reader.ReadRune()
	if err == io.EOF {
		return eof
	}
	if err != nil {
		// NOTE: This is a simplified error handling.
		// A real lexer should emit an error token.
		panic(err)
	}

	return r
}

func (l *lexer) backup() {
	if err := l.reader.UnreadRune(); err != nil {
		panic(err)
	}
}

func (l *lexer) peek() rune {
	peeked := l.next()
	if peeked != eof {
		l.backup()
	}

	return peeked
}

func lexDispatch(l *lexer) stateFn {
	r := l.next()

	// To comply with the r7rs-small standard, handle Unicode
	// whitespace correctly.
	if unicode.IsSpace(r) {
		// Reset the builder just in case, and skip the
		// whitespace.
		l.builder.Reset()

		return lexDispatch
	} else if unicode.IsDigit(r) || r == '-' || r == '+' {
		l.backup()
		return lexReal10
	}

	switch r {
	case eof:
		// End of input.
		l.emit(TokenEOF)

		// Stop the state machine.
		return nil
	case '(':
		l.builder.WriteRune(r)
		l.emit(TokenTypeLparen)
	case ')':
		l.builder.WriteRune(r)
		l.emit(TokenTypeRparen)
	default:
		// Unhandled characters. A real implementation would
		// transition to a state for numbers, identifiers,
		// etc. For now, we ignore them.
		l.builder.Reset()

		return lexDispatch
	}

	return lexDispatch
}

func lexReal10(l *lexer) stateFn {
	r := l.next()

	if r == '-' || r == '+' {
		l.builder.WriteRune(r)

		peeked := l.peek()
		if unicode.IsDigit(peeked) {
			return lexUinteger10
		} else {
			return lexInfnan
		}
	} else if unicode.IsDigit(r) {
		l.backup()

		return lexUinteger10
	} else {
		// This case should ideally not be reached if lexDispatch correctly
		// transitions to lexReal10 only for digits, '+' or '-'.
		panic(fmt.Sprintf("unexpected rune in lexReal10: %c", r))
	}
}

func lexUinteger10(l *lexer) stateFn {
	for {
		r := l.next()
		if unicode.IsDigit(r) {
			l.builder.WriteRune(r)
		} else {
			if r != eof {
				l.backup()
			}
			l.emit(TokenTypeNumber)

			return lexDispatch
		}
	}
}

func lexInfnan(l *lexer) stateFn {
	infnanStrLen := utf8.RuneCountInString("+inf.0")
	for utf8.RuneCountInString(l.builder.String()) < infnanStrLen {
		r := l.next()
		if r == eof {
			break
		}
		l.builder.WriteRune(r)
	}

	v := l.builder.String()
	// +/- is checked in lexReal10
	v = v[1:]
	if v != "inf.0" && v != "nan.0" {
		panic(fmt.Sprintf("invalid number token: %s", v))
	}
	l.emit(TokenTypeNumber)

	return lexDispatch
}
