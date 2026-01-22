// Package foundations provides core types and operations for the Typst runtime.
//
// This package contains the foundational value types that form the basis of
// the Typst language. It corresponds to typst-library/src/foundations/ in
// the Rust implementation.
package foundations

import (
	"fmt"
	"math/big"

	"github.com/boergens/gotypst/syntax"
)

// Value represents a runtime value in the Typst evaluator.
//
// This is a sum type interface - each value kind has a separate concrete type.
// The interface provides common operations that all values support.
type Value interface {
	// Type returns the type of this value.
	Type() Type

	// Display returns the display representation as Content.
	Display() Content

	// Clone creates a shallow copy of the value.
	Clone() Value

	// isValue is a marker method to seal the interface.
	isValue()
}

// ----------------------------------------------------------------------------
// Engine and Context Interfaces
// ----------------------------------------------------------------------------

// Engine provides access to the compilation environment during evaluation.
// This interface matches Rust's Engine struct pattern, allowing native functions
// to access world, route, sink, and call other functions.
//
// The concrete implementation lives in the eval package.
type Engine interface {
	// World returns the compilation environment (files, packages, fonts).
	World() World

	// Route returns the evaluation route for cycle detection and depth tracking.
	Route() Route

	// Sink returns the sink for collecting warnings and traced values.
	Sink() Sink

	// CallFunc calls a function with the given context and arguments.
	// This is the primary way for native functions to call other functions.
	CallFunc(context Context, callee Value, args *Args, span syntax.Span) (Value, error)
}

// Context provides contextual data during evaluation.
// This matches Rust's Context struct which holds location and styles.
type Context interface {
	// Styles returns the currently active styles, or nil if not in context.
	Styles() *Styles

	// Location returns the current location for introspection, or nil if not available.
	Location() *Location
}

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

// Route tracks the evaluation path for detecting cyclic imports and call depth.
type Route interface {
	// CheckCallDepth checks if the call depth limit has been exceeded.
	CheckCallDepth() error

	// EnterCall increments the call depth.
	EnterCall()

	// ExitCall decrements the call depth.
	ExitCall()

	// Contains checks if a file is already in the route.
	Contains(id FileID) bool

	// Push adds a file to the route.
	Push(id FileID)

	// Pop removes the last file from the route.
	Pop()

	// CurrentFile returns the current file being evaluated, or nil.
	CurrentFile() *FileID
}

