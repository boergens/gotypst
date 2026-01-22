// Args type for Typst.
// Translated from foundations/args.rs

package foundations

import (
	"fmt"

	"github.com/boergens/gotypst/syntax"
)

// Args represents captured arguments to a function.
// Matches Rust: pub struct Args
type Args struct {
	// Span is the callsite span for the function (not the argument list itself,
	// but of the whole function call).
	Span syntax.Span
	// Items contains the positional and named arguments.
	Items []Arg
}

// Arg represents a single argument to a function call.
// Matches Rust: pub struct Arg
type Arg struct {
	// Span is the span of the whole argument.
	Span syntax.Span
	// Name is the name of the argument (None/nil for positional arguments).
	Name *Str
	// Value is the value of the argument.
	Value syntax.Spanned[Value]
}

// NewArgs creates positional arguments from a span and values.
// Matches Rust: pub fn new<T: IntoValue>(span: Span, values: impl IntoIterator<Item = T>) -> Self
func NewArgs(span syntax.Span, values ...Value) *Args {
	items := make([]Arg, 0, len(values))
	for _, value := range values {
		items = append(items, Arg{
			Span:  span,
			Name:  nil,
			Value: syntax.Spanned[Value]{V: value, Span: span},
		})
	}
	return &Args{Span: span, Items: items}
}

// Spanned attaches a span to these arguments if they don't already have one.
// Matches Rust: pub fn spanned(mut self, span: Span) -> Self
func (a *Args) Spanned(span syntax.Span) *Args {
	if a.Span.IsDetached() {
		a.Span = span
	}
	return a
}

// Remaining returns the number of remaining positional arguments.
// Matches Rust: pub fn remaining(&self) -> usize
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
// Matches Rust: pub fn insert(&mut self, index: usize, span: Span, value: Value)
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
// Matches Rust: pub fn push(&mut self, span: Span, value: Value)
func (a *Args) Push(span syntax.Span, value Value) {
	a.Items = append(a.Items, Arg{
		Span:  a.Span,
		Name:  nil,
		Value: syntax.Spanned[Value]{V: value, Span: span},
	})
}

// Eat consumes and returns the first positional argument if there is one.
// Matches Rust: pub fn eat<T>(&mut self) -> SourceResult<Option<T>>
func (a *Args) Eat() *syntax.Spanned[Value] {
	for i, item := range a.Items {
		if item.Name == nil {
			value := item.Value
			a.Items = append(a.Items[:i], a.Items[i+1:]...)
			return &value
		}
	}
	return nil
}

// Consume consumes n positional arguments if possible.
// Matches Rust: pub fn consume(&mut self, n: usize) -> SourceResult<Vec<Arg>>
func (a *Args) Consume(n int) ([]Arg, error) {
	var list []Arg
	i := 0
	for i < len(a.Items) && len(list) < n {
		if a.Items[i].Name == nil {
			list = append(list, a.Items[i])
			a.Items = append(a.Items[:i], a.Items[i+1:]...)
		} else {
			i++
		}
	}
	if len(list) < n {
		return nil, fmt.Errorf("not enough arguments")
	}
	return list, nil
}

// Expect consumes and returns the first positional argument.
// Returns a "missing argument: {what}" error if no positional argument is left.
// Matches Rust: pub fn expect<T>(&mut self, what: &str) -> SourceResult<T>
func (a *Args) Expect(what string) (syntax.Spanned[Value], error) {
	result := a.Eat()
	if result == nil {
		// Check if there's a named argument with this name (positional/named confusion)
		for _, item := range a.Items {
			if item.Name != nil && string(*item.Name) == what {
				return syntax.Spanned[Value]{}, &PositionalArgumentError{
					Name: what,
					Span: item.Span,
				}
			}
		}
		return syntax.Spanned[Value]{}, &MissingArgumentError{Name: what, Span: a.Span}
	}
	return *result, nil
}

