package eval

import (
	"github.com/boergens/gotypst/syntax"
)

// MaxCallDepth is the maximum allowed call stack depth.
const MaxCallDepth = 256

// Vm is the virtual machine for evaluating Typst code.
//
// The VM holds all state needed during evaluation, including variable scopes,
// control flow state, and access to the underlying engine and world.
type Vm struct {
	// Engine provides access to the typesetter and world.
	Engine *Engine

	// Flow holds the current control flow event (break/continue/return).
	// This is nil during normal execution and set when a flow event occurs.
	Flow FlowEvent

	// Scopes is the stack of variable scopes.
	Scopes *Scopes

	// Inspected is the span being traced for IDE inspection.
	Inspected *syntax.Span

	// Context provides contextual data during evaluation.
	Context *Context

	// callDepth tracks the current function call depth.
	callDepth int

	// rootSpan is the span of the root evaluation.
	rootSpan syntax.Span
}

// NewVm creates a new VM with the given engine, context, scopes, and root span.
func NewVm(engine *Engine, ctx *Context, scopes *Scopes, rootSpan syntax.Span) *Vm {
	return &Vm{
		Engine:    engine,
		Flow:      nil,
		Scopes:    scopes,
		Inspected: nil,
		Context:   ctx,
		callDepth: 0,
		rootSpan:  rootSpan,
	}
}

// Define binds a value to an identifier in the current scope.
func (vm *Vm) Define(name string, value Value) {
	vm.Scopes.Define(name, value, vm.rootSpan)
}

// DefineWithSpan binds a value to an identifier with a specific span.
func (vm *Vm) DefineWithSpan(name string, value Value, span syntax.Span) {
	vm.Scopes.Define(name, value, span)
}

// Bind inserts a binding into the current scope.
func (vm *Vm) Bind(name string, binding Binding) {
	vm.Scopes.Insert(name, binding)
}

// Get looks up a binding by name.
func (vm *Vm) Get(name string) *Binding {
	return vm.Scopes.Get(name)
}

// GetMut looks up a mutable binding by name.
func (vm *Vm) GetMut(name string) *Binding {
	return vm.Scopes.GetMut(name)
}

// Trace records a value for IDE inspection if the given span matches.
func (vm *Vm) Trace(value Value, span syntax.Span) {
	if vm.Inspected != nil && *vm.Inspected == span {
		// TODO: Record the traced value for IDE integration
	}
}

// World returns the world from the engine.
func (vm *Vm) World() World {
	return vm.Engine.World
}

// CheckCallDepth checks if the call depth limit has been exceeded.
// Returns an error if the limit is exceeded.
func (vm *Vm) CheckCallDepth() error {
	if vm.callDepth >= MaxCallDepth {
		return &CallDepthExceededError{Depth: vm.callDepth}
	}
	return nil
}

// EnterCall increments the call depth.
func (vm *Vm) EnterCall() {
	vm.callDepth++
}

// ExitCall decrements the call depth.
func (vm *Vm) ExitCall() {
	vm.callDepth--
}

