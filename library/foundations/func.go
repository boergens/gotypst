// Func types for Typst.
// Translated from foundations/func.rs

package foundations

import (
	"github.com/boergens/gotypst/syntax"
)

// FuncValue represents a function.
type FuncValue struct {
	// Func is the underlying function.
	Func *Func
}

func (FuncValue) Type() Type         { return TypeFunc }
func (v FuncValue) Display() Content { return Content{} }
func (v FuncValue) Clone() Value     { return v } // Functions are immutable
func (FuncValue) isValue()           {}

// Func represents a callable function.
type Func struct {
	// Name is the optional function name.
	Name *string
	// Span is the source location.
	Span syntax.Span
	// Repr is the function representation.
	Repr FuncRepr
}

// FuncRepr represents different kinds of function implementations.
type FuncRepr interface {
	isFuncRepr()
}

// NativeFunc represents a built-in function implemented in Go.
// This matches Rust's NativeFuncSignature pattern where native functions
// receive Engine and Context explicitly.
type NativeFunc struct {
	// Func is the Go function implementing this native.
	// Receives Engine (world, route, sink) and Context (styles, location) explicitly.
	Func func(engine Engine, context Context, args *Args) (Value, error)
	// Info contains function metadata.
	Info *FuncInfo
	// Scope contains associated methods (e.g., table.cell).
	Scope *Scope
}

func (NativeFunc) isFuncRepr() {}

// ClosureFunc represents a user-defined closure.
type ClosureFunc struct {
	// Closure is the closure data.
	Closure *Closure
}

func (ClosureFunc) isFuncRepr() {}

// Closure represents a user-defined closure.
type Closure struct {
	// Node is the AST node for the closure (Closure or Contextual).
	Node ClosureNode

	// Defaults contains the default values for named parameters.
	Defaults []Value

	// Captured contains the captured variable bindings.
	Captured *Scope

	// NumPosParams is the number of positional parameters.
	NumPosParams int
}

// ClosureNode represents the AST node for a closure.
type ClosureNode interface {
	isClosureNode()
}

// ClosureAstNode wraps a closure AST node.
type ClosureAstNode struct {
	Node *syntax.SyntaxNode
}

func (ClosureAstNode) isClosureNode() {}

// ContextAstNode wraps a contextual expression AST node.
type ContextAstNode struct {
	Node *syntax.SyntaxNode
}

func (ContextAstNode) isClosureNode() {}

// WithFunc represents a function with modified properties.
type WithFunc struct {
	// Func is the wrapped function.
	Func *Func
	// Args are the pre-applied arguments.
	Args *Args
}

func (WithFunc) isFuncRepr() {}

// FuncInfo contains metadata about a function.
type FuncInfo struct {
	Name   string
	Params []ParamInfo
}

// ParamInfo describes a function parameter.
type ParamInfo struct {
	Name     string
	Type     Type
	Default  Value
	Variadic bool
	Named    bool
}

// Scope returns the function's associated scope, if any.
// Only native functions have scopes (containing associated methods).
func (f *Func) Scope() *Scope {
	if f == nil || f.Repr == nil {
		return nil
	}
	if nf, ok := f.Repr.(NativeFunc); ok {
		return nf.Scope
	}
	return nil
}
