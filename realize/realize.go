// Package realize implements the realization subsystem for GoTypst.
//
// Realization is the process of recursively applying styling and, in particular,
// show rules to produce well-known elements that can be processed further.
//
// This is a faithful port of Rust's typst-realize crate.
// Reference: typst-reference/crates/typst-realize/src/lib.rs
package realize

import (
	"errors"
	"regexp"

	"github.com/boergens/gotypst/eval"
	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Realization Kind
// ----------------------------------------------------------------------------

// RealizationKind specifies the context for realization.
// Different kinds affect how content is processed and grouped.
// Matches Rust: typst-library/src/routines/realize.rs RealizationKind
type RealizationKind interface {
	isRealizationKind()
	// IsDocument returns true if this is a document realization.
	IsDocument() bool
	// IsFragment returns true if this is a fragment realization.
	IsFragment() bool
}

// LayoutDocument prepares content for full document layout.
// Matches Rust: RealizationKind::LayoutDocument
type LayoutDocument struct {
	// Info receives document metadata populated during realization.
	Info *DocumentInfo
}

func (LayoutDocument) isRealizationKind() {}
func (LayoutDocument) IsDocument() bool   { return true }
func (LayoutDocument) IsFragment() bool   { return false }

// LayoutFragment prepares content for fragment layout.
// Matches Rust: RealizationKind::LayoutFragment
type LayoutFragment struct {
	// Kind receives the detected fragment kind after realization.
	Kind *FragmentKind
}

func (LayoutFragment) isRealizationKind() {}
func (LayoutFragment) IsDocument() bool   { return false }
func (LayoutFragment) IsFragment() bool   { return true }

// LayoutPar prepares content for paragraph-specific realization.
// Matches Rust: RealizationKind::LayoutPar
type LayoutPar struct{}

func (LayoutPar) isRealizationKind() {}
func (LayoutPar) IsDocument() bool   { return false }
func (LayoutPar) IsFragment() bool   { return false }

// HtmlDocument prepares content for HTML document export.
// Matches Rust: RealizationKind::HtmlDocument
type HtmlDocument struct {
	Info       *DocumentInfo
	IsPhrasing func(eval.ContentElement) bool
}

func (HtmlDocument) isRealizationKind() {}
func (HtmlDocument) IsDocument() bool   { return true }
func (HtmlDocument) IsFragment() bool   { return false }

// HtmlFragment prepares content for HTML fragment export.
// Matches Rust: RealizationKind::HtmlFragment
type HtmlFragment struct {
	Kind       *FragmentKind
	IsPhrasing func(eval.ContentElement) bool
}

func (HtmlFragment) isRealizationKind() {}
func (HtmlFragment) IsDocument() bool   { return false }
func (HtmlFragment) IsFragment() bool   { return true }

// Math prepares content for mathematical typesetting.
// Matches Rust: RealizationKind::Math
type Math struct{}

func (Math) isRealizationKind() {}
func (Math) IsDocument() bool   { return false }
func (Math) IsFragment() bool   { return false }

// DocumentInfo holds document metadata populated during realization.
type DocumentInfo struct {
	Title    *eval.Content
	Author   []string
	Keywords []string
	Date     eval.Value
	Locale   string
}

// FragmentKind indicates the type of fragment detected during realization.
// Matches Rust: typst-library/src/routines/realize.rs FragmentKind
type FragmentKind int

const (
	// FragmentBlock indicates block-level content.
	FragmentBlock FragmentKind = iota
	// FragmentInline indicates inline-level content.
	FragmentInline
)

// ----------------------------------------------------------------------------
// Pair
// ----------------------------------------------------------------------------

// Pair represents a realized element with its associated style chain.
// Matches Rust: typst-library/src/routines/realize.rs Pair
type Pair struct {
	Content eval.ContentElement
	Styles  *eval.StyleChain
}

// ----------------------------------------------------------------------------
// State
// ----------------------------------------------------------------------------

// state maintains mutable context during realization.
// Matches Rust: State struct in typst-realize/src/lib.rs
type state struct {
	// kind specifies the realization context.
	kind RealizationKind
	// engine provides access to the evaluation engine.
	engine *eval.Engine
	// sink collects output pairs.
	sink []Pair
	// rules are the grouping rules for this realization kind.
	rules []*GroupingRule
	// groupings tracks active groupings.
	groupings []grouping
	// outside indicates we're not within any container or show rule output.
	// This is used to determine page styles during layout.
	outside bool
	// mayAttach indicates if following attach spacing can survive.
	mayAttach bool
	// sawParbreak indicates if we visited any paragraph breaks.
	sawParbreak bool
	// locationCounter generates unique locations for elements.
	locationCounter uint64
}

// grouping tracks an active grouping operation.
// Matches Rust: Grouping struct
type grouping struct {
	// start is the position in sink where the group starts.
	start int
	// interrupted indicates the group is interrupted but not finished
	// (for PAR grouping that may be ignored if fully inline).
	interrupted bool
	// rule is the grouping rule being applied.
	rule *GroupingRule
}

// grouped provides context for grouping finishers.
// Matches Rust: Grouped struct
type grouped struct {
	s     *state
	start int
}

// get returns the grouped elements.
func (g *grouped) get() []Pair {
	return g.s.sink[g.start:]
}

// end removes grouped elements and returns state for further visiting.
func (g *grouped) end() *state {
	g.s.sink = g.s.sink[:g.start]
	return g.s
}

// ----------------------------------------------------------------------------
// Verdict and ShowStep
// ----------------------------------------------------------------------------

// verdict determines how to proceed with an element.
// Matches Rust: Verdict struct
type verdict struct {
	// prepared indicates element has already been prepared.
	prepared bool
	// styles are the styles to apply.
	styles *eval.Styles
	// step is the show rule step to apply (if any).
	step *showStep
}

// showStep represents a show rule transformation to apply.
// Matches Rust: ShowStep enum
type showStep struct {
	// recipe is a user-defined show rule (if not builtin).
	recipe *eval.Recipe
	// recipeIndex is the recipe's index for guarding.
	recipeIndex int
	// builtin is the built-in show rule (if not recipe).
	builtin bool
}

// ----------------------------------------------------------------------------
// Main Entry Point
// ----------------------------------------------------------------------------

// Realize transforms content into a flat list of well-known, styled items.
// This is the main entry point for realization.
//
// Matches Rust: pub fn realize() in typst-realize/src/lib.rs
func Realize(
	kind RealizationKind,
	engine *eval.Engine,
	content eval.ContentElement,
	styles *eval.StyleChain,
) ([]Pair, error) {
	if content == nil {
		return nil, nil
	}

	s := &state{
		kind:      kind,
		engine:    engine,
		sink:      make([]Pair, 0),
		rules:     getRulesForKind(kind),
		groupings: make([]grouping, 0, maxGroupNesting),
		outside:   kind.IsDocument(),
		mayAttach: false,
	}

	if err := visit(s, content, styles); err != nil {
		return nil, err
	}

	if err := finish(s); err != nil {
		return nil, err
	}

	return s.sink, nil
}

// getRulesForKind returns the grouping rules for a realization kind.
// Matches Rust: LAYOUT_RULES, LAYOUT_PAR_RULES, HTML_*_RULES, MATH_RULES
func getRulesForKind(kind RealizationKind) []*GroupingRule {
	switch kind.(type) {
	case LayoutDocument, *LayoutDocument:
		return layoutRules
	case LayoutFragment, *LayoutFragment:
		return layoutRules
	case LayoutPar, *LayoutPar:
		return layoutParRules
	case HtmlDocument, *HtmlDocument:
		return htmlDocumentRules
	case HtmlFragment, *HtmlFragment:
		return htmlFragmentRules
	case Math, *Math:
		return mathRules
	default:
		return layoutRules
	}
}

// ----------------------------------------------------------------------------
// Visit Functions
// ----------------------------------------------------------------------------

// visit handles an arbitrary piece of content during realization.
// Matches Rust: fn visit()
func visit(s *state, content eval.ContentElement, styles *eval.StyleChain) error {
	if content == nil {
		return nil
	}

	// Tags can always simply be pushed.
	if _, ok := content.(*eval.TagElem); ok {
		s.sink = append(s.sink, Pair{Content: content, Styles: styles})
		return nil
	}

	// Transformations for content based on the realization kind.
	// Needs to happen before show rules.
	if handled, err := visitKindRules(s, content, styles); err != nil {
		return err
	} else if handled {
		return nil
	}

	// Apply show rules and preparation.
	if handled, err := visitShowRules(s, content, styles); err != nil {
		return err
	} else if handled {
		return nil
	}

	// Recurse into sequences.
	if seq, ok := content.(*eval.SequenceElem); ok {
		for _, elem := range seq.Children {
			if err := visit(s, elem, styles); err != nil {
				return err
			}
		}
		return nil
	}

	// Recurse into styled elements.
	if styled, ok := content.(*eval.StyledElement); ok {
		return visitStyled(s, styled.Child, styled.Styles, styles)
	}

	// Apply grouping rules.
	if handled, err := visitGroupingRules(s, content, styles); err != nil {
		return err
	} else if handled {
		return nil
	}

	// Apply filter rules.
	if handled, err := visitFilterRules(s, content, styles); err != nil {
		return err
	} else if handled {
		return nil
	}

	// No further transformations, push to output.
	s.sink = append(s.sink, Pair{Content: content, Styles: styles})
	return nil
}

// visitKindRules handles transformations based on the realization kind.
// Matches Rust: fn visit_kind_rules()
func visitKindRules(s *state, content eval.ContentElement, styles *eval.StyleChain) (bool, error) {
	switch s.kind.(type) {
	case Math, *Math:
		// In math, transparently recurse into nested equations.
		if eq, ok := content.(*eval.EquationElement); ok {
			return true, visitContent(s, eq.Body, styles)
		}

		// In math, apply regex show rules to symbols per-element.
		// (Simplified - full implementation would handle regex matching)
		return false, nil

	default:
		// Transparently wrap mathy content into equations.
		if isMathy(content) && !isEquation(content) {
			eq := &eval.EquationElement{Body: eval.Content{Elements: []eval.ContentElement{content}}}
			return true, visit(s, eq, styles)
		}

		// Symbols in non-math convert to TextElem.
		if sym, ok := content.(*eval.SymbolElem); ok {
			text := &eval.TextElement{Text: sym.Text}
			return true, visit(s, text, styles)
		}
	}

	return false, nil
}

// visitShowRules tries to apply show rules or prepare content.
// Matches Rust: fn visit_show_rules()
func visitShowRules(s *state, content eval.ContentElement, styles *eval.StyleChain) (bool, error) {
	v := getVerdict(s.engine, content, styles)
	if v == nil {
		return false, nil
	}

	// Create local styles if we have any.
	localStyles := styles
	if v.styles != nil && len(v.styles.Rules) > 0 {
		localStyles = styles.Chain(v.styles)
	}

	// Prepare the element and get tags for introspection.
	// Tags are only created for elements that need them (locatable/tagged).
	startTag, endTag := prepare(s, content, v.styles, localStyles)

	// Push start tag.
	if startTag != nil {
		if err := visit(s, startTag, localStyles); err != nil {
			return false, err
		}
	}

	// Remember state for restoring after show rule application.
	prevOutside := s.outside
	s.outside = s.outside && isContextElem(content)

	// Apply show rule step if there is one.
	var visitErr error
	handled := false
	if v.step != nil {
		if v.step.builtin {
			// Apply built-in show rule.
			output, err := applyBuiltinShowRule(s.engine, content, localStyles)
			if err != nil {
				visitErr = err
			} else if output != nil {
				visitErr = visitContent(s, *output, localStyles)
				handled = true
			}
		} else if v.step.recipe != nil {
			// Apply user-defined show rule.
			output, err := applyRecipe(s.engine, content, v.step.recipe, localStyles)
			if err != nil {
				visitErr = err
			} else if output != nil {
				visitErr = visitContent(s, *output, localStyles)
				handled = true
			}
		}
	}

	// Restore outside state.
	s.outside = prevOutside

	// Push end tag.
	if endTag != nil {
		if err := visit(s, endTag, localStyles); err != nil {
			return handled, err
		}
	}

	if visitErr != nil {
		return handled, visitErr
	}

	return handled, nil
}

// isContextElem returns true if the element is a ContextElem.
func isContextElem(elem eval.ContentElement) bool {
	_, ok := elem.(*eval.ContextElem)
	return ok
}

// getVerdict determines whether and how to proceed with show rule application.
// Matches Rust: fn verdict()
func getVerdict(engine *eval.Engine, elem eval.ContentElement, styles *eval.StyleChain) *verdict {
	// Get recipes from style chain.
	recipes := styles.Recipes()
	if len(recipes) == 0 {
		return nil
	}

	var matchedRecipe *eval.Recipe
	var matchedIndex int
	var showSetStyles *eval.Styles

	// Check each recipe for a match.
	for i, recipe := range recipes {
		if recipe.Selector == nil {
			continue
		}

		if !matchesSelector(elem, *recipe.Selector, styles) {
			continue
		}

		// Handle show-set rules (StyleTransformation).
		if styleTrans, ok := recipe.Transform.(eval.StyleTransformation); ok {
			if showSetStyles == nil {
				showSetStyles = &eval.Styles{}
			}
			showSetStyles.Rules = append(showSetStyles.Rules, styleTrans.Styles.Rules...)
			continue
		}

		// Found a show rule transformation.
		if matchedRecipe == nil {
			matchedRecipe = recipe
			matchedIndex = i
		}
	}

	// If nothing to do, return nil.
	if matchedRecipe == nil && showSetStyles == nil {
		return nil
	}

	v := &verdict{
		prepared: false,
		styles:   showSetStyles,
	}

	if matchedRecipe != nil {
		v.step = &showStep{
			recipe:      matchedRecipe,
			recipeIndex: matchedIndex,
		}
	}

	return v
}

// prepare prepares an element for realization.
// This assigns a location, applies show-set rules, synthesizes fields,
// and returns tags for introspection.
// Matches Rust: fn prepare()
func prepare(s *state, elem eval.ContentElement, showSetStyles *eval.Styles, styles *eval.StyleChain) (startTag, endTag *eval.TagElem) {
	// Determine if this element needs introspection tags.
	// Elements are introspectable if they're "locatable" (like headings, figures)
	// or have labels.
	flags := eval.TagFlags{
		Introspectable: isLocatable(elem) || hasLabel(elem),
		Tagged:         isTagged(elem),
	}

	// Only create tags if any flag is set.
	if !flags.Any() {
		return nil, nil
	}

	// Generate a unique location for this element.
	s.locationCounter++
	loc := eval.TagLocation{
		Hash:    s.locationCounter,
		Variant: 0,
	}

	// Create start and end tags.
	key := s.locationCounter // Use same value for simplicity
	startTag = eval.NewStartTag(elem, loc, flags)
	endTag = eval.NewEndTag(loc, key, flags)

	return startTag, endTag
}

// isLocatable returns true if an element can be located via introspection.
// These elements get start/end tags during realization.
func isLocatable(elem eval.ContentElement) bool {
	switch elem.(type) {
	case *eval.HeadingElement:
		return true
	case *eval.ImageElement:
		return true
	case *eval.EquationElement:
		return true
	case *eval.CiteElement:
		return true
	case *eval.RefElement:
		return true
	case *eval.LinkElement:
		return true
	case *eval.StrongElement:
		return true
	case *eval.EmphElement:
		return true
	default:
		return false
	}
}

// hasLabel returns true if an element has a label attached.
// Labels are used for cross-references in Typst.
// Matches Rust: content.label().is_some()
func hasLabel(elem eval.ContentElement) bool {
	// Check element types that can have labels.
	// Currently only SymbolElem has a Label field in our implementation.
	// More elements will get label support as they are implemented.
	if sym, ok := elem.(*eval.SymbolElem); ok {
		return sym.Label != nil && *sym.Label != ""
	}
	// TODO: Add more element types as they gain label support
	// (e.g., HeadingElement, ImageElement, EquationElement, etc.)
	return false
}

// isTagged returns true if an element is semantically tagged (for accessibility).
func isTagged(elem eval.ContentElement) bool {
	switch elem.(type) {
	case *eval.HeadingElement:
		return true
	case *eval.ParagraphElement:
		return true
	case *eval.ListElement:
		return true
	case *eval.EnumElement:
		return true
	case *eval.TermsElement:
		return true
	case *eval.ImageElement:
		return true
	default:
		return false
	}
}

// visitStyled handles a styled element.
// Matches Rust: fn visit_styled()
func visitStyled(s *state, content eval.Content, local *eval.Styles, outer *eval.StyleChain) error {
	// Nothing to do if styles are empty.
	if local == nil || (len(local.Rules) == 0 && len(local.Recipes) == 0) {
		return visitContent(s, content, outer)
	}

	// Check for page styles.
	pagebreak := false
	for _, rule := range local.Rules {
		if rule.Func != nil && rule.Func.Name != nil {
			name := *rule.Func.Name
			if name == "document" {
				// Document set rules populate document info.
				if info := getDocumentInfo(s.kind); info != nil {
					populateDocumentInfo(info, local)
				} else {
					return errors.New("document set rules are not allowed inside of containers")
				}
			} else if name == "page" {
				switch s.kind.(type) {
				case LayoutDocument, *LayoutDocument:
					pagebreak = true
					s.outside = true
				case HtmlDocument, *HtmlDocument:
					// Warn: page set rule ignored in HTML
				default:
					return errors.New("page configuration is not allowed inside of containers")
				}
			}
		}
	}

	// Chain the styles.
	chained := outer.Chain(local)

	// Generate weak pagebreak if there is a page interruption.
	if pagebreak {
		pb := &eval.PagebreakElem{Weak: true}
		if err := visit(s, pb, chained); err != nil {
			return err
		}
	}

	// Finish any interrupted groupings.
	if err := finishInterrupted(s, local); err != nil {
		return err
	}

	// Visit the content.
	if err := visitContent(s, content, chained); err != nil {
		return err
	}

	// Finish interrupted again after content.
	if err := finishInterrupted(s, local); err != nil {
		return err
	}

	// Generate boundary pagebreak at end.
	if pagebreak {
		pb := &eval.PagebreakElem{Weak: true}
		if err := visit(s, pb, outer); err != nil {
			return err
		}
	}

	return nil
}

// visitGroupingRules tries to group content or start a new group.
// Matches Rust: fn visit_grouping_rules()
func visitGroupingRules(s *state, content eval.ContentElement, styles *eval.StyleChain) (bool, error) {
	// Find a matching rule.
	var matching *GroupingRule
	for _, rule := range s.rules {
		if rule.Trigger(content, s) {
			matching = rule
			break
		}
	}

	// Try to continue or finish existing groupings.
	iterations := 0
	for len(s.groupings) > 0 {
		active := s.groupings[len(s.groupings)-1]

		// Start nested group if a rule with higher priority matches.
		if matching != nil && matching.Priority > active.rule.Priority {
			break
		}

		// If element can be added to active grouping, do it.
		if !active.interrupted && (active.rule.Trigger(content, s) || active.rule.Inner(content)) {
			s.sink = append(s.sink, Pair{Content: content, Styles: styles})
			return true, nil
		}

		// Finish the innermost grouping.
		if err := finishInnermostGrouping(s); err != nil {
			return false, err
		}

		iterations++
		if iterations > 512 {
			return false, errors.New("maximum grouping depth exceeded")
		}
	}

	// Start a new grouping.
	if matching != nil {
		s.groupings = append(s.groupings, grouping{
			start:       len(s.sink),
			interrupted: false,
			rule:        matching,
		})
		s.sink = append(s.sink, Pair{Content: content, Styles: styles})
		return true, nil
	}

	return false, nil
}

// visitFilterRules filters out certain elements based on context.
// Matches Rust: fn visit_filter_rules()
func visitFilterRules(s *state, content eval.ContentElement, styles *eval.StyleChain) (bool, error) {
	switch s.kind.(type) {
	case LayoutPar, *LayoutPar, Math, *Math:
		return false, nil
	}

	// Outside of math and paragraph realization, filter spaces not in paragraph grouper.
	if _, ok := content.(*eval.SpaceElement); ok {
		return true, nil
	}

	// Paragraph breaks are only boundaries, don't store them.
	if _, ok := content.(*eval.ParbreakElement); ok {
		s.mayAttach = false
		s.sawParbreak = true
		return true, nil
	}

	// Attach spacing collapses if not immediately following a paragraph.
	if !s.mayAttach {
		if v, ok := content.(*eval.VElem); ok && v.Attach {
			return true, nil
		}
	}

	// Remember whether following attach spacing can survive.
	_, isPar := content.(*eval.ParagraphElement)
	s.mayAttach = isPar

	return false, nil
}

// visitContent visits each element in Content.
func visitContent(s *state, content eval.Content, styles *eval.StyleChain) error {
	for _, elem := range content.Elements {
		if err := visit(s, elem, styles); err != nil {
			return err
		}
	}
	return nil
}

// ----------------------------------------------------------------------------
// Finish Functions
// ----------------------------------------------------------------------------

// finish finishes all grouping.
// Matches Rust: fn finish()
func finish(s *state) error {
	return finishGroupingWhile(s, func() bool {
		// If fragment realization and all we have is phrasing, don't turn into paragraph.
		if isFullyInline(s) {
			if frag, ok := s.kind.(*LayoutFragment); ok && frag.Kind != nil {
				*frag.Kind = FragmentInline
			}
			if len(s.groupings) > 0 {
				s.groupings = s.groupings[:len(s.groupings)-1]
			}
			collapseSpaces(s.sink, 0)
			return false
		}
		return len(s.groupings) > 0
	})
}

// finishInterrupted finishes groupings interrupted by styles.
// Matches Rust: fn finish_interrupted()
func finishInterrupted(s *state, local *eval.Styles) error {
	for _, rule := range local.Rules {
		if rule.Func == nil || rule.Func.Name == nil {
			continue
		}
		elem := eval.Element{Name: *rule.Func.Name}
		if err := finishGroupingWhile(s, func() bool {
			for _, g := range s.groupings {
				if g.rule.Interrupt(elem) {
					if isFullyInline(s) {
						s.groupings[0].interrupted = true
						return false
					}
					return true
				}
			}
			return false
		}); err != nil {
			return err
		}
	}
	return nil
}

// finishGroupingWhile finishes groupings while condition is true.
// Matches Rust: fn finish_grouping_while()
func finishGroupingWhile(s *state, cond func() bool) error {
	iterations := 0
	for cond() {
		if err := finishInnermostGrouping(s); err != nil {
			return err
		}
		iterations++
		if iterations > 512 {
			return errors.New("maximum grouping depth exceeded")
		}
	}
	return nil
}

// finishInnermostGrouping finishes the currently innermost grouping.
// Matches Rust: fn finish_innermost_grouping()
func finishInnermostGrouping(s *state) error {
	if len(s.groupings) == 0 {
		return nil
	}

	// Pop the grouping.
	g := s.groupings[len(s.groupings)-1]
	s.groupings = s.groupings[:len(s.groupings)-1]

	// Trim trailing non-trigger elements.
	end := len(s.sink)
	for end > g.start && !g.rule.Trigger(s.sink[end-1].Content, s) {
		end--
	}

	// Save tail elements to revisit.
	tail := make([]Pair, len(s.sink)-end)
	copy(tail, s.sink[end:])

	// Truncate sink to end of grouping content.
	s.sink = s.sink[:end]

	// Collect grouped elements (excluding tags if rule doesn't handle them).
	start := g.start
	var tags []Pair
	if !g.rule.Tags {
		newSink := s.sink[:start]
		for i := start; i < len(s.sink); i++ {
			if _, ok := s.sink[i].Content.(*eval.TagElem); ok {
				tags = append(tags, s.sink[i])
			} else {
				newSink = append(newSink, s.sink[i])
			}
		}
		s.sink = newSink
	}

	// Execute the grouping's finisher.
	gr := &grouped{s: s, start: g.start}
	if err := g.rule.Finish(gr); err != nil {
		return err
	}

	// Revisit tags and tail elements.
	for _, pair := range tags {
		if err := visit(s, pair.Content, pair.Styles); err != nil {
			return err
		}
	}
	for _, pair := range tail {
		if err := visit(s, pair.Content, pair.Styles); err != nil {
			return err
		}
	}

	return nil
}

// isFullyInline checks if there's exactly one PAR grouping spanning the whole sink.
// Matches Rust: fn is_fully_inline()
func isFullyInline(s *state) bool {
	if !s.kind.IsFragment() {
		return false
	}
	if s.sawParbreak {
		return false
	}
	if len(s.groupings) != 1 {
		return false
	}
	g := s.groupings[0]
	if g.rule != parRule {
		return false
	}
	// Check that everything before grouping start is tags.
	for i := 0; i < g.start; i++ {
		if _, ok := s.sink[i].Content.(*eval.TagElem); !ok {
			return false
		}
	}
	return true
}

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

// matchesSelector checks if an element matches a selector.
func matchesSelector(elem eval.ContentElement, selector eval.Selector, styles *eval.StyleChain) bool {
	switch sel := selector.(type) {
	case eval.ElemSelector:
		return getElementName(elem) == sel.Element.Name
	case eval.LabelSelector:
		// TODO: Implement label matching
		return false
	case eval.TextSelector:
		if text, ok := elem.(*eval.TextElement); ok {
			if sel.IsRegex {
				re, err := regexp.Compile(sel.Text)
				if err != nil {
					return false
				}
				return re.MatchString(text.Text)
			}
			return text.Text == sel.Text
		}
		return false
	case eval.OrSelector:
		for _, s := range sel.Selectors {
			if matchesSelector(elem, s, styles) {
				return true
			}
		}
		return false
	case eval.AndSelector:
		for _, s := range sel.Selectors {
			if !matchesSelector(elem, s, styles) {
				return false
			}
		}
		return true
	default:
		return false
	}
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
	case *eval.ListElement:
		return "list"
	case *eval.EnumElement:
		return "enum"
	case *eval.TermsElement:
		return "terms"
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
	case *eval.SpaceElement:
		return "space"
	case *eval.HElem:
		return "h"
	case *eval.VElem:
		return "v"
	case *eval.BoxElement:
		return "box"
	case *eval.BlockElement:
		return "block"
	case *eval.AlignElement:
		return "align"
	case *eval.PageElem:
		return "page"
	case *eval.PagebreakElem:
		return "pagebreak"
	case *eval.CiteElement:
		return "cite"
	case *eval.InlineElem:
		return "inline"
	default:
		return ""
	}
}

