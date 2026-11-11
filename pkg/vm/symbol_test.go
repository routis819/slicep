package vm

import (
	"strings"
	"testing"

	"github.com/routis819/slicep/pkg/lex"
)

func TestSymbolTableResolutionAndSafety(t *testing.T) {
	pos := lex.Pos{Line: 1, Column: 1}
	builtins := newSymbolTable(nil, false)
	builtins.defineBuiltin(0, BuiltinDef{Name: "+", MinArity: 0, MaxArity: -1})
	builtins.defineBuiltin(1, BuiltinDef{Name: "not", MinArity: 1, MaxArity: 1})

	globals := newSymbolTable(builtins, false)
	globals.store["g"] = symbol{name: "g", kind: symGlobalVar, index: 0}

	if _, err := globals.defineUserFunc(pos, "fn", 0, 2); err != nil {
		t.Fatalf("defineUserFunc(fn) unexpected error: %v", err)
	}

	if _, err := globals.defineUserFunc(pos, "fn", 1, 2); err == nil {
		t.Fatal("redefining user func in same scope succeeded, want error")
	}

	fnScope := newSymbolTable(globals, true)
	localA := fnScope.defineLocal("a")
	if localA.index != 0 || localA.kind != symLocalVar {
		t.Fatalf("localA = %#v, want index 0 symLocalVar", localA)
	}

	if sym, err := fnScope.resolveVar(pos, "a", nil); err != nil || sym.index != 0 {
		t.Fatalf("resolveVar(a) = %#v, %v", sym, err)
	}

	if sym, err := fnScope.resolveVar(pos, "g", nil); err != nil || sym.kind != symGlobalVar {
		t.Fatalf("resolveVar(g) = %#v, %v", sym, err)
	}

	if _, err := fnScope.resolveVar(pos, "fn", nil); err == nil || !strings.Contains(err.Error(), "cannot be used as a variable") {
		t.Fatalf("resolveVar(fn) = %v, want procedure-as-variable error", err)
	}

	if sym, err := fnScope.resolveProc(pos, "fn", 2, nil); err != nil || sym.kind != symUserFunc {
		t.Fatalf("resolveProc(fn, 2) = %#v, %v", sym, err)
	}

	if _, err := fnScope.resolveProc(pos, "fn", 1, nil); err == nil || !strings.Contains(err.Error(), "arity mismatch") {
		t.Fatalf("resolveProc(fn, 1) = %v, want arity mismatch", err)
	}

	if _, err := fnScope.resolveProc(pos, "a", 1, nil); err == nil || !strings.Contains(err.Error(), "not a procedure") {
		t.Fatalf("resolveProc(a) = %v, want not-a-procedure error", err)
	}

	active := map[string]bool{"rec": true}
	if _, err := fnScope.resolveProc(pos, "rec", 1, active); err == nil || !strings.Contains(err.Error(), "recursive call") {
		t.Fatalf("resolveProc(rec) = %v, want recursive call error", err)
	}

	nestedScope := newSymbolTable(fnScope, true)
	if _, err := nestedScope.resolveVar(pos, "a", nil); err == nil || !strings.Contains(err.Error(), "capturing outer local") {
		t.Fatalf("nested resolveVar(a) = %v, want outer local capture error", err)
	}
}
