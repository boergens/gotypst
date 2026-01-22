// Scope and Binding types for Typst.
// Translated from foundations/scope.rs

package foundations

import (
	"github.com/boergens/gotypst/syntax"
)

// Scope represents a single lexical scope containing variable bindings.
type Scope struct {
	// bindings maps identifier names to their bindings.
	bindings []scopeEntry
	// deduplicate prevents duplicate definitions.
	deduplicate bool
	// category is an optional documentation category.
	category *Category
}

type scopeEntry struct {
	name    string
	binding Binding
}

// Category represents a documentation category.
type Category struct {
	Name string
}

// Binding represents a bound value with metadata.
// Translated from Rust's Binding struct in scope.rs.
type Binding struct {
	// value is the bound value.
	value Value
	// kind determines how the value can be accessed.
	kind BindingKind
	// span is a span associated with the binding.
	span syntax.Span
	// category is the optional documentation category.
	category *Category
	// TODO: deprecation field for deprecation warnings
}

// BindingKind determines how a binding can be accessed.
type BindingKind int

const (
	// BindingNormal is a normal, mutable binding.
	BindingNormal BindingKind = iota
	// BindingCapturedFunction is captured by a function/closure (read-only).
	BindingCapturedFunction
	// BindingCapturedContext is captured by a context expression (read-only).
	BindingCapturedContext
)

// Capturer represents what kind of construct captures a variable.
// Used by CapturesVisitor to determine how to capture variables.
type Capturer int

const (
	// CapturerFunction indicates capture by a function/closure.
	CapturerFunction Capturer = iota
	// CapturerContext indicates capture by a context expression.
	CapturerContext
)

// ToBindingKind converts a Capturer to the corresponding BindingKind.
func (c Capturer) ToBindingKind() BindingKind {
	switch c {
	case CapturerFunction:
		return BindingCapturedFunction
	case CapturerContext:
		return BindingCapturedContext
	default:
		return BindingNormal
	}
}

// NewBinding creates a new binding with a span marking its definition site.
func NewBinding(value Value, span syntax.Span) Binding {
	return Binding{
		value:    value,
		span:     span,
		kind:     BindingNormal,
		category: nil,
	}
}

// NewBindingDetached creates a binding without a span.
func NewBindingDetached(value Value) Binding {
	return NewBinding(value, syntax.Span{})
}

// Read returns the bound value.
func (b *Binding) Read() Value {
	return b.value
}

// ReadChecked returns the bound value, checking for deprecation.
// TODO: implement deprecation checking
func (b *Binding) ReadChecked(span syntax.Span) Value {
	return b.value
}

// Write tries to write to the value.
// Returns an error if the value is a read-only closure capture.
// Matches Rust's Binding::write method.
func (b *Binding) Write(value Value) error {
	switch b.kind {
	case BindingNormal:
		b.value = value
		return nil
	case BindingCapturedFunction:
		return &CapturedVariableError{Capturer: "function"}
	case BindingCapturedContext:
		return &CapturedVariableError{Capturer: "context expression"}
	default:
		b.value = value
		return nil
	}
}

// Slot returns a pointer to the value for in-place mutation.
// Returns an error if the binding is captured (read-only).
// This is the Go equivalent of Rust's write() returning &mut Value.
func (b *Binding) Slot() (*Value, error) {
	switch b.kind {
	case BindingNormal:
		return &b.value, nil
	case BindingCapturedFunction:
		return nil, &CapturedVariableError{Capturer: "function"}
	case BindingCapturedContext:
		return nil, &CapturedVariableError{Capturer: "context expression"}
	default:
		return &b.value, nil
	}
}

// Capture creates a copy of the binding for closure capturing.
// Takes a Capturer to determine the binding kind.
func (b *Binding) Capture(capturer Capturer) Binding {
	return Binding{
		value:    b.value.Clone(),
		kind:     capturer.ToBindingKind(),
		span:     b.span,
		category: b.category,
	}
}

// Span returns the span associated with the binding.
func (b *Binding) Span() syntax.Span {
	return b.span
}

// Category returns the category of the binding.
func (b *Binding) Category() *Category {
	return b.category
}

// Value returns the bound value (for direct access).
func (b *Binding) Value() Value {
	return b.value
}

// Clone creates a copy of the binding.
func (b Binding) Clone() Binding {
	return Binding{
		value:    b.value.Clone(),
		kind:     b.kind,
		span:     b.span,
		category: b.category,
	}
}

// CapturedVariableError is returned when trying to mutate a captured variable.
type CapturedVariableError struct {
	Capturer string
}

func (e *CapturedVariableError) Error() string {
	return "variables from outside the " + e.Capturer + " are read-only and cannot be modified"
}

// NewScope creates a new empty scope.
func NewScope() *Scope {
	return &Scope{bindings: nil}
}

// Define binds a value to an identifier in this scope.
func (s *Scope) Define(name string, value Value, span syntax.Span) {
	s.Insert(name, NewBinding(value, span))
}

// Insert adds a binding to this scope.
func (s *Scope) Insert(name string, binding Binding) {
	if binding.category == nil && s.category != nil {
		binding.category = s.category
	}
	for i, entry := range s.bindings {
		if entry.name == name {
			s.bindings[i].binding = binding
			return
		}
	}
	s.bindings = append(s.bindings, scopeEntry{name: name, binding: binding})
}

// Get retrieves a binding by name.
func (s *Scope) Get(name string) *Binding {
	for i := range s.bindings {
		if s.bindings[i].name == name {
			return &s.bindings[i].binding
		}
	}
	return nil
}

