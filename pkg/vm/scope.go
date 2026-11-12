package vm

type symbolKind uint8

const (
	symGlobal symbolKind = iota
	symLocal
	symFunc
	symBuiltin
)

type symbol struct {
	kind  symbolKind
	index int
}

// funcState tracks the frame layout of the function being compiled.
type funcState struct {
	name      string
	numLocals int
}

// scope is a lexical symbol table. Names become visible only after their
// definition has been compiled, which statically forbids recursion.
//
// Scope chain: builtin scope -> global scope -> function scopes. fn is nil
// for the builtin and global scopes.
type scope struct {
	parent  *scope
	fn      *funcState
	symbols map[string]symbol
	// pending holds names of functions that will be defined later in this
	// scope, used only to produce clearer forward-reference errors.
	pending map[string]bool
}

func newScope(parent *scope, fn *funcState) *scope {
	return &scope{
		parent:  parent,
		fn:      fn,
		symbols: make(map[string]symbol),
		pending: make(map[string]bool),
	}
}

// resolve looks up name along the scope chain and returns the symbol together
// with the scope that defines it.
func (s *scope) resolve(name string) (symbol, *scope, bool) {
	for cur := s; cur != nil; cur = cur.parent {
		if sym, ok := cur.symbols[name]; ok {
			return sym, cur, true
		}
	}

	return symbol{}, nil, false
}

// isPending reports whether name is a function defined later in an
// enclosing scope.
func (s *scope) isPending(name string) bool {
	for cur := s; cur != nil; cur = cur.parent {
		if cur.pending[name] {
			return true
		}
	}

	return false
}

func (s *scope) isGlobal() bool {
	return s.fn == nil && s.parent != nil
}

func (s *scope) allocLocal() int {
	idx := s.fn.numLocals
	s.fn.numLocals++

	return idx
}
