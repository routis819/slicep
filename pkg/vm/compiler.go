package vm

import (
	"errors"
	"fmt"
	"math"

	"github.com/routis819/slicep/pkg/lex"
	"github.com/routis819/slicep/pkg/parse"
)

// CompileOption configures the bytecode compiler.
type CompileOption func(*compilerConfig)

type compilerConfig struct {
	builtins []BuiltinDef
}

// WithBuiltins registers built-in or host-bound procedures in the root scope.
func WithBuiltins(builtins []BuiltinDef) CompileOption {
	return func(cfg *compilerConfig) { cfg.builtins = append(cfg.builtins, builtins...) }
}

type compiler struct {
	prog        *Program
	scope       *symbolTable
	chunk       *Chunk
	activeFuncs map[string]bool
}

// Compile translates a parsed AST ProgramNode into an immutable bytecode Program.
func Compile(ast *parse.ProgramNode, opts ...CompileOption) (*Program, error) {
	if ast == nil || len(ast.Body) == 0 {
		return nil, errors.New("cannot compile empty program")
	}

	var cfg compilerConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	builtinScope := newSymbolTable(nil, false)
	for i, b := range cfg.builtins {
		builtinScope.defineBuiltin(uint16(i), b)
	}

	prog := &Program{GlobalIndex: make(map[string]uint16), Builtins: cfg.builtins}
	c := &compiler{
		prog:        prog,
		scope:       newSymbolTable(builtinScope, false),
		chunk:       &prog.Main,
		activeFuncs: make(map[string]bool),
	}

	stmts := flattenBegin(ast.Body)
	for i, stmt := range stmts {
		if err := c.compileStmt(stmt, i == len(stmts)-1); err != nil {
			return nil, err
		}
	}

	c.chunk.Emit(ast.Pos, OpReturn)

	return prog, nil
}

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

func (c *compiler) compileStmt(node parse.Node, isLast bool) error {
	switch n := node.(type) {
	case *parse.DefineVarNode:
		if err := c.compileExpr(n.Value); err != nil {
			return err
		}

		sym, err := c.defineVarSymbol(n.Pos, n.Name)
		if err != nil {
			return err
		}

		c.chunk.Emit(n.Pos, selectOp(sym.kind == symLocalVar, OpSetLocal, OpSetGlobal), int(sym.index))
		if isLast {
			c.chunk.Emit(n.Pos, OpVoid)
		}
	case *parse.DefineFuncNode:
		if err := c.compileDefineFunc(n); err != nil {
			return err
		}

		if isLast {
			c.chunk.Emit(n.Pos, OpVoid)
		}
	case nil, *parse.ProgramNode, *parse.BeginNode,
		*parse.IntNode, *parse.FloatNode, *parse.StringNode, *parse.BoolNode,
		*parse.IdentifierNode, *parse.SetNode, *parse.IfNode, *parse.ProcedureCallNode:
		if err := c.compileExpr(node); err != nil {
			return err
		}

		if !isLast {
			c.chunk.Emit(c.chunk.PosAt(len(c.chunk.Code)-1), OpPop)
		}
	}

	return nil
}

func (c *compiler) defineVarSymbol(pos lex.Pos, name string) (symbol, error) {
	if existing, exists := c.scope.store[name]; exists {
		if existing.kind != symGlobalVar && existing.kind != symLocalVar {
			return symbol{}, errorfAt(pos, "define: cannot redefine function %q as variable", name)
		}

		return existing, nil
	}

	if c.scope.isFuncScope {
		return c.scope.defineLocal(name), nil
	}

	idx := uint16(len(c.prog.GlobalNames))
	c.prog.GlobalNames = append(c.prog.GlobalNames, name)
	c.prog.GlobalIndex[name] = idx
	c.scope.store[name] = symbol{name: name, kind: symGlobalVar, index: idx}

	return c.scope.store[name], nil
}

func (c *compiler) compileDefineFunc(n *parse.DefineFuncNode) error {
	if len(n.Params) > math.MaxUint8 {
		return errorfAt(n.Pos, "define: function %q exceeds maximum parameter count (%d)", n.Name, math.MaxUint8)
	}

	if _, exists := c.scope.store[n.Name]; exists {
		return errorfAt(n.Pos, "define: %q is already defined in this scope", n.Name)
	}

	fnScope := newSymbolTable(c.scope, true)
	for _, param := range n.Params {
		fnScope.defineLocal(param)
	}

	var fnChunk Chunk
	prevScope, prevChunk := c.scope, c.chunk
	c.scope, c.chunk = fnScope, &fnChunk
	c.activeFuncs[n.Name] = true

	err := c.compileFuncBody(n)

	delete(c.activeFuncs, n.Name)
	c.scope, c.chunk = prevScope, prevChunk
	if err != nil {
		return err
	}

	c.prog.Functions = append(c.prog.Functions, FuncProto{
		Name: n.Name, Arity: len(n.Params), NumLocals: fnScope.numLocals, Chunk: fnChunk,
	})
	_, err = c.scope.defineUserFunc(n.Pos, n.Name, uint16(len(c.prog.Functions)-1), len(n.Params))

	return err
}