// isMathy returns true if content can participate in math.
func isMathy(elem eval.ContentElement) bool {
	switch elem.(type) {
	case *eval.MathFracElement, *eval.MathRootElement, *eval.MathAttachElement,
		*eval.MathDelimitedElement, *eval.MathAlignElement, *eval.MathLimitsElement,
		*eval.MathSymbolElement, *eval.MathAccentElement:
		return true
	default:
		return false
	}
}

// isEquation returns true if content is an equation element.
func isEquation(elem eval.ContentElement) bool {
	_, ok := elem.(*eval.EquationElement)
	return ok
}

// getDocumentInfo returns document info if this is a document realization.
func getDocumentInfo(kind RealizationKind) *DocumentInfo {
	switch k := kind.(type) {
	case LayoutDocument:
		return k.Info
	case *LayoutDocument:
		return k.Info
	case HtmlDocument:
		return k.Info
	case *HtmlDocument:
		return k.Info
	default:
		return nil
	}
}

// populateDocumentInfo fills in document info from styles.
func populateDocumentInfo(info *DocumentInfo, styles *eval.Styles) {
	// TODO: Extract document properties from styles
}

// applyBuiltinShowRule applies a built-in show rule.
func applyBuiltinShowRule(engine *eval.Engine, elem eval.ContentElement, styles *eval.StyleChain) (*eval.Content, error) {
	// Built-in show rules convert elements to their visual representation.
	// For now, return nil to indicate no transformation.
	// TODO: Implement built-in show rules for each element type.
	return nil, nil
}

