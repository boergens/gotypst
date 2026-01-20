package realize

import (
	"testing"

	"github.com/boergens/gotypst/eval"
	"github.com/boergens/gotypst/syntax"
)

func TestRealizeEmpty(t *testing.T) {
	pairs, err := Realize(LayoutDocument{}, nil, nil, EmptyStyleChain())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 0 {
		t.Errorf("expected 0 pairs, got %d", len(pairs))
	}
}

func TestRealizeEmptyContent(t *testing.T) {
	content := &eval.Content{}
	pairs, err := Realize(LayoutDocument{}, nil, content, EmptyStyleChain())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 0 {
		t.Errorf("expected 0 pairs, got %d", len(pairs))
	}
}

func TestRealizeTextElement(t *testing.T) {
	content := &eval.Content{
		Elements: []eval.ContentElement{
			&eval.TextElement{Text: "Hello"},
		},
	}

	pairs, err := Realize(LayoutDocument{}, nil, content, EmptyStyleChain())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With document kind, inline elements get grouped into paragraphs
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair (paragraph), got %d", len(pairs))
	}

	para, ok := pairs[0].Element.(*eval.ParagraphElement)
	if !ok {
		t.Fatalf("expected ParagraphElement, got %T", pairs[0].Element)
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
	content := &eval.Content{
		Elements: []eval.ContentElement{
			&eval.TextElement{Text: "Hello "},
			&eval.TextElement{Text: "World"},
		},
	}

	pairs, err := Realize(LayoutDocument{}, nil, content, EmptyStyleChain())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both text elements should be grouped into one paragraph
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair (paragraph), got %d", len(pairs))
	}

	para, ok := pairs[0].Element.(*eval.ParagraphElement)
	if !ok {
		t.Fatalf("expected ParagraphElement, got %T", pairs[0].Element)
	}

	if len(para.Body.Elements) != 2 {
		t.Fatalf("expected 2 elements in paragraph, got %d", len(para.Body.Elements))
	}
}

func TestRealizeBlockElement(t *testing.T) {
	content := &eval.Content{
		Elements: []eval.ContentElement{
			&eval.HeadingElement{Level: 1, Content: eval.Content{
				Elements: []eval.ContentElement{
					&eval.TextElement{Text: "Title"},
				},
			}},
		},
	}

	pairs, err := Realize(LayoutDocument{}, nil, content, EmptyStyleChain())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Block elements are not grouped
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(pairs))
	}

	_, ok := pairs[0].Element.(*eval.HeadingElement)
	if !ok {
		t.Fatalf("expected HeadingElement, got %T", pairs[0].Element)
	}
}

func TestRealizeParbreak(t *testing.T) {
	content := &eval.Content{
		Elements: []eval.ContentElement{
			&eval.TextElement{Text: "First"},
			&eval.ParbreakElement{},
			&eval.TextElement{Text: "Second"},
		},
	}

	pairs, err := Realize(LayoutDocument{}, nil, content, EmptyStyleChain())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should produce: paragraph, parbreak, paragraph
	if len(pairs) != 3 {
		t.Fatalf("expected 3 pairs, got %d", len(pairs))
	}

	_, ok := pairs[0].Element.(*eval.ParagraphElement)
	if !ok {
		t.Errorf("expected first element to be ParagraphElement, got %T", pairs[0].Element)
	}

	_, ok = pairs[1].Element.(*eval.ParbreakElement)
	if !ok {
		t.Errorf("expected second element to be ParbreakElement, got %T", pairs[1].Element)
	}

	_, ok = pairs[2].Element.(*eval.ParagraphElement)
	if !ok {
		t.Errorf("expected third element to be ParagraphElement, got %T", pairs[2].Element)
	}
}

func TestFragmentKindDetection(t *testing.T) {
	tests := []struct {
		name     string
		elements []eval.ContentElement
		expected FragmentKind
	}{
		{
			name: "inline grouped into paragraph (becomes block)",
			elements: []eval.ContentElement{
				&eval.TextElement{Text: "Hello"},
				&eval.StrongElement{Content: eval.Content{}},
			},
			// Inline elements get grouped into ParagraphElement which is block
			expected: FragmentBlock,
		},
		{
			name: "block only",
			elements: []eval.ContentElement{
				&eval.HeadingElement{Level: 1},
				&eval.ParagraphElement{},
			},
			expected: FragmentBlock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var kind FragmentKind
			content := &eval.Content{Elements: tt.elements}
			_, err := Realize(&LayoutFragment{Kind: &kind}, nil, content, EmptyStyleChain())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if kind != tt.expected {
				t.Errorf("expected kind %v, got %v", tt.expected, kind)
			}
		})
	}
}

func TestStyleChain(t *testing.T) {
	// Create a style chain with some rules
	funcName := "text"
	propName := "size"
	args := &eval.Args{
		Items: []eval.Arg{
			{Name: &propName, Value: syntax.SpannedDetached[eval.Value](eval.FloatValue(12.0))},
		},
	}

	styles := &eval.Styles{
		Rules: []eval.StyleRule{
			{
				Func: &eval.Func{Name: &funcName},
				Args: args,
			},
		},
	}

	chain := NewStyleChain(styles, nil)

	// Test Get
	val, ok := chain.Get("text", "size")
	if !ok {
		t.Fatal("expected to find 'text.size' in chain")
	}

	if f, ok := val.(eval.FloatValue); !ok || float64(f) != 12.0 {
		t.Errorf("expected 12.0, got %v", val)
	}

	// Test not found
	_, ok = chain.Get("text", "fill")
	if ok {
		t.Error("expected 'text.fill' to not be found")
	}
}

func TestStyleChainInheritance(t *testing.T) {
	funcName := "text"
	parentProp := "size"
	childProp := "fill"

	parentArgs := &eval.Args{
		Items: []eval.Arg{
			{Name: &parentProp, Value: syntax.SpannedDetached[eval.Value](eval.FloatValue(12.0))},
		},
	}
	childArgs := &eval.Args{
		Items: []eval.Arg{
			{Name: &childProp, Value: syntax.SpannedDetached[eval.Value](eval.StrValue("red"))},
		},
	}

	parentStyles := &eval.Styles{
		Rules: []eval.StyleRule{
			{Func: &eval.Func{Name: &funcName}, Args: parentArgs},
		},
	}
	childStyles := &eval.Styles{
		Rules: []eval.StyleRule{
			{Func: &eval.Func{Name: &funcName}, Args: childArgs},
		},
	}

	parent := NewStyleChain(parentStyles, nil)
	child := parent.Chain(childStyles)

	// Child should inherit parent's 'size'
	val, ok := child.Get("text", "size")
	if !ok {
		t.Fatal("expected to find 'text.size' from parent")
	}
	if f, ok := val.(eval.FloatValue); !ok || float64(f) != 12.0 {
		t.Errorf("expected 12.0, got %v", val)
	}

	// Child should have its own 'fill'
	val, ok = child.Get("text", "fill")
	if !ok {
		t.Fatal("expected to find 'text.fill' in child")
	}
	if s, ok := val.(eval.StrValue); !ok || string(s) != "red" {
		t.Errorf("expected 'red', got %v", val)
	}
}
