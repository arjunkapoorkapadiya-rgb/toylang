package main

import (
	"fmt"
	"math"
	"os"
)

const maxStackSize = 100000

type VM struct {
	chunk  *Chunk
	ip     int
	stack  []Value
	frames []Frame
}

type Frame struct {
	fn        *FunctionValue
	ip        int
	stackBase int
}

func NewVM() *VM {
	return &VM{}
}

func (v *VM) Run(chunk *Chunk, functions map[string]*FunctionValue) error {
	main := &FunctionValue{
		Name:  "<main>",
		Arity: 0,
		Chunk: chunk,
	}
	v.frames = []Frame{{fn: main, ip: 0, stackBase: 0}}
	v.chunk = chunk
	v.ip = 0

	debug := os.Getenv("TOYLANG_DEBUG") == "1"

	for {
		if v.ip >= len(v.chunk.Code) {
			break
		}
		inst := v.chunk.Code[v.ip]

		if debug {
			fmt.Fprintf(os.Stderr, "[debug] ip=%d op=%s operand=%d stack=%d frames=%d\n",
				v.ip, opName(inst.Op), inst.Operand, len(v.stack), len(v.frames))
		}

		v.ip++

		switch inst.Op {
		case OpConst:
			v.push(v.chunk.Constants[inst.Operand])

		case OpTrue:
			v.push(boolVal(true))

		case OpFalse:
			v.push(boolVal(false))

		case OpNil:
			v.push(nilVal())

		case OpPop:
			v.pop()

		case OpAdd:
			if err := v.binaryOp("+"); err != nil {
				return err
			}

		case OpSub:
			if err := v.binaryOp("-"); err != nil {
				return err
			}

		case OpMul:
			if err := v.binaryOp("*"); err != nil {
				return err
			}

		case OpDiv:
			if err := v.binaryOp("/"); err != nil {
				return err
			}

		case OpMod:
			if err := v.binaryOp("%"); err != nil {
				return err
			}

		case OpEq, OpNeq, OpLt, OpGt, OpLte, OpGte:
			if err := v.compareOp(inst.Op); err != nil {
				return err
			}

		case OpNeg:
			val := v.pop()
			num, ok := val.(NumberValue)
			if !ok {
				return fmt.Errorf("cannot negate %s", val.Type())
			}
			v.push(numberVal(-num.V))

		case OpNot:
			val := v.pop()
			v.push(boolVal(!truthy(val)))

		case OpGetLocal:
			base := v.currentFrameBase()
			slot := base + inst.Operand
			if slot >= len(v.stack) {
				v.push(nilVal())
			} else {
				v.push(v.stack[slot])
			}

		case OpSetLocal:
			val := v.pop()
			base := v.currentFrameBase()
			slot := base + inst.Operand
			for len(v.stack) <= slot {
				v.stack = append(v.stack, nilVal())
			}
			v.stack[slot] = val
			v.push(val)

		case OpJump:
			v.ip = inst.Operand

		case OpJumpIfFalse:
			val := v.pop()
			if !truthy(val) {
				v.ip = inst.Operand
			}

		case OpCall:
			if err := v.callFunction(inst.Operand); err != nil {
				return err
			}

		case OpReturn:
			retVal := v.pop()
			if len(v.frames) > 1 {
				// Pop current frame
				v.frames = v.frames[:len(v.frames)-1]
				frame := v.frames[len(v.frames)-1]

				// Truncate stack to the state it was in before we called this function.
				v.stack = v.stack[:frame.stackBase]

				// Restore caller's chunk + ip
				v.chunk = frame.fn.Chunk
				v.ip = frame.ip

				// Push return value
				v.push(retVal)
				continue
			}
			v.push(retVal)
			return nil

		case OpPrint:
			val := v.pop()
			fmt.Println(val.String())

		case OpHalt:
			return nil
		}
	}
	return nil
}

func (v *VM) currentFrameBase() int {
	if len(v.frames) == 0 {
		return 0
	}
	return v.frames[len(v.frames)-1].stackBase
}

