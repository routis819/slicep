package vm

import (
	"bytes"
	"strings"
	"testing"

	"github.com/routis819/slicep/pkg/lex"
	"github.com/routis819/slicep/pkg/parse"
)

func testBuiltins() []BuiltinDef {
	return []BuiltinDef{
		{Name: "+", MinArity: 0, MaxArity: -1},
		{Name: "-", MinArity: 1, MaxArity: -1},
		{Name: "=", MinArity: 2, MaxArity: -1},
		{Name: "not", MinArity: 1, MaxArity: 1},
	}
}

func makeBytecode(t *testing.T, ops ...any) []byte {
	t.Helper()

	var c Chunk
	for i := 0; i < len(ops); {
		op := ops[i].(Opcode)
		widths := operandWidths(op)
		operands := make([]int, len(widths))
		for j := range widths {
			operands[j] = ops[i+1+j].(int)
		}

		c.Emit(lex.Pos{}, op, operands...)
		i += 1 + len(widths)
	}

	return c.Code
}

func compileSource(src string, opts ...CompileOption) (*Program, error) {
	ast, err := parse.ParseProgram(strings.NewReader(src))
	if err != nil {
		return nil, err
	}

	return Compile(ast, opts...)
}

func TestCompileBytecodeGeneration(t *testing.T) {
	tests := []struct {
		name      string
		src       string
		wantCode  []byte
		wantConst []Value
		wantFuncs int
	}{
		{
			name: "literals and sequence pop",
			src:  `42 #t #f "hi"`,
			wantCode: makeBytecode(t,
				OpConst, 0, OpPop,
				OpTrue, OpPop,
				OpFalse, OpPop,
				OpConst, 1,
				OpReturn,
			),
			wantConst: []Value{IntVal(42), StringVal("hi")},
		},
		{
			name: "global define and set!",
			src:  `(define x 10) (set! x 20) x`,
			wantCode: makeBytecode(t,
				OpConst, 0,
				OpSetGlobal, 0,
				OpConst, 1,
				OpSetGlobal, 0,
				OpVoid,
				OpPop,
				OpGetGlobal, 0,
				OpReturn,
			),
			wantConst: []Value{IntVal(10), IntVal(20)},
		},
		{
			name: "if with and without alternate",
			src:  `(if #t 1 2) (if #f 3)`,
			wantCode: makeBytecode(t,
				OpTrue,
				OpJumpIfFalse, 6,
				OpConst, 0,
				OpJump, 3,
				OpConst, 1,
				OpPop,
				OpFalse,
				OpJumpIfFalse, 6,
				OpConst, 2,
				OpJump, 1,
				OpVoid,
				OpReturn,
			),
			wantConst: []Value{IntVal(1), IntVal(2), IntVal(3)},
		},
		{
			name: "function definition and call",
			src:  `(define (add1 a) (define b 1) (+ a b)) (add1 5)`,
			wantCode: makeBytecode(t,
				OpConst, 1,
				OpCall, 0, 1,
				OpReturn,
			),
			wantConst: []Value{IntVal(1), IntVal(5)},
			wantFuncs: 1,
		},
		{
			name: "top-level trailing definition emits OpVoid",
			src:  `(begin (define x 3.5))`,
			wantCode: makeBytecode(t,
				OpConst, 0,
				OpSetGlobal, 0,
				OpVoid,
				OpReturn,
			),
			wantConst: []Value{FloatVal(3.5)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			prog, err := compileSource(tc.src, WithBuiltins(testBuiltins()))
			if err != nil {
				t.Fatalf("compileSource(%q) unexpected error: %v", tc.src, err)
			}

			if !bytes.Equal(prog.Main.Code, tc.wantCode) {
				t.Errorf("Main.Code = %x, want %x", prog.Main.Code, tc.wantCode)
			}

			if len(prog.Constants) != len(tc.wantConst) {
				t.Fatalf("len(Constants) = %d, want %d", len(prog.Constants), len(tc.wantConst))
			}

			for i, want := range tc.wantConst {
				if prog.Constants[i] != want {
					t.Errorf("Constants[%d] = %#v, want %#v", i, prog.Constants[i], want)
				}
			}

			if len(prog.Functions) != tc.wantFuncs {
				t.Errorf("len(Functions) = %d, want %d", len(prog.Functions), tc.wantFuncs)
			}
		})
	}
}

