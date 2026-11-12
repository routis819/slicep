package vm

import (
	"fmt"

	"github.com/routis819/slicep/pkg/lex"
	"github.com/routis819/slicep/pkg/parse"
)

// CompileError is a compile-time error with its source position.
type CompileError struct {
	Pos lex.Pos
	Msg string
}

func (e *CompileError) Error() string {
	if e.Pos.Line > 0 {
		return fmt.Sprintf("%s: %s", e.Pos, e.Msg)
	}

	return e.Msg
}

func errorfAt(pos lex.Pos, format string, args ...any) error {
	return &CompileError{Pos: pos, Msg: fmt.Sprintf(format, args...)}
}

type compiler struct {
	prog  *Program
	chunk *Chunk
	scope *scope
	// compiling is the stack of function names whose bodies are being
	// compiled, used to report recursion errors.
	compiling []string
}

// Compile compiles an analyzed program into bytecode. builtins are the
// procedures implemented in Go that the program may call; user definitions
// shadow builtins of the same name.
//
// Compile rejects any program that could fail to terminate: a function may
// only call builtins and functions defined before it, so self recursion and
// mutual recursion are compile-time errors, and functions cannot be used as
// values or reassigned with set!.
func Compile(program *parse.ProgramNode, builtins ...Builtin) (*Program, error) {
	if program == nil {
		return nil, errorfAt(lex.Pos{}, "nil program")
	}

	root := newScope(nil, nil)
	for i, b := range builtins {
		if b.Fn == nil {
			return nil, errorfAt(lex.Pos{}, "builtin %q has no implementation", b.Name)
		}
		root.symbols[b.Name] = symbol{kind: symBuiltin, index: i}
	}

	c := &compiler{
		prog:  &Program{Builtins: append([]Builtin(nil), builtins...)},
		scope: newScope(root, nil),
	}
	c.chunk = &c.prog.Main

	items := flattenBegin(program.Body)
	markPending(c.scope, items)
	if err := c.compileSequence(items, program.Pos); err != nil {
		return nil, err
	}
	c.chunk.emit(OpReturn, 0, 0, program.Pos)

	return c.prog, nil
}

// flattenBegin splices top-level (begin ...) forms into a single sequence.
func flattenBegin(nodes []parse.Node) []parse.Node {
	out := make([]parse.Node, 0, len(nodes))
	for _, n := range nodes {
		if b, ok := n.(*parse.BeginNode); ok {
			out = append(out, flattenBegin(b.Body)...)

			continue
		}
		out = append(out, n)
	}

	return out
}

func markPending(s *scope, nodes []parse.Node) {
	for _, n := range nodes {
		if f, ok := n.(*parse.DefineFuncNode); ok {
			s.pending[f.Name] = true
		}
	}
}

func isDefinition(n parse.Node) bool {
	if _, ok := n.(*parse.DefineVarNode); ok {
		return true
	}
	_, ok := n.(*parse.DefineFuncNode)

	return ok
}

// compileSequence compiles definitions and expressions so that exactly one
// value (the last expression's, or Void if the last item is a definition)
// remains on the stack.
func (c *compiler) compileSequence(items []parse.Node, pos lex.Pos) error {
	for i, n := range items {
		last := i == len(items)-1
		if isDefinition(n) {
			if err := c.compileDefinition(n); err != nil {
				return err
			}
			if last {
				c.chunk.emit(OpVoid, 0, 0, pos)
			}

			continue
		}

		if err := c.compileExpr(n); err != nil {
			return err
		}
		if !last {
			c.chunk.emit(OpPop, 0, 0, pos)
		}
	}

	return nil
}

func (c *compiler) compileDefinition(n parse.Node) error {
	switch d := n.(type) {
	case *parse.DefineVarNode:
		return c.compileDefineVar(d)
	case *parse.DefineFuncNode:
		return c.compileDefineFunc(d)
	case nil, *parse.ProgramNode, *parse.IntNode, *parse.FloatNode, *parse.StringNode,
		*parse.BoolNode, *parse.IdentifierNode, *parse.BeginNode, *parse.SetNode,
		*parse.IfNode, *parse.ProcedureCallNode:
	}

	return errorfAt(lex.Pos{}, "internal error: %T is not a definition", n)
}

