package vm

import (
	"encoding/binary"
	"fmt"

	"github.com/routis819/slicep/pkg/lex"
)

// Opcode represents a single-byte VM instruction code.
type Opcode byte

const (
	// OpConst pushes Constants[u16] onto the operand stack.
	OpConst Opcode = iota
	// OpTrue pushes BoolVal(true) onto the operand stack.
	OpTrue
	// OpFalse pushes BoolVal(false) onto the operand stack.
	OpFalse
	// OpVoid pushes VoidVal{} onto the operand stack.
	OpVoid
	// OpPop discards the top value from the operand stack.
	OpPop
	// OpGetGlobal pushes Globals[u16] onto the operand stack.
	OpGetGlobal
	// OpSetGlobal pops the top value from the stack and stores it in Globals[u16].
	OpSetGlobal
	// OpGetLocal pushes Stack[Base+u16] onto the operand stack.
	OpGetLocal
	// OpSetLocal pops the top value from the stack and stores it in Stack[Base+u16].
	OpSetLocal
	// OpJump advances the instruction pointer forward by u16 bytes.
	OpJump
	// OpJumpIfFalse pops the top value and, if it is #f, advances IP forward by u16 bytes.
	OpJumpIfFalse
	// OpCall invokes Functions[u16] with u8 arguments on top of the stack.
	OpCall
	// OpCallBuiltin invokes Builtins[u16] with u8 arguments on top of the stack.
	OpCallBuiltin
	// OpReturn pops the return value, restores the caller frame, and pushes the return value.
	OpReturn
)

// BuiltinFunc is the signature for built-in operators and host-bound procedures.
type BuiltinFunc func(args []Value) (Value, error)

// BuiltinDef describes a built-in or host-bound procedure and its arity constraints.
// If MaxArity < 0, the procedure accepts any number of arguments >= MinArity.
type BuiltinDef struct {
	Name     string
	MinArity int
	MaxArity int
	Fn       BuiltinFunc
}

// Chunk holds a sequence of bytecode instructions and their corresponding source positions.
type Chunk struct {
	Code  []byte
	Lines []lex.Pos
}

// Emit appends an instruction and its encoded operands to the chunk, returning the byte offset of the opcode.
func (c *Chunk) Emit(pos lex.Pos, op Opcode, operands ...int) int {
	offset := len(c.Code)
	widths := operandWidths(op)
	if len(operands) != len(widths) {
		panic(fmt.Sprintf("opcode %d expects %d operands, got %d", op, len(widths), len(operands)))
	}

	c.appendByte(byte(op), pos)
	for i, w := range widths {
		val := operands[i]
		switch w {
		case 1:
			c.appendByte(byte(val), pos)
		case 2:
			var buf [2]byte
			binary.BigEndian.PutUint16(buf[:], uint16(val))
			c.appendByte(buf[0], pos)
			c.appendByte(buf[1], pos)
		default:
			panic(fmt.Sprintf("unsupported operand width %d", w))
		}
	}

	return offset
}

// PatchUint16 overwrites a 2-byte big-endian operand at the given byte offset.
func (c *Chunk) PatchUint16(offset int, val uint16) {
	binary.BigEndian.PutUint16(c.Code[offset:offset+2], val)
}

// PosAt returns the source position associated with the instruction at ip.
func (c *Chunk) PosAt(ip int) lex.Pos {
	if ip >= 0 && ip < len(c.Lines) {
		return c.Lines[ip]
	}

	return lex.Pos{}
}

func (c *Chunk) appendByte(b byte, pos lex.Pos) {
	c.Code = append(c.Code, b)
	c.Lines = append(c.Lines, pos)
}

func operandWidths(op Opcode) []int {
	switch op {
	case OpTrue, OpFalse, OpVoid, OpPop, OpReturn:
		return nil
	case OpConst, OpGetGlobal, OpSetGlobal, OpGetLocal, OpSetLocal, OpJump, OpJumpIfFalse:
		return []int{2}
	case OpCall, OpCallBuiltin:
		return []int{2, 1}
	default:
		panic(fmt.Sprintf("unknown opcode %d", op))
	}
}

// FuncProto represents a compiled user-defined procedure.
type FuncProto struct {
	Name      string
	Arity     int
	NumLocals int
	Chunk     Chunk
}

// Program represents an immutable compiled bytecode program ready for execution.
type Program struct {
	Main        Chunk
	Constants   []Value
	Functions   []FuncProto
	GlobalNames []string
	GlobalIndex map[string]uint16
	Builtins    []BuiltinDef
}
