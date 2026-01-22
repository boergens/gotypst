package realize

import (
	"testing"

	"github.com/boergens/gotypst/eval"
)

// ----------------------------------------------------------------------------
// isPhrasing Tests
// ----------------------------------------------------------------------------

func TestIsPhrasing(t *testing.T) {
	tests := []struct {
		name     string
		elem     eval.ContentElement
		expected bool
	}{
		// Text and basic inline elements
		{"TextElement is phrasing", &eval.TextElement{Text: "hello"}, true},
		{"SpaceElement is phrasing", &eval.SpaceElement{}, true},
		{"SmartQuoteElement is phrasing", &eval.SmartQuoteElement{Double: true}, true},

		// Formatting elements
		{"StrongElement is phrasing", &eval.StrongElement{}, true},
		{"EmphElement is phrasing", &eval.EmphElement{}, true},
		{"RawElement is phrasing", &eval.RawElement{Text: "code"}, true},

		// Links and references
		{"LinkElement is phrasing", &eval.LinkElement{URL: "http://example.com"}, true},
		{"RefElement is phrasing", &eval.RefElement{Target: "fig:1"}, true},

		// Spacing and boxes
		{"HElem is phrasing", &eval.HElem{}, true},
		{"BoxElement is phrasing", &eval.BoxElement{}, true},
		{"InlineElem is phrasing", &eval.InlineElem{}, true},

		// Math
		{"EquationElement is phrasing", &eval.EquationElement{}, true},

		// Line breaks
		{"LinebreakElement is phrasing", &eval.LinebreakElement{}, true},

		// Sequences and styled
		{"SequenceElem is phrasing", &eval.SequenceElem{}, true},
		{"StyledElement is phrasing", &eval.StyledElement{}, true},

		// Tags
		{"TagElem is phrasing", &eval.TagElem{}, true},

		// Block elements are NOT phrasing
		{"HeadingElement is not phrasing", &eval.HeadingElement{Depth: 1}, false},
		{"ParagraphElement is not phrasing", &eval.ParagraphElement{}, false},
		{"ListItemElement is not phrasing", &eval.ListItemElement{}, false},
		{"EnumItemElement is not phrasing", &eval.EnumItemElement{}, false},
		{"TermItemElement is not phrasing", &eval.TermItemElement{}, false},
		{"ListElement is not phrasing", &eval.ListElement{}, false},
		{"EnumElement is not phrasing", &eval.EnumElement{}, false},
		{"TermsElement is not phrasing", &eval.TermsElement{}, false},
		{"BlockElement is not phrasing", &eval.BlockElement{}, false},

		// Breaks
		{"ParbreakElement is not phrasing", &eval.ParbreakElement{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPhrasing(tt.elem)
			if result != tt.expected {
				t.Errorf("isPhrasing() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// isGroupable Tests
// ----------------------------------------------------------------------------

func TestIsGroupable(t *testing.T) {
	tests := []struct {
		name     string
		elem     eval.ContentElement
		expected bool
	}{
		// Groupable elements (trigger their own groups)
		{"ListItemElement is groupable", &eval.ListItemElement{}, true},
		{"EnumItemElement is groupable", &eval.EnumItemElement{}, true},
		{"TermItemElement is groupable", &eval.TermItemElement{}, true},
		{"CiteElement is groupable", &eval.CiteElement{Key: "key"}, true},

		// Non-groupable elements
		{"TextElement is not groupable", &eval.TextElement{Text: "hello"}, false},
		{"SpaceElement is not groupable", &eval.SpaceElement{}, false},
		{"StrongElement is not groupable", &eval.StrongElement{}, false},
		{"ParagraphElement is not groupable", &eval.ParagraphElement{}, false},
		{"HeadingElement is not groupable", &eval.HeadingElement{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGroupable(tt.elem)
			if result != tt.expected {
				t.Errorf("isGroupable() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// GroupingRule Tests
// ----------------------------------------------------------------------------

func TestParRule_Trigger(t *testing.T) {
	// parRule triggers on phrasing content
	state := &state{}

	tests := []struct {
		name     string
		elem     eval.ContentElement
		expected bool
	}{
		{"TextElement triggers", &eval.TextElement{Text: "hello"}, true},
		{"StrongElement triggers", &eval.StrongElement{}, true},
		{"LinkElement triggers", &eval.LinkElement{URL: "http://example.com"}, true},
		{"HeadingElement does not trigger", &eval.HeadingElement{Depth: 1}, false},
		{"ListItemElement does not trigger", &eval.ListItemElement{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parRule.Trigger(tt.elem, state)
			if result != tt.expected {
				t.Errorf("parRule.Trigger() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestParRule_Interrupt(t *testing.T) {
	tests := []struct {
		name     string
		elem     eval.Element
		expected bool
	}{
		{"par interrupts", eval.Element{Name: "par"}, true},
		{"text interrupts", eval.Element{Name: "text"}, true},
		{"heading does not interrupt", eval.Element{Name: "heading"}, false},
		{"list does not interrupt", eval.Element{Name: "list"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parRule.Interrupt(tt.elem)
			if result != tt.expected {
				t.Errorf("parRule.Interrupt() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestListRule_Trigger(t *testing.T) {
	state := &state{}

	tests := []struct {
		name     string
		elem     eval.ContentElement
		expected bool
	}{
		{"ListItemElement triggers", &eval.ListItemElement{}, true},
		{"EnumItemElement does not trigger", &eval.EnumItemElement{}, false},
		{"TermItemElement does not trigger", &eval.TermItemElement{}, false},
		{"TextElement does not trigger", &eval.TextElement{Text: "hello"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := listRule.Trigger(tt.elem, state)
			if result != tt.expected {
				t.Errorf("listRule.Trigger() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestEnumRule_Trigger(t *testing.T) {
	state := &state{}

	tests := []struct {
		name     string
		elem     eval.ContentElement
		expected bool
	}{
		{"EnumItemElement triggers", &eval.EnumItemElement{}, true},
		{"ListItemElement does not trigger", &eval.ListItemElement{}, false},
		{"TermItemElement does not trigger", &eval.TermItemElement{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := enumRule.Trigger(tt.elem, state)
			if result != tt.expected {
				t.Errorf("enumRule.Trigger() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestTermsRule_Trigger(t *testing.T) {
	state := &state{}

	tests := []struct {
		name     string
		elem     eval.ContentElement
		expected bool
	}{
		{"TermItemElement triggers", &eval.TermItemElement{}, true},
		{"ListItemElement does not trigger", &eval.ListItemElement{}, false},
		{"EnumItemElement does not trigger", &eval.EnumItemElement{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := termsRule.Trigger(tt.elem, state)
			if result != tt.expected {
				t.Errorf("termsRule.Trigger() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestCitesRule_Trigger(t *testing.T) {
	state := &state{}

	tests := []struct {
		name     string
		elem     eval.ContentElement
		expected bool
	}{
		{"CiteElement triggers", &eval.CiteElement{Key: "smith2020"}, true},
		{"TextElement does not trigger", &eval.TextElement{Text: "hello"}, false},
		{"RefElement does not trigger", &eval.RefElement{Target: "fig:1"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := citesRule.Trigger(tt.elem, state)
			if result != tt.expected {
				t.Errorf("citesRule.Trigger() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestTextualRule_Trigger(t *testing.T) {
	state := &state{}

	tests := []struct {
		name     string
		elem     eval.ContentElement
		expected bool
	}{
		// Phrasing but not groupable -> triggers
		{"TextElement triggers", &eval.TextElement{Text: "hello"}, true},
		{"SpaceElement triggers", &eval.SpaceElement{}, true},
		{"StrongElement triggers", &eval.StrongElement{}, true},

		// Groupable elements don't trigger (they have their own rules)
		{"ListItemElement does not trigger", &eval.ListItemElement{}, false},
		{"CiteElement does not trigger", &eval.CiteElement{Key: "key"}, false},

		// Non-phrasing doesn't trigger
		{"HeadingElement does not trigger", &eval.HeadingElement{}, false},
		{"ParagraphElement does not trigger", &eval.ParagraphElement{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := textualRule.Trigger(tt.elem, state)
			if result != tt.expected {
				t.Errorf("textualRule.Trigger() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// Rule Priority Tests
// ----------------------------------------------------------------------------

func TestRulePriorities(t *testing.T) {
	// Verify rule priorities match expected hierarchy
	if textualRule.Priority != 0 {
		t.Errorf("textualRule.Priority = %d, expected 0", textualRule.Priority)
	}
	if parRule.Priority != 1 {
		t.Errorf("parRule.Priority = %d, expected 1", parRule.Priority)
	}
	if citesRule.Priority != 2 {
		t.Errorf("citesRule.Priority = %d, expected 2", citesRule.Priority)
	}
	if listRule.Priority != 2 {
		t.Errorf("listRule.Priority = %d, expected 2", listRule.Priority)
	}
	if enumRule.Priority != 2 {
		t.Errorf("enumRule.Priority = %d, expected 2", enumRule.Priority)
	}
	if termsRule.Priority != 2 {
		t.Errorf("termsRule.Priority = %d, expected 2", termsRule.Priority)
	}
}

// ----------------------------------------------------------------------------
// Rule Sets Tests
// ----------------------------------------------------------------------------

func TestLayoutRulesOrder(t *testing.T) {
	// Verify layout rules are in expected order
	if len(layoutRules) != 6 {
		t.Errorf("layoutRules length = %d, expected 6", len(layoutRules))
	}

	expectedOrder := []*GroupingRule{parRule, citesRule, listRule, enumRule, termsRule, textualRule}
	for i, rule := range expectedOrder {
		if layoutRules[i] != rule {
			t.Errorf("layoutRules[%d] mismatch", i)
		}
	}
}

func TestLayoutParRulesOrder(t *testing.T) {
	// Verify layout par rules contain only textual
	if len(layoutParRules) != 1 {
		t.Errorf("layoutParRules length = %d, expected 1", len(layoutParRules))
	}
	if layoutParRules[0] != textualRule {
		t.Error("layoutParRules[0] should be textualRule")
	}
}

func TestMathRulesEmpty(t *testing.T) {
	// Math has no grouping rules
	if len(mathRules) != 0 {
		t.Errorf("mathRules length = %d, expected 0", len(mathRules))
	}
}
