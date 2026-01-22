package realize

import (
	"github.com/boergens/gotypst/eval"
)

// ----------------------------------------------------------------------------
// Grouping Rules
// ----------------------------------------------------------------------------
// Grouping rules define how content elements are grouped during realization.
// Each rule has a priority (higher wins), trigger/inner predicates, an
// interrupt condition, and a finisher function.
//
// This is a faithful port of Rust's grouping rules in typst-realize/src/lib.rs.

// GroupingRule defines a rule for grouping content elements.
// Matches Rust: GroupingRule struct
type GroupingRule struct {
	// Priority determines nesting order (higher = can nest inside lower).
	Priority uint8
	// Tags indicates if the rule handles tags itself (vs. separating them out).
	Tags bool
	// Trigger returns true if an element can start or extend the group.
	Trigger func(elem eval.ContentElement, s *state) bool
	// Inner returns true if an element can be inside the group but not trigger.
	Inner func(elem eval.ContentElement) bool
	// Interrupt returns true if a set-rule element interrupts the group.
	Interrupt func(elem eval.Element) bool
	// Finish is called when the group is complete.
	Finish func(g *grouped) error
}

// Priority constants for grouping rules.
// Matches Rust: static TEXTUAL, PAR, CITES, LIST, ENUM, TERMS priorities.
const (
	priorityTextual = 0 // Lowest priority - text grouping (within paragraphs)
	priorityPar     = 1 // Paragraph grouping
	priorityCites   = 2 // Citation grouping
	priorityList    = 2 // List item grouping (same priority as cites)
	priorityEnum    = 2 // Enum item grouping (same priority as cites)
	priorityTerms   = 2 // Terms item grouping (same priority as cites)
)

// ----------------------------------------------------------------------------
// Static Grouping Rules
// ----------------------------------------------------------------------------

// textualRule groups inline content for text shaping.
// Matches Rust: static TEXTUAL rule
var textualRule = &GroupingRule{
	Priority: priorityTextual,
	Tags:     true, // TEXTUAL handles tags
	Trigger: func(elem eval.ContentElement, s *state) bool {
		return isPhrasing(elem) && !isGroupable(elem)
	},
	Inner: func(elem eval.ContentElement) bool {
		return isPhrasing(elem)
	},
	Interrupt: func(elem eval.Element) bool {
		return elem.Name == "text"
	},
	// Finish is set in init() to avoid initialization cycle
}

// parRule groups content into paragraphs.
// Matches Rust: static PAR rule
var parRule = &GroupingRule{
	Priority: priorityPar,
	Tags:     false, // PAR does not handle tags
	Trigger: func(elem eval.ContentElement, s *state) bool {
		return isPhrasing(elem)
	},
	Inner: func(elem eval.ContentElement) bool {
		return isPhrasing(elem)
	},
	Interrupt: func(elem eval.Element) bool {
		return elem.Name == "par" || elem.Name == "text"
	},
	// Finish is set in init() to avoid initialization cycle
}

// citesRule groups consecutive citations.
// Matches Rust: static CITES rule
var citesRule = &GroupingRule{
	Priority: priorityCites,
	Tags:     false,
	Trigger: func(elem eval.ContentElement, s *state) bool {
		_, ok := elem.(*eval.CiteElement)
		return ok
	},
	Inner: func(elem eval.ContentElement) bool {
		return false
	},
	Interrupt: func(elem eval.Element) bool {
		return elem.Name == "cite"
	},
	// Finish is set in init() to avoid initialization cycle
}

// listRule groups consecutive list items.
// Matches Rust: static LIST rule
var listRule = &GroupingRule{
	Priority: priorityList,
	Tags:     false,
	Trigger: func(elem eval.ContentElement, s *state) bool {
		_, ok := elem.(*eval.ListItemElement)
		return ok
	},
	Inner: func(elem eval.ContentElement) bool {
		return false
	},
	Interrupt: func(elem eval.Element) bool {
		return elem.Name == "list"
	},
	// Finish is set in init() to avoid initialization cycle
}