func (c *compiler) checkRedefinition(name string, pos lex.Pos, isVar bool) error {
	sym, ok := c.scope.symbols[name]
	if !ok {
		return nil
	}
	if isVar && sym.kind == symGlobal {
		// Top-level redefinition of a variable acts as assignment (R7RS 5.3.1).
		return nil
	}

	return errorfAt(pos, "define: %q is already defined in this scope", name)
}

func (c *compiler) compileDefineVar(d *parse.DefineVarNode) error {
	if err := c.checkRedefinition(d.Name, d.Pos, true); err != nil {
		return err
	}

	// The value is compiled before the name is bound, so it cannot refer to
	// the variable being defined.
	if err := c.compileExpr(d.Value); err != nil {
		return err
	}

	if c.scope.isGlobal() {
		sym, ok := c.scope.symbols[d.Name]
		if !ok {
			sym = symbol{kind: symGlobal, index: len(c.prog.Globals)}
			c.prog.Globals = append(c.prog.Globals, d.Name)
			c.scope.symbols[d.Name] = sym
		}
		c.chunk.emit(OpSetGlobal, sym.index, 0, d.Pos)

		return nil
	}

	idx := c.scope.allocLocal()
	c.scope.symbols[d.Name] = symbol{kind: symLocal, index: idx}
	c.chunk.emit(OpSetLocal, idx, 0, d.Pos)

	return nil
}

func (c *compiler) compileDefineFunc(d *parse.DefineFuncNode) error {
	if err := c.checkRedefinition(d.Name, d.Pos, false); err != nil {
		return err
	}

	fs := &funcState{name: d.Name, numLocals: len(d.Params)}
	fnScope := newScope(c.scope, fs)
	for i, p := range d.Params {
		fnScope.symbols[p] = symbol{kind: symLocal, index: i}
	}
	markPending(fnScope, d.Definitions)

	fn := &Function{Name: d.Name, Arity: len(d.Params)}
	outerChunk, outerScope := c.chunk, c.scope
	c.chunk, c.scope = &fn.Chunk, fnScope
	c.compiling = append(c.compiling, d.Name)

	err := c.compileBody(d)

	c.compiling = c.compiling[:len(c.compiling)-1]
	c.chunk, c.scope = outerChunk, outerScope
	if err != nil {
		return err
	}

	fn.NumLocals = fs.numLocals
	idx := len(c.prog.Functions)
	c.prog.Functions = append(c.prog.Functions, fn)
	// The function name becomes visible only now, after its body.
	c.scope.symbols[d.Name] = symbol{kind: symFunc, index: idx}
	delete(c.scope.pending, d.Name)

	return nil
}

func (c *compiler) compileBody(d *parse.DefineFuncNode) error {
	for _, def := range d.Definitions {
		if err := c.compileDefinition(def); err != nil {
			return err
		}
	}
	if len(d.Sequence) == 0 {
		return errorfAt(d.Pos, "function %q: body has no expression", d.Name)
	}
	if err := c.compileSequence(d.Sequence, d.Pos); err != nil {
		return err
	}
	c.chunk.emit(OpReturn, 0, 0, d.Pos)

	return nil
}

// unboundError reports a reference to a name that is not (yet) bound,
// distinguishing recursion and forward references to functions.
func (c *compiler) unboundError(name string, pos lex.Pos) error {
	for _, f := range c.compiling {
		if f == name {
			return errorfAt(pos, "recursive reference to function %q is not allowed", name)
		}
	}
	if c.scope.isPending(name) {
		return errorfAt(pos, "reference to function %q before its definition is not allowed (forward references could introduce mutual recursion)", name)
	}

	return errorfAt(pos, "undefined identifier %q", name)
}

// resolveVariable resolves name as a variable (not a function) slot.
func (c *compiler) resolveVariable(name string, pos lex.Pos, context string) (symbol, error) {
	sym, owner, ok := c.scope.resolve(name)
	if !ok {
		return symbol{}, c.unboundError(name, pos)
	}

	switch sym.kind {
	case symGlobal:
		return sym, nil
	case symLocal:
		if owner.fn != c.scope.fn {
			return symbol{}, errorfAt(pos, "%s: cannot access local variable %q of enclosing function %q (closures are not supported)", context, name, owner.fn.name)
		}

		return sym, nil
	case symFunc, symBuiltin:
		return symbol{}, errorfAt(pos, "%s: function %q cannot be used as a value", context, name)
	}

	return symbol{}, errorfAt(pos, "internal error: unknown symbol kind %d", sym.kind)
}

