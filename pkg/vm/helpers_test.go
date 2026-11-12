package vm

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/routis819/slicep/pkg/parse"
)

// testBuiltins provides minimal arithmetic/comparison procedures so that the
// compiler and VM can be exercised before the standard builtins (#5) exist.
func testBuiltins() []Builtin {
	return []Builtin{
		{Name: "+", MinArgs: 0, MaxArgs: -1, Fn: func(args []Value) (Value, error) {
			var sum Int
			for _, a := range args {
				i, ok := a.(Int)
				if !ok {
					return nil, fmt.Errorf("expected int, got %s", FormatValue(a))
				}
				sum += i
			}

			return sum, nil
		}},
		{Name: "-", MinArgs: 2, MaxArgs: 2, Fn: func(args []Value) (Value, error) {
			return args[0].(Int) - args[1].(Int), nil
		}},
		{Name: "<", MinArgs: 2, MaxArgs: 2, Fn: func(args []Value) (Value, error) {
			return Bool(args[0].(Int) < args[1].(Int)), nil
		}},
		{Name: "fail", MinArgs: 0, MaxArgs: 0, Fn: func(_ []Value) (Value, error) {
			return nil, errors.New("boom")
		}},
	}
}

func compileSource(t *testing.T, src string) (*Program, error) {
	t.Helper()

	ast, err := parse.ParseProgram(strings.NewReader(src))
	if err != nil {
		t.Fatalf("failed to parse %q: %v", src, err)
	}

	return Compile(ast, testBuiltins()...)
}

func mustCompile(t *testing.T, src string) *Program {
	t.Helper()

	prog, err := compileSource(t, src)
	if err != nil {
		t.Fatalf("Compile(%q) error: %v", src, err)
	}

	return prog
}
