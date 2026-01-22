// Package foundations provides core types and operations for the Typst runtime.
//
// This package contains the foundational value types that form the basis of
// the Typst language. It corresponds to typst-library/src/foundations/ in
// the Rust implementation.
//
// File organization matches Rust:
//   - value.go: Value interface, Type enum
//   - primitives.go: None, Auto, Bool, Int, Float
//   - measurements.go: Length, Angle, Ratio, Relative, Fraction
//   - data.go: Str, Bytes, Label, Decimal, Version
//   - datetime.go: Datetime, Duration
//   - content.go: Content, ContentValue
//   - visual.go: Gradient, Tiling, Symbol, Dyn
//   - array.go: Array
//   - dict.go: Dict
//   - func.go: Func, NativeFunc, Closure
//   - scope.go: Scope, Binding
//   - module.go: Module
//   - styles.go: Styles, Recipe
//   - engine.go: Engine, Context, World interfaces
//   - color.go: Color types
//   - args.go: Args
//   - cast.go: Type conversion utilities
package foundations

import "fmt"

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

// typeScopes holds the method scopes for each type.
// This is populated by RegisterTypeScope during initialization.
var typeScopes = make(map[Type]*Scope)

// Scope returns the type's associated scope containing methods.
// Returns nil if the type has no associated scope.
func (t Type) Scope() *Scope {
	return typeScopes[t]
}

// RegisterTypeScope registers a scope for a type.
// This should be called during package initialization to set up type methods.
func RegisterTypeScope(t Type, scope *Scope) {
	typeScopes[t] = scope
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
