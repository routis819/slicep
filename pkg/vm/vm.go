package vm

import (
	"fmt"

	"github.com/routis819/slicep/pkg/lex"
)

// RuntimeError is an error raised while executing bytecode.
type RuntimeError struct {
	Pos lex.Pos
	Err error
}

func (e *RuntimeError) Error() string {
	if e.Pos.Line > 0 {
		return fmt.Sprintf("%s: %v", e.Pos, e.Err)
	}

	return e.Err.Error()
}

func (e *RuntimeError) Unwrap() error {
	return e.Err
}

// callFrame is the activation record of a running chunk. Local slot i of
// the frame lives at stack[fp+i].
type callFrame struct {
	chunk *Chunk
	ip    int
	fp    int
}

// VM executes a compiled Program on a value stack.
type VM struct {
	prog    *Program
	stack   []Value
	frames  []callFrame
	globals []Value
}

// New returns a VM that executes prog.
func New(prog *Program) *VM {
	return &VM{prog: prog}
}

// Global returns the current value of the global variable name.
func (vm *VM) Global(name string) (Value, bool) {
	for i, n := range vm.prog.Globals {
		if n == name && i < len(vm.globals) {
			return vm.globals[i], true
		}
	}

	return nil, false
}

// Run executes the program from the beginning and returns the value of its
// last expression (Void if the program ends with a definition). Global
// variables are reset on each run.
func (vm *VM) Run() (Value, error) {
	vm.globals = make([]Value, len(vm.prog.Globals))
	for i := range vm.globals {
		vm.globals[i] = Void{}
	}
	vm.stack = vm.stack[:0]
	vm.frames = append(vm.frames[:0], callFrame{chunk: &vm.prog.Main})

	return vm.run()
}

func (vm *VM) push(v Value) {
	vm.stack = append(vm.stack, v)
}

func (vm *VM) pop() Value {
	v := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]

	return v
}

func (vm *VM) run() (Value, error) {
	frame := &vm.frames[len(vm.frames)-1]

	for {
		ins := frame.chunk.Code[frame.ip]
		frame.ip++

		switch ins.Op {
		case OpConstant:
			vm.push(frame.chunk.Constants[ins.A])
		case OpVoid:
			vm.push(Void{})
		case OpPop:
			vm.pop()
		case OpGetGlobal:
			vm.push(vm.globals[ins.A])
		case OpSetGlobal:
			vm.globals[ins.A] = vm.pop()
		case OpGetLocal:
			vm.push(vm.stack[frame.fp+ins.A])
		case OpSetLocal:
			vm.stack[frame.fp+ins.A] = vm.pop()
		case OpJump:
			frame.ip = ins.A
		case OpJumpIfFalse:
			if !IsTruthy(vm.pop()) {
				frame.ip = ins.A
			}
		case OpCall:
			fn := vm.prog.Functions[ins.A]
			fp := len(vm.stack) - ins.B
			for i := fn.Arity; i < fn.NumLocals; i++ {
				vm.push(Void{})
			}
			vm.frames = append(vm.frames, callFrame{chunk: &fn.Chunk, fp: fp})
			frame = &vm.frames[len(vm.frames)-1]
		case OpCallBuiltin:
			b := &vm.prog.Builtins[ins.A]
			base := len(vm.stack) - ins.B
			args := make([]Value, ins.B)
			copy(args, vm.stack[base:])
			result, err := b.Fn(args)
			if err != nil {
				return nil, &RuntimeError{
					Pos: frame.chunk.Positions[frame.ip-1],
					Err: fmt.Errorf("%s: %w", b.Name, err),
				}
			}
			if result == nil {
				return nil, &RuntimeError{
					Pos: frame.chunk.Positions[frame.ip-1],
					Err: fmt.Errorf("%s: builtin returned nil value", b.Name),
				}
			}
			vm.stack = vm.stack[:base]
			vm.push(result)
		case OpReturn:
			result := vm.pop()
			fp := frame.fp
			vm.frames = vm.frames[:len(vm.frames)-1]
			if len(vm.frames) == 0 {
				return result, nil
			}
			vm.stack = vm.stack[:fp]
			vm.push(result)
			frame = &vm.frames[len(vm.frames)-1]
		default:
			return nil, &RuntimeError{
				Pos: frame.chunk.Positions[frame.ip-1],
				Err: fmt.Errorf("unknown opcode %s", ins.Op),
			}
		}
	}
}
