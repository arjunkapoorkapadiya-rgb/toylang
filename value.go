package main

import (
	"fmt"
	"strconv"
)

// Value is any runtime value (number, string, bool, nil, function).
type Value interface {
	Type() string
	String() string
	Equal(other Value) bool
}

// NumberValue wraps a float64.
type NumberValue struct {
	V float64
}

func (n NumberValue) Type() string { return "number" }
func (n NumberValue) String() string {
	if n.V == float64(int64(n.V)) {
		return strconv.FormatInt(int64(n.V), 10)
	}
	return strconv.FormatFloat(n.V, 'g', -1, 64)
}
func (n NumberValue) Equal(o Value) bool {
	other, ok := o.(NumberValue)
	return ok && n.V == other.V
}

// StringValue wraps a string.
type StringValue struct {
	V string
}

func (s StringValue) Type() string          { return "string" }
func (s StringValue) String() string        { return s.V }
func (s StringValue) Equal(o Value) bool {
	other, ok := o.(StringValue)
	return ok && s.V == other.V
}

// BoolValue is true/false.
type BoolValue struct {
	V bool
}

func (b BoolValue) Type() string { return "bool" }
func (b BoolValue) String() string {
	if b.V {
		return "true"
	}
	return "false"
}
func (b BoolValue) Equal(o Value) bool {
	other, ok := o.(BoolValue)
	return ok && b.V == other.V
}

// NilValue represents nil.
type NilValue struct{}

func (NilValue) Type() string   { return "nil" }
func (NilValue) String() string { return "nil" }
func (NilValue) Equal(o Value) bool {
	_, ok := o.(NilValue)
	return ok
}

// FunctionValue is a compiled function.
type FunctionValue struct {
	Name     string
	Arity    int
	Chunk    *Chunk
}

func (f *FunctionValue) Type() string   { return "function" }
func (f *FunctionValue) String() string {
	return fmt.Sprintf("<fn %s/%d>", f.Name, f.Arity)
}
func (f *FunctionValue) Equal(o Value) bool {
	other, ok := o.(*FunctionValue)
	return ok && other == f
}

// Helpers

func numberVal(f float64) Value      { return NumberValue{V: f} }
func stringVal(s string) Value       { return StringValue{V: s} }
func boolVal(b bool) Value           { return BoolValue{V: b} }
func nilVal() Value                  { return NilValue{} }

// truthy returns whether a value is truthy (for if/while).
func truthy(v Value) bool {
	switch t := v.(type) {
	case NilValue:
		return false
	case BoolValue:
		return t.V
	case NumberValue:
		return t.V != 0
	}
	return true
}