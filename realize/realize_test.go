package realize

import (
	"testing"

	"github.com/boergens/gotypst/eval"
)

// ----------------------------------------------------------------------------
// Realize Function Tests
// ----------------------------------------------------------------------------

func TestRealizeNil(t *testing.T) {
	pairs, err := Realize(LayoutDocument{}, nil, nil, eval.EmptyStyleChain())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 0 {
		t.Errorf("expected 0 pairs, got %d", len(pairs))
	}
}

func TestRealizeEmptySequence(t *testing.T) {
	content := &eval.SequenceElem{Children: []eval.ContentElement{}}
	pairs, err := Realize(LayoutDocument{}, nil, content, eval.EmptyStyleChain())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 0 {
		t.Errorf("expected 0 pairs, got %d", len(pairs))
	}
}

func TestRealizeTextElement(t *testing.T) {
	content := &eval.SequenceElem{
		Children: []eval.ContentElement{
			&eval.TextElement{Text: "Hello"},
		},
	}

	pairs, err := Realize(LayoutDocument{}, nil, content, eval.EmptyStyleChain())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With document kind, inline elements get grouped into paragraphs
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair (paragraph), got %d", len(pairs))
	}

	para, ok := pairs[0].Content.(*eval.ParagraphElement)
	if !ok {
		t.Fatalf("expected ParagraphElement, got %T", pairs[0].Content)
	}

	if len(para.Body.Elements) != 1 {
		t.Fatalf("expected 1 element in paragraph, got %d", len(para.Body.Elements))
	}

	text, ok := para.Body.Elements[0].(*eval.TextElement)
	if !ok {
		t.Fatalf("expected TextElement in paragraph, got %T", para.Body.Elements[0])
	}

	if text.Text != "Hello" {
		t.Errorf("expected text 'Hello', got '%s'", text.Text)
	}
}

func TestRealizeMultipleTextElements(t *testing.T) {
	content := &eval.SequenceElem{
		Children: []eval.ContentElement{
			&eval.TextElement{Text: "Hello "},
			&eval.TextElement{Text: "World"},
		},
	}

	pairs, err := Realize(LayoutDocument{}, nil, content, eval.EmptyStyleChain())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both text elements should be grouped into one paragraph
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair (paragraph), got %d", len(pairs))
	}

	para, ok := pairs[0].Content.(*eval.ParagraphElement)
	if !ok {
		t.Fatalf("expected ParagraphElement, got %T", pairs[0].Content)
	}

	if len(para.Body.Elements) != 2 {
		t.Fatalf("expected 2 elements in paragraph, got %d", len(para.Body.Elements))
	}
}

func TestRealizeBlockElement(t *testing.T) {
	content := &eval.SequenceElem{
		Children: []eval.ContentElement{
			&eval.HeadingElement{Depth: 1, Content: eval.Content{
				Elements: []eval.ContentElement{
					&eval.TextElement{Text: "Title"},
				},
			}},
		},
	}

	pairs, err := Realize(LayoutDocument{}, nil, content, eval.EmptyStyleChain())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Block elements are not grouped
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(pairs))
	}

	_, ok := pairs[0].Content.(*eval.HeadingElement)
	if !ok {
		t.Fatalf("expected HeadingElement, got %T", pairs[0].Content)
	}
}

func TestRealizeParbreak(t *testing.T) {
	content := &eval.SequenceElem{
		Children: []eval.ContentElement{
			&eval.TextElement{Text: "First"},
			&eval.ParbreakElement{},
			&eval.TextElement{Text: "Second"},
		},
	}

	pairs, err := Realize(LayoutDocument{}, nil, content, eval.EmptyStyleChain())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should produce: paragraph, paragraph (parbreaks are filtered out but separate content)
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs (two paragraphs), got %d", len(pairs))
	}

	_, ok := pairs[0].Content.(*eval.ParagraphElement)
	if !ok {
		t.Errorf("expected first element to be ParagraphElement, got %T", pairs[0].Content)
	}

	_, ok = pairs[1].Content.(*eval.ParagraphElement)
	if !ok {
		t.Errorf("expected second element to be ParagraphElement, got %T", pairs[1].Content)
	}
}

