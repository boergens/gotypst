// Styles types for Typst.
// Translated from typst-library/src/foundations/styles.rs
//
// Selector, Transformation, and Recipe types are in selector.go.

package foundations

import (
	"github.com/boergens/gotypst/syntax"
)

// StylesValue represents a collection of styles as a Value.
type StylesValue struct {
	Styles *Styles
}

func (StylesValue) Type() Type         { return TypeStyles }
func (v StylesValue) Display() Content { return Content{} }
func (v StylesValue) Clone() Value     { return v }
func (StylesValue) isValue()           {}

// Styles represents a collection of style rules and recipes.
// Corresponds to Rust's Styles struct in styles.rs.
type Styles struct {
	// Rules contains the style rules (from set rules).
	Rules []StyleRule
	// Recipes contains the show rule recipes.
	Recipes []*Recipe
}

// NewStyles creates a new empty Styles collection.
func NewStyles() *Styles {
	return &Styles{
		Rules:   nil,
		Recipes: nil,
	}
}

// IsEmpty returns true if there are no rules or recipes.
func (s *Styles) IsEmpty() bool {
	return s == nil || (len(s.Rules) == 0 && len(s.Recipes) == 0)
}

// AddRule adds a style rule.
func (s *Styles) AddRule(rule StyleRule) {
	s.Rules = append(s.Rules, rule)
}

// AddRecipe adds a recipe.
func (s *Styles) AddRecipe(recipe *Recipe) {
	s.Recipes = append(s.Recipes, recipe)
}

// StyleRule represents a single style rule from a set rule.
// Corresponds to Rust's Style::Property variant.
type StyleRule struct {
	// Func is the element function this style applies to.
	Func *Func
	// Args are the style arguments.
	Args *Args
	// Span is the source location of the rule.
	Span syntax.Span
	// Liftable indicates whether this style can be lifted to page level.
	Liftable bool
}

// Style represents a single style entry (rule or recipe).
// Corresponds to Rust's Style enum.
type Style interface {
	isStyle()
	// StyleSpan returns the source span of this style.
	StyleSpan() syntax.Span
}

// PropertyStyle is a set rule style.
type PropertyStyle struct {
	Rule StyleRule
}

func (PropertyStyle) isStyle()                 {}
func (p PropertyStyle) StyleSpan() syntax.Span { return p.Rule.Span }

// RecipeStyle is a show rule style.
type RecipeStyle struct {
	Recipe *Recipe
}

func (RecipeStyle) isStyle()                 {}
func (r RecipeStyle) StyleSpan() syntax.Span { return r.Recipe.Span }

// Element represents an element type (e.g., text, heading, par).
// Corresponds to Rust's Element type.
type Element struct {
	// Name is the element name.
	Name string
}

// ----------------------------------------------------------------------------
// StyleChain
// ----------------------------------------------------------------------------

// StyleChain is a linked-list structure for tracking accumulated styles
// during realization and layout. It allows efficient style lookup without
// copying or merging style lists at each level.
//
// Corresponds to Rust's StyleChain in styles.rs.
//
// The chain is traversed from innermost (most recent) to outermost (base),
// with the first matching property winning.
type StyleChain struct {
	// styles are the styles at this level of the chain.
	styles *Styles
	// parent is the outer/parent chain (nil for root).
	parent *StyleChain
}

// NewStyleChain creates a new style chain with the given root styles.
func NewStyleChain(styles *Styles) *StyleChain {
	return &StyleChain{
		styles: styles,
		parent: nil,
	}
}

// EmptyStyleChain returns an empty style chain.
func EmptyStyleChain() *StyleChain {
	return &StyleChain{
		styles: nil,
		parent: nil,
	}
}

// Chain creates a new chain with the given styles as the innermost level,
// and the current chain as the parent. This is the primary way to "push"
// styles when descending into styled content.
//
// Corresponds to Rust's Chainable::chain method.
func (sc *StyleChain) Chain(inner *Styles) *StyleChain {
	if inner == nil || inner.IsEmpty() {
		return sc
	}
	return &StyleChain{
		styles: inner,
		parent: sc,
	}
}

// IsEmpty returns true if the chain has no styles at any level.
func (sc *StyleChain) IsEmpty() bool {
	if sc == nil {
		return true
	}
	if sc.styles != nil && !sc.styles.IsEmpty() {
		return false
	}
	return sc.parent.IsEmpty()
}

// ToStyles converts the chain to a flat Styles struct.
// This is an alias for AllStyles for compatibility.
func (sc *StyleChain) ToStyles() *Styles {
	return sc.AllStyles()
}