func (c *compiler) compileFuncBody(n *parse.DefineFuncNode) error {
	for _, def := range flattenBegin(n.Definitions) {
		if err := c.compileStmt(def, false); err != nil {
			return err
		}
	}

	for i, expr := range n.Sequence {
		if err := c.compileExpr(expr); err != nil {
			return err
		}

		if i < len(n.Sequence)-1 {
			c.chunk.Emit(c.chunk.PosAt(len(c.chunk.Code)-1), OpPop)
		}
	}

	c.chunk.Emit(n.Pos, OpReturn)

	return nil
}

func (c *compiler) compileExpr(node parse.Node) error {
	switch n := node.(type) {
	case nil:
		return errors.New("nil AST node")
	case *parse.IntNode:
		c.chunk.Emit(n.Pos, OpConst, c.addConst(IntVal(n.Value)))
	case *parse.FloatNode:
		c.chunk.Emit(n.Pos, OpConst, c.addConst(FloatVal(n.Value)))
	case *parse.StringNode:
		c.chunk.Emit(n.Pos, OpConst, c.addConst(StringVal(n.Value)))
	case *parse.BoolNode:
		c.chunk.Emit(n.Pos, selectOp(n.Value, OpTrue, OpFalse))
	case *parse.IdentifierNode:
		sym, err := c.scope.resolveVar(n.Pos, n.Name, c.activeFuncs)
		if err != nil {
			return err
		}

		c.chunk.Emit(n.Pos, selectOp(sym.kind == symLocalVar, OpGetLocal, OpGetGlobal), int(sym.index))
	case *parse.SetNode:
		sym, err := c.scope.resolveVar(n.Pos, n.Name, c.activeFuncs)
		if err != nil {
			return err
		}

		if err := c.compileExpr(n.Value); err != nil {
			return err
		}

		c.chunk.Emit(n.Pos, selectOp(sym.kind == symLocalVar, OpSetLocal, OpSetGlobal), int(sym.index))
		c.chunk.Emit(n.Pos, OpVoid)
	case *parse.IfNode:
		return c.compileIf(n)
	case *parse.ProcedureCallNode:
		return c.compileCall(n)
	case *parse.ProgramNode, *parse.BeginNode, *parse.DefineVarNode, *parse.DefineFuncNode:
		return fmt.Errorf("unexpected non-expression node %T", node)
	}

	return nil
}

func (c *compiler) compileIf(n *parse.IfNode) error {
	if err := c.compileExpr(n.Test); err != nil {
		return err
	}

	jumpFalseOp := c.chunk.Emit(n.Pos, OpJumpIfFalse, 0)
	if err := c.compileExpr(n.Consequent); err != nil {
		return err
	}

	jumpEndOp := c.chunk.Emit(n.Pos, OpJump, 0)
	c.chunk.PatchUint16(jumpFalseOp+1, uint16(len(c.chunk.Code)-(jumpFalseOp+3)))

	if n.Alternate != nil {
		if err := c.compileExpr(n.Alternate); err != nil {
			return err
		}
	} else {
		c.chunk.Emit(n.Pos, OpVoid)
	}

	c.chunk.PatchUint16(jumpEndOp+1, uint16(len(c.chunk.Code)-(jumpEndOp+3)))

	return nil
}

func (c *compiler) compileCall(n *parse.ProcedureCallNode) error {
	if len(n.Operands) > math.MaxUint8 {
		return errorfAt(n.Pos, "call to %q exceeds maximum argument count (%d)", n.Operator, math.MaxUint8)
	}

	sym, err := c.scope.resolveProc(n.Pos, n.Operator, len(n.Operands), c.activeFuncs)
	if err != nil {
		return err
	}

	for _, arg := range n.Operands {
		if err := c.compileExpr(arg); err != nil {
			return err
		}
	}

	c.chunk.Emit(n.Pos, selectOp(sym.kind == symUserFunc, OpCall, OpCallBuiltin), int(sym.index), len(n.Operands))

	return nil
}

func selectOp(cond bool, t, f Opcode) Opcode {
	if cond {
		return t
	}

	return f
}

func (c *compiler) addConst(v Value) int {
	c.prog.Constants = append(c.prog.Constants, v)

	return len(c.prog.Constants) - 1
}
