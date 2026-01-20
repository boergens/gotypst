package realize

import (
	"testing"

	"github.com/boergens/gotypst/eval"
)

// ----------------------------------------------------------------------------
// Paragraph Grouping Tests
// ----------------------------------------------------------------------------

func TestParagraphGroupingRule_Trigger(t *testing.T) {
	rule := &ParagraphGroupingRule{}

	tests := []struct {
		name     string
		elem     eval.ContentElement
		expected bool
	}{
		{"TextElement triggers", &eval.TextElement{Text: "hello"}, true},
		{"StrongElement triggers", &eval.StrongElement{}, true},
		{"EmphElement triggers", &eval.EmphElement{}, true},
		{"LinkElement triggers", &eval.LinkElement{URL: "http://example.com"}, true},
		{"SmartQuoteElement triggers", &eval.SmartQuoteElement{Double: true}, true},
		{"LinebreakElement triggers", &eval.LinebreakElement{}, true},
		{"Inline RawElement triggers", &eval.RawElement{Text: "code", Block: false}, true},
		{"Block RawElement does not trigger", &eval.RawElement{Text: "code", Block: true}, false},
		{"HeadingElement does not trigger", &eval.HeadingElement{Level: 1}, false},
		{"ListItemElement does not trigger", &eval.ListItemElement{}, false},
		{"ParbreakElement does not trigger", &eval.ParbreakElement{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rule.Trigger(tt.elem)
			if result != tt.expected {
				t.Errorf("Trigger() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestParagraphGroupingRule_Interrupt(t *testing.T) {
	rule := &ParagraphGroupingRule{}

	tests := []struct {
		name     string
		elem     eval.ContentElement
		expected bool
	}{
		{"ParbreakElement interrupts", &eval.ParbreakElement{}, true},
		{"HeadingElement interrupts", &eval.HeadingElement{Level: 1}, true},
		{"ListItemElement interrupts", &eval.ListItemElement{}, true},
		{"EnumItemElement interrupts", &eval.EnumItemElement{}, true},
		{"TermItemElement interrupts", &eval.TermItemElement{}, true},
		{"Block RawElement interrupts", &eval.RawElement{Text: "code", Block: true}, true},
		{"ParagraphElement interrupts", &eval.ParagraphElement{}, true},
		{"TextElement does not interrupt", &eval.TextElement{Text: "hello"}, false},
		{"StrongElement does not interrupt", &eval.StrongElement{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rule.Interrupt(tt.elem)
			if result != tt.expected {
				t.Errorf("Interrupt() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestParagraphGroupingRule_Finalize(t *testing.T) {
	rule := &ParagraphGroupingRule{}

	elements := []eval.ContentElement{
		&eval.TextElement{Text: "Hello "},
		&eval.StrongElement{Content: eval.Content{
			Elements: []eval.ContentElement{&eval.TextElement{Text: "world"}},
		}},
		&eval.TextElement{Text: "!"},
	}

	result := rule.Finalize(elements)
	para, ok := result.(*eval.ParagraphElement)
	if !ok {
		t.Fatalf("Expected *ParagraphElement, got %T", result)
	}

	if len(para.Body.Elements) != 3 {
		t.Errorf("Expected 3 elements in paragraph body, got %d", len(para.Body.Elements))
	}
}

// ----------------------------------------------------------------------------
// List Grouping Tests
// ----------------------------------------------------------------------------

func TestBulletListGroupingRule_Trigger(t *testing.T) {
	rule := NewBulletListGroupingRule()

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
			result := rule.Trigger(tt.elem)
			if result != tt.expected {
				t.Errorf("Trigger() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestEnumListGroupingRule_Trigger(t *testing.T) {
	rule := NewEnumListGroupingRule()

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
			result := rule.Trigger(tt.elem)
			if result != tt.expected {
				t.Errorf("Trigger() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestTermListGroupingRule_Trigger(t *testing.T) {
	rule := NewTermListGroupingRule()

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
			result := rule.Trigger(tt.elem)
			if result != tt.expected {
				t.Errorf("Trigger() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestBulletListGroupingRule_Interrupt(t *testing.T) {
	rule := NewBulletListGroupingRule()

	tests := []struct {
		name     string
		elem     eval.ContentElement
		expected bool
	}{
		{"ListItemElement does not interrupt", &eval.ListItemElement{}, false},
		{"EnumItemElement interrupts", &eval.EnumItemElement{}, true},
		{"TermItemElement interrupts", &eval.TermItemElement{}, true},
		{"HeadingElement interrupts", &eval.HeadingElement{Level: 1}, true},
		{"ParagraphElement interrupts", &eval.ParagraphElement{}, true},
		{"ParbreakElement interrupts", &eval.ParbreakElement{}, true},
		{"TextElement interrupts", &eval.TextElement{Text: "hello"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rule.Interrupt(tt.elem)
			if result != tt.expected {
				t.Errorf("Interrupt() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestBulletListGroupingRule_Finalize(t *testing.T) {
	rule := NewBulletListGroupingRule()

	elements := []eval.ContentElement{
		&eval.ListItemElement{Content: eval.Content{
			Elements: []eval.ContentElement{&eval.TextElement{Text: "Item 1"}},
		}},
		&eval.ListItemElement{Content: eval.Content{
			Elements: []eval.ContentElement{&eval.TextElement{Text: "Item 2"}},
		}},
		&eval.ListItemElement{Content: eval.Content{
			Elements: []eval.ContentElement{&eval.TextElement{Text: "Item 3"}},
		}},
	}

	result := rule.Finalize(elements)
	list, ok := result.(*eval.ListElement)
	if !ok {
		t.Fatalf("Expected *ListElement, got %T", result)
	}

	if len(list.Items) != 3 {
		t.Errorf("Expected 3 items in list, got %d", len(list.Items))
	}
}

func TestEnumListGroupingRule_Finalize(t *testing.T) {
	rule := NewEnumListGroupingRule()

	elements := []eval.ContentElement{
		&eval.EnumItemElement{Number: 1},
		&eval.EnumItemElement{Number: 2},
	}

	result := rule.Finalize(elements)
	enum, ok := result.(*eval.EnumElement)
	if !ok {
		t.Fatalf("Expected *EnumElement, got %T", result)
	}

	if len(enum.Items) != 2 {
		t.Errorf("Expected 2 items in enum, got %d", len(enum.Items))
	}
}

func TestTermListGroupingRule_Finalize(t *testing.T) {
	rule := NewTermListGroupingRule()

	elements := []eval.ContentElement{
		&eval.TermItemElement{
			Term:        eval.Content{Elements: []eval.ContentElement{&eval.TextElement{Text: "Term 1"}}},
			Description: eval.Content{Elements: []eval.ContentElement{&eval.TextElement{Text: "Desc 1"}}},
		},
	}

	result := rule.Finalize(elements)
	terms, ok := result.(*eval.TermsElement)
	if !ok {
		t.Fatalf("Expected *TermsElement, got %T", result)
	}

	if len(terms.Items) != 1 {
		t.Errorf("Expected 1 item in terms, got %d", len(terms.Items))
	}
}

// ----------------------------------------------------------------------------
// Citation Grouping Tests
// ----------------------------------------------------------------------------

func TestCitationGroupingRule_Trigger(t *testing.T) {
	rule := &CitationGroupingRule{}

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
			result := rule.Trigger(tt.elem)
			if result != tt.expected {
				t.Errorf("Trigger() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestCitationGroupingRule_Interrupt(t *testing.T) {
	rule := &CitationGroupingRule{}

	tests := []struct {
		name     string
		elem     eval.ContentElement
		expected bool
	}{
		{"CiteElement does not interrupt", &eval.CiteElement{Key: "smith2020"}, false},
		{"TextElement interrupts", &eval.TextElement{Text: "hello"}, true},
		{"SpaceElement interrupts", &eval.TextElement{Text: " "}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rule.Interrupt(tt.elem)
			if result != tt.expected {
				t.Errorf("Interrupt() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestCitationGroupingRule_Finalize(t *testing.T) {
	rule := &CitationGroupingRule{}

	elements := []eval.ContentElement{
		&eval.CiteElement{Key: "smith2020"},
		&eval.CiteElement{Key: "jones2021"},
	}

	result := rule.Finalize(elements)
	group, ok := result.(*eval.CitationGroup)
	if !ok {
		t.Fatalf("Expected *CitationGroup, got %T", result)
	}

	if len(group.Citations) != 2 {
		t.Errorf("Expected 2 citations in group, got %d", len(group.Citations))
	}
}

// ----------------------------------------------------------------------------
// Grouper Integration Tests
// ----------------------------------------------------------------------------

func TestGrouper_ProcessParagraph(t *testing.T) {
	grouper := NewGrouper()

	// Input: text, text, parbreak, text
	elements := []eval.ContentElement{
		&eval.TextElement{Text: "Hello "},
		&eval.TextElement{Text: "world"},
		&eval.ParbreakElement{},
		&eval.TextElement{Text: "New paragraph"},
	}

	result := grouper.Process(elements)

	// Should have: paragraph, parbreak, paragraph
	if len(result) != 3 {
		t.Fatalf("Expected 3 elements, got %d", len(result))
	}

	// First should be a paragraph
	para1, ok := result[0].(*eval.ParagraphElement)
	if !ok {
		t.Errorf("Expected first element to be *ParagraphElement, got %T", result[0])
	} else if len(para1.Body.Elements) != 2 {
		t.Errorf("Expected first paragraph to have 2 elements, got %d", len(para1.Body.Elements))
	}

	// Second should be parbreak
	if _, ok := result[1].(*eval.ParbreakElement); !ok {
		t.Errorf("Expected second element to be *ParbreakElement, got %T", result[1])
	}

	// Third should be a paragraph
	para2, ok := result[2].(*eval.ParagraphElement)
	if !ok {
		t.Errorf("Expected third element to be *ParagraphElement, got %T", result[2])
	} else if len(para2.Body.Elements) != 1 {
		t.Errorf("Expected second paragraph to have 1 element, got %d", len(para2.Body.Elements))
	}
}

func TestGrouper_ProcessList(t *testing.T) {
	grouper := NewGrouper()

	// Input: list item, list item, text
	elements := []eval.ContentElement{
		&eval.ListItemElement{Content: eval.Content{
			Elements: []eval.ContentElement{&eval.TextElement{Text: "Item 1"}},
		}},
		&eval.ListItemElement{Content: eval.Content{
			Elements: []eval.ContentElement{&eval.TextElement{Text: "Item 2"}},
		}},
		&eval.TextElement{Text: "After list"},
	}

	result := grouper.Process(elements)

	// Should have: list, paragraph
	if len(result) != 2 {
		t.Fatalf("Expected 2 elements, got %d", len(result))
	}

	// First should be a list
	list, ok := result[0].(*eval.ListElement)
	if !ok {
		t.Errorf("Expected first element to be *ListElement, got %T", result[0])
	} else if len(list.Items) != 2 {
		t.Errorf("Expected list to have 2 items, got %d", len(list.Items))
	}

	// Second should be a paragraph (from the text)
	if _, ok := result[1].(*eval.ParagraphElement); !ok {
		t.Errorf("Expected second element to be *ParagraphElement, got %T", result[1])
	}
}

func TestGrouper_ProcessMixedLists(t *testing.T) {
	grouper := NewGrouper()

	// Input: bullet item, enum item (different list types should create separate lists)
	elements := []eval.ContentElement{
		&eval.ListItemElement{Content: eval.Content{
			Elements: []eval.ContentElement{&eval.TextElement{Text: "Bullet"}},
		}},
		&eval.EnumItemElement{Content: eval.Content{
			Elements: []eval.ContentElement{&eval.TextElement{Text: "Numbered"}},
		}},
	}

	result := grouper.Process(elements)

	// Should have: bullet list, enum list
	if len(result) != 2 {
		t.Fatalf("Expected 2 elements, got %d", len(result))
	}

	// First should be a bullet list
	if _, ok := result[0].(*eval.ListElement); !ok {
		t.Errorf("Expected first element to be *ListElement, got %T", result[0])
	}

	// Second should be an enum list
	if _, ok := result[1].(*eval.EnumElement); !ok {
		t.Errorf("Expected second element to be *EnumElement, got %T", result[1])
	}
}

func TestGrouper_ProcessCitations(t *testing.T) {
	grouper := NewGrouper()

	// Input: cite, cite, text
	elements := []eval.ContentElement{
		&eval.CiteElement{Key: "smith2020"},
		&eval.CiteElement{Key: "jones2021"},
		&eval.TextElement{Text: " says..."},
	}

	result := grouper.Process(elements)

	// Should have: citation group, paragraph
	if len(result) != 2 {
		t.Fatalf("Expected 2 elements, got %d", len(result))
	}

	// First should be a citation group
	group, ok := result[0].(*eval.CitationGroup)
	if !ok {
		t.Errorf("Expected first element to be *CitationGroup, got %T", result[0])
	} else if len(group.Citations) != 2 {
		t.Errorf("Expected citation group to have 2 citations, got %d", len(group.Citations))
	}

	// Second should be a paragraph
	if _, ok := result[1].(*eval.ParagraphElement); !ok {
		t.Errorf("Expected second element to be *ParagraphElement, got %T", result[1])
	}
}

func TestGrouper_ProcessHeading(t *testing.T) {
	grouper := NewGrouper()

	// Input: text, heading, text
	// Headings should not be grouped into paragraphs
	elements := []eval.ContentElement{
		&eval.TextElement{Text: "Before"},
		&eval.HeadingElement{Level: 1, Content: eval.Content{
			Elements: []eval.ContentElement{&eval.TextElement{Text: "Title"}},
		}},
		&eval.TextElement{Text: "After"},
	}

	result := grouper.Process(elements)

	// Should have: paragraph, heading, paragraph
	if len(result) != 3 {
		t.Fatalf("Expected 3 elements, got %d", len(result))
	}

	// First should be a paragraph
	if _, ok := result[0].(*eval.ParagraphElement); !ok {
		t.Errorf("Expected first element to be *ParagraphElement, got %T", result[0])
	}

	// Second should be heading (unchanged)
	if _, ok := result[1].(*eval.HeadingElement); !ok {
		t.Errorf("Expected second element to be *HeadingElement, got %T", result[1])
	}

	// Third should be a paragraph
	if _, ok := result[2].(*eval.ParagraphElement); !ok {
		t.Errorf("Expected third element to be *ParagraphElement, got %T", result[2])
	}
}

func TestGrouper_EmptyInput(t *testing.T) {
	grouper := NewGrouper()

	result := grouper.Process(nil)

	if len(result) != 0 {
		t.Errorf("Expected 0 elements for nil input, got %d", len(result))
	}

	result = grouper.Process([]eval.ContentElement{})

	if len(result) != 0 {
		t.Errorf("Expected 0 elements for empty input, got %d", len(result))
	}
}

func TestGrouper_Reset(t *testing.T) {
	grouper := NewGrouper()

	// Process some elements
	elements := []eval.ContentElement{
		&eval.TextElement{Text: "Hello"},
	}
	grouper.Process(elements)

	// Reset
	grouper.Reset()

	// Process again should work
	result := grouper.Process(elements)
	if len(result) != 1 {
		t.Errorf("Expected 1 element after reset, got %d", len(result))
	}
}

