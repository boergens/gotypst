// Package realize implements the realization subsystem for GoTypst.
//
// Realization is the process of recursively applying styling and show rules
// to transform content into well-known elements suitable for layout and rendering.
//
// Pipeline position:
//
//	Source → Parse → Evaluate → REALIZE → Layout → Render
//
// The realize step sits between parsing/evaluation and layout. It doesn't produce
// visual output directly; instead, it transforms the content tree by applying
// rules and grouping related elements.
package realize

import (
	"github.com/boergens/gotypst/eval"
)

// RealizationKind specifies the context for realization.
// Different kinds affect how content is processed and grouped.
type RealizationKind interface {
	isRealizationKind()
}

// LayoutDocument prepares content for full document layout.
type LayoutDocument struct{}

func (LayoutDocument) isRealizationKind() {}

// LayoutFragment prepares content for fragment layout (block/inline detection).
type LayoutFragment struct {
	// Kind receives the detected fragment kind after realization.
	Kind *FragmentKind
}

func (LayoutFragment) isRealizationKind() {}

// LayoutPar prepares content for paragraph-specific realization.
type LayoutPar struct{}

func (LayoutPar) isRealizationKind() {}

// HtmlDocument prepares content for HTML document export.
type HtmlDocument struct{}

func (HtmlDocument) isRealizationKind() {}

// HtmlFragment prepares content for HTML fragment export.
type HtmlFragment struct{}

func (HtmlFragment) isRealizationKind() {}

// Math prepares content for mathematical typesetting.
type Math struct{}

func (Math) isRealizationKind() {}

// FragmentKind indicates the type of fragment detected during realization.
type FragmentKind int

const (
	// FragmentBlock indicates block-level content.
	FragmentBlock FragmentKind = iota
	// FragmentInline indicates inline-level content.
	FragmentInline
	// FragmentMixed indicates mixed block and inline content.
	FragmentMixed
)

// Pair represents a realized element with its associated style chain.
type Pair struct {
	// Element is the realized content element.
	Element eval.ContentElement
	// Styles contains the cascading styles for this element.
	Styles *StyleChain
}

// StyleChain represents a chain of cascading styles.
// Styles are inherited from parent to child, with child styles taking precedence.
type StyleChain struct {
	// Styles contains the current level's styles.
	Styles *eval.Styles
	// Parent is the parent style chain (or nil for root).
	Parent *StyleChain
}

// NewStyleChain creates a new style chain with the given styles and parent.
func NewStyleChain(styles *eval.Styles, parent *StyleChain) *StyleChain {
	return &StyleChain{
		Styles: styles,
		Parent: parent,
	}
}

// EmptyStyleChain returns an empty style chain.
func EmptyStyleChain() *StyleChain {
	return &StyleChain{Styles: &eval.Styles{}}
}

// Get retrieves a style value for the given function and property name.
// It searches up the chain until a value is found.
func (s *StyleChain) Get(funcName string, propName string) (eval.Value, bool) {
	if s == nil {
		return nil, false
	}

	// Check current styles
	if s.Styles != nil {
		for _, rule := range s.Styles.Rules {
			if rule.Func != nil && rule.Func.Name != nil && *rule.Func.Name == funcName {
				if rule.Args != nil {
					// Peek at args without removing (Args.Find removes the arg)
					if val := peekArg(rule.Args, propName); val != nil {
						return val, true
					}
				}
			}
		}
	}

	// Check parent chain
	return s.Parent.Get(funcName, propName)
}

// peekArg looks up a named argument without removing it from Args.
// Unlike Args.Find which consumes the argument, this is read-only.
func peekArg(args *eval.Args, name string) eval.Value {
	if args == nil {
		return nil
	}
	for _, arg := range args.Items {
		if arg.Name != nil && *arg.Name == name {
			return arg.Value.V
		}
	}
	return nil
}