// applyRecipe applies a user-defined show rule recipe.
func applyRecipe(engine *eval.Engine, elem eval.ContentElement, recipe *eval.Recipe, styles *eval.StyleChain) (*eval.Content, error) {
	switch t := recipe.Transform.(type) {
	case eval.NoneTransformation:
		// Hide the element.
		return &eval.Content{}, nil

	case eval.ContentTransformation:
		// Replace with content.
		return &t.Content, nil

	case eval.FuncTransformation:
		// Apply function transformation.
		if t.Func == nil {
			return nil, nil
		}

		// Create a VM to execute the transformation function.
		// We use the engine from the state, and create a fresh context and scopes.
		vm := eval.NewVm(engine, eval.NewContext(), eval.NewScopes(nil), syntax.Detached())

		// Use the ApplyTransformation helper from eval package.
		result, err := eval.ApplyTransformation(t, elem, vm)
		if err != nil {
			return nil, err
		}

		// Convert result to Content.
		return &eval.Content{Elements: result}, nil

	default:
		return nil, nil
	}
}

// Maximum number of nested groups (corresponds to unique priority levels).
const maxGroupNesting = 3

// ----------------------------------------------------------------------------
// Regex Show Rules for Grouped Text
// ----------------------------------------------------------------------------
// These functions implement regex show rules across grouped textual elements.
// Matches Rust: visit_textual, find_regex_match_in_elems, find_regex_match_in_str,
// visit_regex_match in typst-realize/src/lib.rs

