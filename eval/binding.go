package eval

import (
	"fmt"

	"github.com/boergens/gotypst/syntax"
)

// Binding represents a variable binding in a scope.
//
// A binding holds a value along with metadata about where it was defined
// and how it can be accessed.
type Binding struct {
	// Value is the bound value.
	Value Value

	// Span is the source location where this binding was defined.
	Span syntax.Span

	// Kind describes the kind of binding.
	Kind BindingKind

	// Mutable indicates whether this binding can be reassigned.
	Mutable bool

	// Category is an optional documentation category.
	Category *Category
}

// BindingKind describes how a binding was created.
type BindingKind int

const (
	// BindingNormal is a regular variable binding (let x = ...).
	BindingNormal BindingKind = iota

	// BindingClosure is a closure binding (let f(x) = ...).
	BindingClosure

	// BindingModule is a module import binding.
	BindingModule
)

// NewBinding creates a new normal binding.
func NewBinding(value Value, span syntax.Span) Binding {
	return Binding{
		Value:   value,
		Span:    span,
		Kind:    BindingNormal,
		Mutable: false,
	}
}

// NewMutableBinding creates a new mutable binding.
func NewMutableBinding(value Value, span syntax.Span) Binding {
	return Binding{
		Value:   value,
		Span:    span,
		Kind:    BindingNormal,
		Mutable: true,
	}
}

// NewClosureBinding creates a new closure binding.
func NewClosureBinding(value Value, span syntax.Span) Binding {
	return Binding{
		Value:   value,
		Span:    span,
		Kind:    BindingClosure,
		Mutable: false,
	}
}

// NewModuleBinding creates a new module import binding.
func NewModuleBinding(value Value, span syntax.Span) Binding {
	return Binding{
		Value:   value,
		Span:    span,
		Kind:    BindingModule,
		Mutable: false,
	}
}

// Read returns the bound value, checking that it's readable.
// Returns an error if the value is not yet initialized.
func (b *Binding) Read() (Value, error) {
	return b.Value, nil
}

// ReadChecked returns the bound value, performing accessibility checks.
// The span is used for error reporting.
func (b *Binding) ReadChecked(span syntax.Span) (Value, error) {
	// Future: Add checks for uninitialized variables, etc.
	return b.Value, nil
}

// Write updates the bound value if mutable.
// Returns an error if the binding is not mutable.
func (b *Binding) Write(value Value) error {
	if !b.Mutable {
		return &ImmutableBindingError{}
	}
	b.Value = value
	return nil
}

// Clone creates a copy of the binding.
func (b Binding) Clone() Binding {
	return Binding{
		Value:    b.Value.Clone(),
		Span:     b.Span,
		Kind:     b.Kind,
		Mutable:  b.Mutable,
		Category: b.Category,
	}
}

// Category represents a documentation category for bindings.
type Category struct {
	// Name is the category name.
	Name string
}

// ImmutableBindingError is returned when trying to mutate an immutable binding.
type ImmutableBindingError struct{}

func (e *ImmutableBindingError) Error() string {
	return "cannot mutate immutable binding"
}

// ----------------------------------------------------------------------------
// Destructuring
// ----------------------------------------------------------------------------

// BindingFunc is a callback function used during destructuring.
// It receives the expression being bound to and the value to bind.
type BindingFunc func(vm *Vm, expr syntax.Expr, value Value) error

// Destructure destructures a value into a pattern, creating new bindings.
// This is used for let bindings: let (a, b) = expr
func Destructure(vm *Vm, pattern syntax.Pattern, value Value) error {
	return DestructureImpl(vm, pattern, value, func(vm *Vm, expr syntax.Expr, value Value) error {
		ident, ok := expr.(*syntax.IdentExpr)
		if !ok {
			return &CannotAssignError{Span: expr.ToUntyped().Span()}
		}
		vm.DefineWithSpan(ident.Get(), value, ident.ToUntyped().Span())
		return nil
	})
}

// DestructureAssign destructures a value into a pattern, assigning to existing bindings.
// This is used for destructuring assignments: (a, b) = expr
func DestructureAssign(vm *Vm, pattern syntax.Pattern, value Value) error {
	return DestructureImpl(vm, pattern, value, func(vm *Vm, expr syntax.Expr, value Value) error {
		location, err := AccessExpr(vm, expr)
		if err != nil {
			return err
		}
		*location = value
		return nil
	})
}