// GetRecipes returns all show rule recipes from the chain.
func (s *StyleChain) GetRecipes() []*eval.Recipe {
	if s == nil {
		return nil
	}

	var recipes []*eval.Recipe

	// Collect from parent first (lower precedence)
	if s.Parent != nil {
		recipes = append(recipes, s.Parent.GetRecipes()...)
	}

	// Add current recipes (higher precedence)
	if s.Styles != nil {
		recipes = append(recipes, s.Styles.Recipes...)
	}

	return recipes
}

// Chain creates a new style chain with additional styles.
func (s *StyleChain) Chain(styles *eval.Styles) *StyleChain {
	if styles == nil || (len(styles.Rules) == 0 && len(styles.Recipes) == 0) {
		return s
	}
	return NewStyleChain(styles, s)
}

// GetWithDefault retrieves a style value with a default fallback.
// If the property is not found in the chain, the default value is returned.
func (s *StyleChain) GetWithDefault(funcName string, propName string, defaultVal eval.Value) eval.Value {
	if val, ok := s.Get(funcName, propName); ok {
		return val
	}
	return defaultVal
}

// GetProperty retrieves all style values for a function's property.
// Unlike Get which returns the first match, this returns all values in the chain
// from most recent (child) to oldest (parent), allowing for style merging.
func (s *StyleChain) GetProperty(funcName string, propName string) []eval.Value {
	if s == nil {
		return nil
	}

	var values []eval.Value

	// Check current styles first (higher precedence)
	if s.Styles != nil {
		for _, rule := range s.Styles.Rules {
			if rule.Func != nil && rule.Func.Name != nil && *rule.Func.Name == funcName {
				if rule.Args != nil {
					// Peek at args without removing (Args.Find removes the arg)
					if val := peekArg(rule.Args, propName); val != nil {
						values = append(values, val)
					}
				}
			}
		}
	}

	// Add parent values (lower precedence)
	if s.Parent != nil {
		values = append(values, s.Parent.GetProperty(funcName, propName)...)
	}

	return values
}

// GetRecipesFor returns recipes that could apply to the given element.
// This filters recipes by checking their selectors against the element type.
func (s *StyleChain) GetRecipesFor(elem eval.ContentElement) []*eval.Recipe {
	if s == nil || elem == nil {
		return nil
	}

	allRecipes := s.GetRecipes()
	if len(allRecipes) == 0 {
		return nil
	}

	var matching []*eval.Recipe
	for _, recipe := range allRecipes {
		if recipe.Selector == nil {
			// Recipes without selectors apply to everything
			matching = append(matching, recipe)
		} else if matchesSelector(elem, *recipe.Selector) {
			matching = append(matching, recipe)
		}
	}

	return matching
}

// IsEmpty returns true if this style chain has no styles or recipes.
func (s *StyleChain) IsEmpty() bool {
	if s == nil {
		return true
	}
	if s.Styles != nil && (len(s.Styles.Rules) > 0 || len(s.Styles.Recipes) > 0) {
		return false
	}
	if s.Parent != nil {
		return s.Parent.IsEmpty()
	}
	return true
}

// Depth returns the number of levels in the style chain.
func (s *StyleChain) Depth() int {
	if s == nil {
		return 0
	}
	return 1 + s.Parent.Depth()
}

// GetAllRules returns all style rules for a function from the entire chain.
// Rules are returned from most recent (child) to oldest (parent).
func (s *StyleChain) GetAllRules(funcName string) []eval.StyleRule {
	if s == nil {
		return nil
	}

	var rules []eval.StyleRule

	// Current rules first (higher precedence)
	if s.Styles != nil {
		for _, rule := range s.Styles.Rules {
			if rule.Func != nil && rule.Func.Name != nil && *rule.Func.Name == funcName {
				rules = append(rules, rule)
			}
		}
	}

	// Parent rules (lower precedence)
	if s.Parent != nil {
		rules = append(rules, s.Parent.GetAllRules(funcName)...)
	}

	return rules
}