// regexMatch holds information about a regex match found in grouped text.
// Matches Rust: RegexMatch struct
type regexMatch struct {
	// offset is the byte offset in the combined text where the match starts.
	offset int
	// text is the matched text.
	text string
	// styles is the style chain of the matching grouping.
	styles *eval.StyleChain
	// recipe is the recipe that matched.
	recipe *eval.Recipe
	// recipeIndex is the index of the recipe (for revocation).
	recipeIndex int
}

// visitTextual handles a completed TEXTUAL grouping by searching for regex matches.
// If a match is found, it splits the elements and applies the transformation.
// Otherwise, it just collapses spaces.
// Matches Rust: visit_textual()
func visitTextual(s *state, pairs []Pair, start int) error {
	if len(pairs) == 0 {
		collapseSpaces(pairs, start)
		return nil
	}

	// Try to find a regex match.
	m := findRegexMatchInElems(pairs)
	if m == nil {
		// No regex match, just collapse spaces.
		collapseSpaces(pairs, start)
		return nil
	}

	// Found a match - apply it.
	return visitRegexMatch(s, pairs, m)
}

// findRegexMatchInElems finds the leftmost regex match across grouped elements.
// Matches Rust: find_regex_match_in_elems()
func findRegexMatchInElems(pairs []Pair) *regexMatch {
	var buf []byte
	base := 0
	var leftmost *regexMatch
	var currentStyles *eval.StyleChain
	state := StateDestructive

	for _, pair := range pairs {
		newState, text := collapseStateTextual(pair.Content, pair.Styles)
		switch newState {
		case StateInvisible:
			continue
		case StateDestructive:
			if state == StateSpace && len(buf) > 0 {
				// Remove trailing space
				buf = buf[:len(buf)-1]
			}
			state = StateDestructive
		case StateSupportive:
			state = StateSupportive
		case StateSpace:
			if state != StateSupportive {
				continue
			}
			state = StateSpace
		}

		// If styles differ, search before adding new text.
		if pair.Styles != currentStyles && len(buf) > 0 {
			leftmost = findRegexMatchInStr(string(buf), currentStyles)
			if leftmost != nil {
				break
			}
			base += len(buf)
			buf = buf[:0]
		}

		currentStyles = pair.Styles
		buf = append(buf, text...)
	}

	// Search in remaining buffer.
	if leftmost == nil && len(buf) > 0 {
		leftmost = findRegexMatchInStr(string(buf), currentStyles)
	}

	// Adjust offset by base.
	if leftmost != nil {
		leftmost.offset = base + leftmost.offset
	}

	return leftmost
}