func (v *VM) push(val Value) {
	if len(v.stack) > maxStackSize {
		panic(fmt.Sprintf("stack overflow (depth %d) — possible infinite recursion", len(v.stack)))
	}
	v.stack = append(v.stack, val)
}

func (v *VM) pop() Value {
	if len(v.stack) == 0 {
		return nilVal()
	}
	val := v.stack[len(v.stack)-1]
	v.stack = v.stack[:len(v.stack)-1]
	return val
}

func (v *VM) binaryOp(op string) error {
	right := v.pop()
	left := v.pop()

	if op == "+" {
		if ls, ok := left.(StringValue); ok {
			if rs, ok := right.(StringValue); ok {
				v.push(stringVal(ls.V + rs.V))
				return nil
			}
			v.push(stringVal(ls.V + right.String()))
			return nil
		}
	}

	ln, lok := left.(NumberValue)
	rn, rok := right.(NumberValue)
	if !lok || !rok {
		return fmt.Errorf("cannot apply %s to %s and %s", op, left.Type(), right.Type())
	}

	var result float64
	switch op {
	case "+":
		result = ln.V + rn.V
	case "-":
		result = ln.V - rn.V
	case "*":
		result = ln.V * rn.V
	case "/":
		if rn.V == 0 {
			return fmt.Errorf("division by zero")
		}
		result = ln.V / rn.V
	case "%":
		result = math.Mod(ln.V, rn.V)
	}
	v.push(numberVal(result))
	return nil
}

func (v *VM) compareOp(op OpCode) error {
	right := v.pop()
	left := v.pop()

	if ln, ok := left.(NumberValue); ok {
		if rn, ok := right.(NumberValue); ok {
			var result bool
			switch op {
			case OpEq:
				result = ln.V == rn.V
			case OpNeq:
				result = ln.V != rn.V
			case OpLt:
				result = ln.V < rn.V
			case OpGt:
				result = ln.V > rn.V
			case OpLte:
				result = ln.V <= rn.V
			case OpGte:
				result = ln.V >= rn.V
			}
			v.push(boolVal(result))
			return nil
		}
	}

	if ls, ok := left.(StringValue); ok {
		if rs, ok := right.(StringValue); ok {
			var result bool
			switch op {
			case OpEq:
				result = ls.V == rs.V
			case OpNeq:
				result = ls.V != rs.V
			case OpLt:
				result = ls.V < rs.V
			case OpGt:
				result = ls.V > rs.V
			case OpLte:
				result = ls.V <= rs.V
			case OpGte:
				result = ls.V >= rs.V
			}
			v.push(boolVal(result))
			return nil
		}
	}

	return fmt.Errorf("cannot compare %s with %s", left.Type(), right.Type())
}

func (v *VM) callFunction(argCount int) error {
	// Pop args
	args := make([]Value, argCount)
	for i := argCount - 1; i >= 0; i-- {
		args[i] = v.pop()
	}

	// Pop callee
	callee := v.pop()
	fn, ok := callee.(*FunctionValue)
	if !ok {
		return fmt.Errorf("cannot call %s (not a function)", callee.Type())
	}
	if fn.Arity != argCount {
		return fmt.Errorf("%s expects %d args, got %d", fn.Name, fn.Arity, argCount)
	}
	if fn.Chunk == nil {
		return fmt.Errorf("function %s has no body", fn.Name)
	}

	// Save return address in caller frame
	callerFrame := &v.frames[len(v.frames)-1]
	callerFrame.ip = v.ip

	// The new frame's stack base is the top of stack *after args are popped*.
	// Push args back onto the stack as the callee's locals.
	newBase := len(v.stack)
	for _, arg := range args {
		v.stack = append(v.stack, arg)
	}

	v.frames = append(v.frames, Frame{
		fn:        fn,
		ip:        0,
		stackBase: newBase,
	})

	v.chunk = fn.Chunk
	v.ip = 0
	return nil
}