// enumRule groups consecutive enum items.
// Matches Rust: static ENUM rule
var enumRule = &GroupingRule{
	Priority: priorityEnum,
	Tags:     false,
	Trigger: func(elem eval.ContentElement, s *state) bool {
		_, ok := elem.(*eval.EnumItemElement)
		return ok
	},
	Inner: func(elem eval.ContentElement) bool {
		return false
	},
	Interrupt: func(elem eval.Element) bool {
		return elem.Name == "enum"
	},
	// Finish is set in init() to avoid initialization cycle
}

// termsRule groups consecutive term items.
// Matches Rust: static TERMS rule
var termsRule = &GroupingRule{
	Priority: priorityTerms,
	Tags:     false,
	Trigger: func(elem eval.ContentElement, s *state) bool {
		_, ok := elem.(*eval.TermItemElement)
		return ok
	},
	Inner: func(elem eval.ContentElement) bool {
		return false
	},
	Interrupt: func(elem eval.Element) bool {
		return elem.Name == "terms"
	},
	// Finish is set in init() to avoid initialization cycle
}

// init sets up the Finish functions for grouping rules.
// This is done in init() to avoid initialization cycles.
func init() {
	// TEXTUAL finish: collapse spaces
	textualRule.Finish = func(g *grouped) error {
		pairs := g.get()
		collapseSpaces(pairs, g.start)
		return nil
	}

	// PAR finish: create paragraph from grouped inline content
	parRule.Finish = func(g *grouped) error {
		pairs := g.get()
		if len(pairs) == 0 {
			return nil
		}

		// Collapse spaces within the paragraph.
		collapseSpaces(pairs, g.start)

		// Take the styles from the first element.
		var styles *eval.StyleChain
		if len(pairs) > 0 {
			styles = pairs[0].Styles
		}

		// Create paragraph element containing the grouped content.
		var children []eval.ContentElement
		for _, p := range pairs {
			children = append(children, p.Content)
		}

		par := &eval.ParagraphElement{
			Body: eval.Content{Elements: children},
		}

		// End the group (remove from sink) and visit the paragraph.
		st := g.end()
		return visit(st, par, styles)
	}

	// CITES finish: create citation group
	citesRule.Finish = func(g *grouped) error {
		pairs := g.get()
		if len(pairs) == 0 {
			return nil
		}

		// Collect citations.
		var cites []*eval.CiteElement
		for _, p := range pairs {
			if cite, ok := p.Content.(*eval.CiteElement); ok {
				cites = append(cites, cite)
			}
		}

		if len(cites) == 0 {
			return nil
		}

		// Take styles from first citation.
		styles := pairs[0].Styles

		// Create citation group.
		group := &eval.CitationGroup{Citations: cites}

		// End the group and visit the citation group.
		st := g.end()
		return visit(st, group, styles)
	}

	// LIST finish: create list from items
	listRule.Finish = func(g *grouped) error {
		pairs := g.get()
		if len(pairs) == 0 {
			return nil
		}

		// Collect list items.
		var items []*eval.ListItemElement
		for _, p := range pairs {
			if item, ok := p.Content.(*eval.ListItemElement); ok {
				items = append(items, item)
			}
		}

		if len(items) == 0 {
			return nil
		}

		// Take styles from first item.
		styles := pairs[0].Styles

		// Create list element.
		tight := true
		list := &eval.ListElement{
			Items: items,
			Tight: &tight,
		}

		// End the group and visit the list.
		st := g.end()
		return visit(st, list, styles)
	}

	// ENUM finish: create enum from items
	enumRule.Finish = func(g *grouped) error {
		pairs := g.get()
		if len(pairs) == 0 {
			return nil
		}

		// Collect enum items.
		var items []*eval.EnumItemElement
		for _, p := range pairs {
			if item, ok := p.Content.(*eval.EnumItemElement); ok {
				items = append(items, item)
			}
		}

		if len(items) == 0 {
			return nil
		}

		// Assign numbers if not already set.
		for i, item := range items {
			if item.Number == 0 {
				item.Number = i + 1
			}
		}

		// Take styles from first item.
		styles := pairs[0].Styles

		// Create enum element.
		tight := true
		enum := &eval.EnumElement{
			Items: items,
			Tight: &tight,
		}

		// End the group and visit the enum.
		st := g.end()
		return visit(st, enum, styles)
	}

	// TERMS finish: create terms from items
	termsRule.Finish = func(g *grouped) error {
		pairs := g.get()
		if len(pairs) == 0 {
			return nil
		}

		// Collect term items.
		var items []*eval.TermItemElement
		for _, p := range pairs {
			if item, ok := p.Content.(*eval.TermItemElement); ok {
				items = append(items, item)
			}
		}

		if len(items) == 0 {
			return nil
		}

		// Take styles from first item.
		styles := pairs[0].Styles

		// Create terms element.
		terms := &eval.TermsElement{Items: items}

		// End the group and visit the terms.
		st := g.end()
		return visit(st, terms, styles)
	}
}