// Get retrieves a property value from the style chain for a given element
// function and property name. It walks the chain from innermost to outermost,
// returning the first matching value.
//
// Returns nil if the property is not found in any level.
func (sc *StyleChain) Get(funcName string, propName string) Value {
	for chain := sc; chain != nil; chain = chain.parent {
		if chain.styles == nil {
			continue
		}
		// Search rules in reverse order within this level (later rules take precedence)
		for i := len(chain.styles.Rules) - 1; i >= 0; i-- {
			rule := chain.styles.Rules[i]
			if rule.Func != nil && rule.Func.Name != nil && *rule.Func.Name == funcName {
				if rule.Args != nil {
					if val := rule.Args.GetNamed(propName); val != nil {
						return val.V
					}
				}
			}
		}
	}
	return nil
}

// GetWithDefault retrieves a property value, returning a default if not found.
func (sc *StyleChain) GetWithDefault(funcName string, propName string, defaultVal Value) Value {
	if val := sc.Get(funcName, propName); val != nil {
		return val
	}
	return defaultVal
}

// GetFloat retrieves a float property, returning the default if not found or wrong type.
func (sc *StyleChain) GetFloat(funcName string, propName string, defaultVal float64) float64 {
	val := sc.Get(funcName, propName)
	if val == nil {
		return defaultVal
	}
	switch v := val.(type) {
	case Int:
		return float64(v)
	case Float:
		return float64(v)
	case LengthValue:
		return v.Length.Points
	default:
		return defaultVal
	}
}

// GetInt retrieves an integer property, returning the default if not found or wrong type.
func (sc *StyleChain) GetInt(funcName string, propName string, defaultVal int64) int64 {
	val := sc.Get(funcName, propName)
	if val == nil {
		return defaultVal
	}
	if iv, ok := val.(Int); ok {
		return int64(iv)
	}
	return defaultVal
}

// GetStr retrieves a string property, returning the default if not found or wrong type.
func (sc *StyleChain) GetStr(funcName string, propName string, defaultVal string) string {
	val := sc.Get(funcName, propName)
	if val == nil {
		return defaultVal
	}
	if sv, ok := val.(Str); ok {
		return string(sv)
	}
	return defaultVal
}

// GetBool retrieves a boolean property, returning the default if not found or wrong type.
func (sc *StyleChain) GetBool(funcName string, propName string, defaultVal bool) bool {
	val := sc.Get(funcName, propName)
	if val == nil {
		return defaultVal
	}
	if bv, ok := val.(Bool); ok {
		return bool(bv)
	}
	return defaultVal
}

// Recipes returns all recipes from the entire chain, from outermost to innermost.
// This is used for show rule matching during realization.
func (sc *StyleChain) Recipes() []*Recipe {
	var result []*Recipe

	// Collect from outermost to innermost (parent first)
	var levels []*StyleChain
	for chain := sc; chain != nil; chain = chain.parent {
		levels = append(levels, chain)
	}

	// Reverse to get outermost first
	for i := len(levels) - 1; i >= 0; i-- {
		if levels[i].styles != nil {
			result = append(result, levels[i].styles.Recipes...)
		}
	}

	return result
}

// AllStyles returns a flattened Styles containing all rules from the chain.
// Rules are ordered from outermost to innermost.
func (sc *StyleChain) AllStyles() *Styles {
	if sc == nil {
		return nil
	}

	var allRules []StyleRule
	var allRecipes []*Recipe

	// Collect from outermost to innermost
	var levels []*StyleChain
	for chain := sc; chain != nil; chain = chain.parent {
		levels = append(levels, chain)
	}

	for i := len(levels) - 1; i >= 0; i-- {
		if levels[i].styles != nil {
			allRules = append(allRules, levels[i].styles.Rules...)
			allRecipes = append(allRecipes, levels[i].styles.Recipes...)
		}
	}

	if len(allRules) == 0 && len(allRecipes) == 0 {
		return nil
	}

	return &Styles{
		Rules:   allRules,
		Recipes: allRecipes,
	}
}

// ----------------------------------------------------------------------------
// Text Style Helpers
// ----------------------------------------------------------------------------

// TextSize returns the text size from the style chain, or the default 11pt.
func (sc *StyleChain) TextSize() float64 {
	return sc.GetFloat("text", "size", 11.0)
}

// TextFont returns the font family from the style chain, or empty string for default.
func (sc *StyleChain) TextFont() string {
	return sc.GetStr("text", "font", "")
}

// TextWeight returns the font weight from the style chain, or 400 (normal).
func (sc *StyleChain) TextWeight() int64 {
	return sc.GetInt("text", "weight", 400)
}

// TextStyle returns the font style from the style chain, or "normal".
func (sc *StyleChain) TextStyle() string {
	return sc.GetStr("text", "style", "normal")
}

// TextFill returns the text fill color from the style chain, or nil for default.
func (sc *StyleChain) TextFill() Color {
	val := sc.Get("text", "fill")
	if val == nil {
		return nil
	}
	if c, ok := val.(Color); ok {
		return c
	}
	return nil
}
