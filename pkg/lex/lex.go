// Package lex provides lexer for slicep language.
package lex

import (
	"bufio"
	"io"
	"strings"
	"unicode"
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

func lexDispatch(l *lexer) stateFn {
	r := l.next()

	// To comply with the r7rs-small standard, handle Unicode
	// whitespace correctly.
	if unicode.IsSpace(r) {
		// Reset the builder just in case, and skip the
		// whitespace.
		l.builder.Reset()

		return lexDispatch
	} else if unicode.IsDigit(r) {
		l.backup()
		return lexUinteger10
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
