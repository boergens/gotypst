package eval

import "github.com/boergens/gotypst/syntax"

// Scope represents a single lexical scope containing variable bindings.
//
// A scope maps identifiers to bindings. Scopes are organized in a stack
// where inner scopes can shadow outer scope bindings.
type Scope struct {
	// bindings maps identifier names to their bindings.
	// Uses a slice to preserve insertion order for iteration.
	bindings []scopeEntry

	// deduplicate prevents duplicate definitions in this scope.
	deduplicate bool

	// category is an optional documentation category for this scope.
	category *Category
}

// scopeEntry represents a single binding entry in a scope.
type scopeEntry struct {
	name    string
	binding Binding
}

// NewScope creates a new empty scope.
func NewScope() *Scope {
	return &Scope{
		bindings:    nil,
		deduplicate: false,
	}
}

// NewScopeWithCategory creates a new scope with a documentation category.
func NewScopeWithCategory(category *Category) *Scope {
	return &Scope{
		bindings: nil,
		category: category,
	}
}

// Define binds a value to an identifier in this scope.
func (s *Scope) Define(name string, value Value, span syntax.Span) {
	s.Insert(name, NewBinding(value, span))
}

// DefineFunc binds a function to an identifier in this scope.
func (s *Scope) DefineFunc(name string, f *Func) {
	s.Insert(name, NewClosureBinding(FuncValue{Func: f}, f.Span))
}

// Insert adds a binding to this scope.
// If a binding with the same name already exists, it is replaced.
func (s *Scope) Insert(name string, binding Binding) {
	// Apply scope category if binding doesn't have one
	if binding.Category == nil && s.category != nil {
		binding.Category = s.category
	}

	// Check for existing binding
	for i, entry := range s.bindings {
		if entry.name == name {
			s.bindings[i].binding = binding
			return
		}
	}

	// Add new binding
	s.bindings = append(s.bindings, scopeEntry{name: name, binding: binding})
}

// Get retrieves a binding by name.
// Returns nil if not found.
func (s *Scope) Get(name string) *Binding {
	for i := range s.bindings {
		if s.bindings[i].name == name {
			return &s.bindings[i].binding
		}
	}
	return nil
}

// GetMut retrieves a mutable reference to a binding by name.
// Returns nil if not found.
func (s *Scope) GetMut(name string) *Binding {
	return s.Get(name) // In Go, all references are mutable if the struct is
}

// Contains returns true if the scope contains a binding for the given name.
func (s *Scope) Contains(name string) bool {
	return s.Get(name) != nil
}

// Remove removes a binding from the scope.
// Returns true if a binding was removed.
func (s *Scope) Remove(name string) bool {
	for i, entry := range s.bindings {
		if entry.name == name {
			s.bindings = append(s.bindings[:i], s.bindings[i+1:]...)
			return true
		}
	}
	return false
}

// Names returns all binding names in insertion order.
func (s *Scope) Names() []string {
	names := make([]string, len(s.bindings))
	for i, entry := range s.bindings {
		names[i] = entry.name
	}
	return names
}

// Len returns the number of bindings in the scope.
func (s *Scope) Len() int {
	return len(s.bindings)
}

// IsEmpty returns true if the scope has no bindings.
func (s *Scope) IsEmpty() bool {
	return len(s.bindings) == 0
}

// Iter returns an iterator over the bindings.
func (s *Scope) Iter() []scopeEntry {
	return s.bindings
}

// Clone creates a shallow copy of the scope.
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

// SetDeduplicate sets whether this scope prevents duplicate definitions.
func (s *Scope) SetDeduplicate(deduplicate bool) {
	s.deduplicate = deduplicate
}

// SetCategory sets the documentation category for this scope.
func (s *Scope) SetCategory(category *Category) {
	s.category = category
}

// ----------------------------------------------------------------------------
// Scopes - Stack of Scopes
// ----------------------------------------------------------------------------

// Scopes represents a stack of lexical scopes.
//
// The scope stack supports nested scopes with variable shadowing.
// The `top` scope is the currently active scope, while `scopes` contains
// the outer (parent) scopes. An optional `base` scope provides access
// to the standard library.
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

// SetTop replaces the top scope.
func (s *Scopes) SetTop(top *Scope) {
	s.top = top
}

// Enter pushes a new scope onto the stack.
func (s *Scopes) Enter() {
	s.scopes = append(s.scopes, s.top)
	s.top = NewScope()
}

// Exit pops the current scope from the stack.
// Panics if there are no scopes to pop.
func (s *Scopes) Exit() {
	if len(s.scopes) == 0 {
		panic("cannot exit: no scopes on stack")
	}
	s.top = s.scopes[len(s.scopes)-1]
	s.scopes = s.scopes[:len(s.scopes)-1]
}

// Get looks up a binding by identifier, searching from top to bottom.
// Returns nil if not found in any scope.
func (s *Scopes) Get(name string) *Binding {
	// Check top scope first
	if binding := s.top.Get(name); binding != nil {
		return binding
	}

	// Check intermediate scopes from most recent to oldest
	for i := len(s.scopes) - 1; i >= 0; i-- {
		if binding := s.scopes[i].Get(name); binding != nil {
			return binding
		}
	}

	// Check base scope
	if s.base != nil {
		return s.base.Get(name)
	}

	return nil
}

// GetMut looks up a mutable binding by identifier.
// Returns nil if not found in any scope.
func (s *Scopes) GetMut(name string) *Binding {
	return s.Get(name) // In Go, returned pointers are mutable
}

// GetInScopes looks up a binding only in the local scopes (excluding base).
func (s *Scopes) GetInScopes(name string) *Binding {
	// Check top scope first
	if binding := s.top.Get(name); binding != nil {
		return binding
	}

	// Check intermediate scopes from most recent to oldest
	for i := len(s.scopes) - 1; i >= 0; i-- {
		if binding := s.scopes[i].Get(name); binding != nil {
			return binding
		}
	}

	return nil
}

// Define binds a value in the top scope.
func (s *Scopes) Define(name string, value Value, span syntax.Span) {
	s.top.Define(name, value, span)
}

// DefineFunc binds a function in the top scope.
func (s *Scopes) DefineFunc(name string, f *Func) {
	s.top.DefineFunc(name, f)
}

// Insert adds a binding to the top scope.
func (s *Scopes) Insert(name string, binding Binding) {
	s.top.Insert(name, binding)
}

// Contains returns true if any scope contains a binding for the given name.
func (s *Scopes) Contains(name string) bool {
	return s.Get(name) != nil
}

// Depth returns the current nesting depth (number of scopes on stack + 1 for top).
func (s *Scopes) Depth() int {
	return len(s.scopes) + 1
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

// FlattenToScope flattens all scopes into a single scope.
// This is useful for closure capture.
func (s *Scopes) FlattenToScope() *Scope {
	result := NewScope()

	// Add base scope first (lowest priority)
	if s.base != nil {
		for _, entry := range s.base.Iter() {
			result.Insert(entry.name, entry.binding)
		}
	}

	// Add intermediate scopes (oldest first)
	for _, scope := range s.scopes {
		for _, entry := range scope.Iter() {
			result.Insert(entry.name, entry.binding)
		}
	}

	// Add top scope (highest priority)
	for _, entry := range s.top.Iter() {
		result.Insert(entry.name, entry.binding)
	}

	return result
}