// Fold applies a function to each style value in the chain, accumulating a result.
// This is useful for merging style values where later values should override earlier ones.
// The function is called with (accumulated value, new value) and should return the merged result.
func (s *StyleChain) Fold(funcName string, propName string, initial eval.Value, fold func(acc, val eval.Value) eval.Value) eval.Value {
	values := s.GetProperty(funcName, propName)
	if len(values) == 0 {
		return initial
	}

	// Apply values from oldest to newest (reverse order)
	result := initial
	for i := len(values) - 1; i >= 0; i-- {
		result = fold(result, values[i])
	}
	return result
}

// State maintains mutable context during realization.
type State struct {
	// Engine provides access to the evaluation/layout engine.
	Engine *eval.Engine

	// Output collects realized pairs.
	Output []Pair

	// Kind specifies the realization context.
	Kind RealizationKind

	// groupings tracks active element groupings (paragraphs, lists, etc.)
	groupings []Grouping

	// config holds configuration flags
	config Config
}

// Config holds realization configuration.
type Config struct {
	// CollapseSpaces enables space collapsing.
	CollapseSpaces bool
	// ProcessShowRules enables show rule processing.
	ProcessShowRules bool
}

// DefaultConfig returns the default realization configuration.
func DefaultConfig() Config {
	return Config{
		CollapseSpaces:   true,
		ProcessShowRules: true,
	}
}

// Grouping defines how related elements are collected for unified processing.
type Grouping interface {
	isGrouping()
	// Trigger returns true if this element triggers the grouping.
	Trigger(elem eval.ContentElement) bool
	// Inner returns true if this element belongs inside the group.
	Inner(elem eval.ContentElement) bool
	// Interrupt returns true if this element interrupts/ends the group.
	Interrupt(elem eval.ContentElement) bool
	// Finalize processes the collected elements and returns the grouped result.
	Finalize(elements []eval.ContentElement) eval.ContentElement
}

// ParagraphGrouping collects inline content into paragraph elements.
type ParagraphGrouping struct {
	elements []eval.ContentElement
}

func (ParagraphGrouping) isGrouping() {}

func (g *ParagraphGrouping) Trigger(elem eval.ContentElement) bool {
	return isInlineElement(elem)
}

func (g *ParagraphGrouping) Inner(elem eval.ContentElement) bool {
	return isInlineElement(elem)
}

func (g *ParagraphGrouping) Interrupt(elem eval.ContentElement) bool {
	return isBlockElement(elem) || isParbreak(elem)
}

func (g *ParagraphGrouping) Finalize(elements []eval.ContentElement) eval.ContentElement {
	if len(elements) == 0 {
		return nil
	}
	return &eval.ParagraphElement{
		Body: eval.Content{Elements: elements},
	}
}

// isBlockElement returns true if the element is block-level.
func isBlockElement(elem eval.ContentElement) bool {
	switch elem.(type) {
	case *eval.ParagraphElement, *eval.HeadingElement,
		*eval.ListItemElement, *eval.EnumItemElement, *eval.TermItemElement:
		return true
	default:
		return false
	}
}

// isParbreak returns true if the element is a paragraph break.
func isParbreak(elem eval.ContentElement) bool {
	_, ok := elem.(*eval.ParbreakElement)
	return ok
}

// Realize transforms content with styles into realized pairs.
//
// This is the core realization function that:
//  1. Applies show rules (user-defined and built-in)
//  2. Groups related elements (paragraphs, lists, citations)
//  3. Collapses spaces according to typesetting rules
//  4. Supports multiple output contexts (document, fragment, HTML, math)
//
// Parameters:
//   - kind: Specifies the realization context
//   - engine: Layout/evaluation engine state
//   - content: Input content tree
//   - styles: Cascading style information
//
// Returns the realized pairs ready for layout.
func Realize(
	kind RealizationKind,
	engine *eval.Engine,
	content *eval.Content,
	styles *StyleChain,
) ([]Pair, error) {
	if content == nil {
		return nil, nil
	}

	state := &State{
		Engine: engine,
		Kind:   kind,
		config: DefaultConfig(),
	}

	// Process each element in the content
	for _, elem := range content.Elements {
		if err := state.realizeElement(elem, styles); err != nil {
			return nil, err
		}
	}

	// Finalize any pending groupings
	state.finalizeGroupings()

	// Collapse spaces if enabled
	if state.config.CollapseSpaces {
		state.Output = collapseSpaces(state.Output)
	}

	// Detect fragment kind for LayoutFragment
	if frag, ok := kind.(*LayoutFragment); ok && frag.Kind != nil {
		*frag.Kind = detectFragmentKind(state.Output)
	}

	return state.Output, nil
}

