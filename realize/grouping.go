// Package realize implements the realization subsystem for Typst.
// Realization transforms evaluated content into a form suitable for layout
// by applying show rules, grouping related elements, and collapsing spaces.
package realize

import (
	"github.com/boergens/gotypst/eval"
)

// GroupingRule defines how related elements are collected for unified processing.
// Each rule specifies when to start grouping, what belongs inside, what breaks
// the group, and how to finalize the collected elements.
type GroupingRule interface {
	// Trigger returns true if this element should start a new group.
	// For example, a ListItemElement triggers list grouping.
	Trigger(elem eval.ContentElement) bool

	// Inner returns true if this element belongs inside an active group.
	// For example, consecutive ListItemElements belong inside a list.
	Inner(elem eval.ContentElement) bool

	// Interrupt returns true if this element breaks an active group.
	// For example, a heading interrupts a paragraph group.
	Interrupt(elem eval.ContentElement) bool

	// Finalize creates the grouped element from collected inner elements.
	// For example, creates a ListElement from collected ListItemElements.
	Finalize(elements []eval.ContentElement) eval.ContentElement
}

// GroupingState tracks the state of an active grouping operation.
type GroupingState struct {
	// Rule is the grouping rule being applied.
	Rule GroupingRule

	// Elements are the collected elements.
	Elements []eval.ContentElement
}

// NewGroupingState creates a new grouping state with the given rule and initial element.
func NewGroupingState(rule GroupingRule, initial eval.ContentElement) *GroupingState {
	return &GroupingState{
		Rule:     rule,
		Elements: []eval.ContentElement{initial},
	}
}

// Add adds an element to the group.
func (g *GroupingState) Add(elem eval.ContentElement) {
	g.Elements = append(g.Elements, elem)
}

// Finalize completes the grouping and returns the grouped element.
func (g *GroupingState) Finalize() eval.ContentElement {
	return g.Rule.Finalize(g.Elements)
}

// ----------------------------------------------------------------------------
// Paragraph Grouping Rule
// ----------------------------------------------------------------------------

// ParagraphGroupingRule collects inline content between parbreaks into paragraphs.
// Inline content includes text, strong, emph, links, refs, and other inline elements.
type ParagraphGroupingRule struct{}

// Trigger returns true if the element starts a new paragraph.
// Any inline element can trigger a paragraph.
func (r *ParagraphGroupingRule) Trigger(elem eval.ContentElement) bool {
	return isInlineElement(elem)
}

// Inner returns true if the element belongs inside a paragraph.
// All inline elements belong inside paragraphs.
func (r *ParagraphGroupingRule) Inner(elem eval.ContentElement) bool {
	return isInlineElement(elem)
}

// Interrupt returns true if the element breaks a paragraph.
// Block-level elements and parbreaks interrupt paragraphs.
func (r *ParagraphGroupingRule) Interrupt(elem eval.ContentElement) bool {
	switch elem.(type) {
	case *eval.ParbreakElement:
		return true
	case *eval.HeadingElement:
		return true
	case *eval.ListItemElement:
		return true
	case *eval.EnumItemElement:
		return true
	case *eval.TermItemElement:
		return true
	case *eval.RawElement:
		// Block-level raw elements interrupt
		if raw, ok := elem.(*eval.RawElement); ok {
			return raw.Block
		}
		return false
	case *eval.ParagraphElement:
		// Explicit paragraph elements interrupt implicit paragraphs
		return true
	default:
		return false
	}
}

// Finalize creates a ParagraphElement from the collected inline content.
func (r *ParagraphGroupingRule) Finalize(elements []eval.ContentElement) eval.ContentElement {
	return &eval.ParagraphElement{
		Body: eval.Content{Elements: elements},
	}
}

