// Package foundations provides core types and operations for the Typst runtime.
package foundations

import (
	"fmt"
	"math"
	"strconv"
)

// Value represents a runtime value in the Typst language.
// This interface is implemented by all concrete value types.
type Value interface {
	// valueMarker is an unexported method to seal the interface.
	valueMarker()
	// Type returns the type name of this value.
	Type() string
	// String returns a string representation for display.
	String() string
}

// Ensure all value types implement Value.
var (
	_ Value = NoneValue{}
	_ Value = AutoValue{}
	_ Value = Bool(false)
	_ Value = Int(0)
	_ Value = Float(0)
	_ Value = Str("")
	_ Value = (*Array)(nil)
	_ Value = (*Dict)(nil)
)

// NoneValue represents the absence of a value.
type NoneValue struct{}

func (NoneValue) valueMarker() {}
func (NoneValue) Type() string { return "none" }
func (NoneValue) String() string { return "none" }

// None is the singleton none value.
var None = NoneValue{}

// AutoValue represents automatic behavior.
type AutoValue struct{}

func (AutoValue) valueMarker() {}
func (AutoValue) Type() string { return "auto" }
func (AutoValue) String() string { return "auto" }

// Auto is the singleton auto value.
var Auto = AutoValue{}

// Bool represents a boolean value.
type Bool bool

func (Bool) valueMarker() {}
func (Bool) Type() string { return "bool" }
func (b Bool) String() string {
	if b {
		return "true"
	}
	return "false"
}

// Int represents an integer value.
type Int int64

func (Int) valueMarker() {}
func (Int) Type() string { return "int" }
func (i Int) String() string { return strconv.FormatInt(int64(i), 10) }

// Float represents a floating-point value.
type Float float64

func (Float) valueMarker() {}
func (Float) Type() string { return "float" }
func (f Float) String() string {
	v := float64(f)
	if math.IsInf(v, 1) {
		return "float.inf"
	}
	if math.IsInf(v, -1) {
		return "-float.inf"
	}
	if math.IsNaN(v) {
		return "float.nan"
	}
	return strconv.FormatFloat(v, 'g', -1, 64)
}

// Str represents a string value.
type Str string

func (Str) valueMarker() {}
func (Str) Type() string { return "str" }
func (s Str) String() string { return fmt.Sprintf("%q", string(s)) }

// Array represents an array of values.
type Array struct {
	items []Value
}

func (*Array) valueMarker() {}
func (*Array) Type() string { return "array" }
func (a *Array) String() string {
	if a == nil || len(a.items) == 0 {
		return "()"
	}
	result := "("
	for i, item := range a.items {
		if i > 0 {
			result += ", "
		}
		result += item.String()
	}
	return result + ")"
}

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

func (*Dict) valueMarker() {}
func (*Dict) Type() string { return "dictionary" }
func (d *Dict) String() string {
	if d == nil || len(d.keys) == 0 {
		return "(:)"
	}
	result := "("
	for i, key := range d.keys {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%s: %s", key, d.values[i].String())
	}
	return result + ")"
}

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