func (c *compiler) compileExpr(n parse.Node) error {
	switch v := n.(type) {
	case nil:
		return errorfAt(lex.Pos{}, "internal error: nil expression")
	case *parse.IntNode:
		c.emitConstant(Int(v.Value), v.Pos)
	case *parse.FloatNode:
		c.emitConstant(Float(v.Value), v.Pos)
	case *parse.StringNode:
		c.emitConstant(String(v.Value), v.Pos)
	case *parse.BoolNode:
		c.emitConstant(Bool(v.Value), v.Pos)
	case *parse.IdentifierNode:
		return c.compileIdentifier(v)
	case *parse.SetNode:
		return c.compileSet(v)
	case *parse.IfNode:
		return c.compileIf(v)
	case *parse.ProcedureCallNode:
		return c.compileCall(v)
	case *parse.ProgramNode:
		return errorfAt(v.Pos, "program is not allowed in expression context")
	case *parse.BeginNode:
		return errorfAt(v.Pos, "begin is not allowed in expression context")
	case *parse.DefineVarNode:
		return errorfAt(v.Pos, "define is not allowed in expression context")
	case *parse.DefineFuncNode:
		return errorfAt(v.Pos, "define is not allowed in expression context")
	}

	return nil
}

func (c *compiler) emitConstant(v Value, pos lex.Pos) {
	c.chunk.emit(OpConstant, c.chunk.addConstant(v), 0, pos)
}

func (c *compiler) compileIdentifier(v *parse.IdentifierNode) error {
	sym, err := c.resolveVariable(v.Name, v.Pos, "reference")
	if err != nil {
		return err
	}

	if sym.kind == symGlobal {
		c.chunk.emit(OpGetGlobal, sym.index, 0, v.Pos)
	} else {
		c.chunk.emit(OpGetLocal, sym.index, 0, v.Pos)
	}

	return nil
}

func (c *compiler) compileSet(v *parse.SetNode) error {
	sym, err := c.resolveVariable(v.Name, v.Pos, "set!")
	if err != nil {
		return err
	}

	if err := c.compileExpr(v.Value); err != nil {
		return err
	}

	if sym.kind == symGlobal {
		c.chunk.emit(OpSetGlobal, sym.index, 0, v.Pos)
	} else {
		c.chunk.emit(OpSetLocal, sym.index, 0, v.Pos)
	}
	c.chunk.emit(OpVoid, 0, 0, v.Pos)

	return nil
}

func (c *compiler) compileIf(v *parse.IfNode) error {
	if err := c.compileExpr(v.Test); err != nil {
		return err
	}
	jumpIfFalse := c.chunk.emit(OpJumpIfFalse, -1, 0, v.Pos)

	if err := c.compileExpr(v.Consequent); err != nil {
		return err
	}
	jumpEnd := c.chunk.emit(OpJump, -1, 0, v.Pos)

	c.chunk.Code[jumpIfFalse].A = len(c.chunk.Code)
	if v.Alternate == nil {
		c.chunk.emit(OpVoid, 0, 0, v.Pos)
	} else if err := c.compileExpr(v.Alternate); err != nil {
		return err
	}
	c.chunk.Code[jumpEnd].A = len(c.chunk.Code)

	return nil
}

func (c *compiler) compileCall(v *parse.ProcedureCallNode) error {
	sym, _, ok := c.scope.resolve(v.Operator)
	if !ok {
		return c.unboundError(v.Operator, v.Pos)
	}

	argc := len(v.Operands)
	var op Opcode
	switch sym.kind {
	case symFunc:
		fn := c.prog.Functions[sym.index]
		if fn.Arity != argc {
			return errorfAt(v.Pos, "function %q expects %d argument(s), got %d", v.Operator, fn.Arity, argc)
		}
		op = OpCall
	case symBuiltin:
		b := &c.prog.Builtins[sym.index]
		if !b.acceptsArgc(argc) {
			return errorfAt(v.Pos, "builtin %q: wrong number of arguments: %d", v.Operator, argc)
		}
		op = OpCallBuiltin
	case symGlobal, symLocal:
		return errorfAt(v.Pos, "%q is a variable, not a function", v.Operator)
	}

	for _, operand := range v.Operands {
		if err := c.compileExpr(operand); err != nil {
			return err
		}
	}
	c.chunk.emit(op, sym.index, argc, v.Pos)

	return nil
}
