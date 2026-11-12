package vm

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/routis819/slicep/pkg/parse"
)

func TestCompile_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		// Static termination guarantees.
		{
			name:    "Direct recursion",
			input:   "(define (f x) (f x))",
			wantErr: `recursive reference to function "f"`,
		},
		{
			name:    "Direct recursion in nested expression",
			input:   "(define (fact n) (if (< n 1) 1 (+ n (fact (- n 1)))))",
			wantErr: `recursive reference to function "fact"`,
		},
		{
			name:    "Direct recursion as value",
			input:   "(define (f x) f)",
			wantErr: `recursive reference to function "f"`,
		},
		{
			name:    "Mutual recursion",
			input:   "(define (even? n) (odd? n)) (define (odd? n) (even? n))",
			wantErr: `reference to function "odd?" before its definition`,
		},
		{
			name:    "Indirect recursion through three functions",
			input:   "(define (a) (b)) (define (b) (c)) (define (c) (a))",
			wantErr: `reference to function "b" before its definition`,
		},
		{
			name:    "Forward reference inside begin",
			input:   "(begin (define (a) (b))) (define (b) 1)",
			wantErr: `reference to function "b" before its definition`,
		},
		{
			name:    "Recursion of internal function",
			input:   "(define (f x) (define (g y) (g y)) (g x))",
			wantErr: `recursive reference to function "g"`,
		},
		{
			name:    "Internal function calls enclosing function",
			input:   "(define (f x) (define (g y) (f y)) (g x))",
			wantErr: `recursive reference to function "f"`,
		},
		{
			name:    "Mutual recursion of internal functions",
			input:   "(define (f x) (define (g y) (h y)) (define (h y) (g y)) (g x))",
			wantErr: `reference to function "h" before its definition`,
		},
		{
			name:    "Function used as value",
			input:   "(define (f) 1) (define x f)",
			wantErr: `function "f" cannot be used as a value`,
		},
		{
			name:    "Function passed as argument",
			input:   "(define (f) 1) (define (g x) x) (g f)",
			wantErr: `function "f" cannot be used as a value`,
		},
		{
			name:    "Builtin used as value",
			input:   "(define x +)",
			wantErr: `function "+" cannot be used as a value`,
		},
		{
			name:    "set! function",
			input:   "(define (f) 1) (define (g) 2) (set! f 3)",
			wantErr: `set!: function "f" cannot be used as a value`,
		},
		{
			name:    "set! builtin",
			input:   "(set! + 1)",
			wantErr: `set!: function "+" cannot be used as a value`,
		},
		{
			name:    "Redefine function",
			input:   "(define (f) 1) (define (f) 2)",
			wantErr: `"f" is already defined`,
		},
		{
			name:    "Redefine function as variable",
			input:   "(define (f) 1) (define f 2)",
			wantErr: `"f" is already defined`,
		},
		{
			name:    "Redefine variable as function",
			input:   "(define f 1) (define (f) 2)",
			wantErr: `"f" is already defined`,
		},
		{
			name:    "Call variable",
			input:   "(define x 1) (x)",
			wantErr: `"x" is a variable, not a function`,
		},
		{
			name:    "Call parameter",
			input:   "(define (f g) (g 1))",
			wantErr: `"g" is a variable, not a function`,
		},
		// Scoping and binding.
		{
			name:    "Undefined identifier",
			input:   "x",
			wantErr: `undefined identifier "x"`,
		},
		{
			name:    "Undefined function",
			input:   "(foo 1)",
			wantErr: `undefined identifier "foo"`,
		},
		{
			name:    "Self-referencing variable",
			input:   "(define x x)",
			wantErr: `undefined identifier "x"`,
		},
		{
			name:    "set! undefined",
			input:   "(set! x 1)",
			wantErr: `undefined identifier "x"`,
		},
		{
			name:    "Use before define",
			input:   "(define y x) (define x 1)",
			wantErr: `undefined identifier "x"`,
		},
		{
			name:    "Function body references later global",
			input:   "(define (f) x) (define x 1)",
			wantErr: `undefined identifier "x"`,
		},
		{
			name:    "Local not visible outside function",
			input:   "(define (f a) a) a",
			wantErr: `undefined identifier "a"`,
		},
		{
			name:    "Internal function not visible outside",
			input:   "(define (f) (define (g) 1) (g)) (g)",
			wantErr: `undefined identifier "g"`,
		},
		{
			name:    "Closure capture",
			input:   "(define (f x) (define (g) x) (g))",
			wantErr: `cannot access local variable "x" of enclosing function "f"`,
		},
		{
			name:    "Closure set!",
			input:   "(define (f x) (define (g) (set! x 1)) (g))",
			wantErr: `cannot access local variable "x" of enclosing function "f"`,
		},
		{
			name:    "Duplicate internal definition",
			input:   "(define (f) (define a 1) (define a 2) a)",
			wantErr: `"a" is already defined`,
		},
		{
			name:    "Internal definition shadows parameter",
			input:   "(define (f a) (define a 2) a)",
			wantErr: `"a" is already defined`,
		},
		// Arity.
		{
			name:    "Too many arguments",
			input:   "(define (f a) a) (f 1 2)",
			wantErr: `function "f" expects 1 argument(s), got 2`,
		},
		{
			name:    "Too few arguments",
			input:   "(define (f a b) a) (f 1)",
			wantErr: `function "f" expects 2 argument(s), got 1`,
		},
		{
			name:    "Builtin arity",
			input:   "(- 1)",
			wantErr: `builtin "-": wrong number of arguments: 1`,
		},
		{
			name:    "Error inside operand",
			input:   "(+ 1 (undefined-fn))",
			wantErr: `undefined identifier "undefined-fn"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := compileSource(t, tt.input)
			if err == nil {
				t.Fatalf("Compile(%q) succeeded, want error containing %q", tt.input, tt.wantErr)
			}

			var cerr *CompileError
			if !errors.As(err, &cerr) {
				t.Errorf("error %v is not *CompileError", err)
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Compile(%q) error = %q, want it to contain %q", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestCompile_ErrorPosition(t *testing.T) {
	_, err := compileSource(t, "(define (f x)\n  (f x))")
	if err == nil {
		t.Fatal("Compile succeeded, want error")
	}

	if !strings.HasPrefix(err.Error(), "2:3: ") {
		t.Errorf("error = %q, want prefix %q", err, "2:3: ")
	}
}

func TestCompile_InvalidInput(t *testing.T) {
	if _, err := Compile(nil); err == nil {
		t.Error("Compile(nil) succeeded, want error")
	}

	ast := &parse.ProgramNode{Body: []parse.Node{&parse.IntNode{Value: 1}}}
	if _, err := Compile(ast, Builtin{Name: "f"}); err == nil {
		t.Error("Compile with nil builtin Fn succeeded, want error")
	}
}

func TestCompile_Bytecode(t *testing.T) {
	prog := mustCompile(t, "(define x 1) (define (f a) (define b a) (if a b)) (set! x (f x)) x")

	wantMain := []Instruction{
		{Op: OpConstant, A: 0},
		{Op: OpSetGlobal, A: 0},
		{Op: OpGetGlobal, A: 0},
		{Op: OpCall, A: 0, B: 1},
		{Op: OpSetGlobal, A: 0},
		{Op: OpVoid},
		{Op: OpPop},
		{Op: OpGetGlobal, A: 0},
		{Op: OpReturn},
	}
	if !reflect.DeepEqual(prog.Main.Code, wantMain) {
		t.Errorf("main code =\n%s", prog.Main.Disassemble())
	}

	if len(prog.Functions) != 1 {
		t.Fatalf("len(Functions) = %d, want 1", len(prog.Functions))
	}

	fn := prog.Functions[0]
	if fn.Name != "f" || fn.Arity != 1 || fn.NumLocals != 2 {
		t.Errorf("function = {%q, arity %d, locals %d}, want {\"f\", 1, 2}", fn.Name, fn.Arity, fn.NumLocals)
	}

	wantFn := []Instruction{
		{Op: OpGetLocal, A: 0},
		{Op: OpSetLocal, A: 1},
		{Op: OpGetLocal, A: 0},
		{Op: OpJumpIfFalse, A: 6},
		{Op: OpGetLocal, A: 1},
		{Op: OpJump, A: 7},
		{Op: OpVoid},
		{Op: OpReturn},
	}
	if !reflect.DeepEqual(fn.Chunk.Code, wantFn) {
		t.Errorf("function code =\n%s", fn.Chunk.Disassemble())
	}

	if !reflect.DeepEqual(prog.Globals, []string{"x"}) {
		t.Errorf("Globals = %v, want [x]", prog.Globals)
	}
}

func TestDisassemble(t *testing.T) {
	prog := mustCompile(t, `(define (f a) a) (if #t (f "s") (+ 1 2))`)

	got := prog.Main.Disassemble()
	for _, want := range []string{
		`CONSTANT      0 (#t)`,
		`JUMP_IF_FALSE`,
		`CALL          0 1`,
		`CALL_BUILTIN  0 2`,
		`RETURN`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Disassemble() missing %q:\n%s", want, got)
		}
	}

	if got := Opcode(255).String(); got != "Opcode(255)" {
		t.Errorf("Opcode(255).String() = %q", got)
	}
}