// realizeElement processes a single content element.
func (s *State) realizeElement(elem eval.ContentElement, styles *StyleChain) error {
	if elem == nil {
		return nil
	}

	// Check for show rule match
	if s.config.ProcessShowRules {
		transformed, matched, err := s.applyShowRules(elem, styles)
		if err != nil {
			return err
		}
		if matched {
			// Recursively realize the transformed content
			for _, t := range transformed {
				if err := s.realizeElement(t, styles); err != nil {
					return err
				}
			}
			return nil
		}
	}

	// Handle grouping
	if s.handleGrouping(elem, styles) {
		return nil
	}

	// Add to output
	s.Output = append(s.Output, Pair{
		Element: elem,
		Styles:  styles,
	})

	return nil
}

// applyShowRules checks if any show rule matches and returns transformed content.
func (s *State) applyShowRules(elem eval.ContentElement, styles *StyleChain) ([]eval.ContentElement, bool, error) {
	recipes := styles.GetRecipes()
	if len(recipes) == 0 {
		return nil, false, nil
	}

	for _, recipe := range recipes {
		if recipe.Selector == nil {
			continue
		}

		if matchesSelector(elem, *recipe.Selector) {
			transformed, err := applyTransformation(s.Engine, elem, recipe.Transform, styles)
			if err != nil {
				return nil, false, err
			}
			return transformed, true, nil
		}
	}

	return nil, false, nil
}

