// Package eval provides the evaluation engine for Typst.
//
// This package is a Go translation of typst-eval from the original Typst
// compiler. It implements a tree-walking interpreter that transforms parsed
// AST nodes into runtime values.
package eval

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

// BoolValue represents a boolean value.
type BoolValue bool

func (BoolValue) Type() Type     { return TypeBool }
func (v BoolValue) Display() Content { return Content{} }
func (v BoolValue) Clone() Value { return v }
func (BoolValue) isValue()       {}

// True and False are the boolean singleton values.
var (
	True  = BoolValue(true)
	False = BoolValue(false)
)

// Bool converts a Go bool to a BoolValue.
func Bool(b bool) BoolValue {
	if b {
		return True
	}
	return False
}

// IntValue represents a 64-bit signed integer.
type IntValue int64

func (IntValue) Type() Type     { return TypeInt }
func (v IntValue) Display() Content { return Content{} }
func (v IntValue) Clone() Value { return v }
func (IntValue) isValue()       {}

// Int creates an IntValue from a Go int64.
func Int(i int64) IntValue {
	return IntValue(i)
}

// FloatValue represents a 64-bit floating point number.
type FloatValue float64

func (FloatValue) Type() Type     { return TypeFloat }
func (v FloatValue) Display() Content { return Content{} }
func (v FloatValue) Clone() Value { return v }
func (FloatValue) isValue()       {}

// Float creates a FloatValue from a Go float64.
func Float(f float64) FloatValue {
	return FloatValue(f)
}

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

func (LengthValue) Type() Type     { return TypeLength }
func (v LengthValue) Display() Content { return Content{} }
func (v LengthValue) Clone() Value { return v }
func (LengthValue) isValue()       {}

// Angle represents an angle value.
type Angle struct {
	// Radians is the angle in radians.
	Radians float64
}

// AngleValue represents an angle as a Value.
type AngleValue struct {
	Angle Angle
}

func (AngleValue) Type() Type     { return TypeAngle }
func (v AngleValue) Display() Content { return Content{} }
func (v AngleValue) Clone() Value { return v }
func (AngleValue) isValue()       {}

// Ratio represents a ratio (percentage) value.
type Ratio struct {
	// Value is the ratio as a fraction (0.5 = 50%).
	Value float64
}

// RatioValue represents a ratio as a Value.
type RatioValue struct {
	Ratio Ratio
}

func (RatioValue) Type() Type     { return TypeRatio }
func (v RatioValue) Display() Content { return Content{} }
func (v RatioValue) Clone() Value { return v }
func (RatioValue) isValue()       {}

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

func (RelativeValue) Type() Type     { return TypeRelative }
func (v RelativeValue) Display() Content { return Content{} }
func (v RelativeValue) Clone() Value { return v }
func (RelativeValue) isValue()       {}

// Fraction represents a fraction of remaining space.
type Fraction struct {
	// Value is the number of fractions (1fr = 1.0).
	Value float64
}

// FractionValue represents a fraction as a Value.
type FractionValue struct {
	Fraction Fraction
}

func (FractionValue) Type() Type     { return TypeFraction }
func (v FractionValue) Display() Content { return Content{} }
func (v FractionValue) Clone() Value { return v }
func (FractionValue) isValue()       {}

// ----------------------------------------------------------------------------
// Data Values
// ----------------------------------------------------------------------------

// StrValue represents a string value.
type StrValue string

func (StrValue) Type() Type     { return TypeStr }
func (v StrValue) Display() Content { return Content{} }
func (v StrValue) Clone() Value { return v }
func (StrValue) isValue()       {}

// Str creates a StrValue from a Go string.
func Str(s string) StrValue {
	return StrValue(s)
}

// BytesValue represents a sequence of bytes.
type BytesValue []byte

func (BytesValue) Type() Type     { return TypeBytes }
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

func (LabelValue) Type() Type     { return TypeLabel }
func (v LabelValue) Display() Content { return Content{} }
func (v LabelValue) Clone() Value { return v }
func (LabelValue) isValue()       {}

// DatetimeValue represents a date and time.
type DatetimeValue struct {
	Year   int
	Month  int
	Day    int
	Hour   int
	Minute int
	Second int
}

func (DatetimeValue) Type() Type     { return TypeDatetime }
func (v DatetimeValue) Display() Content { return Content{} }
func (v DatetimeValue) Clone() Value { return v }
func (DatetimeValue) isValue()       {}

// DurationValue represents a duration of time.
type DurationValue struct {
	// Nanoseconds is the duration in nanoseconds.
	Nanoseconds int64
}

