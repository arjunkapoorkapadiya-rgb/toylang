package main

import "fmt"

// Compiler turns an AST into bytecode.
type Compiler struct {
	chunk     *Chunk
	locals    []map[string]int
	nextLocal int
	functions map[string]*FunctionValue
}

// NewCompiler creates a compiler.
func NewCompiler() *Compiler {
	return &Compiler{
		chunk:     NewChunk(),
		locals:    []map[string]int{{}},
		functions: map[string]*FunctionValue{},
	}
}

// CompileProgram compiles a whole program.
func (c *Compiler) CompileProgram(p *Program) (*Chunk, map[string]*FunctionValue, error) {
	// Pre-register all top-level functions
	for _, stmt := range p.Statements {
		if fn, ok := stmt.(FuncDecl); ok {
			c.functions[fn.Name] = &FunctionValue{
				Name:  fn.Name,
				Arity: len(fn.Params),
			}
		}
	}

	for _, stmt := range p.Statements {
		if err := c.compileStatement(stmt); err != nil {
			return nil, nil, err
		}
	}

	c.chunk.Emit(OpHalt, 0, 0)
	return c.chunk, c.functions, nil
}

// ==== Statements ====

func (c *Compiler) compileStatement(stmt Node) error {
	switch s := stmt.(type) {
	case LetStmt:
		if err := c.compileExpr(s.Value); err != nil {
			return err
		}
		slot := c.declareLocal(s.Name)
		c.chunk.Emit(OpSetLocal, slot, 0)
		c.chunk.Emit(OpPop, 0, 0)
		return nil

	case AssignStmt:
		if err := c.compileExpr(s.Value); err != nil {
			return err
		}
		slot, ok := c.resolveLocal(s.Name)
		if !ok {
			return fmt.Errorf("undefined variable: %s", s.Name)
		}
		c.chunk.Emit(OpSetLocal, slot, 0)
		c.chunk.Emit(OpPop, 0, 0)
		return nil

	case ExprStmt:
		if err := c.compileExpr(s.Expr); err != nil {
			return err
		}
		c.chunk.Emit(OpPop, 0, 0)
		return nil

	case ReturnStmt:
		if s.Value != nil {
			if err := c.compileExpr(s.Value); err != nil {
				return err
			}
		} else {
			c.chunk.Emit(OpNil, 0, 0)
		}
		c.chunk.Emit(OpReturn, 0, 0)
		return nil

	case IfStmt:
		return c.compileIf(s)

	case WhileStmt:
		return c.compileWhile(s)

	case FuncDecl:
		return c.compileFunction(s)
	}
	return fmt.Errorf("unsupported statement: %T", stmt)
}

func (c *Compiler) compileIf(s IfStmt) error {
	if err := c.compileExpr(s.Cond); err != nil {
		return err
	}
	jumpIfFalseIdx := c.chunk.Emit(OpJumpIfFalse, 0, 0)

	for _, stmt := range s.Then {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}

	if len(s.Else) > 0 {
		jumpOverElseIdx := c.chunk.Emit(OpJump, 0, 0)
		c.chunk.PatchJump(jumpIfFalseIdx, len(c.chunk.Code))
		for _, stmt := range s.Else {
			if err := c.compileStatement(stmt); err != nil {
				return err
			}
		}
		c.chunk.PatchJump(jumpOverElseIdx, len(c.chunk.Code))
	} else {
		c.chunk.PatchJump(jumpIfFalseIdx, len(c.chunk.Code))
	}
	return nil
}

func (c *Compiler) compileWhile(s WhileStmt) error {
	startIdx := len(c.chunk.Code)
	if err := c.compileExpr(s.Cond); err != nil {
		return err
	}
	jumpIfFalseIdx := c.chunk.Emit(OpJumpIfFalse, 0, 0)
	for _, stmt := range s.Body {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}
	c.chunk.Emit(OpJump, startIdx, 0)
	c.chunk.PatchJump(jumpIfFalseIdx, len(c.chunk.Code))
	return nil
}

func (c *Compiler) compileFunction(s FuncDecl) error {
	// Save current state
	outerChunk := c.chunk
	outerLocals := c.locals
	outerNextLocal := c.nextLocal

	// New chunk for this function
	fnChunk := NewChunk()
	c.chunk = fnChunk
	c.locals = []map[string]int{{}}
	c.nextLocal = 0

	// Params become locals in slots 0..n-1
	for _, p := range s.Params {
		c.declareLocal(p)
	}

	// Compile body
	for _, stmt := range s.Body {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}

	// Implicit return nil at end
	fnChunk.Emit(OpNil, 0, 0)
	fnChunk.Emit(OpReturn, 0, 0)

	// Attach chunk
	if fnVal, ok := c.functions[s.Name]; ok {
		fnVal.Chunk = fnChunk
	}

	// Restore
	c.chunk = outerChunk
	c.locals = outerLocals
	c.nextLocal = outerNextLocal

	return nil
}

