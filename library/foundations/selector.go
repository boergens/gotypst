// Selector types for Typst.
// Translated from typst-library/src/foundations/selector.rs

package foundations

import (
	"github.com/boergens/gotypst/syntax"
)

// Selector represents a filter for selecting elements within the document.
// This is used for show rules and queries.
//
// Matches Rust's Selector enum in selector.rs.
type Selector interface {
	isSelector()
}

// ElemSelector matches content of a specific element type.
// Corresponds to Rust's Selector::Elem variant.
type ElemSelector struct {
	// Element is the element type to match.
	Element Element

	// Where is an optional filter function for field matching.
	Where *Func
}

func (ElemSelector) isSelector() {}

// LabelSelector matches content with a specific label.
// Corresponds to Rust's Selector::Label variant.
type LabelSelector struct {
	Label string
}

func (LabelSelector) isSelector() {}

// RegexSelector matches text content using a regular expression.
// Corresponds to Rust's Selector::Regex variant.
type RegexSelector struct {
	// Pattern is the regex pattern to match.
	Pattern string
}

func (RegexSelector) isSelector() {}

// LocationSelector matches the element at a specific location.
// Corresponds to Rust's Selector::Location variant.
type LocationSelector struct {
	Location *Location
}

func (LocationSelector) isSelector() {}

// OrSelector matches content that matches any of the selectors.
// Corresponds to Rust's Selector::Or variant.
type OrSelector struct {
	Selectors []Selector
}

func (OrSelector) isSelector() {}

// AndSelector matches content that matches all of the selectors.
// Corresponds to Rust's Selector::And variant.
type AndSelector struct {
	Selectors []Selector
}

func (AndSelector) isSelector() {}

// BeforeSelector matches all matches of selector before end.
// Corresponds to Rust's Selector::Before variant.
type BeforeSelector struct {
	Selector  Selector
	End       Selector
	Inclusive bool
}

func (BeforeSelector) isSelector() {}

// AfterSelector matches all matches of selector after start.
// Corresponds to Rust's Selector::After variant.
type AfterSelector struct {
	Selector  Selector
	Start     Selector
	Inclusive bool
}

func (AfterSelector) isSelector() {}

// LocatableSelector is a selector that can be used with query.
// Corresponds to Rust's LocatableSelector struct.
type LocatableSelector struct {
	Selector Selector
}

// ShowableSelector is a selector that can be used with show rules.
// Corresponds to Rust's ShowableSelector struct.
type ShowableSelector struct {
	Selector Selector
}

// ----------------------------------------------------------------------------
// Transformation Types
// ----------------------------------------------------------------------------

// Transformation represents how content should be transformed in a show rule.
// Corresponds to Rust's Transformation enum in styles.rs.
type Transformation interface {
	isTransformation()
}

// StyleTransformation applies styles to content.
// Corresponds to Rust's Transformation::Style variant.
type StyleTransformation struct {
	Styles *Styles
}

func (StyleTransformation) isTransformation() {}

// FuncTransformation applies a function to transform content.
// Corresponds to Rust's Transformation::Func variant.
type FuncTransformation struct {
	Func *Func
}

func (FuncTransformation) isTransformation() {}

// ContentTransformation replaces content directly.
// Corresponds to Rust's Transformation::Content variant.
type ContentTransformation struct {
	Content Content
}

func (ContentTransformation) isTransformation() {}

// NoneTransformation hides the matched content.
// This is a Go extension for explicit none handling.
type NoneTransformation struct{}

func (NoneTransformation) isTransformation() {}

// ----------------------------------------------------------------------------
// Recipe (Show Rule)
// ----------------------------------------------------------------------------

// Recipe represents a show rule recipe that defines how to transform content.
// Corresponds to Rust's Recipe struct in styles.rs.
type Recipe struct {
	// Selector optionally restricts which content the recipe applies to.
	// If nil, this is an eager show rule that applies immediately.
	Selector Selector

	// Transform defines how to transform matching content.
	Transform Transformation

	// Span is the source location of the recipe.
	Span syntax.Span

	// Outside indicates if the recipe applies from outside the element.
	Outside bool
}

// NewRecipe creates a new recipe with the given components.
func NewRecipe(selector Selector, transform Transformation, span syntax.Span) *Recipe {
	return &Recipe{
		Selector:  selector,
		Transform: transform,
		Span:      span,
		Outside:   false,
	}
}

// RecipeIndex identifies a show rule recipe from the top of the chain.
// Corresponds to Rust's RecipeIndex struct.
type RecipeIndex struct {
	Index int
}