func (DurationValue) Type() Type     { return TypeDuration }
func (v DurationValue) Display() Content { return Content{} }
func (v DurationValue) Clone() Value { return v }
func (DurationValue) isValue()       {}

// DecimalValue represents an arbitrary-precision decimal number.
type DecimalValue struct {
	Value *big.Rat
}

func (DecimalValue) Type() Type { return TypeDecimal }
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

// Color represents a color value.
type Color struct {
	R, G, B, A uint8
}

// ColorValue represents a color as a Value.
type ColorValue struct {
	Color Color
}

func (ColorValue) Type() Type     { return TypeColor }
func (v ColorValue) Display() Content { return Content{} }
func (v ColorValue) Clone() Value { return v }
func (ColorValue) isValue()       {}

// GradientValue represents a gradient.
type GradientValue struct {
	// Stops contains the color stops.
	Stops []GradientStop
}

// GradientStop represents a single stop in a gradient.
type GradientStop struct {
	Color  Color
	Offset float64
}

func (GradientValue) Type() Type     { return TypeGradient }
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

func (TilingValue) Type() Type     { return TypeTiling }
func (v TilingValue) Display() Content { return Content{} }
func (v TilingValue) Clone() Value { return v }
func (TilingValue) isValue()       {}

// SymbolValue represents a symbol character.
type SymbolValue struct {
	// Char is the symbol character.
	Char rune
}

func (SymbolValue) Type() Type     { return TypeSymbol }
func (v SymbolValue) Display() Content { return Content{} }
func (v SymbolValue) Clone() Value { return v }
func (SymbolValue) isValue()       {}

// ----------------------------------------------------------------------------
// Collection Values
// ----------------------------------------------------------------------------

// Content represents typeset content.
// This is a placeholder that will be expanded in the library package.
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

func (ContentValue) Type() Type     { return TypeContent }
func (v ContentValue) Display() Content { return v.Content }
func (v ContentValue) Clone() Value { return v } // TODO: deep clone
func (ContentValue) isValue()       {}

// ArrayValue represents an array of values.
type ArrayValue []Value

func (ArrayValue) Type() Type     { return TypeArray }
func (v ArrayValue) Display() Content { return Content{} }
func (v ArrayValue) Clone() Value {
	if v == nil {
		return ArrayValue(nil)
	}
	clone := make([]Value, len(v))
	for i, elem := range v {
		clone[i] = elem.Clone()
	}
	return ArrayValue(clone)
}
func (ArrayValue) isValue() {}

// DictValue represents a dictionary mapping strings to values.
type DictValue struct {
	// entries preserves insertion order.
	entries []dictEntry
}

type dictEntry struct {
	Key   string
	Value Value
}

func (DictValue) Type() Type     { return TypeDict }
func (v DictValue) Display() Content { return Content{} }
func (v DictValue) Clone() Value {
	if v.entries == nil {
		return DictValue{}
	}
	clone := make([]dictEntry, len(v.entries))
	for i, e := range v.entries {
		clone[i] = dictEntry{Key: e.Key, Value: e.Value.Clone()}
	}
	return DictValue{entries: clone}
}
func (DictValue) isValue() {}

// NewDict creates a new empty dictionary.
func NewDict() DictValue {
	return DictValue{entries: nil}
}

// Get retrieves a value from the dictionary.
func (d *DictValue) Get(key string) (Value, bool) {
	for _, e := range d.entries {
		if e.Key == key {
			return e.Value, true
		}
	}
	return nil, false
}

// Set sets a value in the dictionary.
func (d *DictValue) Set(key string, value Value) {
	for i, e := range d.entries {
		if e.Key == key {
			d.entries[i].Value = value
			return
		}
	}
	d.entries = append(d.entries, dictEntry{Key: key, Value: value})
}

// Len returns the number of entries in the dictionary.
func (d *DictValue) Len() int {
	return len(d.entries)
}

// Keys returns the keys in insertion order.
func (d *DictValue) Keys() []string {
	keys := make([]string, len(d.entries))
	for i, e := range d.entries {
		keys[i] = e.Key
	}
	return keys
}

// ----------------------------------------------------------------------------
// Callable Values
// ----------------------------------------------------------------------------

// FuncValue represents a function.
type FuncValue struct {
	// Func is the underlying function.
	Func *Func
}

