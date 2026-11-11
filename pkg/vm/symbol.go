package vm

import (
	"errors"
	"fmt"

	"github.com/routis819/slicep/pkg/lex"
)

type symbolKind uint8

const (
	symGlobalVar symbolKind = iota
	symLocalVar
	symUserFunc
	symBuiltin
)

type symbol struct {
	name     string
	kind     symbolKind
	index    uint16
	minArity int
	maxArity int
}

type symbolTable struct {
	outer       *symbolTable
	store       map[string]symbol
	isFuncScope bool
	numLocals   int
}

func newSymbolTable(outer *symbolTable, isFuncScope bool) *symbolTable {
	return &symbolTable{
		outer:       outer,
		store:       make(map[string]symbol),
		isFuncScope: isFuncScope,
	}
}

func (s *symbolTable) defineBuiltin(idx uint16, def BuiltinDef) {
	s.store[def.Name] = symbol{
		name:     def.Name,
		kind:     symBuiltin,
		index:    idx,
		minArity: def.MinArity,
		maxArity: def.MaxArity,
	}
}

func (s *symbolTable) defineLocal(name string) symbol {
	sym := symbol{
		name:  name,
		kind:  symLocalVar,
		index: uint16(s.numLocals),
	}
	s.numLocals++
	s.store[name] = sym

	return sym
}

func (s *symbolTable) defineUserFunc(pos lex.Pos, name string, idx uint16, arity int) (symbol, error) {
	if _, exists := s.store[name]; exists {
		return symbol{}, errorfAt(pos, "define: %q is already defined in this scope", name)
	}

	sym := symbol{
		name:     name,
		kind:     symUserFunc,
		index:    idx,
		minArity: arity,
		maxArity: arity,
	}
	s.store[name] = sym

	return sym, nil
}

func (s *symbolTable) resolve(pos lex.Pos, name string) (symbol, bool, error) {
	crossedFunc := false
	for cur := s; cur != nil; cur = cur.outer {
		if found, exists := cur.store[name]; exists {
			if found.kind == symLocalVar && crossedFunc {
				return symbol{}, false, errorfAt(pos, "capturing outer local variable %q is not supported", name)
			}

			return found, true, nil
		}

		if cur.isFuncScope {
			crossedFunc = true
		}
	}

	return symbol{}, false, nil
}

func (s *symbolTable) resolveVar(pos lex.Pos, name string, activeFuncs map[string]bool) (symbol, error) {
	sym, ok, err := s.resolve(pos, name)
	if err != nil {
		return symbol{}, err
	}

	if !ok {
		if activeFuncs[name] {
			return symbol{}, errorfAt(pos, "procedure %q cannot be used as a variable", name)
		}

		return symbol{}, errorfAt(pos, "undefined variable %q", name)
	}

	if sym.kind == symUserFunc || sym.kind == symBuiltin {
		return symbol{}, errorfAt(pos, "procedure %q cannot be used as a variable", name)
	}

	return sym, nil
}

func (s *symbolTable) resolveProc(pos lex.Pos, name string, argc int, activeFuncs map[string]bool) (symbol, error) {
	sym, ok, err := s.resolve(pos, name)
	if err != nil {
		return symbol{}, err
	}

	if !ok {
		if activeFuncs[name] {
			return symbol{}, errorfAt(pos, "recursive call to %q is not allowed", name)
		}

		return symbol{}, errorfAt(pos, "undefined procedure %q", name)
	}

	if sym.kind != symUserFunc && sym.kind != symBuiltin {
		return symbol{}, errorfAt(pos, "variable %q is not a procedure", name)
	}

	if argc < sym.minArity || (sym.maxArity >= 0 && argc > sym.maxArity) {
		return symbol{}, errorfAt(pos, "procedure %q: arity mismatch (got %d arguments)", name, argc)
	}

	return sym, nil
}

func errorfAt(pos lex.Pos, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	if pos.Line > 0 {
		return fmt.Errorf("%s: %s", pos, msg)
	}

	return errors.New(msg)
}
