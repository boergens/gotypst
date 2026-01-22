// Dict type for Typst.
// Translated from foundations/dict.rs

package foundations

import (
	"fmt"

	"github.com/boergens/gotypst/syntax"
)

// DictConstruct converts a value to a dictionary.
// Supports: Module.
//
// Note: This is only for conversion of dictionary-like values to a dictionary,
// not for creation of a dictionary from individual pairs. Use dict syntax `(key: value)` instead.
//
// This matches Rust's dict::construct function.
func DictConstruct(args *Args) (Value, error) {
	spanned, err := args.Expect("value")
	if err != nil {
		return nil, err
	}
	value := spanned.V

	if err := args.Finish(); err != nil {
		return nil, err
	}

	switch v := value.(type) {
	case *Dict:
		return v, nil

	case ModuleValue:
		// Convert module scope to dictionary
		result := NewDict()
		if v.Module != nil && v.Module.Scope != nil {
			v.Module.Scope.Iter(func(name string, binding Binding) {
				result.Set(name, binding.Read())
			})
		}
		return result, nil

	default:
		return nil, &ConstructorError{
			Message: fmt.Sprintf("expected dictionary or module, found %s", value.Type().String()),
			Span:    spanned.Span,
		}
	}
}

// Dict represents a map from string keys to values.
//
// You can construct a dictionary by enclosing comma-separated key: value pairs
// in parentheses. The values do not have to be of the same type. Since empty
// parentheses already yield an empty array, you have to use the special (:)
// syntax to create an empty dictionary.
//
// A dictionary is conceptually similar to an array, but it is indexed by
// strings instead of integers. You can access and create dictionary entries
// with the .at() method. If you know the key statically, you can alternatively
// use field access notation (.key) to access the value.
type Dict struct {
	// We use parallel slices to preserve insertion order,
	// similar to Rust's IndexMap.
	keys   []string
	values []Value
}

func (*Dict) Type() Type         { return TypeDict }
func (d *Dict) Display() Content { return Content{} }
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

// IsEmpty returns true if the dictionary is empty.
func (d *Dict) IsEmpty() bool {
	return d == nil || len(d.keys) == 0
}

// Len returns the number of entries in the dictionary.
func (d *Dict) Len() int {
	if d == nil {
		return 0
	}
	return len(d.keys)
}

// Get retrieves a value by key.
// Returns the value and true if found, nil and false otherwise.
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

// AtMut returns a mutable pointer to the value at the given key.
// Returns an error if the key doesn't exist.
// Matches Rust's at_mut method.
func (d *Dict) AtMut(key string) (*Value, error) {
	if d == nil {
		return nil, &OpError{Message: fmt.Sprintf("dictionary does not contain key %q", key)}
	}
	for i, k := range d.keys {
		if k == key {
			return &d.values[i], nil
		}
	}
	return nil, &OpError{Message: fmt.Sprintf("dictionary does not contain key %q; use insert to add or update values", key)}
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

// Insert is an alias for Set, matching Rust's insert method.
func (d *Dict) Insert(key string, value Value) {
	d.Set(key, value)
}

// Take removes and returns the value for the given key.
// Returns an error if the key doesn't exist.
// Matches Rust's take method.
func (d *Dict) Take(key string) (Value, error) {
	if d == nil {
		return nil, &OpError{Message: fmt.Sprintf("dictionary does not contain key %q", key)}
	}
	for i, k := range d.keys {
		if k == key {
			v := d.values[i]
			// Remove by shifting
			d.keys = append(d.keys[:i], d.keys[i+1:]...)
			d.values = append(d.values[:i], d.values[i+1:]...)
			return v, nil
		}
	}
	return nil, &OpError{Message: fmt.Sprintf("dictionary does not contain key %q", key)}
}

// Remove removes and returns the value for the given key.
// If the key doesn't exist and a default is provided, returns the default.
// Matches Rust's remove method with optional default.
func (d *Dict) Remove(key string, def *syntax.Spanned[Value]) (Value, error) {
	if d == nil {
		if def != nil {
			return def.V, nil
		}
		return nil, &OpError{Message: fmt.Sprintf("dictionary does not contain key %q", key)}
	}
	for i, k := range d.keys {
		if k == key {
			v := d.values[i]
			// Remove by shifting
			d.keys = append(d.keys[:i], d.keys[i+1:]...)
			d.values = append(d.values[:i], d.values[i+1:]...)
			return v, nil
		}
	}
	if def != nil {
		return def.V, nil
	}
	return nil, &OpError{Message: fmt.Sprintf("dictionary does not contain key %q", key)}
}

// Contains returns true if the dictionary contains the given key.
func (d *Dict) Contains(key string) bool {
	_, ok := d.Get(key)
	return ok
}

// Clear removes all entries from the dictionary.
func (d *Dict) Clear() {
	if d == nil {
		return
	}
	d.keys = nil
	d.values = nil
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

// Iter returns an iterator over (key, value) pairs.
// Returns parallel slices for keys and values.
func (d *Dict) Iter() ([]string, []Value) {
	if d == nil {
		return nil, nil
	}
	return d.keys, d.values
}