// CallDepth returns the current call depth.
func (vm *Vm) CallDepth() int {
	return vm.callDepth
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

// ----------------------------------------------------------------------------
// Engine
// ----------------------------------------------------------------------------

// Engine provides access to the typesetter and external resources.
type Engine struct {
	// World provides access to files, packages, and fonts.
	World World

	// Route tracks the evaluation path for cycle detection.
	Route *Route

	// Sink collects warnings and traced values.
	Sink *Sink

	// Traced tracks spans for IDE inspection.
	Traced *Traced

	// Routines provides access to introspection routines.
	Routines *Routines
}

// NewEngine creates a new engine with the given world.
func NewEngine(world World) *Engine {
	return &Engine{
		World:    world,
		Route:    NewRoute(),
		Sink:     NewSink(),
		Traced:   nil,
		Routines: nil,
	}
}

// ----------------------------------------------------------------------------
// World Interface
// ----------------------------------------------------------------------------

// World provides access to the external environment during evaluation.
type World interface {
	// Library returns the standard library scope.
	Library() *Scope

	// MainFile returns the main source file ID.
	MainFile() FileID

	// Source returns the source content for a file.
	Source(id FileID) (*syntax.Source, error)

	// File returns the raw bytes of a file.
	File(id FileID) ([]byte, error)

	// Today returns the current date.
	Today(offset *int) Date
}

// FileID uniquely identifies a file.
type FileID struct {
	// Package is the optional package specification.
	Package *PackageSpec

	// Path is the file path within the package or project.
	Path string
}

// PackageSpec identifies a package.
type PackageSpec struct {
	Namespace string
	Name      string
	Version   Version
}

// Version represents a semantic version.
type Version struct {
	Major int
	Minor int
	Patch int
}

// Date represents a date value.
type Date struct {
	Year  int
	Month int
	Day   int
}

// ----------------------------------------------------------------------------
// Context
// ----------------------------------------------------------------------------

// Context provides contextual data during evaluation.
type Context struct {
	// Styles are the currently active styles.
	Styles *Styles

	// Location is the current location for introspection.
	Location *Location
}

// NewContext creates a new empty context.
func NewContext() *Context {
	return &Context{
		Styles:   nil,
		Location: nil,
	}
}

// Location represents a location in the document for introspection.
type Location struct {
	// Page is the current page number.
	Page int

	// Position is the position on the page.
	Position Position
}

// Position represents a position on a page.
type Position struct {
	X, Y Length
}

// ----------------------------------------------------------------------------
// Route (Cycle Detection)
// ----------------------------------------------------------------------------

// Route tracks the evaluation path for detecting cyclic imports.
type Route struct {
	// files contains the file IDs currently being evaluated.
	files []FileID
}

// NewRoute creates a new empty route.
func NewRoute() *Route {
	return &Route{files: nil}
}

// Contains checks if a file is already in the route.
func (r *Route) Contains(id FileID) bool {
	for _, f := range r.files {
		if f.Path == id.Path {
			// TODO: Also compare Package
			return true
		}
	}
	return false
}

// Push adds a file to the route.
func (r *Route) Push(id FileID) {
	r.files = append(r.files, id)
}

// Pop removes the last file from the route.
func (r *Route) Pop() {
	if len(r.files) > 0 {
		r.files = r.files[:len(r.files)-1]
	}
}

// Clone creates a copy of the route.
func (r *Route) Clone() *Route {
	if r == nil {
		return nil
	}
	clone := &Route{files: make([]FileID, len(r.files))}
	copy(clone.files, r.files)
	return clone
}

// ----------------------------------------------------------------------------
// Sink (Warning/Trace Collection)
// ----------------------------------------------------------------------------

// Sink collects warnings and traced values during evaluation.
type Sink struct {
	// Warnings contains collected warnings.
	Warnings []SourceDiagnostic

	// TracedValues contains values traced for IDE inspection.
	TracedValues []TracedValue
}

// NewSink creates a new empty sink.
func NewSink() *Sink {
	return &Sink{
		Warnings:     nil,
		TracedValues: nil,
	}
}

// Warn adds a warning to the sink.
func (s *Sink) Warn(warning SourceDiagnostic) {
	s.Warnings = append(s.Warnings, warning)
}

// TraceValue records a traced value.
func (s *Sink) TraceValue(value TracedValue) {
	s.TracedValues = append(s.TracedValues, value)
}

// TracedValue represents a value traced for IDE inspection.
type TracedValue struct {
	Span  syntax.Span
	Value Value
}

// SourceDiagnostic represents a diagnostic message with source location.
type SourceDiagnostic struct {
	// Span is the source location.
	Span syntax.Span

	// Severity indicates the severity level.
	Severity DiagnosticSeverity

	// Message is the diagnostic message.
	Message string

	// Hints are optional hints for resolving the issue.
	Hints []string
}

// DiagnosticSeverity indicates the severity of a diagnostic.
type DiagnosticSeverity int

const (
	SeverityError DiagnosticSeverity = iota
	SeverityWarning
)

// ----------------------------------------------------------------------------
// Traced (IDE Support)
// ----------------------------------------------------------------------------

// Traced tracks spans for IDE inspection.
type Traced struct {
	// Spans are the spans being traced.
	Spans []syntax.Span
}

// Routines provides access to introspection routines.
type Routines struct {
	// TODO: Add introspection routine fields
}

// ----------------------------------------------------------------------------
// Closure
// ----------------------------------------------------------------------------

// Closure represents a user-defined closure with captured variables.
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

// ----------------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------------

// CallDepthExceededError is returned when the call stack depth limit is exceeded.
type CallDepthExceededError struct {
	Depth int
}

func (e *CallDepthExceededError) Error() string {
	return "maximum call depth exceeded"
}

// UndefinedVariableError is returned when an undefined variable is accessed.
type UndefinedVariableError struct {
	Name string
	Span syntax.Span
}

func (e *UndefinedVariableError) Error() string {
	return "unknown variable: " + e.Name
}

// CyclicImportError is returned when a cyclic import is detected.
type CyclicImportError struct {
	File FileID
	Span syntax.Span
}

func (e *CyclicImportError) Error() string {
	return "cyclic import detected: " + e.File.Path
}