// DestructureImpl is the core recursive destructuring function.
// It traverses the pattern and applies the binding function to each binding.
func DestructureImpl(vm *Vm, pattern syntax.Pattern, value Value, f BindingFunc) error {
	if pattern == nil {
		return nil
	}

	switch p := pattern.(type) {
	case *syntax.NormalPattern:
		// Normal identifier pattern - apply the binding function
		// Create a proper IdentExpr from the NormalPattern's node
		return f(vm, syntax.ExprFromNode(p.ToUntyped()), value)

	case *syntax.PlaceholderPattern:
		// Placeholder pattern (_) - discard the value
		return nil

	case *syntax.ParenthesizedPattern:
		// Parenthesized pattern - unwrap and recurse
		return DestructureImpl(vm, p.Pattern(), value, f)

	case *syntax.DestructuringPattern:
		// Destructuring pattern - handle arrays and dicts
		switch v := value.(type) {
		case ArrayValue:
			return destructureArray(vm, p, v, f)
		case DictValue:
			return destructureDict(vm, p, v, f)
		case *DictValue:
			return destructureDict(vm, p, *v, f)
		default:
			return &CannotDestructureError{
				Type: value.Type(),
				Span: p.ToUntyped().Span(),
			}
		}

	default:
		return &UnknownPatternError{Span: pattern.ToUntyped().Span()}
	}
}

// destructureArray handles array destructuring patterns.
func destructureArray(vm *Vm, destruct *syntax.DestructuringPattern, value ArrayValue, f BindingFunc) error {
	items := destruct.Items()
	length := len(value)
	var idx int

	for _, item := range items {
		switch it := item.(type) {
		case *syntax.DestructuringBinding:
			// Simple pattern binding
			if idx >= length {
				return wrongNumberOfElements(destruct, length)
			}
			if err := DestructureImpl(vm, it.Pattern(), value[idx], f); err != nil {
				return err
			}
			idx++

		case *syntax.DestructuringSpread:
			// Spread pattern - collect remaining elements
			patternCount := countPatterns(items)
			sinkSize := length + 1 - patternCount
			if sinkSize < 0 || idx+sinkSize > length {
				return wrongNumberOfElements(destruct, length)
			}

			if sink := it.Sink(); sink != nil {
				// Create array from the sink elements
				sinkArray := make(ArrayValue, sinkSize)
				copy(sinkArray, value[idx:idx+sinkSize])
				if err := DestructureImpl(vm, sink, sinkArray, f); err != nil {
					return err
				}
			}
			idx += sinkSize

		case *syntax.DestructuringNamed:
			// Named patterns are not valid for arrays
			return &CannotDestructureNamedFromArrayError{Span: destruct.ToUntyped().Span()}
		}
	}

	if idx < length {
		return wrongNumberOfElements(destruct, length)
	}

	return nil
}

// destructureDict handles dictionary destructuring patterns.
func destructureDict(vm *Vm, destruct *syntax.DestructuringPattern, dict DictValue, f BindingFunc) error {
	items := destruct.Items()
	var sink syntax.Pattern
	used := make(map[string]bool)

	for _, item := range items {
		switch it := item.(type) {
		case *syntax.DestructuringBinding:
			// Check if this is a simple identifier (shorthand for name: name)
			pattern := it.Pattern()
			if normalPat, ok := pattern.(*syntax.NormalPattern); ok {
				name := normalPat.Name()
				val, ok := dict.Get(name)
				if !ok {
					return &KeyNotFoundError{Key: name, Span: normalPat.ToUntyped().Span()}
				}
				expr := syntax.ExprFromNode(normalPat.ToUntyped())
				if err := f(vm, expr, val); err != nil {
					return err
				}
				used[name] = true
			} else {
				// Non-identifier patterns require named syntax for dicts
				return &CannotDestructureUnnamedFromDictError{Span: pattern.ToUntyped().Span()}
			}

		case *syntax.DestructuringNamed:
			// Named pattern: name: pattern
			nameIdent := it.Name()
			if nameIdent == nil {
				continue
			}
			name := nameIdent.Get()
			val, ok := dict.Get(name)
			if !ok {
				return &KeyNotFoundError{Key: name, Span: nameIdent.ToUntyped().Span()}
			}
			if err := DestructureImpl(vm, it.Pattern(), val, f); err != nil {
				return err
			}
			used[name] = true

		case *syntax.DestructuringSpread:
			// Record the sink for later processing
			sink = it.Sink()
		}
	}

	// Handle spread sink - collect unused keys
	if sink != nil {
		sinkDict := NewDict()
		for _, key := range dict.Keys() {
			if !used[key] {
				val, _ := dict.Get(key)
				sinkDict.Set(key, val)
			}
		}
		expr := syntax.ExprFromNode(sink.ToUntyped())
		if err := f(vm, expr, sinkDict); err != nil {
			return err
		}
	}

	return nil
}