// isInlineElement returns true if the element is inline content.
func isInlineElement(elem eval.ContentElement) bool {
	switch e := elem.(type) {
	case *eval.TextElement:
		return true
	case *eval.StrongElement:
		return true
	case *eval.EmphElement:
		return true
	case *eval.LinkElement:
		return true
	case *eval.RefElement:
		return true
	case *eval.SmartQuoteElement:
		return true
	case *eval.LinebreakElement:
		return true
	case *eval.RawElement:
		// Inline raw elements (not block)
		return !e.Block
	default:
		return false
	}
}

// ----------------------------------------------------------------------------
// List Grouping Rule
// ----------------------------------------------------------------------------

// ListGroupingRule groups consecutive list items into a list element.
// It handles bullet lists (ListItemElement), numbered lists (EnumItemElement),
// and term lists (TermItemElement) separately.
type ListGroupingRule struct {
	// listType tracks what kind of list is being grouped.
	// Values: "bullet", "enum", "term"
	listType string
}

// NewBulletListGroupingRule creates a rule for grouping bullet list items.
func NewBulletListGroupingRule() *ListGroupingRule {
	return &ListGroupingRule{listType: "bullet"}
}

// NewEnumListGroupingRule creates a rule for grouping enumerated list items.
func NewEnumListGroupingRule() *ListGroupingRule {
	return &ListGroupingRule{listType: "enum"}
}

// NewTermListGroupingRule creates a rule for grouping term list items.
func NewTermListGroupingRule() *ListGroupingRule {
	return &ListGroupingRule{listType: "term"}
}

// Trigger returns true if this element starts a list of the matching type.
func (r *ListGroupingRule) Trigger(elem eval.ContentElement) bool {
	switch r.listType {
	case "bullet":
		_, ok := elem.(*eval.ListItemElement)
		return ok
	case "enum":
		_, ok := elem.(*eval.EnumItemElement)
		return ok
	case "term":
		_, ok := elem.(*eval.TermItemElement)
		return ok
	default:
		return false
	}
}

// Inner returns true if this element belongs inside the list.
func (r *ListGroupingRule) Inner(elem eval.ContentElement) bool {
	// Same as trigger - only matching item types belong inside
	return r.Trigger(elem)
}

// Interrupt returns true if this element breaks the list.
// Different list types interrupt each other, as do block elements.
func (r *ListGroupingRule) Interrupt(elem eval.ContentElement) bool {
	switch r.listType {
	case "bullet":
		switch elem.(type) {
		case *eval.ListItemElement:
			return false // More bullet items continue the list
		case *eval.EnumItemElement, *eval.TermItemElement:
			return true // Different list types interrupt
		case *eval.HeadingElement, *eval.ParagraphElement, *eval.ParbreakElement:
			return true // Block elements interrupt
		default:
			return true // Non-list elements interrupt
		}
	case "enum":
		switch elem.(type) {
		case *eval.EnumItemElement:
			return false
		case *eval.ListItemElement, *eval.TermItemElement:
			return true
		case *eval.HeadingElement, *eval.ParagraphElement, *eval.ParbreakElement:
			return true
		default:
			return true
		}
	case "term":
		switch elem.(type) {
		case *eval.TermItemElement:
			return false
		case *eval.ListItemElement, *eval.EnumItemElement:
			return true
		case *eval.HeadingElement, *eval.ParagraphElement, *eval.ParbreakElement:
			return true
		default:
			return true
		}
	default:
		return true
	}
}

// Finalize creates the appropriate list element from collected items.
func (r *ListGroupingRule) Finalize(elements []eval.ContentElement) eval.ContentElement {
	switch r.listType {
	case "bullet":
		items := make([]*eval.ListItemElement, 0, len(elements))
		for _, e := range elements {
			if item, ok := e.(*eval.ListItemElement); ok {
				items = append(items, item)
			}
		}
		return &eval.ListElement{Items: items}
	case "enum":
		items := make([]*eval.EnumItemElement, 0, len(elements))
		for _, e := range elements {
			if item, ok := e.(*eval.EnumItemElement); ok {
				items = append(items, item)
			}
		}
		return &eval.EnumElement{Items: items}
	case "term":
		items := make([]*eval.TermItemElement, 0, len(elements))
		for _, e := range elements {
			if item, ok := e.(*eval.TermItemElement); ok {
				items = append(items, item)
			}
		}
		return &eval.TermsElement{Items: items}
	default:
		// Fallback: return first element
		if len(elements) > 0 {
			return elements[0]
		}
		return nil
	}
}

