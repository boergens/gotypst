package symbols

import (
	"strings"
)

// Symbol represents a named symbol that may have multiple variants.
// A symbol with no modifiers has a default variant.
// Additional variants are accessed via dot-separated modifiers.
type Symbol struct {
	// Name is the base name of the symbol (e.g., "arrow").
	Name string
	// Variants maps modifier paths to Unicode strings.
	// The empty string key "" is the default variant.
	// Keys like "r" map to arrow.r, "r.double" maps to arrow.r.double.
	Variants map[string]string
}

// Get returns the Unicode string for the given modifiers.
// Returns empty string if the variant doesn't exist.
func (s *Symbol) Get(modifiers ...string) string {
	key := strings.Join(modifiers, ".")
	return s.Variants[key]
}

// Default returns the default (unmodified) variant.
func (s *Symbol) Default() string {
	return s.Variants[""]
}

// HasVariant returns true if the symbol has the specified variant.
func (s *Symbol) HasVariant(modifiers ...string) bool {
	key := strings.Join(modifiers, ".")
	_, ok := s.Variants[key]
	return ok
}

// AllVariants returns all variant keys for this symbol.
func (s *Symbol) AllVariants() []string {
	keys := make([]string, 0, len(s.Variants))
	for k := range s.Variants {
		keys = append(keys, k)
	}
	return keys
}

// Module represents a named module containing symbols and submodules.
type Module struct {
	// Name is the module name.
	Name string
	// Symbols maps symbol names to their definitions.
	Symbols map[string]*Symbol
	// Submodules maps submodule names to their definitions.
	Submodules map[string]*Module
}

// Get looks up a symbol by name.
// Returns nil if not found.
func (m *Module) Get(name string) *Symbol {
	return m.Symbols[name]
}

// GetSubmodule looks up a submodule by name.
// Returns nil if not found.
func (m *Module) GetSubmodule(name string) *Module {
	return m.Submodules[name]
}

// Lookup resolves a dotted path to a symbol value.
// For example, "arrow.r.double" looks up the arrow symbol and returns
// the "r.double" variant.
// Returns the Unicode string and true if found, empty string and false otherwise.
func (m *Module) Lookup(path string) (string, bool) {
	parts := strings.SplitN(path, ".", 2)
	if len(parts) == 0 {
		return "", false
	}

	name := parts[0]

	// First, check if it's a submodule access
	if sub := m.Submodules[name]; sub != nil {
		if len(parts) == 1 {
			return "", false // Can't get value of a submodule directly
		}
		return sub.Lookup(parts[1])
	}

	// Look up the symbol
	sym := m.Symbols[name]
	if sym == nil {
		return "", false
	}

	// If no modifiers, return default
	if len(parts) == 1 {
		return sym.Default(), true
	}

	// Look up the variant
	variant := sym.Variants[parts[1]]
	if variant == "" {
		return "", false
	}
	return variant, true
}

// newSymbol creates a new symbol with the given name and variants.
func newSymbol(name string, variants map[string]string) *Symbol {
	return &Symbol{
		Name:     name,
		Variants: variants,
	}
}

// singleSymbol creates a symbol with only a default variant.
func singleSymbol(name, value string) *Symbol {
	return &Symbol{
		Name:     name,
		Variants: map[string]string{"": value},
	}
}

// Sym is the main symbol module containing general and math symbols.
var Sym *Module

// Emoji is the emoji symbol module.
var Emoji *Module

// Root is the root module containing both sym and emoji.
var Root *Module

func init() {
	Sym = buildSymModule()
	Emoji = buildEmojiModule()
	Root = &Module{
		Name:    "",
		Symbols: nil,
		Submodules: map[string]*Module{
			"sym":   Sym,
			"emoji": Emoji,
		},
	}
}