// collapseStateTextual returns the space state and text for a textual element.
// Matches Rust: collapse_state_textual()
func collapseStateTextual(elem eval.ContentElement, styles *eval.StyleChain) (SpaceState, string) {
	switch e := elem.(type) {
	case *eval.TagElem:
		return StateInvisible, ""
	case *eval.LinebreakElement:
		return StateDestructive, "\n"
	case *eval.SpaceElement:
		return StateSpace, " "
	case *eval.TextElement:
		return StateSupportive, e.Text
	case *eval.SmartQuoteElement:
		if e.Double {
			return StateSupportive, "\""
		}
		return StateSupportive, "'"
	default:
		// Should not happen for textual elements
		return StateSupportive, ""
	}
}

// findRegexMatchInStr searches for regex show rules in the style chain.
// Returns the leftmost match found.
// Matches Rust: find_regex_match_in_str()
func findRegexMatchInStr(text string, styles *eval.StyleChain) *regexMatch {
	if styles == nil || text == "" {
		return nil
	}

	recipes := styles.Recipes()
	if len(recipes) == 0 {
		return nil
	}

	var leftmost *regexMatch

	for i, recipe := range recipes {
		if recipe.Selector == nil {
			continue
		}

		// Check if selector is a regex TextSelector.
		textSel, ok := (*recipe.Selector).(eval.TextSelector)
		if !ok || !textSel.IsRegex {
			continue
		}

		// Try to compile and match the regex.
		re, err := regexp.Compile(textSel.Text)
		if err != nil {
			continue
		}

		loc := re.FindStringIndex(text)
		if loc == nil {
			continue
		}

		// Skip empty matches.
		if loc[0] == loc[1] {
			continue
		}

		// Check if this is more to the left than current best.
		if leftmost != nil && leftmost.offset <= loc[0] {
			continue
		}

		leftmost = &regexMatch{
			offset:      loc[0],
			text:        text[loc[0]:loc[1]],
			styles:      styles,
			recipe:      recipe,
			recipeIndex: i,
		}
	}

	return leftmost
}

