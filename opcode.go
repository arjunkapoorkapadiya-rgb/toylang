package main

import "fmt"

// OpCode is a single bytecode instruction.
type OpCode byte

const (
	OpConst OpCode = iota // push constant from pool onto stack
	OpTrue
	OpFalse
	OpNil
	OpPop      // discard top of stack
	OpAdd      // pop 2, push sum
	OpSub
	OpMul
	OpDiv
	OpMod
	OpEq       // pop 2, push equality result
	OpNeq
	OpLt
	OpGt
	OpLte
	OpGte
	OpNeg      // pop 1, push negation
	OpNot      // pop 1, push logical not
	OpGetLocal // push local variable
	OpSetLocal // pop and store into local
	OpJump     // unconditional jump
	OpJumpIfFalse
	OpCall     // call a function
	OpReturn   // return from function
	OpPrint    // built-in: print top of stack
	OpHalt     // stop execution
)

// Instruction is an opcode + optional operand.
type Instruction struct {
	Op     OpCode
	Operand int
	Line   int
}

// String formats an instruction for debugging.
func (i Instruction) String() string {
	base := opName(i.Op)
	if i.Operand != 0 {
		return fmt.Sprintf("%-15s %d", base, i.Operand)
	}
	return base
}

func opName(op OpCode) string {
	switch op {
	case OpConst:
		return "CONST"
	case OpTrue:
		return "TRUE"
	case OpFalse:
		return "FALSE"
	case OpNil:
		return "NIL"
	case OpPop:
		return "POP"
	case OpAdd:
		return "ADD"
	case OpSub:
		return "SUB"
	case OpMul:
		return "MUL"
	case OpDiv:
		return "DIV"
	case OpMod:
		return "MOD"
	case OpEq:
		return "EQ"
	case OpNeq:
		return "NEQ"
	case OpLt:
		return "LT"
	case OpGt:
		return "GT"
	case OpLte:
		return "LTE"
	case OpGte:
		return "GTE"
	case OpNeg:
		return "NEG"
	case OpNot:
		return "NOT"
	case OpGetLocal:
		return "GET_LOCAL"
	case OpSetLocal:
		return "SET_LOCAL"
	case OpJump:
		return "JUMP"
	case OpJumpIfFalse:
		return "JUMP_IF_FALSE"
	case OpCall:
		return "CALL"
	case OpReturn:
		return "RETURN"
	case OpPrint:
		return "PRINT"
	case OpHalt:
		return "HALT"
	}
	return "UNKNOWN"
}