// ----------------------------------------------------------------------------
// Rule Sets for Different Realization Kinds
// ----------------------------------------------------------------------------

// layoutRules are used for document and fragment layout realization.
// Order matters: first matching rule wins within same priority tier.
// Matches Rust: LAYOUT_RULES
var layoutRules = []*GroupingRule{
	parRule,     // Priority 1: paragraphs
	citesRule,   // Priority 2: citations (must be before list/enum/terms)
	listRule,    // Priority 2: lists
	enumRule,    // Priority 2: enums
	termsRule,   // Priority 2: terms
	textualRule, // Priority 0: text grouping (lowest)
}

// layoutParRules are used for paragraph-level realization.
// Matches Rust: LAYOUT_PAR_RULES
var layoutParRules = []*GroupingRule{
	textualRule, // Only text grouping within paragraphs
}

// htmlDocumentRules are used for HTML document realization.
// Matches Rust: HTML_DOCUMENT_RULES
var htmlDocumentRules = []*GroupingRule{
	parRule,
	citesRule,
	listRule,
	enumRule,
	termsRule,
	textualRule,
}

// htmlFragmentRules are used for HTML fragment realization.
// Matches Rust: HTML_FRAGMENT_RULES
var htmlFragmentRules = []*GroupingRule{
	parRule,
	citesRule,
	listRule,
	enumRule,
	termsRule,
	textualRule,
}

// mathRules are used for math realization.
// Matches Rust: MATH_RULES
var mathRules = []*GroupingRule{
	// Math has no grouping rules (elements are processed directly).
}

// ----------------------------------------------------------------------------
// Predicates
// ----------------------------------------------------------------------------

// isPhrasing returns true if an element is inline/phrasing content.
// Matches Rust: is_phrasing() in typst-realize/src/lib.rs
func isPhrasing(elem eval.ContentElement) bool {
	switch elem.(type) {
	// Text and basic inline elements
	case *eval.TextElement, *eval.SpaceElement, *eval.SmartQuoteElement:
		return true

	// Formatting elements
	case *eval.StrongElement, *eval.EmphElement, *eval.RawElement:
		return true

	// Links and references
	case *eval.LinkElement, *eval.RefElement:
		return true

	// Spacing
	case *eval.HElem:
		return true

	// Inline boxes
	case *eval.BoxElement, *eval.InlineElem:
		return true

	// Math (equations are inline unless display mode)
	case *eval.EquationElement:
		return true

	// Line breaks are inline but break lines
	case *eval.LinebreakElement:
		return true

	// Sequences recurse
	case *eval.SequenceElem:
		return true

	// Styled content recurses
	case *eval.StyledElement:
		return true

	// Tags are considered phrasing
	case *eval.TagElem:
		return true

	default:
		return false
	}
}

// isGroupable returns true if an element participates in grouping.
// These elements can trigger their own groups.
// Matches Rust: logic in TEXTUAL trigger
func isGroupable(elem eval.ContentElement) bool {
	switch elem.(type) {
	case *eval.ListItemElement, *eval.EnumItemElement, *eval.TermItemElement:
		return true
	case *eval.CiteElement:
		return true
	default:
		return false
	}
}