// ==== Expressions ====

func (c *Compiler) compileExpr(expr Node) error {
	switch e := expr.(type) {
	case NumberLit:
		var f float64
		if _, err := fmt.Sscanf(e.Value, "%f", &f); err != nil {
			return fmt.Errorf("invalid number: %s", e.Value)
		}
		idx := c.chunk.AddConstant(numberVal(f))
		c.chunk.Emit(OpConst, idx, 0)
		return nil

	case StringLit:
		idx := c.chunk.AddConstant(stringVal(e.Value))
		c.chunk.Emit(OpConst, idx, 0)
		return nil

	case BoolLit:
		if e.Value {
			c.chunk.Emit(OpTrue, 0, 0)
		} else {
			c.chunk.Emit(OpFalse, 0, 0)
		}
		return nil

	case NilLit:
		c.chunk.Emit(OpNil, 0, 0)
		return nil

	case Ident:
		if slot, ok := c.resolveLocal(e.Name); ok {
			c.chunk.Emit(OpGetLocal, slot, 0)
			return nil
		}
		if fn, ok := c.functions[e.Name]; ok {
			idx := c.chunk.AddConstant(fn)
			c.chunk.Emit(OpConst, idx, 0)
			return nil
		}
		return fmt.Errorf("undefined variable: %s", e.Name)

	case Binary:
		return c.compileBinary(e)

	case Unary:
		return c.compileUnary(e)

	case Call:
		return c.compileCall(e)
	}
	return fmt.Errorf("unsupported expression: %T", expr)
}

func (c *Compiler) compileBinary(b Binary) error {
	if err := c.compileExpr(b.Left); err != nil {
		return err
	}
	if err := c.compileExpr(b.Right); err != nil {
		return err
	}
	switch b.Op {
	case "+":
		c.chunk.Emit(OpAdd, 0, 0)
	case "-":
		c.chunk.Emit(OpSub, 0, 0)
	case "*":
		c.chunk.Emit(OpMul, 0, 0)
	case "/":
		c.chunk.Emit(OpDiv, 0, 0)
	case "%":
		c.chunk.Emit(OpMod, 0, 0)
	case "==":
		c.chunk.Emit(OpEq, 0, 0)
	case "!=":
		c.chunk.Emit(OpNeq, 0, 0)
	case "<":
		c.chunk.Emit(OpLt, 0, 0)
	case ">":
		c.chunk.Emit(OpGt, 0, 0)
	case "<=":
		c.chunk.Emit(OpLte, 0, 0)
	case ">=":
		c.chunk.Emit(OpGte, 0, 0)
	default:
		return fmt.Errorf("unknown binary op: %s", b.Op)
	}
	return nil
}

func (c *Compiler) compileUnary(u Unary) error {
	if err := c.compileExpr(u.Right); err != nil {
		return err
	}
	switch u.Op {
	case "-":
		c.chunk.Emit(OpNeg, 0, 0)
	case "not":
		c.chunk.Emit(OpNot, 0, 0)
	default:
		return fmt.Errorf("unknown unary op: %s", u.Op)
	}
	return nil
}

func (c *Compiler) compileCall(call Call) error {
	// Special-case print (a built-in)
	if ident, ok := call.Callee.(Ident); ok && ident.Name == "print" {
		for _, arg := range call.Args {
			if err := c.compileExpr(arg); err != nil {
				return err
			}
			c.chunk.Emit(OpPrint, 0, 0)
		}
		c.chunk.Emit(OpNil, 0, 0)
		return nil
	}

	// Regular function call: push callee, then args
	if err := c.compileExpr(call.Callee); err != nil {
		return err
	}
	for _, arg := range call.Args {
		if err := c.compileExpr(arg); err != nil {
			return err
		}
	}
	c.chunk.Emit(OpCall, len(call.Args), 0)
	return nil
}

// ==== Scope ====

func (c *Compiler) declareLocal(name string) int {
	slot := c.nextLocal
	c.nextLocal++
	c.locals[len(c.locals)-1][name] = slot
	return slot
}

func (c *Compiler) resolveLocal(name string) (int, bool) {
	for i := len(c.locals) - 1; i >= 0; i-- {
		if slot, ok := c.locals[i][name]; ok {
			return slot, true
		}
	}
	return 0, false
}