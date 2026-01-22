// Virtual machine for evaluating Typst code.
// Translated from typst-eval/src/vm.rs

package eval

import (
	"github.com/boergens/gotypst/library/foundations"
	"github.com/boergens/gotypst/syntax"
)

// Vm is the virtual machine for evaluating Typst code.
//
// Holds the state needed to evaluate Typst sources. A new virtual machine
// is created for each module evaluation and function call.
//
// Matches Rust's Vm struct in vm.rs.
type Vm struct {
	// Engine is the underlying virtual typesetter.
	Engine *foundations.Engine

	// Flow is a control flow event that is currently happening.
	Flow FlowEvent

	// Scopes is the stack of scopes.
	Scopes *foundations.Scopes

	// Inspected is a span that is currently under inspection.
	Inspected *syntax.Span

	// Context provides data that is contextually made accessible to code.
	Context *foundations.Context
}

// NewVm creates a new virtual machine.
func NewVm(engine *foundations.Engine, context *foundations.Context, scopes *foundations.Scopes, target syntax.Span) *Vm {
	var inspected *syntax.Span
	if engine.Traced != nil {
		if id := target.Id(); id != nil {
			inspected = engine.Traced.Get(*id)
		}
	}
	return &Vm{
		Engine:    engine,
		Context:   context,
		Flow:      nil,
		Scopes:    scopes,
		Inspected: inspected,
	}
}

// World returns the underlying world.
func (vm *Vm) World() foundations.World {
	return vm.Engine.World
}

// Define binds a value to an identifier.
// This will create a Binding with the value and the identifier's span.
func (vm *Vm) Define(ident *syntax.IdentExpr, value foundations.Value) {
	vm.Bind(ident, foundations.NewBinding(value, ident.ToUntyped().Span()))
}

// Bind inserts a binding into the current scope.
// This will insert the value into the top-most scope and make it available
// for dynamic tracing, assisting IDE functionality.
func (vm *Vm) Bind(ident *syntax.IdentExpr, binding foundations.Binding) {
	span := ident.ToUntyped().Span()
	if vm.Inspected != nil && *vm.Inspected == span {
		vm.Trace(binding.Read())
	}

	// TODO: Warn if variable name is "is" (future keyword)

	vm.Scopes.Top().Insert(ident.Get(), binding)
}

// DefineSimple binds a value to an identifier name without an AST node.
func (vm *Vm) DefineSimple(name string, value foundations.Value) {
	vm.Scopes.Define(name, value, syntax.Span{})
}

// Trace records a value for IDE inspection.
func (vm *Vm) Trace(value foundations.Value) {
	if vm.Engine.Sink != nil {
		styles := vm.Context.Styles
		var stylesPtr *foundations.Styles
		if styles != nil {
			stylesPtr = styles.ToStyles()
		}
		vm.Engine.Sink.TraceValue(value, stylesPtr)
	}
}

// Get looks up a binding by name.
func (vm *Vm) Get(name string) *foundations.Binding {
	return vm.Scopes.Get(name)
}

// GetMut looks up a mutable binding by name.
func (vm *Vm) GetMut(name string) *foundations.Binding {
	return vm.Scopes.GetMut(name)
}

// GetInMath looks up a binding for a math identifier.
// Matches Rust: vm.scopes.get_in_math(&self)
func (vm *Vm) GetInMath(name string) *foundations.Binding {
	return vm.Scopes.GetInMath(name)
}

// EnterScope pushes a new scope onto the stack.
func (vm *Vm) EnterScope() {
	vm.Scopes.Enter()
}

// ExitScope pops the current scope from the stack.
func (vm *Vm) ExitScope() {
	vm.Scopes.Exit()
}

// HasFlow returns true if there is a pending flow event.
func (vm *Vm) HasFlow() bool {
	return vm.Flow != nil
}

// ClearFlow clears the current flow event.
func (vm *Vm) ClearFlow() {
	vm.Flow = nil
}

// SetFlow sets a flow event.
func (vm *Vm) SetFlow(flow FlowEvent) {
	vm.Flow = flow
}

// TakeFlow takes the current flow event and clears it.
func (vm *Vm) TakeFlow() FlowEvent {
	flow := vm.Flow
	vm.Flow = nil
	return flow
}

// MaxCallDepth is the maximum call stack depth before a stack overflow error.
const MaxCallDepth = 64

// callDepth tracks the current call stack depth (stored in Engine for sharing).
// For now we track it locally in a package variable since Engine doesn't have it yet.
var callDepth int

// CheckCallDepth checks if we've exceeded the maximum call depth.
// Returns an error if the call stack is too deep.
func (vm *Vm) CheckCallDepth() error {
	if callDepth >= MaxCallDepth {
		return &CallStackOverflowError{}
	}
	return nil
}

// EnterCall increments the call depth counter.
func (vm *Vm) EnterCall() {
	callDepth++
}

// ExitCall decrements the call depth counter.
func (vm *Vm) ExitCall() {
	if callDepth > 0 {
		callDepth--
	}
}

// CallStackOverflowError is returned when the call stack is too deep.
type CallStackOverflowError struct{}

func (e *CallStackOverflowError) Error() string {
	return "maximum function call depth exceeded"
}

