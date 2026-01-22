// Primitive value types for Typst.
// Translated from foundations/none.rs, foundations/auto.rs, foundations/bool.rs

package foundations

import "fmt"

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