func TestCompileStaticTerminationAndSafetyErrors(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr string
	}{
		{
			name:    "direct recursion",
			src:     `(define (f x) (f x))`,
			wantErr: `recursive call to "f" is not allowed`,
		},
		{
			name:    "indirect/mutual recursion",
			src:     `(define (f x) (g x)) (define (g x) (f x))`,
			wantErr: `undefined procedure "g"`,
		},
		{
			name:    "nested function calling enclosing function recursively",
			src:     `(define (outer x) (define (inner y) (outer y)) (inner x))`,
			wantErr: `recursive call to "outer" is not allowed`,
		},
		{
			name:    "referencing user function as value",
			src:     `(define (f x) x) (define g f)`,
			wantErr: `procedure "f" cannot be used as a variable`,
		},
		{
			name:    "referencing active function as value inside own body",
			src:     `(define (f x) f)`,
			wantErr: `procedure "f" cannot be used as a variable`,
		},
		{
			name:    "referencing builtin as value",
			src:     `(define op +)`,
			wantErr: `procedure "+" cannot be used as a variable`,
		},
		{
			name:    "passing user function as argument",
			src:     `(define (f x) x) (+ f 1)`,
			wantErr: `procedure "f" cannot be used as a variable`,
		},
		{
			name:    "mutating user function with set!",
			src:     `(define (f x) x) (set! f 10)`,
			wantErr: `procedure "f" cannot be used as a variable`,
		},
		{
			name:    "mutating active function with set! inside own body",
			src:     `(define (f x) (set! f 10))`,
			wantErr: `procedure "f" cannot be used as a variable`,
		},
		{
			name:    "mutating builtin with set!",
			src:     `(set! + 10)`,
			wantErr: `procedure "+" cannot be used as a variable`,
		},
		{
			name:    "redefining function as variable in same scope",
			src:     `(define (f x) x) (define f 10)`,
			wantErr: `define: cannot redefine function "f" as variable`,
		},
		{
			name:    "redefining variable as function in same scope",
			src:     `(define f 10) (define (f x) x)`,
			wantErr: `define: "f" is already defined in this scope`,
		},
		{
			name:    "redefining function in same scope",
			src:     `(define (f x) x) (define (f x) (+ x 1))`,
			wantErr: `define: "f" is already defined in this scope`,
		},
		{
			name:    "calling global variable as procedure",
			src:     `(define x 10) (x 1)`,
			wantErr: `variable "x" is not a procedure`,
		},
		{
			name:    "calling local parameter shadowing builtin as procedure",
			src:     `(define (f +) (+ 1 2))`,
			wantErr: `variable "+" is not a procedure`,
		},
		{
			name:    "user function arity mismatch",
			src:     `(define (f a b) (+ a b)) (f 1)`,
			wantErr: `procedure "f": arity mismatch (got 1 arguments)`,
		},
		{
			name:    "builtin arity mismatch (too few)",
			src:     `(-)`,
			wantErr: `procedure "-": arity mismatch (got 0 arguments)`,
		},
		{
			name:    "builtin arity mismatch (too many)",
			src:     `(not #t #f)`,
			wantErr: `procedure "not": arity mismatch (got 2 arguments)`,
		},
		{
			name:    "undefined variable",
			src:     `(+ unknown 1)`,
			wantErr: `undefined variable "unknown"`,
		},
		{
			name:    "self-referencing variable in own initializer",
			src:     `(define x (+ x 1))`,
			wantErr: `undefined variable "x"`,
		},
		{
			name:    "capturing outer local variable in nested function",
			src:     `(define (outer a) (define (inner b) (+ a b)) (inner 1))`,
			wantErr: `capturing outer local variable "a" is not supported`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := compileSource(tc.src, WithBuiltins(testBuiltins()))
			if err == nil {
				t.Fatalf("compileSource(%q) succeeded, want error containing %q", tc.src, tc.wantErr)
			}

			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("compileSource(%q) error = %q, want substring %q", tc.src, err.Error(), tc.wantErr)
			}
		})
	}
}