// ----------------------------------------------------------------------------
// Citation Grouping Rule
// ----------------------------------------------------------------------------

// CitationGroupingRule collects citations for unified bibliography handling.
// Multiple citations can be grouped together for proper formatting.
type CitationGroupingRule struct{}

// Trigger returns true if this element starts a citation group.
func (r *CitationGroupingRule) Trigger(elem eval.ContentElement) bool {
	_, ok := elem.(*eval.CiteElement)
	return ok
}

// Inner returns true if this element belongs inside a citation group.
// Only cite elements belong inside.
func (r *CitationGroupingRule) Inner(elem eval.ContentElement) bool {
	return r.Trigger(elem)
}

// Interrupt returns true if this element breaks a citation group.
// Any non-citation element interrupts.
func (r *CitationGroupingRule) Interrupt(elem eval.ContentElement) bool {
	_, ok := elem.(*eval.CiteElement)
	return !ok
}

// Finalize creates a citation group from collected citations.
func (r *CitationGroupingRule) Finalize(elements []eval.ContentElement) eval.ContentElement {
	cites := make([]*eval.CiteElement, 0, len(elements))
	for _, e := range elements {
		if cite, ok := e.(*eval.CiteElement); ok {
			cites = append(cites, cite)
		}
	}
	return &eval.CitationGroup{Citations: cites}
}

// ----------------------------------------------------------------------------
// Grouper
// ----------------------------------------------------------------------------

// Grouper applies grouping rules to a sequence of content elements.
type Grouper struct {
	// rules are the grouping rules to apply, in priority order.
	rules []GroupingRule

	// active is the currently active grouping state, if any.
	active *GroupingState

	// output collects the grouped output elements.
	output []eval.ContentElement
}

// NewGrouper creates a new grouper with the standard grouping rules.
// Rules are applied in order: lists first, then paragraphs, then citations.
func NewGrouper() *Grouper {
	return &Grouper{
		rules: []GroupingRule{
			NewBulletListGroupingRule(),
			NewEnumListGroupingRule(),
			NewTermListGroupingRule(),
			&CitationGroupingRule{},
			&ParagraphGroupingRule{},
		},
		active: nil,
		output: nil,
	}
}

// NewGrouperWithRules creates a new grouper with custom rules.
func NewGrouperWithRules(rules []GroupingRule) *Grouper {
	return &Grouper{
		rules:  rules,
		active: nil,
		output: nil,
	}
}

// Process processes a sequence of content elements, applying grouping rules.
func (g *Grouper) Process(elements []eval.ContentElement) []eval.ContentElement {
	g.output = nil
	g.active = nil

	for _, elem := range elements {
		g.processElement(elem)
	}

	// Finalize any remaining active group
	if g.active != nil {
		g.output = append(g.output, g.active.Finalize())
		g.active = nil
	}

	return g.output
}

// processElement handles a single element, applying grouping rules.
func (g *Grouper) processElement(elem eval.ContentElement) {
	// If we have an active group, check if this element continues or interrupts it
	if g.active != nil {
		if g.active.Rule.Inner(elem) {
			g.active.Add(elem)
			return
		}
		if g.active.Rule.Interrupt(elem) {
			g.output = append(g.output, g.active.Finalize())
			g.active = nil
		}
	}

	// Check if any rule triggers on this element
	for _, rule := range g.rules {
		if rule.Trigger(elem) {
			// Start a new group
			g.active = NewGroupingState(rule, elem)
			return
		}
	}

	// No grouping applies - output directly
	g.output = append(g.output, elem)
}

// Reset clears the grouper state for reuse.
func (g *Grouper) Reset() {
	g.active = nil
	g.output = nil
}