func (FuncValue) Type() Type     { return TypeFunc }
func (v FuncValue) Display() Content { return Content{} }
func (v FuncValue) Clone() Value { return v } // Functions are immutable
func (FuncValue) isValue()       {}

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
type NativeFunc struct {
	// Func is the Go function implementing this native.
	Func func(vm *Vm, args *Args) (Value, error)
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

func (ArgsValue) Type() Type     { return TypeArgs }
func (v ArgsValue) Display() Content { return Content{} }
func (v ArgsValue) Clone() Value { return v } // TODO: deep clone
func (ArgsValue) isValue()       {}

// Args represents a collection of function arguments.
type Args struct {
	// Span is the source location of the arguments.
	Span syntax.Span
	// Items contains the argument items.
	Items []Arg
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

// TypeValue represents a type as a value.
type TypeValue struct {
	Inner Type
}

func (TypeValue) Type() Type       { return TypeType }
func (v TypeValue) Display() Content   { return Content{} }
func (v TypeValue) Clone() Value   { return v }
func (TypeValue) isValue()         {}

// Get returns the wrapped type.
func (v TypeValue) Get() Type { return v.Inner }

// ModuleValue represents an evaluated module.
type ModuleValue struct {
	Module *Module
}

func (ModuleValue) Type() Type     { return TypeModule }
func (v ModuleValue) Display() Content { return Content{} }
func (v ModuleValue) Clone() Value { return v }
func (ModuleValue) isValue()       {}

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
// Dynamic Value
// ----------------------------------------------------------------------------

// DynValue represents a dynamically-typed value.
// This allows extending the value system with custom types.
type DynValue struct {
	// Inner is the underlying dynamic value.
	Inner interface{}
	// TypeName is the name of the dynamic type.
	TypeName string
}

func (DynValue) Type() Type       { return TypeDyn }
func (v DynValue) Display() Content   { return Content{} }
func (v DynValue) Clone() Value   { return v } // Shallow clone
func (DynValue) isValue()         {}

// StylesValue represents a collection of styles.
type StylesValue struct {
	Styles *Styles
}

func (StylesValue) Type() Type     { return TypeStyles }
func (v StylesValue) Display() Content { return Content{} }
func (v StylesValue) Clone() Value { return v }
func (StylesValue) isValue()       {}

// Styles represents a collection of style rules and recipes.
type Styles struct {
	// Rules contains the style rules (from set rules).
	Rules []StyleRule
	// Recipes contains the show rule recipes.
	Recipes []*Recipe
}

// StyleRule represents a single style rule.
// This corresponds to Typst's Property type in styles.rs.
type StyleRule struct {
	// Func is the function this style applies to.
	Func *Func
	// Args are the style arguments.
	Args *Args
	// Span is the source location of the rule.
	Span syntax.Span
	// Liftable indicates whether this style can be lifted to page level.
	// Set rules produce liftable styles, constructor calls do not.
	// This affects whether styles propagate into page headers/footers.
	Liftable bool
}

// VersionValue represents a semantic version.
type VersionValue struct {
	Major int
	Minor int
	Patch int
}

func (VersionValue) Type() Type     { return TypeVersion }
func (v VersionValue) Display() Content { return Content{} }
func (v VersionValue) Clone() Value { return v }
func (VersionValue) isValue()       {}

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
		return "bool"
	case TypeInt:
		return "int"
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
		return "str"
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
	if b, ok := v.(BoolValue); ok {
		return bool(b), true
	}
	return false, false
}

// AsInt attempts to convert a value to an int64.
func AsInt(v Value) (int64, bool) {
	if i, ok := v.(IntValue); ok {
		return int64(i), true
	}
	return 0, false
}

// AsFloat attempts to convert a value to a float64.
func AsFloat(v Value) (float64, bool) {
	switch v := v.(type) {
	case FloatValue:
		return float64(v), true
	case IntValue:
		return float64(v), true
	}
	return 0, false
}

// AsStr attempts to convert a value to a string.
func AsStr(v Value) (string, bool) {
	if s, ok := v.(StrValue); ok {
		return string(s), true
	}
	return "", false
}

// AsArray attempts to convert a value to an array.
func AsArray(v Value) (ArrayValue, bool) {
	if a, ok := v.(ArrayValue); ok {
		return a, true
	}
	return nil, false
}

// AsDict attempts to convert a value to a dictionary.
func AsDict(v Value) (*DictValue, bool) {
	if d, ok := v.(*DictValue); ok {
		return d, true
	}
	if d, ok := v.(DictValue); ok {
		return &d, true
	}
	return nil, false
}

// AsFunc attempts to convert a value to a function.
func AsFunc(v Value) (*Func, bool) {
	if f, ok := v.(FuncValue); ok {
		return f.Func, true
	}
	// TypeValue is callable as a constructor
	if t, ok := v.(TypeValue); ok {
		return typeConstructorFunc(t.Inner), true
	}
	return nil, false
}