// Named retrieves and removes the value for the given named argument.
// When multiple matches exist, removes all of them and returns the last one.
// Matches Rust: pub fn named<T>(&mut self, name: &str) -> SourceResult<Option<T>>
func (a *Args) Named(name string) *syntax.Spanned[Value] {
	var found *syntax.Spanned[Value]
	i := 0
	for i < len(a.Items) {
		if a.Items[i].Name != nil && string(*a.Items[i].Name) == name {
			value := a.Items[i].Value
			a.Items = append(a.Items[:i], a.Items[i+1:]...)
			found = &value
		} else {
			i++
		}
	}
	return found
}

// Take takes out all arguments into a new instance.
// Matches Rust: pub fn take(&mut self) -> Self
func (a *Args) Take() *Args {
	items := a.Items
	a.Items = nil
	return &Args{Span: a.Span, Items: items}
}

// Finish returns an "unexpected argument" error if there is any remaining argument.
// Matches Rust: pub fn finish(self) -> SourceResult<()>
func (a *Args) Finish() error {
	if len(a.Items) > 0 {
		arg := a.Items[0]
		if arg.Name != nil {
			return fmt.Errorf("unexpected argument: %s", *arg.Name)
		}
		return fmt.Errorf("unexpected argument")
	}
	return nil
}

// IsEmpty returns true if there are no arguments.
// Matches Rust: pub fn is_empty(&self) -> bool
func (a *Args) IsEmpty() bool {
	return len(a.Items) == 0
}

// Len returns the number of arguments.
// Matches Rust: #[func(title = "Length")] pub fn len(&self) -> usize
func (a *Args) Len() int {
	return len(a.Items)
}

// Pos returns the positional arguments as an Array.
// Matches Rust: #[func(name = "pos")] pub fn to_pos(&self) -> Array
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
// Matches Rust: #[func(name = "named")] pub fn to_named(&self) -> Dict
func (a *Args) ToNamed() *Dict {
	result := NewDict()
	for _, item := range a.Items {
		if item.Name != nil {
			result.Set(string(*item.Name), item.Value.V)
		}
	}
	return result
}

// Clone creates a deep copy of the Args.
func (a *Args) Clone() *Args {
	if a == nil {
		return nil
	}
	items := make([]Arg, len(a.Items))
	for i, item := range a.Items {
		var nameCopy *Str
		if item.Name != nil {
			n := *item.Name
			nameCopy = &n
		}
		items[i] = Arg{
			Span:  item.Span,
			Name:  nameCopy,
			Value: syntax.Spanned[Value]{V: item.Value.V.Clone(), Span: item.Value.Span},
		}
	}
	return &Args{Span: a.Span, Items: items}
}

// All returns all remaining positional argument values.
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

// ----------------------------------------------------------------------------
// ArgsValue - Value wrapper
// ----------------------------------------------------------------------------

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

// ----------------------------------------------------------------------------
// Error Types
// ----------------------------------------------------------------------------

// MissingArgumentError is returned when a required argument is missing.
// Matches Rust: error!(self.span, "missing argument: {what}")
type MissingArgumentError struct {
	Name string
	Span syntax.Span
}

func (e *MissingArgumentError) Error() string {
	return fmt.Sprintf("missing argument: %s", e.Name)
}

// PositionalArgumentError is returned when a named argument should be positional.
// Matches Rust: error!(item.span, "the argument `{what}` is positional"; hint: ...)
type PositionalArgumentError struct {
	Name string
	Span syntax.Span
}

func (e *PositionalArgumentError) Error() string {
	return fmt.Sprintf("the argument `%s` is positional", e.Name)
}

// UnexpectedArgumentError is returned when an unexpected argument is provided.
// Matches Rust: bail!(arg.span, "unexpected argument: {name}")
type UnexpectedArgumentError struct {
	Arg Arg
}

func (e *UnexpectedArgumentError) Error() string {
	if e.Arg.Name != nil {
		return fmt.Sprintf("unexpected argument: %s", *e.Arg.Name)
	}
	return "unexpected argument"
}