// visitRegexMatch applies a regex match transformation.
// It splits the elements around the match and applies the recipe.
// Matches Rust: visit_regex_match()
func visitRegexMatch(s *state, pairs []Pair, m *regexMatch) error {
	matchStart := m.offset
	matchEnd := m.offset + len(m.text)

	// Create the replacement piece.
	piece := &eval.TextElement{Text: m.text}

	// Apply the recipe transformation.
	output, err := applyRecipe(s.engine, piece, m.recipe, m.styles)
	if err != nil {
		return err
	}

	cursor := 0
	outputConsumed := false

	for _, pair := range pairs {
		// Forward tags unchanged.
		if _, ok := pair.Content.(*eval.TagElem); ok {
			if err := visit(s, pair.Content, pair.Styles); err != nil {
				return err
			}
			continue
		}

		// Get the length of this element's text.
		var elemLen int
		switch e := pair.Content.(type) {
		case *eval.TextElement:
			elemLen = len(e.Text)
		case *eval.SpaceElement:
			elemLen = 1
		case *eval.LinebreakElement:
			elemLen = 1
		case *eval.SmartQuoteElement:
			elemLen = 1
		default:
			elemLen = 1
		}

		elemStart := cursor
		elemEnd := cursor + elemLen

		// If element starts before match, visit it fully or sliced.
		if elemStart < matchStart {
			if elemEnd <= matchStart {
				// Entirely before match.
				if err := visit(s, pair.Content, pair.Styles); err != nil {
					return err
				}
			} else {
				// Overlaps with match start - slice the text.
				if text, ok := pair.Content.(*eval.TextElement); ok {
					sliceEnd := matchStart - elemStart
					if sliceEnd > 0 && sliceEnd <= len(text.Text) {
						sliced := &eval.TextElement{Text: text.Text[:sliceEnd]}
						if err := visit(s, sliced, pair.Styles); err != nil {
							return err
						}
					}
				}
			}
		}

		// When the match starts before this element ends, visit the output.
		if matchStart < elemEnd && !outputConsumed {
			if output != nil {
				if err := visitContent(s, *output, m.styles); err != nil {
					return err
				}
			}
			outputConsumed = true
		}

		// If element ends after match, visit it fully or sliced.
		if elemEnd > matchEnd {
			if elemStart >= matchEnd {
				// Entirely after match.
				if err := visit(s, pair.Content, pair.Styles); err != nil {
					return err
				}
			} else {
				// Overlaps with match end - slice the text.
				if text, ok := pair.Content.(*eval.TextElement); ok {
					sliceStart := matchEnd - elemStart
					if sliceStart >= 0 && sliceStart < len(text.Text) {
						sliced := &eval.TextElement{Text: text.Text[sliceStart:]}
						if err := visit(s, sliced, pair.Styles); err != nil {
							return err
						}
					}
				}
			}
		}

		cursor = elemEnd
	}

	// Consume output if not yet done.
	if !outputConsumed && output != nil {
		if err := visitContent(s, *output, m.styles); err != nil {
			return err
		}
	}

	return nil
}