// countPatterns counts the number of non-spread patterns in a destructuring.
func countPatterns(items []syntax.DestructuringItem) int {
	count := 0
	for _, item := range items {
		switch item.(type) {
		case *syntax.DestructuringBinding:
			count++
		case *syntax.DestructuringNamed:
			// Named items don't count for array destructuring
		case *syntax.DestructuringSpread:
			// Spread doesn't count
		}
	}
	return count
}

// wrongNumberOfElements creates an error for mismatched destructuring length.
func wrongNumberOfElements(destruct *syntax.DestructuringPattern, length int) error {
	items := destruct.Items()
	count := 0
	hasSpread := false

	for _, item := range items {
		switch item.(type) {
		case *syntax.DestructuringBinding:
			count++
		case *syntax.DestructuringSpread:
			hasSpread = true
		case *syntax.DestructuringNamed:
			// Named items don't count for arrays
		}
	}

	var quantifier string
	if length > count {
		quantifier = "too many"
	} else {
		quantifier = "not enough"
	}

	var expected string
	if hasSpread {
		if count == 1 {
			expected = "at least 1 element"
		} else {
			expected = fmt.Sprintf("at least %d elements", count)
		}
	} else {
		if count == 0 {
			expected = "an empty array"
		} else if count == 1 {
			expected = "a single element"
		} else {
			expected = fmt.Sprintf("%d elements", count)
		}
	}

	return &WrongNumberOfElementsError{
		Quantifier: quantifier,
		Expected:   expected,
		Got:        length,
		Span:       destruct.ToUntyped().Span(),
	}
}

// ----------------------------------------------------------------------------
// Destructuring Error Types
// ----------------------------------------------------------------------------

// CannotAssignError is returned when trying to assign to a non-assignable expression.
type CannotAssignError struct {
	Span syntax.Span
}

func (e *CannotAssignError) Error() string {
	return "cannot assign to this expression"
}

// CannotDestructureError is returned when a value cannot be destructured.
type CannotDestructureError struct {
	Type Type
	Span syntax.Span
}

func (e *CannotDestructureError) Error() string {
	return fmt.Sprintf("cannot destructure %s", e.Type)
}

// UnknownPatternError is returned for unrecognized pattern types.
type UnknownPatternError struct {
	Span syntax.Span
}

func (e *UnknownPatternError) Error() string {
	return "unknown pattern type"
}

// CannotDestructureNamedFromArrayError is returned when using named patterns on arrays.
type CannotDestructureNamedFromArrayError struct {
	Span syntax.Span
}

func (e *CannotDestructureNamedFromArrayError) Error() string {
	return "cannot destructure named pattern from an array"
}

// CannotDestructureUnnamedFromDictError is returned when using unnamed patterns on dicts.
type CannotDestructureUnnamedFromDictError struct {
	Span syntax.Span
}

func (e *CannotDestructureUnnamedFromDictError) Error() string {
	return "cannot destructure unnamed pattern from dictionary"
}

// WrongNumberOfElementsError is returned when destructuring length doesn't match.
type WrongNumberOfElementsError struct {
	Quantifier string
	Expected   string
	Got        int
	Span       syntax.Span
}

func (e *WrongNumberOfElementsError) Error() string {
	return fmt.Sprintf("%s elements to destructure; the provided array has a length of %d, but the pattern expects %s",
		e.Quantifier, e.Got, e.Expected)
}