// GetMut retrieves a mutable binding by name.
// Returns nil if not found.
func (s *Scope) GetMut(name string) *Binding {
	return s.Get(name) // In Go, pointers are always mutable
}

// Contains returns true if the scope contains a binding for the given name.
func (s *Scope) Contains(name string) bool {
	return s.Get(name) != nil
}

// Clone creates a copy of the scope.
func (s *Scope) Clone() *Scope {
	if s == nil {
		return nil
	}
	clone := &Scope{
		bindings:    make([]scopeEntry, len(s.bindings)),
		deduplicate: s.deduplicate,
		category:    s.category,
	}
	for i, entry := range s.bindings {
		clone.bindings[i] = scopeEntry{
			name:    entry.name,
			binding: entry.binding.Clone(),
		}
	}
	return clone
}

// Iter iterates over all bindings in the scope, calling f for each.
func (s *Scope) Iter(f func(name string, binding Binding)) {
	if s == nil {
		return
	}
	for _, entry := range s.bindings {
		f(entry.name, entry.binding)
	}
}

// Len returns the number of bindings in the scope.
func (s *Scope) Len() int {
	if s == nil {
		return 0
	}
	return len(s.bindings)
}

// Entries returns the raw binding entries for iteration.
func (s *Scope) Entries() []scopeEntry {
	if s == nil {
		return nil
	}
	return s.bindings
}

// SetCategory sets the documentation category for this scope.
func (s *Scope) SetCategory(category *Category) {
	s.category = category
}

// ----------------------------------------------------------------------------
// Scopes - Stack of Scopes
// ----------------------------------------------------------------------------

// Scopes represents a stack of lexical scopes.
// Translated from Rust's Scopes struct in scope.rs.
type Scopes struct {
	// top is the currently active scope.
	top *Scope
	// scopes contains the outer scopes (parents of top).
	scopes []*Scope
	// base is the optional standard library scope.
	base *Scope
}

// NewScopes creates a new scope stack with an optional base scope.
func NewScopes(base *Scope) *Scopes {
	return &Scopes{
		top:    NewScope(),
		scopes: nil,
		base:   base,
	}
}

// Top returns the current top scope.
func (s *Scopes) Top() *Scope {
	return s.top
}

// Enter pushes a new scope onto the stack.
func (s *Scopes) Enter() {
	s.scopes = append(s.scopes, s.top)
	s.top = NewScope()
}

// Exit pops the current scope from the stack.
func (s *Scopes) Exit() {
	if len(s.scopes) == 0 {
		panic("cannot exit: no scopes on stack")
	}
	s.top = s.scopes[len(s.scopes)-1]
	s.scopes = s.scopes[:len(s.scopes)-1]
}

// Get looks up a binding by identifier, searching from top to bottom.
func (s *Scopes) Get(name string) *Binding {
	if binding := s.top.Get(name); binding != nil {
		return binding
	}
	for i := len(s.scopes) - 1; i >= 0; i-- {
		if binding := s.scopes[i].Get(name); binding != nil {
			return binding
		}
	}
	if s.base != nil {
		return s.base.Get(name)
	}
	return nil
}

// GetMut looks up a mutable binding by identifier.
func (s *Scopes) GetMut(name string) *Binding {
	return s.Get(name)
}

// Define binds a value in the top scope.
func (s *Scopes) Define(name string, value Value, span syntax.Span) {
	s.top.Define(name, value, span)
}

// Insert adds a binding to the top scope.
func (s *Scopes) Insert(name string, binding Binding) {
	s.top.Insert(name, binding)
}

// Clone creates a copy of the scope stack.
func (s *Scopes) Clone() *Scopes {
	if s == nil {
		return nil
	}
	clone := &Scopes{
		top:    s.top.Clone(),
		scopes: make([]*Scope, len(s.scopes)),
		base:   s.base, // Base is shared, not cloned
	}
	for i, scope := range s.scopes {
		clone.scopes[i] = scope.Clone()
	}
	return clone
}

// SetTop replaces the current top scope.
// This is used when restoring captured scope for closure calls.
func (s *Scopes) SetTop(scope *Scope) {
	s.top = scope
}

// FlattenToScope returns all bindings in the scope stack as a single scope.
// This is useful for capturing all accessible variables for closures.
func (s *Scopes) FlattenToScope() *Scope {
	result := NewScope()

	// Add base scope bindings first.
	if s.base != nil {
		s.base.Iter(func(name string, binding Binding) {
			result.Insert(name, binding.Clone())
		})
	}

	// Add parent scope bindings (bottom to top, so later scopes shadow earlier).
	for _, scope := range s.scopes {
		scope.Iter(func(name string, binding Binding) {
			result.Insert(name, binding.Clone())
		})
	}

	// Add top scope bindings last (highest priority).
	s.top.Iter(func(name string, binding Binding) {
		result.Insert(name, binding.Clone())
	})

	return result
}

// GetInMath looks up a binding for a math identifier.
// Math identifiers are multi-letter identifiers like "foo" in math mode.
// Single-letter identifiers in math are not captured (they're interpreted as symbols).
// This matches Rust's Scopes::get_in_math method.
func (s *Scopes) GetInMath(name string) *Binding {
	// Math identifiers behave the same as regular identifiers for now
	// The difference is in which identifiers are considered "captures"
	// - single letter idents in math mode are not captured (they're symbols)
	// - multi-letter idents in math mode are captured
	return s.Get(name)
}

// Bind adds a detached binding to the top scope.
// This is used by CapturesVisitor to track internal bindings.
func (s *Scope) Bind(name string, binding Binding) {
	s.Insert(name, binding)
}