func TestFragmentKindDetection(t *testing.T) {
	tests := []struct {
		name     string
		elements []eval.ContentElement
		expected FragmentKind
	}{
		{
			name: "fully inline content stays inline",
			elements: []eval.ContentElement{
				&eval.TextElement{Text: "Hello"},
				&eval.StrongElement{Content: eval.Content{}},
			},
			// Per Rust is_fully_inline: fragment with only phrasing content,
			// no parbreaks, single PAR grouping spanning whole sink stays inline.
			expected: FragmentInline,
		},
		{
			name: "block only",
			elements: []eval.ContentElement{
				&eval.HeadingElement{Depth: 1},
				&eval.ParagraphElement{},
			},
			expected: FragmentBlock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var kind FragmentKind
			content := &eval.SequenceElem{Children: tt.elements}
			_, err := Realize(&LayoutFragment{Kind: &kind}, nil, content, eval.EmptyStyleChain())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if kind != tt.expected {
				t.Errorf("expected kind %v, got %v", tt.expected, kind)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// RealizationKind Tests
// ----------------------------------------------------------------------------

func TestRealizationKindMethods(t *testing.T) {
	tests := []struct {
		name       string
		kind       RealizationKind
		isDocument bool
		isFragment bool
	}{
		{"LayoutDocument", LayoutDocument{}, true, false},
		{"LayoutFragment", LayoutFragment{}, false, true},
		{"LayoutPar", LayoutPar{}, false, false},
		{"HtmlDocument", HtmlDocument{}, true, false},
		{"HtmlFragment", HtmlFragment{}, false, true},
		{"Math", Math{}, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.kind.IsDocument() != tt.isDocument {
				t.Errorf("IsDocument() = %v, expected %v", tt.kind.IsDocument(), tt.isDocument)
			}
			if tt.kind.IsFragment() != tt.isFragment {
				t.Errorf("IsFragment() = %v, expected %v", tt.kind.IsFragment(), tt.isFragment)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// Pair Tests
// ----------------------------------------------------------------------------

func TestPairStruct(t *testing.T) {
	elem := &eval.TextElement{Text: "test"}
	styles := eval.EmptyStyleChain()

	pair := Pair{
		Content: elem,
		Styles:  styles,
	}

	if pair.Content != elem {
		t.Error("pair.Content mismatch")
	}
	if pair.Styles != styles {
		t.Error("pair.Styles mismatch")
	}
}

// ----------------------------------------------------------------------------
// Helper Function Tests
// ----------------------------------------------------------------------------

func TestGetElementName(t *testing.T) {
	tests := []struct {
		elem     eval.ContentElement
		expected string
	}{
		{&eval.TextElement{}, "text"},
		{&eval.ParagraphElement{}, "par"},
		{&eval.StrongElement{}, "strong"},
		{&eval.EmphElement{}, "emph"},
		{&eval.HeadingElement{}, "heading"},
		{&eval.ListItemElement{}, "list.item"},
		{&eval.EnumItemElement{}, "enum.item"},
		{&eval.TermItemElement{}, "terms.item"},
		{&eval.ListElement{}, "list"},
		{&eval.EnumElement{}, "enum"},
		{&eval.TermsElement{}, "terms"},
		{&eval.LinkElement{}, "link"},
		{&eval.RefElement{}, "ref"},
		{&eval.LinebreakElement{}, "linebreak"},
		{&eval.ParbreakElement{}, "parbreak"},
		{&eval.SmartQuoteElement{}, "smartquote"},
		{&eval.EquationElement{}, "equation"},
		{&eval.ImageElement{}, "image"},
		{&eval.SpaceElement{}, "space"},
		{&eval.HElem{}, "h"},
		{&eval.VElem{}, "v"},
		{&eval.BoxElement{}, "box"},
		{&eval.BlockElement{}, "block"},
		{&eval.AlignElement{}, "align"},
		{&eval.PageElem{}, "page"},
		{&eval.PagebreakElem{}, "pagebreak"},
		{&eval.CiteElement{}, "cite"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := getElementName(tt.elem)
			if result != tt.expected {
				t.Errorf("getElementName() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestMatchesSelector(t *testing.T) {
	styles := eval.EmptyStyleChain()

	tests := []struct {
		name     string
		elem     eval.ContentElement
		selector eval.Selector
		expected bool
	}{
		{
			name:     "ElemSelector matches text",
			elem:     &eval.TextElement{Text: "hello"},
			selector: eval.ElemSelector{Element: eval.Element{Name: "text"}},
			expected: true,
		},
		{
			name:     "ElemSelector doesn't match wrong type",
			elem:     &eval.TextElement{Text: "hello"},
			selector: eval.ElemSelector{Element: eval.Element{Name: "heading"}},
			expected: false,
		},
		{
			name:     "TextSelector exact match",
			elem:     &eval.TextElement{Text: "hello"},
			selector: eval.TextSelector{Text: "hello", IsRegex: false},
			expected: true,
		},
		{
			name:     "TextSelector no match",
			elem:     &eval.TextElement{Text: "hello"},
			selector: eval.TextSelector{Text: "world", IsRegex: false},
			expected: false,
		},
		{
			name: "OrSelector matches first",
			elem: &eval.TextElement{Text: "hello"},
			selector: eval.OrSelector{Selectors: []eval.Selector{
				eval.ElemSelector{Element: eval.Element{Name: "text"}},
				eval.ElemSelector{Element: eval.Element{Name: "heading"}},
			}},
			expected: true,
		},
		{
			name: "OrSelector matches second",
			elem: &eval.HeadingElement{},
			selector: eval.OrSelector{Selectors: []eval.Selector{
				eval.ElemSelector{Element: eval.Element{Name: "text"}},
				eval.ElemSelector{Element: eval.Element{Name: "heading"}},
			}},
			expected: true,
		},
		{
			name: "OrSelector no match",
			elem: &eval.ParagraphElement{},
			selector: eval.OrSelector{Selectors: []eval.Selector{
				eval.ElemSelector{Element: eval.Element{Name: "text"}},
				eval.ElemSelector{Element: eval.Element{Name: "heading"}},
			}},
			expected: false,
		},
		{
			name: "AndSelector matches all",
			elem: &eval.TextElement{Text: "hello"},
			selector: eval.AndSelector{Selectors: []eval.Selector{
				eval.ElemSelector{Element: eval.Element{Name: "text"}},
				eval.TextSelector{Text: "hello", IsRegex: false},
			}},
			expected: true,
		},
		{
			name: "AndSelector fails one",
			elem: &eval.TextElement{Text: "hello"},
			selector: eval.AndSelector{Selectors: []eval.Selector{
				eval.ElemSelector{Element: eval.Element{Name: "text"}},
				eval.TextSelector{Text: "world", IsRegex: false},
			}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesSelector(tt.elem, tt.selector, styles)
			if result != tt.expected {
				t.Errorf("matchesSelector() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
