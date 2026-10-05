package main

import (
	"fmt"
	"strings"
)

// Node is any AST node.
type Node interface {
	Format(indent int) string
}

// === Expressions ===

type NumberLit struct {
	Value string
}

func (n NumberLit) Format(indent int) string {
	return pad(indent) + "NumberLit(" + n.Value + ")"
}

type StringLit struct {
	Value string
}

func (n StringLit) Format(indent int) string {
	return pad(indent) + fmt.Sprintf("StringLit(%q)", n.Value)
}

type BoolLit struct {
	Value bool
}

func (n BoolLit) Format(indent int) string {
	return pad(indent) + fmt.Sprintf("BoolLit(%v)", n.Value)
}

type NilLit struct{}

func (n NilLit) Format(indent int) string {
	return pad(indent) + "NilLit"
}

type Ident struct {
	Name string
}

func (n Ident) Format(indent int) string {
	return pad(indent) + "Ident(" + n.Name + ")"
}

type Binary struct {
	Op    string
	Left  Node
	Right Node
}

func (n Binary) Format(indent int) string {
	var sb strings.Builder
	sb.WriteString(pad(indent) + "Binary(" + n.Op + ")\n")
	sb.WriteString(n.Left.Format(indent + 1) + "\n")
	sb.WriteString(n.Right.Format(indent + 1))
	return sb.String()
}

type Unary struct {
	Op    string
	Right Node
}

func (n Unary) Format(indent int) string {
	var sb strings.Builder
	sb.WriteString(pad(indent) + "Unary(" + n.Op + ")\n")
	sb.WriteString(n.Right.Format(indent + 1))
	return sb.String()
}

type Call struct {
	Callee Node
	Args   []Node
}

func (n Call) Format(indent int) string {
	var sb strings.Builder
	sb.WriteString(pad(indent) + "Call\n")
	sb.WriteString(pad(indent+1) + "Callee:\n")
	sb.WriteString(n.Callee.Format(indent + 2) + "\n")
	if len(n.Args) > 0 {
		sb.WriteString(pad(indent+1) + "Args:\n")
		for i, arg := range n.Args {
			sb.WriteString(arg.Format(indent + 2))
			if i < len(n.Args)-1 {
				sb.WriteString("\n")
			}
		}
	}
	return sb.String()
}

// === Statements ===

type LetStmt struct {
	Name  string
	Value Node
}

func (n LetStmt) Format(indent int) string {
	var sb strings.Builder
	sb.WriteString(pad(indent) + "Let(" + n.Name + ")\n")
	sb.WriteString(n.Value.Format(indent + 1))
	return sb.String()
}

type ExprStmt struct {
	Expr Node
}

func (n ExprStmt) Format(indent int) string {
	return pad(indent) + "ExprStmt\n" + n.Expr.Format(indent+1)
}

type ReturnStmt struct {
	Value Node // may be nil
}

func (n ReturnStmt) Format(indent int) string {
	if n.Value == nil {
		return pad(indent) + "Return"
	}
	return pad(indent) + "Return\n" + n.Value.Format(indent+1)
}

type AssignStmt struct {
	Name  string
	Value Node
}

func (n AssignStmt) Format(indent int) string {
	var sb strings.Builder
	sb.WriteString(pad(indent) + "Assign(" + n.Name + ")\n")
	sb.WriteString(n.Value.Format(indent + 1))
	return sb.String()
}

type IfStmt struct {
	Cond    Node
	Then    []Node
	Else    []Node
}

func (n IfStmt) Format(indent int) string {
	var sb strings.Builder
	sb.WriteString(pad(indent) + "If\n")
	sb.WriteString(pad(indent+1) + "Cond:\n")
	sb.WriteString(n.Cond.Format(indent + 2) + "\n")
	sb.WriteString(pad(indent+1) + "Then:\n")
	for _, stmt := range n.Then {
		sb.WriteString(stmt.Format(indent + 2) + "\n")
	}
	if len(n.Else) > 0 {
		sb.WriteString(pad(indent+1) + "Else:\n")
		for _, stmt := range n.Else {
			sb.WriteString(stmt.Format(indent + 2) + "\n")
		}
	}
	return strings.TrimRight(sb.String(), "\n")
}

type WhileStmt struct {
	Cond Node
	Body []Node
}

func (n WhileStmt) Format(indent int) string {
	var sb strings.Builder
	sb.WriteString(pad(indent) + "While\n")
	sb.WriteString(pad(indent+1) + "Cond:\n")
	sb.WriteString(n.Cond.Format(indent + 2) + "\n")
	sb.WriteString(pad(indent+1) + "Body:\n")
	for _, stmt := range n.Body {
		sb.WriteString(stmt.Format(indent + 2) + "\n")
	}
	return strings.TrimRight(sb.String(), "\n")
}

type FuncDecl struct {
	Name   string
	Params []string
	Body   []Node
}

func (n FuncDecl) Format(indent int) string {
	var sb strings.Builder
	sb.WriteString(pad(indent) + "Func(" + n.Name + ")\n")
	sb.WriteString(pad(indent+1) + "Params: " + strings.Join(n.Params, ", ") + "\n")
	sb.WriteString(pad(indent+1) + "Body:\n")
	for _, stmt := range n.Body {
		sb.WriteString(stmt.Format(indent + 2) + "\n")
	}
	return strings.TrimRight(sb.String(), "\n")
}

// Program is a list of top-level statements.
type Program struct {
	Statements []Node
}

func (p Program) Format(indent int) string {
	var sb strings.Builder
	for _, stmt := range p.Statements {
		sb.WriteString(stmt.Format(indent) + "\n")
	}
	return sb.String()
}

// FormatProgram formats a program for display.
func FormatProgram(p *Program) string {
	return p.Format(0)
}

// pad returns indent * "  ".
func pad(indent int) string {
	return strings.Repeat("  ", indent)
}