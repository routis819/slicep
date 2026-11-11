package vm

import (
	"bytes"
	"testing"

	"github.com/routis819/slicep/pkg/lex"
)

func TestValueTruthinessAndString(t *testing.T) {
	tests := []struct {
		val        Value
		wantTruthy bool
		wantStr    string
	}{
		{IntVal(42), true, "42"},
		{IntVal(0), true, "0"},
		{FloatVal(3.14), true, "3.14"},
		{FloatVal(0.0), true, "0"},
		{StringVal("hello"), true, "hello"},
		{StringVal(""), true, ""},
		{BoolVal(true), true, "#t"},
		{BoolVal(false), false, "#f"},
		{VoidVal{}, true, "#<void>"},
	}

	for _, tc := range tests {
		if got := IsTruthy(tc.val); got != tc.wantTruthy {
			t.Errorf("IsTruthy(%#v) = %v, want %v", tc.val, got, tc.wantTruthy)
		}

		if s, ok := tc.val.(interface{ String() string }); ok {
			if got := s.String(); got != tc.wantStr {
				t.Errorf("%#v.String() = %q, want %q", tc.val, got, tc.wantStr)
			}
		}
	}

	if IsTruthy(nil) {
		t.Errorf("IsTruthy(nil) = true, want false")
	}
}

func TestChunkEmitAndPatch(t *testing.T) {
	var c Chunk

	p1 := lex.Pos{Line: 1, Column: 1}
	p2 := lex.Pos{Line: 2, Column: 5}
	p3 := lex.Pos{Line: 3, Column: 10}

	off0 := c.Emit(p1, OpConst, 0x0102)
	if off0 != 0 {
		t.Fatalf("off0 = %d, want 0", off0)
	}

	off1 := c.Emit(p2, OpJumpIfFalse, 0)
	if off1 != 3 {
		t.Fatalf("off1 = %d, want 3", off1)
	}

	off2 := c.Emit(p3, OpCall, 0x0005, 3)
	if off2 != 6 {
		t.Fatalf("off2 = %d, want 6", off2)
	}

	off3 := c.Emit(p3, OpReturn)
	if off3 != 10 {
		t.Fatalf("off3 = %d, want 10", off3)
	}

	c.PatchUint16(off1+1, 0x0007)

	wantCode := []byte{
		byte(OpConst), 0x01, 0x02,
		byte(OpJumpIfFalse), 0x00, 0x07,
		byte(OpCall), 0x00, 0x05, 0x03,
		byte(OpReturn),
	}
	if !bytes.Equal(c.Code, wantCode) {
		t.Errorf("c.Code = %x, want %x", c.Code, wantCode)
	}

	if got := c.PosAt(off0); got != p1 {
		t.Errorf("PosAt(%d) = %v, want %v", off0, got, p1)
	}

	if got := c.PosAt(off1); got != p2 {
		t.Errorf("PosAt(%d) = %v, want %v", off1, got, p2)
	}

	if got := c.PosAt(off2); got != p3 {
		t.Errorf("PosAt(%d) = %v, want %v", off2, got, p3)
	}

	if got := c.PosAt(999); got != (lex.Pos{}) {
		t.Errorf("PosAt(999) = %v, want zero Pos", got)
	}
}
