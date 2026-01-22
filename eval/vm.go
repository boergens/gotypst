package eval

import (
	"github.com/boergens/gotypst/library/foundations"
	"github.com/boergens/gotypst/syntax"
)

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
func (vm *Vm) WorldInternal() WorldInternal {
	return vm.Engine.world
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
	// world provides access to files, packages, and fonts.
	world WorldInternal

	// route tracks the evaluation path for cycle detection.
	route *Route

	// sink collects warnings and traced values.
	sink *Sink

	// traced tracks spans for IDE inspection.
	traced *Traced

	// routines provides access to introspection routines.
	routines *Routines
}

// NewEngine creates a new engine with the given world.
func NewEngine(world WorldInternal) *Engine {
	return &Engine{
		world:    world,
		route:    NewRoute(),
		sink:     NewSink(),
		traced:   nil,
		routines: nil,
	}
}

// callFuncInternal calls a function with eval-specific types.
// This is the internal implementation used by the VM.
func (e *Engine) callFuncInternal(context *Context, callee Value, args *Args, span syntax.Span) (Value, error) {
	fn, ok := AsFunc(callee)
	if !ok {
		return nil, &InvalidCalleeError{Value: callee, Span: span}
	}

	// Check and update call depth
	if err := e.route.CheckCallDepth(); err != nil {
		return nil, err
	}
	e.route.EnterCall()
	defer e.route.ExitCall()

	// Dispatch based on function representation
	switch repr := fn.Repr.(type) {
	case NativeFunc:
		return repr.Func(e, context, args)

	case ClosureFunc:
		// Create a temporary Vm for closure evaluation
		scopes := NewScopes(e.world.Library())
		if repr.Closure.Captured != nil {
			scopes.SetTop(repr.Closure.Captured.Clone())
		}
		vm := NewVm(e, context, scopes, fn.Span)
		return callClosure(vm, fn, repr.Closure, args)

	case WithFunc:
		// Merge pre-applied args with new args
		merged := &Args{Span: args.Span}
		merged.Items = append(merged.Items, repr.Args.Items...)
		merged.Items = append(merged.Items, args.Items...)
		return e.callFuncInternal(context, FuncValue{Func: repr.Func}, merged, span)

	default:
		return nil, &InvalidCalleeError{Value: callee, Span: span}
	}
}

// ----------------------------------------------------------------------------
// foundations.Engine Interface Implementation
// ----------------------------------------------------------------------------

// Verify Engine implements foundations.Engine interface.
var _ foundations.Engine = (*Engine)(nil)

// World returns the world as a foundations.World interface.
func (e *Engine) World() foundations.World {
	return &worldAdapter{e.world}
}

// Route returns the route as a foundations.Route interface.
func (e *Engine) Route() foundations.Route {
	return &routeAdapter{e.route}
}

// Sink returns the sink as a foundations.Sink interface.
func (e *Engine) Sink() foundations.Sink {
	return &sinkAdapter{e.sink}
}

// sinkAdapter wraps eval.Sink to implement foundations.Sink.
type sinkAdapter struct {
	s *Sink
}

func (a *sinkAdapter) Warn(warning foundations.SourceDiagnostic) {
	a.s.Warn(SourceDiagnostic{
		Span:     warning.Span,
		Severity: DiagnosticSeverity(warning.Severity),
		Message:  warning.Message,
		Hints:    warning.Hints,
	})
}

// routeAdapter wraps eval.Route to implement foundations.Route.
type routeAdapter struct {
	r *Route
}

func (a *routeAdapter) CheckCallDepth() error {
	return a.r.CheckCallDepth()
}

func (a *routeAdapter) EnterCall() {
	a.r.EnterCall()
}

func (a *routeAdapter) ExitCall() {
	a.r.ExitCall()
}

func (a *routeAdapter) Contains(id foundations.FileID) bool {
	return a.r.Contains(FileID{Path: id.Path, Package: convertPackageSpec(id.Package)})
}

func (a *routeAdapter) Push(id foundations.FileID) {
	a.r.Push(FileID{Path: id.Path, Package: convertPackageSpec(id.Package)})
}

func (a *routeAdapter) Pop() {
	a.r.Pop()
}

func (a *routeAdapter) CurrentFile() *foundations.FileID {
	f := a.r.CurrentFile()
	if f == nil {
		return nil
	}
	return &foundations.FileID{Path: f.Path}
}

// convertPackageSpec converts foundations.PackageSpec to eval.PackageSpec.
func convertPackageSpec(spec *foundations.PackageSpec) *PackageSpec {
	if spec == nil {
		return nil
	}
	return &PackageSpec{
		Namespace: spec.Namespace,
		Name:      spec.Name,
		Version:   Version{Major: spec.Version.Major, Minor: spec.Version.Minor, Patch: spec.Version.Patch},
	}
}

// CallFunc calls a function with foundations interface types.
// This implements foundations.Engine.CallFunc().
func (e *Engine) CallFunc(context foundations.Context, callee foundations.Value, args *foundations.Args, span syntax.Span) (foundations.Value, error) {
	// Convert foundations types to eval types
	ctx := contextFromFoundations(context)
	evalArgs := argsFromFoundations(args)
	evalCallee := valueFromFoundations(callee)

	result, err := e.callFuncInternal(ctx, evalCallee, evalArgs, span)
	if err != nil {
		return nil, err
	}
	return valueToFoundations(result), nil
}

// worldAdapter wraps an eval.WorldInternal to implement foundations.World.
type worldAdapter struct {
	w WorldInternal
}

func (a *worldAdapter) Library() *foundations.Scope {
	// For now, return nil - full integration requires scope migration
	return nil
}

func (a *worldAdapter) MainFile() foundations.FileID {
	f := a.w.MainFile()
	return foundations.FileID{Path: f.Path}
}

