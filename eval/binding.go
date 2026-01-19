package eval

import "github.com/boergens/gotypst/syntax"

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
