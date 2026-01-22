// Engine and related types for Typst.
// Translated from typst-library/src/engine.rs

package foundations

import (
	"github.com/boergens/gotypst/syntax"
)

// Engine holds all data needed during compilation.
// This matches Rust's Engine struct in engine.rs.
type Engine struct {
	// Routines provides implementation of various Typst compiler routines.
	// This is essentially dynamic linking to avoid circular dependencies.
	Routines Routines

	// World provides access to the compilation environment.
	World World

	// Route tracks the evaluation path for cycle detection.
	Route *Route

	// Sink collects warnings, delayed errors, and traced values.
	Sink *Sink

	// Traced tracks spans for IDE inspection.
	Traced *Traced
}

// NewEngine creates a new engine with the given world and routines.
func NewEngine(world World, routines Routines) *Engine {
	return &Engine{
		Routines: routines,
		World:    world,
		Route:    NewRoute(),
		Sink:     NewSink(),
		Traced:   nil,
	}
}

// ----------------------------------------------------------------------------
// Routines (Virtual Function Table)
// ----------------------------------------------------------------------------

// Routines defines implementation of various Typst compiler routines.
// This is essentially a vtable pattern to allow crate/package splitting
// without circular dependencies.
//
// Matches Rust's Routines struct in routines.rs.
type Routines interface {
	// EvalClosure calls a closure with the given context and arguments.
	EvalClosure(engine *Engine, context *Context, fn *Func, closure *Closure, args *Args) (Value, error)
}

// ----------------------------------------------------------------------------
// Route (Cycle Detection)
// ----------------------------------------------------------------------------

// Route tracks the evaluation path for detecting cyclic imports and excessive nesting.
// Matches Rust's Route struct in engine.rs.
type Route struct {
	// files contains the file IDs currently being evaluated.
	files []syntax.FileId
	// len is the nesting depth for this route segment.
	len int
}

// MaxShowRuleDepth is the maximum show rule nesting depth.
const MaxShowRuleDepth = 64

// MaxLayoutDepth is the maximum layout nesting depth.
const MaxLayoutDepth = 72

// MaxCallDepth is the maximum function call nesting depth.
const MaxCallDepth = 80

// NewRoute creates a new empty route.
func NewRoute() *Route {
	return &Route{files: nil, len: 0}
}

// CheckShowDepth ensures we are within the maximum show rule depth.
func (r *Route) CheckShowDepth() error {
	if r.len > MaxShowRuleDepth {
		return &DepthExceededError{Kind: "show rule", Depth: r.len, Max: MaxShowRuleDepth}
	}
	return nil
}

// CheckLayoutDepth ensures we are within the maximum layout depth.
func (r *Route) CheckLayoutDepth() error {
	if r.len > MaxLayoutDepth {
		return &DepthExceededError{Kind: "layout", Depth: r.len, Max: MaxLayoutDepth}
	}
	return nil
}

// CheckCallDepth ensures we are within the maximum function call depth.
func (r *Route) CheckCallDepth() error {
	if r.len > MaxCallDepth {
		return &DepthExceededError{Kind: "function call", Depth: r.len, Max: MaxCallDepth}
	}
	return nil
}

// Increase increments the nesting depth.
func (r *Route) Increase() {
	r.len++
}

// Decrease decrements the nesting depth.
func (r *Route) Decrease() {
	r.len--
}

// Len returns the current nesting depth.
func (r *Route) Len() int {
	return r.len
}

// Contains checks if a file is already in the route.
func (r *Route) Contains(id syntax.FileId) bool {
	for _, f := range r.files {
		if f == id {
			return true
		}
	}
	return false
}

// Push adds a file to the route.
func (r *Route) Push(id syntax.FileId) {
	r.files = append(r.files, id)
}

// Pop removes the last file from the route.
func (r *Route) Pop() {
	if len(r.files) > 0 {
		r.files = r.files[:len(r.files)-1]
	}
}