func (a *worldAdapter) Source(id foundations.FileID) (*syntax.Source, error) {
	return a.w.Source(FileID{Path: id.Path})
}

func (a *worldAdapter) File(id foundations.FileID) ([]byte, error) {
	return a.w.File(FileID{Path: id.Path})
}

func (a *worldAdapter) Today(offset *int) foundations.Date {
	d := a.w.Today(offset)
	return foundations.Date{Year: d.Year, Month: d.Month, Day: d.Day}
}

// contextFromFoundations converts foundations.Context to *Context.
func contextFromFoundations(ctx foundations.Context) *Context {
	if ctx == nil {
		return nil
	}
	// For now, create empty context - full integration requires styles migration
	return &Context{}
}

// argsFromFoundations converts *foundations.Args to *Args.
func argsFromFoundations(args *foundations.Args) *Args {
	if args == nil {
		return nil
	}
	// For now, create minimal conversion
	result := &Args{Span: args.Span}
	for _, item := range args.Items {
		result.Items = append(result.Items, Arg{
			Span:  item.Span,
			Name:  item.Name,
			Value: syntax.Spanned[Value]{V: valueFromFoundations(item.Value.V), Span: item.Value.Span},
		})
	}
	return result
}

// valueFromFoundations converts foundations.Value to eval.Value.
// This is a temporary adapter during migration.
func valueFromFoundations(v foundations.Value) Value {
	if v == nil {
		return nil
	}
	// Type switch on common types
	switch val := v.(type) {
	case foundations.NoneValue:
		return NoneValue{}
	case foundations.AutoValue:
		return AutoValue{}
	case foundations.Bool:
		return BoolValue(val)
	case foundations.Int:
		return IntValue(val)
	case foundations.Float:
		return FloatValue(val)
	case foundations.Str:
		return StrValue(val)
	default:
		// Wrap unknown types in DynValue
		return DynValue{Inner: v, TypeName: v.Type().String()}
	}
}

// valueToFoundations converts eval.Value to foundations.Value.
// This is a temporary adapter during migration.
func valueToFoundations(v Value) foundations.Value {
	if v == nil {
		return nil
	}
	// Type switch on common types
	switch val := v.(type) {
	case NoneValue:
		return foundations.NoneValue{}
	case AutoValue:
		return foundations.AutoValue{}
	case BoolValue:
		return foundations.Bool(val)
	case IntValue:
		return foundations.Int(val)
	case FloatValue:
		return foundations.Float(val)
	case StrValue:
		return foundations.Str(val)
	default:
		// Wrap unknown types in DynValue
		return foundations.DynValue{Inner: v, TypeName: v.Type().String()}
	}
}

// ----------------------------------------------------------------------------
// World Interface
// ----------------------------------------------------------------------------

// WorldInternal provides access to the external environment during evaluation.
// This is the internal interface using eval types. The Engine.World() method
// returns a foundations.World adapter for external use.
type WorldInternal interface {
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

// Verify Context implements foundations.Context interface.
var _ foundations.Context = (*Context)(nil)

// Context provides contextual data during evaluation.
type Context struct {
	// styles are the currently active styles.
	styles *Styles

	// location is the current location for introspection.
	location *LocationInternal
}

// NewContext creates a new empty context.
func NewContext() *Context {
	return &Context{
		styles:   nil,
		location: nil,
	}
}

// Styles returns the currently active styles as foundations.Styles.
// Implements foundations.Context.Styles().
func (c *Context) Styles() *foundations.Styles {
	if c == nil || c.styles == nil {
		return nil
	}
	// For now, return nil - full integration requires styles migration
	return nil
}

// Location returns the current location as foundations.Location.
// Implements foundations.Context.Location().
func (c *Context) Location() *foundations.Location {
	if c == nil || c.location == nil {
		return nil
	}
	return &foundations.Location{
		Page: c.location.Page,
		Position: foundations.Point{
			X: foundations.Length{Points: c.location.Position.X.Points},
			Y: foundations.Length{Points: c.location.Position.Y.Points},
		},
	}
}

// LocationInternal represents a location in the document for introspection.
type LocationInternal struct {
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

// Route tracks the evaluation path for detecting cyclic imports and call depth.
// This matches Rust's Route which tracks both file cycles and call depth.
type Route struct {
	// files contains the file IDs currently being evaluated.
	files []FileID
	// callDepth tracks the current function call depth.
	callDepth int
}

// MaxCallDepth is the maximum allowed call stack depth.
const MaxCallDepth = 256

// NewRoute creates a new empty route.
func NewRoute() *Route {
	return &Route{files: nil, callDepth: 0}
}

// CheckCallDepth checks if the call depth limit has been exceeded.
// Returns an error if the limit is exceeded.
func (r *Route) CheckCallDepth() error {
	if r.callDepth >= MaxCallDepth {
		return &CallDepthExceededError{Depth: r.callDepth}
	}
	return nil
}

// EnterCall increments the call depth.
func (r *Route) EnterCall() {
	r.callDepth++
}

// ExitCall decrements the call depth.
func (r *Route) ExitCall() {
	r.callDepth--
}

// CallDepth returns the current call depth.
func (r *Route) CallDepth() int {
	return r.callDepth
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

// CurrentFile returns the current file being evaluated, or nil.
func (r *Route) CurrentFile() *FileID {
	if r == nil || len(r.files) == 0 {
		return nil
	}
	return &r.files[len(r.files)-1]
}

// Clone creates a copy of the route.
func (r *Route) Clone() *Route {
	if r == nil {
		return nil
	}
	clone := &Route{
		files:     make([]FileID, len(r.files)),
		callDepth: r.callDepth,
	}
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