// Sink collects warnings and traced values during evaluation.
type Sink interface {
	// Warn adds a warning to the sink.
	Warn(warning SourceDiagnostic)
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

// Version represents a semantic version (for packages).
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
// Primitive Values
// ----------------------------------------------------------------------------

// NoneValue represents the absence of a meaningful value.
type NoneValue struct{}

func (NoneValue) Type() Type       { return TypeNone }
func (NoneValue) Display() Content { return Content{} }
func (NoneValue) Clone() Value     { return NoneValue{} }
func (NoneValue) isValue()         {}

// None is the singleton none value.
var None = NoneValue{}

// AutoValue represents a value that is automatically determined.
type AutoValue struct{}

func (AutoValue) Type() Type       { return TypeAuto }
func (AutoValue) Display() Content { return Content{} }
func (AutoValue) Clone() Value     { return AutoValue{} }
func (AutoValue) isValue()         {}

// Auto is the singleton auto value.
var Auto = AutoValue{}

// Bool represents a boolean value.
type Bool bool

func (Bool) Type() Type         { return TypeBool }
func (v Bool) Display() Content { return Content{} }
func (v Bool) Clone() Value     { return v }
func (Bool) isValue()           {}

// True and False are the boolean singleton values.
var (
	True  = Bool(true)
	False = Bool(false)
)


// Int represents a 64-bit signed integer.
type Int int64

func (Int) Type() Type         { return TypeInt }
func (v Int) Display() Content { return Content{} }
func (v Int) Clone() Value     { return v }
func (Int) isValue()           {}

// String returns the string representation of the integer.
func (v Int) String() string {
	return fmt.Sprintf("%d", v)
}

// Float represents a 64-bit floating point number.
type Float float64

func (Float) Type() Type         { return TypeFloat }
func (v Float) Display() Content { return Content{} }
func (v Float) Clone() Value     { return v }
func (Float) isValue()           {}


// ----------------------------------------------------------------------------
// Measurement Values
// ----------------------------------------------------------------------------

// Length represents a physical length value.
type Length struct {
	// Points is the length in typographic points (1/72 inch).
	Points float64
}

// LengthValue represents a length as a Value.
type LengthValue struct {
	Length Length
}

func (LengthValue) Type() Type         { return TypeLength }
func (v LengthValue) Display() Content { return Content{} }
func (v LengthValue) Clone() Value     { return v }
func (LengthValue) isValue()           {}

// Angle represents an angle value.
type Angle struct {
	// Radians is the angle in radians.
	Radians float64
}

// AngleValue represents an angle as a Value.
type AngleValue struct {
	Angle Angle
}

func (AngleValue) Type() Type         { return TypeAngle }
func (v AngleValue) Display() Content { return Content{} }
func (v AngleValue) Clone() Value     { return v }
func (AngleValue) isValue()           {}

// Ratio represents a ratio (percentage) value.
type Ratio struct {
	// Value is the ratio as a fraction (0.5 = 50%).
	Value float64
}

// RatioValue represents a ratio as a Value.
type RatioValue struct {
	Ratio Ratio
}

func (RatioValue) Type() Type         { return TypeRatio }
func (v RatioValue) Display() Content { return Content{} }
func (v RatioValue) Clone() Value     { return v }
func (RatioValue) isValue()           {}

// Relative represents a combination of absolute length and ratio.
type Relative struct {
	// Abs is the absolute component.
	Abs Length
	// Rel is the relative component.
	Rel Ratio
}

// RelativeValue represents a relative length as a Value.
type RelativeValue struct {
	Relative Relative
}

func (RelativeValue) Type() Type         { return TypeRelative }
func (v RelativeValue) Display() Content { return Content{} }
func (v RelativeValue) Clone() Value     { return v }
func (RelativeValue) isValue()           {}

// Fraction represents a fraction of remaining space.
type Fraction struct {
	// Value is the number of fractions (1fr = 1.0).
	Value float64
}

// FractionValue represents a fraction as a Value.
type FractionValue struct {
	Fraction Fraction
}

func (FractionValue) Type() Type         { return TypeFraction }
func (v FractionValue) Display() Content { return Content{} }
func (v FractionValue) Clone() Value     { return v }
func (FractionValue) isValue()           {}

// ----------------------------------------------------------------------------
// Data Values
// ----------------------------------------------------------------------------

// Str represents a string value.
type Str string

func (Str) Type() Type         { return TypeStr }
func (v Str) Display() Content { return Content{} }
func (v Str) Clone() Value     { return v }
func (Str) isValue()           {}


// BytesValue represents a sequence of bytes.
type BytesValue []byte

func (BytesValue) Type() Type         { return TypeBytes }
func (v BytesValue) Display() Content { return Content{} }
func (v BytesValue) Clone() Value {
	if v == nil {
		return BytesValue(nil)
	}
	clone := make([]byte, len(v))
	copy(clone, v)
	return BytesValue(clone)
}
func (BytesValue) isValue() {}

// LabelValue represents a label for referencing content.
type LabelValue string

func (LabelValue) Type() Type         { return TypeLabel }
func (v LabelValue) Display() Content { return Content{} }
func (v LabelValue) Clone() Value     { return v }
func (LabelValue) isValue()           {}

// Datetime represents a date, time, or datetime value.
// Components are optional: nil means not specified.
// This matches the Rust Datetime type.
type Datetime struct {
	year   *int
	month  *int
	day    *int
	hour   *int
	minute *int
	second *int
}

func (*Datetime) Type() Type         { return TypeDatetime }
func (v *Datetime) Display() Content { return Content{} }
func (v *Datetime) Clone() Value     { return v } // Shallow clone is fine for immutable data
func (*Datetime) isValue()           {}

// Year returns the year component or nil if not set.
func (dt *Datetime) Year() *int {
	if dt == nil {
		return nil
	}
	return dt.year
}

// YearOr returns the year component or a default value.
func (dt *Datetime) YearOr(def int) int {
	if dt == nil || dt.year == nil {
		return def
	}
	return *dt.year
}

// Month returns the month component or nil if not set.
func (dt *Datetime) Month() *int {
	if dt == nil {
		return nil
	}
	return dt.month
}

// MonthOr returns the month component or a default value.
func (dt *Datetime) MonthOr(def int) int {
	if dt == nil || dt.month == nil {
		return def
	}
	return *dt.month
}

// Day returns the day component or nil if not set.
func (dt *Datetime) Day() *int {
	if dt == nil {
		return nil
	}
	return dt.day
}

// DayOr returns the day component or a default value.
func (dt *Datetime) DayOr(def int) int {
	if dt == nil || dt.day == nil {
		return def
	}
	return *dt.day
}

// Hour returns the hour component or nil if not set.
func (dt *Datetime) Hour() *int {
	if dt == nil {
		return nil
	}
	return dt.hour
}

// HourOr returns the hour component or a default value.
func (dt *Datetime) HourOr(def int) int {
	if dt == nil || dt.hour == nil {
		return def
	}
	return *dt.hour
}

// Minute returns the minute component or nil if not set.
func (dt *Datetime) Minute() *int {
	if dt == nil {
		return nil
	}
	return dt.minute
}

// MinuteOr returns the minute component or a default value.
func (dt *Datetime) MinuteOr(def int) int {
	if dt == nil || dt.minute == nil {
		return def
	}
	return *dt.minute
}

// Second returns the second component or nil if not set.
func (dt *Datetime) Second() *int {
	if dt == nil {
		return nil
	}
	return dt.second
}

// SecondOr returns the second component or a default value.
func (dt *Datetime) SecondOr(def int) int {
	if dt == nil || dt.second == nil {
		return def
	}
	return *dt.second
}

// HasDate returns true if the datetime has date components.
func (dt *Datetime) HasDate() bool {
	return dt != nil && (dt.year != nil || dt.month != nil || dt.day != nil)
}

// HasTime returns true if the datetime has time components.
func (dt *Datetime) HasTime() bool {
	return dt != nil && (dt.hour != nil || dt.minute != nil || dt.second != nil)
}

// Duration represents a duration of time in nanoseconds.
// Positive values represent forward time, negative values represent backward time.
type Duration int64

func (Duration) Type() Type         { return TypeDuration }
func (v Duration) Display() Content { return Content{} }
func (v Duration) Clone() Value     { return v }
func (Duration) isValue()           {}

// Nanoseconds returns the duration in nanoseconds.
func (d Duration) Nanoseconds() int64 {
	return int64(d)
}

// Seconds returns the duration expressed in seconds (as a float).
func (d Duration) Seconds() float64 {
	return float64(d) / 1e9
}

// Minutes returns the duration expressed in minutes (as a float).
func (d Duration) Minutes() float64 {
	return float64(d) / (60 * 1e9)
}

// Hours returns the duration expressed in hours (as a float).
func (d Duration) Hours() float64 {
	return float64(d) / (3600 * 1e9)
}

// Days returns the duration expressed in days (as a float).
func (d Duration) Days() float64 {
	return float64(d) / (86400 * 1e9)
}

// Weeks returns the duration expressed in weeks (as a float).
func (d Duration) Weeks() float64 {
	return float64(d) / (604800 * 1e9)
}

// DecimalValue represents an arbitrary-precision decimal number.
type DecimalValue struct {
	Value *big.Rat
}

func (DecimalValue) Type() Type         { return TypeDecimal }
func (v DecimalValue) Display() Content { return Content{} }
func (v DecimalValue) Clone() Value {
	if v.Value == nil {
		return DecimalValue{}
	}
	return DecimalValue{Value: new(big.Rat).Set(v.Value)}
}
func (DecimalValue) isValue() {}

// ----------------------------------------------------------------------------
// Visual Values
// ----------------------------------------------------------------------------

// Note: Color is defined in color.go as an interface with multiple color space
// implementations (Luma, Rgba, Oklab, etc.). Color types implement Value directly.

// GradientValue represents a gradient.
type GradientValue struct {
	// Stops contains the color stops.
	Stops []GradientStop
}

// GradientStop represents a single stop in a gradient.
type GradientStop struct {
	// Color is the color at this stop (implements Color interface from color.go).
	Color Value
	Offset float64
}

func (GradientValue) Type() Type         { return TypeGradient }
func (v GradientValue) Display() Content { return Content{} }
func (v GradientValue) Clone() Value {
	if v.Stops == nil {
		return GradientValue{}
	}
	stops := make([]GradientStop, len(v.Stops))
	copy(stops, v.Stops)
	return GradientValue{Stops: stops}
}
func (GradientValue) isValue() {}

// TilingValue represents a tiling pattern.
type TilingValue struct {
	// Content is the pattern content.
	Content Content
}

func (TilingValue) Type() Type         { return TypeTiling }
func (v TilingValue) Display() Content { return Content{} }
func (v TilingValue) Clone() Value     { return v }
func (TilingValue) isValue()           {}

// SymbolValue represents a symbol character.
type SymbolValue struct {
	// Char is the symbol character.
	Char rune
}

func (SymbolValue) Type() Type         { return TypeSymbol }
func (v SymbolValue) Display() Content { return Content{} }
func (v SymbolValue) Clone() Value     { return v }
func (SymbolValue) isValue()           {}

// ----------------------------------------------------------------------------
// Collection Values
// ----------------------------------------------------------------------------

// Content represents typeset content.
type Content struct {
	// Elements contains the content elements.
	Elements []ContentElement
}

// ContentElement is a placeholder interface for content elements.
// IsContentElement is exported to allow cross-package type assertions.
type ContentElement interface {
	IsContentElement()
}

// ContentValue represents content as a Value.
type ContentValue struct {
	Content Content
}

func (ContentValue) Type() Type         { return TypeContent }
func (v ContentValue) Display() Content { return v.Content }
func (v ContentValue) Clone() Value     { return v } // TODO: deep clone
func (ContentValue) isValue()           {}

// Array represents an array of values.
type Array struct {
	items []Value
}

func (*Array) Type() Type         { return TypeArray }
func (v *Array) Display() Content { return Content{} }
func (v *Array) Clone() Value {
	if v == nil || v.items == nil {
		return &Array{}
	}
	clone := make([]Value, len(v.items))
	for i, elem := range v.items {
		clone[i] = elem.Clone()
	}
	return &Array{items: clone}
}
func (*Array) isValue() {}

// NewArray creates a new array from values.
func NewArray(items ...Value) *Array {
	return &Array{items: items}
}

// Len returns the number of items in the array.
func (a *Array) Len() int {
	if a == nil {
		return 0
	}
	return len(a.items)
}

// At returns the item at the given index.
func (a *Array) At(i int) Value {
	if a == nil || i < 0 || i >= len(a.items) {
		return nil
	}
	return a.items[i]
}

// Items returns the underlying slice.
func (a *Array) Items() []Value {
	if a == nil {
		return nil
	}
	return a.items
}

// Dict represents a dictionary mapping strings to values.
type Dict struct {
	// We use parallel slices to preserve insertion order.
	keys   []string
	values []Value
}

func (*Dict) Type() Type         { return TypeDict }
func (v *Dict) Display() Content { return Content{} }
func (d *Dict) Clone() Value {
	if d == nil {
		return &Dict{}
	}
	keys := make([]string, len(d.keys))
	copy(keys, d.keys)
	values := make([]Value, len(d.values))
	for i, v := range d.values {
		values[i] = v.Clone()
	}
	return &Dict{keys: keys, values: values}
}
func (*Dict) isValue() {}

// NewDict creates a new empty dictionary.
func NewDict() *Dict {
	return &Dict{}
}

// Len returns the number of entries in the dictionary.
func (d *Dict) Len() int {
	if d == nil {
		return 0
	}
	return len(d.keys)
}

// Get retrieves a value by key.
func (d *Dict) Get(key string) (Value, bool) {
	if d == nil {
		return nil, false
	}
	for i, k := range d.keys {
		if k == key {
			return d.values[i], true
		}
	}
	return nil, false
}

// Set inserts or updates a key-value pair.
func (d *Dict) Set(key string, value Value) {
	if d == nil {
		return
	}
	for i, k := range d.keys {
		if k == key {
			d.values[i] = value
			return
		}
	}
	d.keys = append(d.keys, key)
	d.values = append(d.values, value)
}

// Keys returns all keys in insertion order.
func (d *Dict) Keys() []string {
	if d == nil {
		return nil
	}
	return d.keys
}

// Values returns all values in insertion order.
func (d *Dict) Values() []Value {
	if d == nil {
		return nil
	}
	return d.values
}

// Contains checks if a key exists.
func (d *Dict) Contains(key string) bool {
	_, ok := d.Get(key)
	return ok
}

// ----------------------------------------------------------------------------
// Callable Values
// ----------------------------------------------------------------------------

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

// ArgsValue represents function arguments.
type ArgsValue struct {
	Args *Args
}

func (ArgsValue) Type() Type         { return TypeArgs }
func (v ArgsValue) Display() Content { return Content{} }
func (v ArgsValue) Clone() Value     { return v } // TODO: deep clone
func (ArgsValue) isValue()           {}

// Args represents a collection of function arguments.
type Args struct {
	// Span is the source location of the arguments.
	Span syntax.Span
	// Items contains the argument items.
	Items []Arg
}

// NewArgs creates a new empty Args with the given span.
func NewArgs(span syntax.Span) *Args {
	return &Args{Span: span, Items: nil}
}

// Push adds a positional argument.
func (a *Args) Push(value Value, span syntax.Span) {
	a.Items = append(a.Items, Arg{
		Span:  span,
		Name:  nil,
		Value: syntax.Spanned[Value]{V: value, Span: span},
	})
}

// PushNamed adds a named argument.
func (a *Args) PushNamed(name string, value Value, span syntax.Span) {
	a.Items = append(a.Items, Arg{
		Span:  span,
		Name:  &name,
		Value: syntax.Spanned[Value]{V: value, Span: span},
	})
}

// Expect retrieves and removes the next positional argument.
func (a *Args) Expect(name string) (syntax.Spanned[Value], error) {
	for i, item := range a.Items {
		if item.Name == nil {
			a.Items = append(a.Items[:i], a.Items[i+1:]...)
			return item.Value, nil
		}
	}
	return syntax.Spanned[Value]{}, &MissingArgumentError{Name: name, Span: a.Span}
}

// Find retrieves and removes a named argument if present.
func (a *Args) Find(name string) *syntax.Spanned[Value] {
	for i, item := range a.Items {
		if item.Name != nil && *item.Name == name {
			a.Items = append(a.Items[:i], a.Items[i+1:]...)
			return &item.Value
		}
	}
	return nil
}

// Finish checks that all arguments have been consumed.
func (a *Args) Finish() error {
	if len(a.Items) > 0 {
		return &UnexpectedArgumentError{Arg: a.Items[0]}
	}
	return nil
}

// Arg represents a single argument.
type Arg struct {
	// Span is the source location.
	Span syntax.Span
	// Name is the optional argument name (for named arguments).
	Name *string
	// Value is the argument value.
	Value syntax.Spanned[Value]
}

// MissingArgumentError is returned when a required argument is missing.
type MissingArgumentError struct {
	Name string
	Span syntax.Span
}

func (e *MissingArgumentError) Error() string {
	return fmt.Sprintf("missing argument: %s", e.Name)
}

// UnexpectedArgumentError is returned when an unexpected argument is provided.
type UnexpectedArgumentError struct {
	Arg Arg
}

func (e *UnexpectedArgumentError) Error() string {
	if e.Arg.Name != nil {
		return fmt.Sprintf("unexpected argument: %s", *e.Arg.Name)
	}
	return "unexpected positional argument"
}

// TypeValue represents a type as a value.
type TypeValue struct {
	Inner Type
}

func (TypeValue) Type() Type         { return TypeType }
func (v TypeValue) Display() Content { return Content{} }
func (v TypeValue) Clone() Value     { return v }
func (TypeValue) isValue()           {}

// Get returns the wrapped type.
func (v TypeValue) Get() Type { return v.Inner }

// ModuleValue represents an evaluated module.
type ModuleValue struct {
	Module *Module
}

func (ModuleValue) Type() Type         { return TypeModule }
func (v ModuleValue) Display() Content { return Content{} }
func (v ModuleValue) Clone() Value     { return v }
func (ModuleValue) isValue()           {}

// Module represents an evaluated Typst module.
type Module struct {
	// Name is the module name.
	Name string
	// Scope contains the module's exported bindings.
	Scope *Scope
	// Content is the module's evaluated content.
	Content Content
}

// ----------------------------------------------------------------------------
// Scope
// ----------------------------------------------------------------------------

// Scope represents a single lexical scope containing variable bindings.
type Scope struct {
	// bindings maps identifier names to their bindings.
	bindings []scopeEntry
	// deduplicate prevents duplicate definitions.
	deduplicate bool
	// category is an optional documentation category.
	category *Category
}

type scopeEntry struct {
	name    string
	binding Binding
}

// Category represents a documentation category.
type Category struct {
	Name string
}

// Binding represents a variable binding.
type Binding struct {
	// Value is the bound value.
	Value Value
	// Span is the source location of the binding.
	Span syntax.Span
	// Category is an optional documentation category.
	Category *Category
	// Kind is the binding kind (e.g., normal, closure).
	Kind BindingKind
}

// BindingKind represents the kind of a binding.
type BindingKind int

const (
	BindingNormal BindingKind = iota
	BindingClosure
)

// NewScope creates a new empty scope.
func NewScope() *Scope {
	return &Scope{bindings: nil}
}

// NewBinding creates a new normal binding.
func NewBinding(value Value, span syntax.Span) Binding {
	return Binding{Value: value, Span: span, Kind: BindingNormal}
}

// NewClosureBinding creates a new closure binding.
func NewClosureBinding(value Value, span syntax.Span) Binding {
	return Binding{Value: value, Span: span, Kind: BindingClosure}
}

// Define binds a value to an identifier in this scope.
func (s *Scope) Define(name string, value Value, span syntax.Span) {
	s.Insert(name, NewBinding(value, span))
}

// Insert adds a binding to this scope.
func (s *Scope) Insert(name string, binding Binding) {
	if binding.Category == nil && s.category != nil {
		binding.Category = s.category
	}
	for i, entry := range s.bindings {
		if entry.name == name {
			s.bindings[i].binding = binding
			return
		}
	}
	s.bindings = append(s.bindings, scopeEntry{name: name, binding: binding})
}

// Get retrieves a binding by name.
func (s *Scope) Get(name string) *Binding {
	for i := range s.bindings {
		if s.bindings[i].name == name {
			return &s.bindings[i].binding
		}
	}
	return nil
}

// Contains returns true if the scope contains a binding for the given name.
func (s *Scope) Contains(name string) bool {
	return s.Get(name) != nil
}

// ----------------------------------------------------------------------------
// Dynamic Value
// ----------------------------------------------------------------------------

// DynValue represents a dynamically-typed value.
type DynValue struct {
	// Inner is the underlying dynamic value.
	Inner interface{}
	// TypeName is the name of the dynamic type.
	TypeName string
}

func (DynValue) Type() Type         { return TypeDyn }
func (v DynValue) Display() Content { return Content{} }
func (v DynValue) Clone() Value     { return v }
func (DynValue) isValue()           {}

// StylesValue represents a collection of styles.
type StylesValue struct {
	Styles *Styles
}

func (StylesValue) Type() Type         { return TypeStyles }
func (v StylesValue) Display() Content { return Content{} }
func (v StylesValue) Clone() Value     { return v }
func (StylesValue) isValue()           {}

// Styles represents a collection of style rules and recipes.
type Styles struct {
	// Rules contains the style rules (from set rules).
	Rules []StyleRule
	// Recipes contains the show rule recipes.
	Recipes []*Recipe
}

// StyleRule represents a single style rule.
type StyleRule struct {
	// Func is the function this style applies to.
	Func *Func
	// Args are the style arguments.
	Args *Args
	// Span is the source location of the rule.
	Span syntax.Span
	// Liftable indicates whether this style can be lifted to page level.
	Liftable bool
}

// Recipe represents a show rule recipe.
type Recipe struct {
	// Span is the source location.
	Span syntax.Span
	// Selector is the selector for matching content.
	Selector Selector
	// Transform is the transformation to apply.
	Transform Transform
}

// Selector represents a content selector.
type Selector interface {
	isSelector()
}

// Transform represents a content transformation.
type Transform interface {
	isTransform()
}

// VersionValue represents a semantic version.
type VersionValue struct {
	Major int
	Minor int
	Patch int
}

func (VersionValue) Type() Type         { return TypeVersion }
func (v VersionValue) Display() Content { return Content{} }
func (v VersionValue) Clone() Value     { return v }
func (VersionValue) isValue()           {}

// ----------------------------------------------------------------------------
// Type System
// ----------------------------------------------------------------------------

// Type represents a Typst type.
type Type int

const (
	TypeNone Type = iota
	TypeAuto
	TypeBool
	TypeInt
	TypeFloat
	TypeLength
	TypeAngle
	TypeRatio
	TypeRelative
	TypeFraction
	TypeStr
	TypeBytes
	TypeLabel
	TypeDatetime
	TypeDuration
	TypeDecimal
	TypeColor
	TypeGradient
	TypeTiling
	TypeSymbol
	TypeContent
	TypeArray
	TypeDict
	TypeFunc
	TypeArgs
	TypeType
	TypeModule
	TypeDyn
	TypeStyles
	TypeVersion
)

// String returns the type name.
func (t Type) String() string {
	switch t {
	case TypeNone:
		return "none"
	case TypeAuto:
		return "auto"
	case TypeBool:
		return "boolean"
	case TypeInt:
		return "integer"
	case TypeFloat:
		return "float"
	case TypeLength:
		return "length"
	case TypeAngle:
		return "angle"
	case TypeRatio:
		return "ratio"
	case TypeRelative:
		return "relative"
	case TypeFraction:
		return "fraction"
	case TypeStr:
		return "string"
	case TypeBytes:
		return "bytes"
	case TypeLabel:
		return "label"
	case TypeDatetime:
		return "datetime"
	case TypeDuration:
		return "duration"
	case TypeDecimal:
		return "decimal"
	case TypeColor:
		return "color"
	case TypeGradient:
		return "gradient"
	case TypeTiling:
		return "tiling"
	case TypeSymbol:
		return "symbol"
	case TypeContent:
		return "content"
	case TypeArray:
		return "array"
	case TypeDict:
		return "dictionary"
	case TypeFunc:
		return "function"
	case TypeArgs:
		return "arguments"
	case TypeType:
		return "type"
	case TypeModule:
		return "module"
	case TypeDyn:
		return "dynamic"
	case TypeStyles:
		return "styles"
	case TypeVersion:
		return "version"
	default:
		return fmt.Sprintf("Type(%d)", t)
	}
}

// Ident returns the short identifier for the type.
func (t Type) Ident() string {
	switch t {
	case TypeBool:
		return "bool"
	case TypeInt:
		return "int"
	case TypeStr:
		return "str"
	default:
		return t.String()
	}
}

// ----------------------------------------------------------------------------
// Value Conversion Helpers
// ----------------------------------------------------------------------------

// IsNone returns true if the value is none.
func IsNone(v Value) bool {
	_, ok := v.(NoneValue)
	return ok
}

// IsAuto returns true if the value is auto.
func IsAuto(v Value) bool {
	_, ok := v.(AutoValue)
	return ok
}

// AsBool attempts to convert a value to a bool.
func AsBool(v Value) (bool, bool) {
	if b, ok := v.(Bool); ok {
		return bool(b), true
	}
	return false, false
}

// AsInt attempts to convert a value to an int64.
func AsInt(v Value) (int64, bool) {
	if i, ok := v.(Int); ok {
		return int64(i), true
	}
	return 0, false
}

// AsFloat attempts to convert a value to a float64.
func AsFloat(v Value) (float64, bool) {
	switch v := v.(type) {
	case Float:
		return float64(v), true
	case Int:
		return float64(v), true
	}
	return 0, false
}

// AsStr attempts to convert a value to a string.
func AsStr(v Value) (string, bool) {
	if s, ok := v.(Str); ok {
		return string(s), true
	}
	return "", false
}

// AsArray attempts to convert a value to an array.
func AsArray(v Value) (*Array, bool) {
	if a, ok := v.(*Array); ok {
		return a, true
	}
	return nil, false
}

// AsDict attempts to convert a value to a dictionary.
func AsDict(v Value) (*Dict, bool) {
	if d, ok := v.(*Dict); ok {
		return d, true
	}
	return nil, false
}

// AsFunc attempts to convert a value to a function.
func AsFunc(v Value) (*Func, bool) {
	if f, ok := v.(FuncValue); ok {
		return f.Func, true
	}
	return nil, false
}
