// Array type for Typst.
// Translated from foundations/array.rs

package foundations

import (
	"fmt"

	"github.com/boergens/gotypst/syntax"
)

// Spanned is an alias for syntax.Spanned for convenience.
type Spanned[T any] = syntax.Spanned[T]

// ArrayConstruct converts a value to an array.
// Supports: Array, Bytes, Version.
//
// Note: This is only for conversion of collection-like values to an array,
// not for creation of an array from individual items. Use array syntax `(1, 2, 3)` instead.
//
// This matches Rust's array::construct function.
func ArrayConstruct(args *Args) (Value, error) {
	spanned, err := args.Expect("value")
	if err != nil {
		return nil, err
	}
	value := spanned.V

	if err := args.Finish(); err != nil {
		return nil, err
	}

	switch v := value.(type) {
	case *Array:
		return v, nil

	case BytesValue:
		// Convert bytes to array of integers
		items := make([]Value, len(v))
		for i, b := range v {
			items[i] = Int(b)
		}
		return NewArray(items...), nil

	case VersionValue:
		// Convert version to array of integers
		return NewArray(Int(int64(v.Major)), Int(int64(v.Minor)), Int(int64(v.Patch))), nil

	default:
		return nil, &ConstructorError{
			Message: fmt.Sprintf("expected array, bytes, or version, found %s", value.Type().String()),
			Span:    spanned.Span,
		}
	}
}

// Array represents a sequence of values.
//
// You can construct an array by enclosing a comma-separated sequence of values
// in parentheses. The values do not have to be of the same type.
//
// You can access and update array items with the .at() method. Indices are
// zero-based and negative indices wrap around to the end of the array.
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

// WithCapacity creates a new array with the given capacity.
func ArrayWithCapacity(capacity int) *Array {
	return &Array{items: make([]Value, 0, capacity)}
}

// IsEmpty returns true if the length is 0.
func (a *Array) IsEmpty() bool {
	return a == nil || len(a.items) == 0
}

// Len returns the number of items in the array.
func (a *Array) Len() int {
	if a == nil {
		return 0
	}
	return len(a.items)
}

// AsSlice returns the items as a slice (like Rust's as_slice).
func (a *Array) AsSlice() []Value {
	if a == nil {
		return nil
	}
	return a.items
}

// Items returns the underlying slice (alias for AsSlice).
func (a *Array) Items() []Value {
	return a.AsSlice()
}

// At returns the item at the given index.
// Returns nil if index is out of bounds.
func (a *Array) At(i int) Value {
	if a == nil || i < 0 || i >= len(a.items) {
		return nil
	}
	return a.items[i]
}

// AtMut returns a mutable pointer to the item at the given index.
// Supports negative indices that wrap around from the end.
// Matches Rust's at_mut method.
func (a *Array) AtMut(index int64) (*Value, error) {
	if a == nil {
		return nil, &OpError{Message: "array is empty"}
	}
	len64 := int64(len(a.items))
	// Handle negative indices
	idx := index
	if idx < 0 {
		idx = len64 + idx
	}
	if idx < 0 || idx >= len64 {
		return nil, &OpError{Message: fmt.Sprintf("array index out of bounds: index %d, length %d", index, len64)}
	}
	return &a.items[idx], nil
}

// FirstMut returns a mutable pointer to the first item.
// Matches Rust's first_mut method.
func (a *Array) FirstMut() (*Value, error) {
	if a == nil || len(a.items) == 0 {
		return nil, &OpError{Message: "array is empty"}
	}
	return &a.items[0], nil
}

// LastMut returns a mutable pointer to the last item.
// Matches Rust's last_mut method.
func (a *Array) LastMut() (*Value, error) {
	if a == nil || len(a.items) == 0 {
		return nil, &OpError{Message: "array is empty"}
	}
	return &a.items[len(a.items)-1], nil
}

// First returns the first item, or an error if empty.
func (a *Array) First() (Value, error) {
	if a == nil || len(a.items) == 0 {
		return nil, &OpError{Message: "array is empty"}
	}
	return a.items[0], nil
}

// Last returns the last item, or an error if empty.
func (a *Array) Last() (Value, error) {
	if a == nil || len(a.items) == 0 {
		return nil, &OpError{Message: "array is empty"}
	}
	return a.items[len(a.items)-1], nil
}

// Push appends a value to the array.
func (a *Array) Push(v Value) {
	if a == nil {
		return
	}
	a.items = append(a.items, v)
}

// Pop removes and returns the last value, or an error if empty.
func (a *Array) Pop() (Value, error) {
	if a == nil || len(a.items) == 0 {
		return nil, &OpError{Message: "array is empty"}
	}
	last := a.items[len(a.items)-1]
	a.items = a.items[:len(a.items)-1]
	return last, nil
}

// Insert inserts a value at the given index.
func (a *Array) Insert(index int64, v Value) error {
	if a == nil {
		return &OpError{Message: "array is nil"}
	}
	len64 := int64(len(a.items))
	idx := index
	if idx < 0 {
		idx = len64 + idx + 1
	}
	if idx < 0 || idx > len64 {
		return &OpError{Message: fmt.Sprintf("array index out of bounds: index %d, length %d", index, len64)}
	}
	a.items = append(a.items[:idx], append([]Value{v}, a.items[idx:]...)...)
	return nil
}

// Remove removes and returns the value at the given index.
// If the index is out of bounds and a default is provided, returns the default.
// Matches Rust's remove method with optional default.
func (a *Array) Remove(index int64, def *Spanned[Value]) (Value, error) {
	if a == nil || len(a.items) == 0 {
		if def != nil {
			return def.V, nil
		}
		return nil, &OpError{Message: "array is empty"}
	}
	len64 := int64(len(a.items))
	idx := index
	if idx < 0 {
		idx = len64 + idx
	}
	if idx < 0 || idx >= len64 {
		if def != nil {
			return def.V, nil
		}
		return nil, &OpError{Message: fmt.Sprintf("array index out of bounds: index %d, length %d", index, len64)}
	}
	v := a.items[idx]
	a.items = append(a.items[:idx], a.items[idx+1:]...)
	return v, nil
}

// Contains returns true if the array contains the given value.
func (a *Array) Contains(v Value) bool {
	if a == nil {
		return false
	}
	for _, item := range a.items {
		if Equal(item, v) {
			return true
		}
	}
	return false
}

// Iter returns an iterator over the array items.
func (a *Array) Iter() []Value {
	if a == nil {
		return nil
	}
	return a.items
}
