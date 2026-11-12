package vm

import (
	"fmt"
	"strings"

	"github.com/routis819/slicep/pkg/lex"
)

// Opcode identifies a VM instruction.
type Opcode uint8

// Instruction set. Operands A and B are interpreted per opcode; unused
// operands are zero.
const (
	// OpConstant pushes Constants[A].
	OpConstant Opcode = iota
	// OpVoid pushes Void.
	OpVoid
	// OpPop discards the top of the stack.
	OpPop
	// OpGetGlobal pushes global slot A.
	OpGetGlobal
	// OpSetGlobal pops a value into global slot A.
	OpSetGlobal
	// OpGetLocal pushes local slot A of the current frame (fp + A).
	OpGetLocal
	// OpSetLocal pops a value into local slot A of the current frame.
	OpSetLocal
	// OpJump jumps to instruction index A.
	OpJump
	// OpJumpIfFalse pops a value and jumps to instruction index A if it is #f.
	OpJumpIfFalse
	// OpCall calls Functions[A] with the top B stack values as arguments.
	OpCall
	// OpCallBuiltin calls Builtins[A] with the top B stack values as arguments.
	OpCallBuiltin
	// OpReturn pops the return value, discards the current frame and pushes
	// the value onto the caller's stack. Returning from the main chunk ends
	// execution.
	OpReturn
)

var opcodeNames = [...]string{
	OpConstant:    "CONSTANT",
	OpVoid:        "VOID",
	OpPop:         "POP",
	OpGetGlobal:   "GET_GLOBAL",
	OpSetGlobal:   "SET_GLOBAL",
	OpGetLocal:    "GET_LOCAL",
	OpSetLocal:    "SET_LOCAL",
	OpJump:        "JUMP",
	OpJumpIfFalse: "JUMP_IF_FALSE",
	OpCall:        "CALL",
	OpCallBuiltin: "CALL_BUILTIN",
	OpReturn:      "RETURN",
}

func (op Opcode) String() string {
	if int(op) < len(opcodeNames) {
		return opcodeNames[op]
	}

	return fmt.Sprintf("Opcode(%d)", op)
}

// Instruction is a fixed-width VM instruction.
type Instruction struct {
	Op Opcode
	A  int
	B  int
}

// Chunk is a sequence of instructions together with its constant pool.
// Positions[i] is the source position of Code[i], used for error reporting.
type Chunk struct {
	Code      []Instruction
	Constants []Value
	Positions []lex.Pos
}

func (c *Chunk) emit(op Opcode, a, b int, pos lex.Pos) int {
	c.Code = append(c.Code, Instruction{Op: op, A: a, B: b})
	c.Positions = append(c.Positions, pos)

	return len(c.Code) - 1
}

func (c *Chunk) addConstant(v Value) int {
	c.Constants = append(c.Constants, v)

	return len(c.Constants) - 1
}

// Disassemble returns a human-readable listing of the chunk.
func (c *Chunk) Disassemble() string {
	var b strings.Builder
	for i, ins := range c.Code {
		fmt.Fprintf(&b, "%04d %-13s", i, ins.Op)
		switch ins.Op {
		case OpConstant:
			fmt.Fprintf(&b, " %d (%s)", ins.A, FormatValue(c.Constants[ins.A]))
		case OpGetGlobal, OpSetGlobal, OpGetLocal, OpSetLocal, OpJump, OpJumpIfFalse:
			fmt.Fprintf(&b, " %d", ins.A)
		case OpCall, OpCallBuiltin:
			fmt.Fprintf(&b, " %d %d", ins.A, ins.B)
		case OpVoid, OpPop, OpReturn:
		}
		b.WriteByte('\n')
	}

	return b.String()
}

// Function is a compiled user-defined function.
// Its frame holds Arity parameter slots followed by internal definition
// slots, NumLocals slots in total.
type Function struct {
	Name      string
	Arity     int
	NumLocals int
	Chunk     Chunk
}

// BuiltinFunc is the Go implementation of a builtin procedure.
type BuiltinFunc func(args []Value) (Value, error)

// Builtin is a procedure implemented in Go (builtin operators or host
// functions). MaxArgs < 0 means the procedure is variadic.
type Builtin struct {
	Name    string
	MinArgs int
	MaxArgs int
	Fn      BuiltinFunc
}

func (b *Builtin) acceptsArgc(argc int) bool {
	return argc >= b.MinArgs && (b.MaxArgs < 0 || argc <= b.MaxArgs)
}

// Program is the compiled form of a slicep <program>.
type Program struct {
	// Main is the top-level code.
	Main Chunk
	// Functions is the user-defined function table indexed by OpCall.
	Functions []*Function
	// Builtins is the builtin procedure table indexed by OpCallBuiltin.
	Builtins []Builtin
	// Globals holds the names of global variable slots.
	Globals []string
}