// matchesSelector checks if an element matches a selector.
func matchesSelector(elem eval.ContentElement, selector eval.Selector) bool {
	switch sel := selector.(type) {
	case eval.ElemSelector:
		return matchesElementSelector(elem, sel)
	case eval.LabelSelector:
		return matchesLabelSelector(elem, sel)
	case eval.TextSelector:
		return matchesTextSelector(elem, sel)
	case eval.OrSelector:
		for _, s := range sel.Selectors {
			if matchesSelector(elem, s) {
				return true
			}
		}
		return false
	case eval.AndSelector:
		for _, s := range sel.Selectors {
			if !matchesSelector(elem, s) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// matchesElementSelector checks if an element matches an element type selector.
func matchesElementSelector(elem eval.ContentElement, sel eval.ElemSelector) bool {
	elemName := getElementName(elem)
	return elemName == sel.Element.Name
}

// matchesLabelSelector checks if an element has a matching label.
func matchesLabelSelector(elem eval.ContentElement, sel eval.LabelSelector) bool {
	// Labels are typically attached to elements via metadata
	// For now, return false as label handling needs more infrastructure
	return false
}

// matchesTextSelector checks if a text element matches a text pattern.
func matchesTextSelector(elem eval.ContentElement, sel eval.TextSelector) bool {
	text, ok := elem.(*eval.TextElement)
	if !ok {
		return false
	}

	if sel.IsRegex {
		// TODO: Implement regex matching
		return false
	}

	return text.Text == sel.Text
}

// getElementName returns the element type name.
func getElementName(elem eval.ContentElement) string {
	switch elem.(type) {
	case *eval.TextElement:
		return "text"
	case *eval.ParagraphElement:
		return "par"
	case *eval.StrongElement:
		return "strong"
	case *eval.EmphElement:
		return "emph"
	case *eval.RawElement:
		return "raw"
	case *eval.HeadingElement:
		return "heading"
	case *eval.ListItemElement:
		return "list.item"
	case *eval.EnumItemElement:
		return "enum.item"
	case *eval.TermItemElement:
		return "terms.item"
	case *eval.LinkElement:
		return "link"
	case *eval.RefElement:
		return "ref"
	case *eval.LinebreakElement:
		return "linebreak"
	case *eval.ParbreakElement:
		return "parbreak"
	case *eval.SmartQuoteElement:
		return "smartquote"
	case *eval.EquationElement:
		return "equation"
	case *eval.ImageElement:
		return "image"
	default:
		return ""
	}
}

// applyTransformation applies a show rule transformation to an element.
func applyTransformation(
	engine *eval.Engine,
	elem eval.ContentElement,
	transform eval.Transformation,
	styles *StyleChain,
) ([]eval.ContentElement, error) {
	switch t := transform.(type) {
	case eval.NoneTransformation:
		// Hide the element
		return nil, nil

	case eval.ContentTransformation:
		// Replace with content
		return t.Content.Elements, nil

	case eval.StyleTransformation:
		// Apply styles (element remains, styles are modified)
		// Return the original element - styles will be applied via chain
		return []eval.ContentElement{elem}, nil

	case eval.FuncTransformation:
		// Apply function transformation
		// This requires calling the function with the element as argument
		// For now, return the original element as placeholder
		// TODO: Implement function transformation
		return []eval.ContentElement{elem}, nil

	default:
		return []eval.ContentElement{elem}, nil
	}
}

// handleGrouping checks if element should be grouped and handles accordingly.
func (s *State) handleGrouping(elem eval.ContentElement, styles *StyleChain) bool {
	// Check active groupings first
	for i := len(s.groupings) - 1; i >= 0; i-- {
		g := s.groupings[i]
		if g.Interrupt(elem) {
			// Finalize the grouping
			s.finalizeGrouping(i)
			return false
		}
		if g.Inner(elem) {
			// Add to grouping
			if pg, ok := g.(*ParagraphGrouping); ok {
				pg.elements = append(pg.elements, elem)
			}
			return true
		}
	}

	// Check if this element triggers a new grouping
	// For LayoutDocument/LayoutFragment, collect inline content into paragraphs
	switch s.Kind.(type) {
	case LayoutDocument, *LayoutFragment:
		if isInlineElement(elem) {
			pg := &ParagraphGrouping{
				elements: []eval.ContentElement{elem},
			}
			s.groupings = append(s.groupings, pg)
			return true
		}
	}

	return false
}

// finalizeGrouping finalizes the grouping at the given index.
func (s *State) finalizeGrouping(index int) {
	if index < 0 || index >= len(s.groupings) {
		return
	}

	g := s.groupings[index]

	// Get elements from grouping
	var elements []eval.ContentElement
	if pg, ok := g.(*ParagraphGrouping); ok {
		elements = pg.elements
	}

	// Finalize and add to output
	if result := g.Finalize(elements); result != nil {
		s.Output = append(s.Output, Pair{
			Element: result,
			Styles:  EmptyStyleChain(), // TODO: Preserve styles
		})
	}

	// Remove the grouping
	s.groupings = append(s.groupings[:index], s.groupings[index+1:]...)
}

// finalizeGroupings finalizes all pending groupings.
func (s *State) finalizeGroupings() {
	for len(s.groupings) > 0 {
		s.finalizeGrouping(len(s.groupings) - 1)
	}
}

// detectFragmentKind determines the kind of content in the output.
func detectFragmentKind(pairs []Pair) FragmentKind {
	hasBlock := false
	hasInline := false

	for _, p := range pairs {
		if isBlockElement(p.Element) {
			hasBlock = true
		} else if isInlineElement(p.Element) {
			hasInline = true
		}
	}

	if hasBlock && hasInline {
		return FragmentMixed
	} else if hasBlock {
		return FragmentBlock
	}
	return FragmentInline
}
