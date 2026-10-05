package main

import "fmt"

// Chunk is compiled bytecode + its constants pool.
type Chunk struct {
	Code      []Instruction
	Constants []Value
}

// NewChunk creates an empty chunk.
func NewChunk() *Chunk {
	return &Chunk{}
}

// Emit adds an instruction to the chunk.
func (c *Chunk) Emit(op OpCode, operand int, line int) int {
	c.Code = append(c.Code, Instruction{Op: op, Operand: operand, Line: line})
	return len(c.Code) - 1
}

// AddConstant adds a value to the constants pool and returns its index.
func (c *Chunk) AddConstant(v Value) int {
	// Dedupe identical values (optional optimization)
	for i, existing := range c.Constants {
		if existing.Equal(v) {
			return i
		}
	}
	c.Constants = append(c.Constants, v)
	return len(c.Constants) - 1
}

// PatchJump fills in the operand of a jump instruction once target is known.
func (c *Chunk) PatchJump(index int, target int) {
	c.Code[index].Operand = target
}

// Disassemble prints a chunk for debugging.
func (c *Chunk) Disassemble() string {
	out := ""
	out += fmt.Sprintf("=== Constants (%d) ===\n", len(c.Constants))
	for i, cst := range c.Constants {
		out += fmt.Sprintf("  %d: %s\n", i, cst.String())
	}
	out += fmt.Sprintf("\n=== Code (%d) ===\n", len(c.Code))
	for i, inst := range c.Code {
		out += fmt.Sprintf("  %04d  %s\n", i, inst)
	}
	return out
}