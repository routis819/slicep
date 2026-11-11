package vm

import (
	"errors"
	"strings"
	"testing"
)

func evalTestBuiltins() []BuiltinDef {
	return []BuiltinDef{
		{
			Name:     "+",
			MinArity: 0,
			MaxArity: -1,
			Fn: func(args []Value) (Value, error) {
				var sum int64
				for _, a := range args {
					iv, ok := a.(IntVal)
					if !ok {
						return nil, errors.New("expected int")
					}

					sum += int64(iv)
				}

				return IntVal(sum), nil
			},
		},
		{
			Name:     "-",
			MinArity: 1,
			MaxArity: -1,
			Fn: func(args []Value) (Value, error) {
				first, ok := args[0].(IntVal)
				if !ok {
					return nil, errors.New("expected int")
				}

				if len(args) == 1 {
					return -first, nil
				}

				res := int64(first)
				for _, a := range args[1:] {
					iv, ok := a.(IntVal)
					if !ok {
						return nil, errors.New("expected int")
					}

					res -= int64(iv)
				}

				return IntVal(res), nil
			},
		},
		{
			Name:     "=",
			MinArity: 2,
			MaxArity: 2,
			Fn: func(args []Value) (Value, error) {
				return BoolVal(args[0] == args[1]), nil
			},
		},
		{
			Name:     "fail!",
			MinArity: 0,
			MaxArity: 0,
			Fn: func(_ []Value) (Value, error) {
				return nil, errors.New("boom")
			},
		},
	}
}

func evalSource(t *testing.T, src string) (Value, error) {
	t.Helper()

	prog, err := compileSource(src, WithBuiltins(evalTestBuiltins()))
	if err != nil {
		return nil, err
	}

	machine := New(prog)

	return machine.Run()
}

func TestVMExecution(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want Value
	}{
		{
			name: "self-evaluating literals",
			src:  `10 3.25 #t "slicep"`,
			want: StringVal("slicep"),
		},
		{
			name: "global define and set! state mutation",
			src: `
				(define hp 100)
				(define dmg 35)
				(set! hp (- hp dmg))
				hp
			`,
			want: IntVal(65),
		},
		{
			name: "set! expression evaluates to VoidVal",
			src: `
				(define x 1)
				(set! x 2)
			`,
			want: VoidVal{},
		},
		{
			name: "if conditional branches and R7RS truthiness",
			src: `
				(define a (if #t 10 20))
				(define b (if #f 10 20))
				(define c (if 0 30 40))
				(define d (if "" 50 60))
				(+ a b c d)
			`,
			want: IntVal(110),
		},
		{
			name: "if without alternate returns VoidVal on false",
			src:  `(if #f 99)`,
			want: VoidVal{},
		},
		{
			name: "if without alternate mutates state on true",
			src: `
				(define bonus 0)
				(if #t (set! bonus 15))
				(if #f (set! bonus 999))
				bonus
			`,
			want: IntVal(15),
		},
		{
			name: "begin top-level splicing",
			src: `
				(begin
					(define a 7)
					(define b 8)
					(+ a b))
			`,
			want: IntVal(15),
		},
		{
			name: "user-defined function with local defines and local set!",
			src: `
				(define (compute base bonus)
					(define step1 (+ base bonus))
					(define step2 10)
					(set! step2 (+ step1 step2))
					step2)
				(compute 20 5)
			`,
			want: IntVal(35),
		},
		{
			name: "non-recursive DAG call chain (f -> g -> h) and nested helper",
			src: `
				(define global-offset 3)
				(define (double x) (+ x x))
				(define (quadruple x)
					(define (add-offset v) (+ v global-offset))
					(add-offset (double (double x))))
				(quadruple 10)
			`,
			want: IntVal(43),
		},
		{
			name: "function parameter shadows global variable without mutating global",
			src: `
				(define x 100)
				(define (modify x)
					(set! x 5)
					x)
				(+ (modify 20) x)
			`,
			want: IntVal(105),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := evalSource(t, tc.src)
			if err != nil {
				t.Fatalf("evalSource(%q) unexpected error: %v", tc.src, err)
			}

			if got != tc.want {
				t.Errorf("evalSource(%q) = %#v, want %#v", tc.src, got, tc.want)
			}
		})
	}
}

func TestVMRuntimeErrorWithPos(t *testing.T) {
	_, err := evalSource(t, "\n(fail!)")
	if err == nil {
		t.Fatal("expected runtime error from fail!, got nil")
	}

	if !strings.Contains(err.Error(), "2:1: fail!: boom") {
		t.Errorf("error = %q, want substring %q", err.Error(), "2:1: fail!: boom")
	}
}
