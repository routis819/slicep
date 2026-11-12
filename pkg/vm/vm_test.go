package vm

import (
	"errors"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Value
	}{
		{"Int literal", "42", Int(42)},
		{"Float literal", "3.5", Float(3.5)},
		{"String literal", `"hi"`, String("hi")},
		{"Bool literal", "#f", Bool(false)},
		{"Last expression wins", "1 2 3", Int(3)},
		{"Program ending with definition", "(define x 1)", Void{}},
		{"Global define", "(define x 10) x", Int(10)},
		{"Global define with expression", "(define x (+ 1 2)) (define y (+ x 4)) y", Int(7)},
		{"Global redefinition acts as assignment", "(define x 1) (define x (+ x 1)) x", Int(2)},
		{"set! global", "(define x 1) (set! x 5) x", Int(5)},
		{"set! returns void", "(define x 1) (set! x 5)", Void{}},
		{"set! uses old value", "(define x 1) (set! x (+ x 10)) x", Int(11)},
		{"if true", "(if #t 1 2)", Int(1)},
		{"if false", "(if #f 1 2)", Int(2)},
		{"if zero is truthy", "(if 0 1 2)", Int(1)},
		{"if empty string is truthy", `(if "" 1 2)`, Int(1)},
		{"if without alternate, true", "(if #t 1)", Int(1)},
		{"if without alternate, false", "(if #f 1)", Void{}},
		{"if with comparison", "(define x 3) (if (< x 5) \"small\" \"big\")", String("small")},
		{"nested if", "(if #f 1 (if #t 2 3))", Int(2)},
		{"if only evaluates chosen branch", "(define x 0) (if #t (set! x 1) (set! x 2)) x", Int(1)},
		{"if in operand position", "(+ 1 (if #f 10 0) (if #t 5 6))", Int(6)},
		{"begin", "(begin (define x 1) (set! x (+ x 1))) x", Int(2)},
		{"nested begin", "(begin (begin 1 2) (begin 3))", Int(3)},
		{"builtin call", "(+ 1 2 3)", Int(6)},
		{"builtin call no args", "(+)", Int(0)},
		{
			name:  "Function call",
			input: "(define (add a b) (+ a b)) (add 2 3)",
			want:  Int(5),
		},
		{
			name:  "Function with no params",
			input: "(define (five) 5) (five)",
			want:  Int(5),
		},
		{
			name:  "Function with body sequence",
			input: "(define (f a) 1 2 (+ a 3)) (f 4)",
			want:  Int(7),
		},
		{
			name:  "Function with internal definitions",
			input: "(define (f a) (define b (+ a 1)) (define c (+ b 1)) (+ a b c)) (f 1)",
			want:  Int(6),
		},
		{
			name:  "Function set! local and param",
			input: "(define (f a) (define b 1) (set! a (+ a 10)) (set! b (+ b a)) b) (f 1)",
			want:  Int(12),
		},
		{
			name:  "Function reads and mutates global",
			input: "(define counter 0) (define (inc! n) (set! counter (+ counter n)) counter) (inc! 2) (inc! 3) counter",
			want:  Int(5),
		},
		{
			name:  "Function calls previously defined function",
			input: "(define (double x) (+ x x)) (define (quad x) (double (double x))) (quad 3)",
			want:  Int(12),
		},
		{
			name:  "Nested calls as arguments",
			input: "(define (add a b) (+ a b)) (add (add 1 2) (add 3 4))",
			want:  Int(10),
		},
		{
			name:  "Internal function definition",
			input: "(define (f x) (define (sq y) (+ y y)) (sq (sq x))) (f 3)",
			want:  Int(12),
		},
		{
			name:  "Internal function calls global function",
			input: "(define (g y) (+ y 1)) (define (f x) (define (h y) (g y)) (h x)) (f 1)",
			want:  Int(2),
		},
		{
			name:  "Function with if",
			input: "(define (abs x) (if (< x 0) (- 0 x) x)) (+ (abs -3) (abs 4))",
			want:  Int(7),
		},
		{
			name:  "Parameter shadows global",
			input: "(define x 100) (define (f x) x) (+ (f 1) x)",
			want:  Int(101),
		},
		{
			name:  "User function shadows builtin",
			input: "(define (+ a b) (- a b)) (+ 5 3)",
			want:  Int(2),
		},
		{
			name:  "Locals do not leak between calls",
			input: "(define (f a) (define b a) b) (f 1) (f 2)",
			want:  Int(2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(mustCompile(t, tt.input)).Run()
			if err != nil {
				t.Fatalf("Run() error: %v", err)
			}

			if got != tt.want {
				t.Errorf("Run() = %s, want %s", FormatValue(got), FormatValue(tt.want))
			}
		})
	}
}

func TestRun_Global(t *testing.T) {
	v := New(mustCompile(t, "(define x 1) (define (f) (set! x 42)) (f)"))
	if _, err := v.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	got, ok := v.Global("x")
	if !ok || got != Int(42) {
		t.Errorf("Global(x) = %v, %v; want 42, true", got, ok)
	}

	if _, ok := v.Global("missing"); ok {
		t.Errorf("Global(missing) found, want not found")
	}

	// Running again resets globals and yields the same result.
	if _, err := v.Run(); err != nil {
		t.Fatalf("second Run() error: %v", err)
	}
	if got, _ := v.Global("x"); got != Int(42) {
		t.Errorf("Global(x) after rerun = %v, want 42", got)
	}
}

func TestRun_StackBalance(t *testing.T) {
	v := New(mustCompile(t, `
(define (f a) (define b (if #f 1)) (if a 1) (set! a 2) a)
(define x 0)
(if #t (f 1))
(set! x (f #f))
(f x)`))

	got, err := v.Run()
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if got != Int(2) {
		t.Errorf("Run() = %s, want 2", FormatValue(got))
	}
	if len(v.stack) != 0 {
		t.Errorf("stack not empty after Run: %v", v.stack)
	}
	if len(v.frames) != 0 {
		t.Errorf("frames not empty after Run: %v", v.frames)
	}
}

func TestRun_BuiltinError(t *testing.T) {
	_, err := New(mustCompile(t, "(define (f) (fail))\n(f)")).Run()
	if err == nil {
		t.Fatal("Run() succeeded, want error")
	}

	var rerr *RuntimeError
	if !errors.As(err, &rerr) {
		t.Fatalf("error %v is not *RuntimeError", err)
	}
	if rerr.Pos.Line != 1 {
		t.Errorf("error position = %v, want line 1", rerr.Pos)
	}
	if !strings.Contains(err.Error(), "fail: boom") {
		t.Errorf("error = %q, want it to mention %q", err, "fail: boom")
	}
}

func TestRun_BuiltinNilResult(t *testing.T) {
	prog := mustCompile(t, "(+ 1)")
	prog.Builtins[0].Fn = func(_ []Value) (Value, error) { return nil, nil }

	if _, err := New(prog).Run(); err == nil || !strings.Contains(err.Error(), "nil value") {
		t.Errorf("Run() error = %v, want nil value error", err)
	}
}