// CurrentFile returns the current file being evaluated, or nil.
func (r *Route) CurrentFile() *syntax.FileId {
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
		files: make([]syntax.FileId, len(r.files)),
		len:   r.len,
	}
	copy(clone.files, r.files)
	return clone
}

// ----------------------------------------------------------------------------
// Sink (Warning/Trace Collection)
// ----------------------------------------------------------------------------

// Sink is a push-only sink for recorded introspections, delayed errors,
// warnings, and traced values.
// Matches Rust's Sink struct in engine.rs.
type Sink struct {
	// Delayed contains delayed errors that can be ignored until the last iteration.
	Delayed []SourceDiagnostic

	// Warnings contains warnings emitted during compilation.
	Warnings []SourceDiagnostic

	// Values contains traced values for IDE inspection.
	Values []TracedValue
}

// MaxTracedValues is the maximum number of traced values to store.
const MaxTracedValues = 10

// NewSink creates a new empty sink.
func NewSink() *Sink {
	return &Sink{
		Delayed:  nil,
		Warnings: nil,
		Values:   nil,
	}
}

// Delay pushes delayed errors.
func (s *Sink) Delay(errors []SourceDiagnostic) {
	s.Delayed = append(s.Delayed, errors...)
}

// Warn adds a warning to the sink.
func (s *Sink) Warn(warning SourceDiagnostic) {
	// TODO: Deduplicate warnings by hash of span+message
	s.Warnings = append(s.Warnings, warning)
}

// TraceValue records a traced value and optional styles.
func (s *Sink) TraceValue(value Value, styles *Styles) {
	if len(s.Values) < MaxTracedValues {
		s.Values = append(s.Values, TracedValue{Value: value, Styles: styles})
	}
}

// TakeDelayed returns and clears the delayed errors.
func (s *Sink) TakeDelayed() []SourceDiagnostic {
	delayed := s.Delayed
	s.Delayed = nil
	return delayed
}

// TracedValue represents a value traced for IDE inspection.
type TracedValue struct {
	Value  Value
	Styles *Styles
}

// ----------------------------------------------------------------------------
// Traced (IDE Support)
// ----------------------------------------------------------------------------

// Traced may hold a span that is currently under inspection.
// Matches Rust's Traced struct in engine.rs.
type Traced struct {
	span *syntax.Span
}

// NewTraced creates a Traced with the given span.
func NewTraced(span syntax.Span) *Traced {
	return &Traced{span: &span}
}

// Get returns the traced span if it is part of the given source file.
// The span is hidden if it isn't in the given file so that only results for
// the file with the traced span are invalidated.
func (t *Traced) Get(id syntax.FileId) *syntax.Span {
	if t == nil || t.span == nil {
		return nil
	}
	// Check if the span's file ID matches
	spanID := t.span.Id()
	if spanID != nil && *spanID == id {
		return t.span
	}
	return nil
}

// ----------------------------------------------------------------------------
// World Interface
// ----------------------------------------------------------------------------

// World provides access to the external environment during compilation.
// Matches Rust's World trait in lib.rs.
type World interface {
	// Library returns the standard library scope.
	Library() *Scope

	// MainFile returns the main source file ID.
	MainFile() syntax.FileId

	// Source returns the source content for a file.
	Source(id syntax.FileId) (*syntax.Source, error)

	// File returns the raw bytes of a file.
	File(id syntax.FileId) ([]byte, error)

	// Today returns the current date.
	Today(offset *int) *Datetime
}

// ----------------------------------------------------------------------------
// Diagnostics
// ----------------------------------------------------------------------------

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
// Location (Introspection)
// ----------------------------------------------------------------------------

// Location represents a location in the document for introspection.
type Location struct {
	// Page is the current page number.
	Page int

	// Position is the position on the page.
	Position Point
}

// Point represents a position on a page.
type Point struct {
	X, Y Length
}

// ----------------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------------

// DepthExceededError is returned when a nesting depth limit is exceeded.
type DepthExceededError struct {
	Kind  string
	Depth int
	Max   int
}

func (e *DepthExceededError) Error() string {
	return "maximum " + e.Kind + " depth exceeded"
}
