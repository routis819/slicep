package vm

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type callFrame struct {
	chunk *Chunk
	ip    int
	base  int
}

// VM executes a compiled bytecode Program on a stack-based virtual machine.
type VM struct {
	prog    *Program
	stack   []Value
	globals []Value
	frames  []callFrame
}

// New creates a new VM instance initialized for the given compiled Program.
func New(prog *Program) *VM {
	var globals []Value
	if prog != nil && len(prog.GlobalNames) > 0 {
		globals = make([]Value, len(prog.GlobalNames))
		for i := range globals {
			globals[i] = VoidVal{}
		}
	}

	return &VM{
		prog:    prog,
		stack:   make([]Value, 0, 64),
		globals: globals,
		frames:  make([]callFrame, 0, 16),
	}
}

// Run executes the program from its top-level entry point and returns the final Value.
func (v *VM) Run() (Value, error) {
	if v.prog == nil {
		return nil, errors.New("cannot run nil program")
	}

	v.stack = v.stack[:0]
	v.frames = append(v.frames[:0], callFrame{
		chunk: &v.prog.Main,
		ip:    0,
		base:  0,
	})

	for len(v.frames) > 0 {
		frame := &v.frames[len(v.frames)-1]
		opPos := frame.chunk.PosAt(frame.ip)
		op := Opcode(frame.chunk.Code[frame.ip])
		frame.ip++

		switch op {
		case OpConst:
			idx := readUint16(frame)
			v.push(v.prog.Constants[idx])

		case OpTrue:
			v.push(BoolVal(true))

		case OpFalse:
			v.push(BoolVal(false))

		case OpVoid:
			v.push(VoidVal{})

		case OpPop:
			v.pop()

		case OpGetGlobal:
			slot := readUint16(frame)
			v.push(v.globals[slot])

		case OpSetGlobal:
			slot := readUint16(frame)
			v.globals[slot] = v.pop()

		case OpGetLocal:
			slot := readUint16(frame)
			v.push(v.stack[frame.base+int(slot)])

		case OpSetLocal:
			slot := readUint16(frame)
			v.stack[frame.base+int(slot)] = v.pop()

		case OpJump:
			offset := readUint16(frame)
			frame.ip += int(offset)

		case OpJumpIfFalse:
			offset := readUint16(frame)
			if !IsTruthy(v.pop()) {
				frame.ip += int(offset)
			}

		case OpCall:
			fnIdx := readUint16(frame)
			argc := int(frame.chunk.Code[frame.ip])
			frame.ip++

			fn := &v.prog.Functions[fnIdx]
			calleeBase := len(v.stack) - argc
			for range fn.NumLocals - argc {
				v.stack = append(v.stack, VoidVal{})
			}

			v.frames = append(v.frames, callFrame{
				chunk: &fn.Chunk,
				ip:    0,
				base:  calleeBase,
			})

		case OpCallBuiltin:
			bIdx := readUint16(frame)
			argc := int(frame.chunk.Code[frame.ip])
			frame.ip++

			b := v.prog.Builtins[bIdx]
			argStart := len(v.stack) - argc
			res, err := b.Fn(v.stack[argStart:])
			if err != nil {
				if opPos.Line > 0 {
					return nil, fmt.Errorf("%s: %s: %w", opPos, b.Name, err)
				}

				return nil, fmt.Errorf("%s: %w", b.Name, err)
			}

			v.stack = append(v.stack[:argStart], res)

		case OpReturn:
			ret := v.pop()
			calleeBase := frame.base
			v.frames = v.frames[:len(v.frames)-1]
			if len(v.frames) == 0 {
				v.stack = v.stack[:0]

				return ret, nil
			}

			v.stack = append(v.stack[:calleeBase], ret)

		default:
			return nil, errorfAt(opPos, "unknown opcode %d", op)
		}
	}

	return VoidVal{}, nil
}

func readUint16(frame *callFrame) uint16 {
	val := binary.BigEndian.Uint16(frame.chunk.Code[frame.ip : frame.ip+2])
	frame.ip += 2

	return val
}

func (v *VM) push(val Value) {
	v.stack = append(v.stack, val)
}

func (v *VM) pop() Value {
	idx := len(v.stack) - 1
	val := v.stack[idx]
	v.stack = v.stack[:idx]

	return val
}
