package eval

// ----------------------------------------------------------------------------
// Style Chain
// ----------------------------------------------------------------------------
//
// A StyleChain is a linked-list-like structure for tracking accumulated styles
// during realization and layout. It allows efficient style lookup without
// copying or merging style lists at each level.
//
// This matches Rust's StyleChain in typst-library/src/foundations/styles.rs.
//
// Example usage:
//   chain := NewStyleChain(baseStyles)
//   innerChain := chain.Chain(localStyles)
//   value := innerChain.Get("text", "size")
//
// The chain is traversed from innermost (most recent) to outermost (base),
// with the first matching property winning.

// StyleChain represents a chain of styles for property lookup.
// It's a linked list where each node contains a Styles reference
// and optionally points to a parent chain.
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
// This matches Rust's Chainable::chain method.
func (sc *StyleChain) Chain(inner *Styles) *StyleChain {
	if inner == nil || (len(inner.Rules) == 0 && len(inner.Recipes) == 0) {
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
	if sc.styles != nil && (len(sc.styles.Rules) > 0 || len(sc.styles.Recipes) > 0) {
		return false
	}
	return sc.parent.IsEmpty()
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
	case IntValue:
		return float64(v)
	case FloatValue:
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
	if iv, ok := AsInt(val); ok {
		return iv
	}
	return defaultVal
}

// GetStr retrieves a string property, returning the default if not found or wrong type.
func (sc *StyleChain) GetStr(funcName string, propName string, defaultVal string) string {
	val := sc.Get(funcName, propName)
	if val == nil {
		return defaultVal
	}
	if sv, ok := AsStr(val); ok {
		return sv
	}
	return defaultVal
}

// GetBool retrieves a boolean property, returning the default if not found or wrong type.
func (sc *StyleChain) GetBool(funcName string, propName string, defaultVal bool) bool {
	val := sc.Get(funcName, propName)
	if val == nil {
		return defaultVal
	}
	if bv, ok := AsBool(val); ok {
		return bv
	}
	return defaultVal
}

// Recipes returns all recipes from the entire chain, from outermost to innermost.
// This is used for show rule matching during realization.
func (sc *StyleChain) Recipes() []*Recipe {
	var result []*Recipe

	// Collect from outermost to innermost (parent first)
	// Build a list of chain levels first
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
// This is useful when you need to pass accumulated styles to a subsystem.
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
	return sc.GetFloat("text", "size", 11.0) // Default: 11pt
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
func (sc *StyleChain) TextFill() *Color {
	val := sc.Get("text", "fill")
	if val == nil {
		return nil
	}
	if cv, ok := val.(ColorValue); ok {
		return &cv.Color
	}
	return nil
}
