// Args type for Typst.
// Translated from foundations/args.rs

package foundations

import (
	"fmt"

	"github.com/boergens/gotypst/syntax"
)

// Args represents captured arguments to a function.
//
// Like built-in functions, custom functions can also take a variable number of
// arguments. You can specify an argument sink which collects all excess
// arguments as ..sink. The resulting sink value is of the arguments type.
type Args struct {
	// Span is the callsite span for the function.
	Span syntax.Span
	// Items contains the positional and named arguments.
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

// NewArgs creates a new empty Args with the given span.
func NewArgs(span syntax.Span) *Args {
	return &Args{Span: span, Items: nil}
}

// NewArgsFrom creates Args from a slice of argument items.
func NewArgsFrom(span syntax.Span, items []Arg) *Args {
	return &Args{Span: span, Items: items}
}

// Spanned attaches a span to these arguments if they don't already have one.
func (a *Args) Spanned(span syntax.Span) *Args {
	if a.Span.IsDetached() {
		a.Span = span
	}
	return a
}

// Remaining returns the number of remaining positional arguments.
func (a *Args) Remaining() int {
	count := 0
	for _, item := range a.Items {
		if item.Name == nil {
			count++
		}
	}
	return count
}

// Insert inserts a positional argument at a specific index.
func (a *Args) Insert(index int, span syntax.Span, value Value) {
	arg := Arg{
		Span:  a.Span,
		Name:  nil,
		Value: syntax.Spanned[Value]{V: value, Span: span},
	}
	if index >= len(a.Items) {
		a.Items = append(a.Items, arg)
	} else {
		a.Items = append(a.Items[:index+1], a.Items[index:]...)
		a.Items[index] = arg
	}
}

// Push adds a positional argument.
func (a *Args) Push(value Value, span syntax.Span) {
	a.Items = append(a.Items, Arg{
		Span:  a.Span,
		Name:  nil,
		Value: syntax.Spanned[Value]{V: value, Span: span},
	})
}

// PushNamed adds a named argument.
func (a *Args) PushNamed(name string, value Value, span syntax.Span) {
	a.Items = append(a.Items, Arg{
		Span:  a.Span,
		Name:  &name,
		Value: syntax.Spanned[Value]{V: value, Span: span},
	})
}

// Eat consumes and returns the first positional argument if there is one.
func (a *Args) Eat() *syntax.Spanned[Value] {
	for i, item := range a.Items {
		if item.Name == nil {
			a.Items = append(a.Items[:i], a.Items[i+1:]...)
			return &item.Value
		}
	}
	return nil
}

// Expect consumes and returns the first positional argument.
// Returns a "missing argument: {what}" error if no positional argument is left.
func (a *Args) Expect(what string) (syntax.Spanned[Value], error) {
	result := a.Eat()
	if result == nil {
		return syntax.Spanned[Value]{}, &MissingArgumentError{Name: what, Span: a.Span}
	}
	return *result, nil
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

// Named retrieves and removes the value for the given named argument.
// When multiple matches exist, removes all of them and returns the last one.
func (a *Args) Named(name string) *syntax.Spanned[Value] {
	var found *syntax.Spanned[Value]
	i := 0
	for i < len(a.Items) {
		if a.Items[i].Name != nil && *a.Items[i].Name == name {
			value := a.Items[i].Value
			a.Items = append(a.Items[:i], a.Items[i+1:]...)
			found = &value
		} else {
			i++
		}
	}
	return found
}

// NamedOrDefault retrieves a named argument or returns the default.
func (a *Args) NamedOrDefault(name string, def Value) syntax.Spanned[Value] {
	if found := a.Named(name); found != nil {
		return *found
	}
	return syntax.Spanned[Value]{V: def, Span: a.Span}
}

// All returns all remaining positional arguments.
func (a *Args) All() []syntax.Spanned[Value] {
	var result []syntax.Spanned[Value]
	i := 0
	for i < len(a.Items) {
		if a.Items[i].Name == nil {
			result = append(result, a.Items[i].Value)
			a.Items = append(a.Items[:i], a.Items[i+1:]...)
		} else {
			i++
		}
	}
	return result
}

// Take takes out all arguments into a new instance.
func (a *Args) Take() *Args {
	items := a.Items
	a.Items = nil
	return &Args{Span: a.Span, Items: items}
}

// Peek returns the first positional argument without consuming it.
// Returns nil if there are no positional arguments.
func (a *Args) Peek() *syntax.Spanned[Value] {
	for _, item := range a.Items {
		if item.Name == nil {
			return &item.Value
		}
	}
	return nil
}

// Finish returns an "unexpected argument" error if there is any remaining argument.
func (a *Args) Finish() error {
	if len(a.Items) > 0 {
		return &UnexpectedArgumentError{Arg: a.Items[0]}
	}
	return nil
}

// IsEmpty returns true if there are no arguments.
func (a *Args) IsEmpty() bool {
	return len(a.Items) == 0
}

// Len returns the number of arguments.
func (a *Args) Len() int {
	return len(a.Items)
}

// HasPositional returns true if there are any positional arguments.
func (a *Args) HasPositional() bool {
	for _, item := range a.Items {
		if item.Name == nil {
			return true
		}
	}
	return false
}

// HasNamed returns true if there is a named argument with the given name.
func (a *Args) HasNamed(name string) bool {
	for _, item := range a.Items {
		if item.Name != nil && *item.Name == name {
			return true
		}
	}
	return false
}

// GetNamed retrieves a named argument without removing it.
func (a *Args) GetNamed(name string) *syntax.Spanned[Value] {
	for _, item := range a.Items {
		if item.Name != nil && *item.Name == name {
			return &item.Value
		}
	}
	return nil
}

// Clone creates a deep copy of the Args.
func (a *Args) Clone() *Args {
	if a == nil {
		return nil
	}
	items := make([]Arg, len(a.Items))
	for i, item := range a.Items {
		items[i] = Arg{
			Span:  item.Span,
			Name:  item.Name,
			Value: syntax.Spanned[Value]{V: item.Value.V.Clone(), Span: item.Value.Span},
		}
	}
	return &Args{Span: a.Span, Items: items}
}

// Pos returns the positional arguments as an Array.
func (a *Args) Pos() *Array {
	result := NewArray()
	for _, item := range a.Items {
		if item.Name == nil {
			result.Push(item.Value.V)
		}
	}
	return result
}

// ToNamed returns the named arguments as a Dict.
func (a *Args) ToNamed() *Dict {
	result := NewDict()
	for _, item := range a.Items {
		if item.Name != nil {
			result.Set(*item.Name, item.Value.V)
		}
	}
	return result
}

// ArgsValue wraps Args as a Value.
type ArgsValue struct {
	Args *Args
}

func (ArgsValue) Type() Type         { return TypeArgs }
func (v ArgsValue) Display() Content { return Content{} }
func (v ArgsValue) Clone() Value {
	return ArgsValue{Args: v.Args.Clone()}
}
func (ArgsValue) isValue() {}

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
