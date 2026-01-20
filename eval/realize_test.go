package eval

import (
	"testing"

	"github.com/boergens/gotypst/syntax"
)

// Helper to convert Selector interface to pointer for recipe
func selPtr(s Selector) *Selector {
	return &s
}

// ----------------------------------------------------------------------------
// Verdict Determination Tests
// ----------------------------------------------------------------------------

func TestDetermineVerdict_NoSelector(t *testing.T) {
	// Recipe with no selector matches everything
	recipe := &Recipe{
		Selector:  nil,
		Transform: NoneTransformation{},
		Span:      syntax.Detached(),
	}

	elem := &TextElement{Text: "hello"}
	ctx := NewRealizeContext(nil, nil)

	verdict, val := DetermineVerdict(recipe, elem, ctx)

	if verdict != VerdictAccept {
		t.Errorf("expected VerdictAccept, got %v", verdict)
	}
	if val == nil {
		t.Error("expected non-nil matched value")
	}
}

func TestDetermineVerdict_ElemSelector_Match(t *testing.T) {
	// Recipe matching text elements
	sel := ElemSelector{Element: Element{Name: "text"}}
	recipe := &Recipe{
		Selector:  selPtr(sel),
		Transform: NoneTransformation{},
		Span:      syntax.Detached(),
	}

	elem := &TextElement{Text: "hello"}
	ctx := NewRealizeContext(nil, nil)

	verdict, val := DetermineVerdict(recipe, elem, ctx)

	if verdict != VerdictAccept {
		t.Errorf("expected VerdictAccept, got %v", verdict)
	}
	if val == nil {
		t.Error("expected non-nil matched value")
	}
}

func TestDetermineVerdict_ElemSelector_NoMatch(t *testing.T) {
	// Recipe matching heading elements (should not match text)
	sel := ElemSelector{Element: Element{Name: "heading"}}
	recipe := &Recipe{
		Selector:  selPtr(sel),
		Transform: NoneTransformation{},
		Span:      syntax.Detached(),
	}

	elem := &TextElement{Text: "hello"}
	ctx := NewRealizeContext(nil, nil)

	verdict, _ := DetermineVerdict(recipe, elem, ctx)

	if verdict != VerdictNone {
		t.Errorf("expected VerdictNone, got %v", verdict)
	}
}

func TestDetermineVerdict_TextSelector_Match(t *testing.T) {
	// Recipe matching literal text
	sel := TextSelector{Text: "hello", IsRegex: false}
	recipe := &Recipe{
		Selector:  selPtr(sel),
		Transform: NoneTransformation{},
		Span:      syntax.Detached(),
	}

	elem := &TextElement{Text: "hello world"}
	ctx := NewRealizeContext(nil, nil)

	verdict, val := DetermineVerdict(recipe, elem, ctx)

	if verdict != VerdictAccept {
		t.Errorf("expected VerdictAccept, got %v", verdict)
	}
	if val == nil {
		t.Error("expected non-nil matched value")
	}
}

func TestDetermineVerdict_TextSelector_NoMatch(t *testing.T) {
	sel := TextSelector{Text: "goodbye", IsRegex: false}
	recipe := &Recipe{
		Selector:  selPtr(sel),
		Transform: NoneTransformation{},
		Span:      syntax.Detached(),
	}

	elem := &TextElement{Text: "hello world"}
	ctx := NewRealizeContext(nil, nil)

	verdict, _ := DetermineVerdict(recipe, elem, ctx)

	if verdict != VerdictNone {
		t.Errorf("expected VerdictNone, got %v", verdict)
	}
}

func TestDetermineVerdict_TextSelector_Regex(t *testing.T) {
	sel := TextSelector{Text: "h[ae]llo", IsRegex: true}
	recipe := &Recipe{
		Selector:  selPtr(sel),
		Transform: NoneTransformation{},
		Span:      syntax.Detached(),
	}

	elem := &TextElement{Text: "hallo world"}
	ctx := NewRealizeContext(nil, nil)

	verdict, val := DetermineVerdict(recipe, elem, ctx)

	if verdict != VerdictAccept {
		t.Errorf("expected VerdictAccept, got %v", verdict)
	}
	if val == nil {
		t.Error("expected non-nil matched value")
	}
	// Check matched value is the regex match
	if s, ok := val.(StrValue); ok {
		if string(s) != "hallo" {
			t.Errorf("expected matched text 'hallo', got %q", string(s))
		}
	}
}

func TestDetermineVerdict_LabelSelector_Match(t *testing.T) {
	sel := LabelSelector{Label: "fig:example"}
	recipe := &Recipe{
		Selector:  selPtr(sel),
		Transform: NoneTransformation{},
		Span:      syntax.Detached(),
	}

	elem := &RefElement{Target: "fig:example"}
	ctx := NewRealizeContext(nil, nil)

	verdict, _ := DetermineVerdict(recipe, elem, ctx)

	if verdict != VerdictAccept {
		t.Errorf("expected VerdictAccept, got %v", verdict)
	}
}

func TestDetermineVerdict_LabelSelector_NoMatch(t *testing.T) {
	sel := LabelSelector{Label: "fig:example"}
	recipe := &Recipe{
		Selector:  selPtr(sel),
		Transform: NoneTransformation{},
		Span:      syntax.Detached(),
	}

	elem := &RefElement{Target: "fig:other"}
	ctx := NewRealizeContext(nil, nil)

	verdict, _ := DetermineVerdict(recipe, elem, ctx)

	if verdict != VerdictNone {
		t.Errorf("expected VerdictNone, got %v", verdict)
	}
}

func TestDetermineVerdict_OrSelector(t *testing.T) {
	// Match either text or heading
	sel := OrSelector{
		Selectors: []Selector{
			ElemSelector{Element: Element{Name: "text"}},
			ElemSelector{Element: Element{Name: "heading"}},
		},
	}
	recipe := &Recipe{
		Selector:  selPtr(sel),
		Transform: NoneTransformation{},
		Span:      syntax.Detached(),
	}

	// Test with text element
	elem1 := &TextElement{Text: "hello"}
	ctx := NewRealizeContext(nil, nil)

	verdict1, _ := DetermineVerdict(recipe, elem1, ctx)
	if verdict1 != VerdictAccept {
		t.Errorf("text element: expected VerdictAccept, got %v", verdict1)
	}

	// Test with heading element
	elem2 := &HeadingElement{Level: 1, Content: Content{}}
	verdict2, _ := DetermineVerdict(recipe, elem2, ctx)
	if verdict2 != VerdictAccept {
		t.Errorf("heading element: expected VerdictAccept, got %v", verdict2)
	}

	// Test with raw element (should not match)
	elem3 := &RawElement{Text: "code"}
	verdict3, _ := DetermineVerdict(recipe, elem3, ctx)
	if verdict3 != VerdictNone {
		t.Errorf("raw element: expected VerdictNone, got %v", verdict3)
	}
}

func TestDetermineVerdict_AndSelector(t *testing.T) {
	// This is a bit artificial since most and-selectors use where clauses
	sel := AndSelector{
		Selectors: []Selector{
			ElemSelector{Element: Element{Name: "text"}},
		},
	}
	recipe := &Recipe{
		Selector:  selPtr(sel),
		Transform: NoneTransformation{},
		Span:      syntax.Detached(),
	}

	elem := &TextElement{Text: "hello"}
	ctx := NewRealizeContext(nil, nil)

	verdict, _ := DetermineVerdict(recipe, elem, ctx)
	if verdict != VerdictAccept {
		t.Errorf("expected VerdictAccept, got %v", verdict)
	}
}

// ----------------------------------------------------------------------------
// Element Name Tests
// ----------------------------------------------------------------------------

func TestGetElementName(t *testing.T) {
	tests := []struct {
		elem ContentElement
		want string
	}{
		{&TextElement{}, "text"},
		{&LinebreakElement{}, "linebreak"},
		{&ParbreakElement{}, "parbreak"},
		{&ParagraphElement{}, "par"},
		{&StrongElement{}, "strong"},
		{&EmphElement{}, "emph"},
		{&RawElement{}, "raw"},
		{&LinkElement{}, "link"},
		{&RefElement{}, "ref"},
		{&HeadingElement{}, "heading"},
		{&ListItemElement{}, "list.item"},
		{&EnumItemElement{}, "enum.item"},
		{&TermItemElement{}, "terms.item"},
		{&ListElement{}, "list"},
		{&EnumElement{}, "enum"},
		{&TermsElement{}, "terms"},
		{&SmartQuoteElement{}, "smartquote"},
		{&EquationElement{}, "math.equation"},
	}

	for _, tt := range tests {
		got := getElementName(tt.elem)
		if got != tt.want {
			t.Errorf("getElementName(%T) = %q, want %q", tt.elem, got, tt.want)
		}
	}
}

// ----------------------------------------------------------------------------
// Transformation Tests
// ----------------------------------------------------------------------------

func TestApplyTransform_None(t *testing.T) {
	prep := &PreparedTransform{
		Transform:    NoneTransformation{},
		MatchedValue: ContentValue{Content: Content{Elements: []ContentElement{&TextElement{Text: "hello"}}}},
	}

	ctx := NewRealizeContext(nil, nil)
	result, err := ApplyTransform(prep, ctx)

	if err != nil {
		t.Fatalf("ApplyTransform error: %v", err)
	}
	if len(result.Elements) != 0 {
		t.Errorf("expected empty content, got %d elements", len(result.Elements))
	}
}

func TestApplyTransform_Content(t *testing.T) {
	replacement := Content{Elements: []ContentElement{&TextElement{Text: "replaced"}}}
	prep := &PreparedTransform{
		Transform:    ContentTransformation{Content: replacement},
		MatchedValue: ContentValue{Content: Content{Elements: []ContentElement{&TextElement{Text: "original"}}}},
	}

	ctx := NewRealizeContext(nil, nil)
	result, err := ApplyTransform(prep, ctx)

	if err != nil {
		t.Fatalf("ApplyTransform error: %v", err)
	}
	if len(result.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(result.Elements))
	}
	if text, ok := result.Elements[0].(*TextElement); !ok || text.Text != "replaced" {
		t.Errorf("expected TextElement with 'replaced', got %T", result.Elements[0])
	}
}

func TestApplyTransform_Style(t *testing.T) {
	styles := &Styles{Rules: []StyleRule{}}
	prep := &PreparedTransform{
		Transform:    StyleTransformation{Styles: styles},
		MatchedValue: ContentValue{Content: Content{Elements: []ContentElement{&TextElement{Text: "styled"}}}},
	}

	ctx := NewRealizeContext(nil, nil)
	result, err := ApplyTransform(prep, ctx)

	if err != nil {
		t.Fatalf("ApplyTransform error: %v", err)
	}
	if len(result.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(result.Elements))
	}
	if _, ok := result.Elements[0].(*StyledElement); !ok {
		t.Errorf("expected StyledElement, got %T", result.Elements[0])
	}
}

// ----------------------------------------------------------------------------
// RealizeContext Tests
// ----------------------------------------------------------------------------

func TestNewRealizeContext(t *testing.T) {
	recipes := []*Recipe{
		{Transform: NoneTransformation{}},
	}

	ctx := NewRealizeContext(nil, recipes)

	if ctx.VM != nil {
		t.Error("expected nil VM")
	}
	if len(ctx.Recipes) != 1 {
		t.Errorf("expected 1 recipe, got %d", len(ctx.Recipes))
	}
	if ctx.Depth != 0 {
		t.Errorf("expected depth 0, got %d", ctx.Depth)
	}
	if ctx.MaxDepth != 64 {
		t.Errorf("expected maxDepth 64, got %d", ctx.MaxDepth)
	}
}

// ----------------------------------------------------------------------------
// RealizeContent Tests
// ----------------------------------------------------------------------------

func TestRealizeContent_NoRecipes(t *testing.T) {
	content := Content{Elements: []ContentElement{
		&TextElement{Text: "hello"},
		&TextElement{Text: " world"},
	}}

	ctx := NewRealizeContext(nil, nil)
	result, err := RealizeContent(content, ctx)

	if err != nil {
		t.Fatalf("RealizeContent error: %v", err)
	}
	if len(result.Elements) != 2 {
		t.Errorf("expected 2 elements, got %d", len(result.Elements))
	}
}

func TestRealizeContent_WithRecipe_ReplaceText(t *testing.T) {
	// Recipe to replace all text with "replaced"
	sel := ElemSelector{Element: Element{Name: "text"}}
	recipe := &Recipe{
		Selector:  selPtr(sel),
		Transform: ContentTransformation{Content: Content{Elements: []ContentElement{&TextElement{Text: "replaced"}}}},
		Span:      syntax.Detached(),
	}

	content := Content{Elements: []ContentElement{
		&TextElement{Text: "original"},
	}}

	ctx := NewRealizeContext(nil, []*Recipe{recipe})
	result, err := RealizeContent(content, ctx)

	if err != nil {
		t.Fatalf("RealizeContent error: %v", err)
	}
	if len(result.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(result.Elements))
	}
	if text, ok := result.Elements[0].(*TextElement); !ok || text.Text != "replaced" {
		t.Errorf("expected 'replaced', got %v", result.Elements[0])
	}
}

func TestRealizeContent_WithRecipe_HideText(t *testing.T) {
	// Recipe to hide all text elements
	sel := ElemSelector{Element: Element{Name: "text"}}
	recipe := &Recipe{
		Selector:  selPtr(sel),
		Transform: NoneTransformation{},
		Span:      syntax.Detached(),
	}

	content := Content{Elements: []ContentElement{
		&TextElement{Text: "hidden"},
	}}

	ctx := NewRealizeContext(nil, []*Recipe{recipe})
	result, err := RealizeContent(content, ctx)

	if err != nil {
		t.Fatalf("RealizeContent error: %v", err)
	}
	if len(result.Elements) != 0 {
		t.Errorf("expected 0 elements (hidden), got %d", len(result.Elements))
	}
}

func TestRealizeContent_NestedContent(t *testing.T) {
	// Strong element containing text
	content := Content{Elements: []ContentElement{
		&StrongElement{Content: Content{Elements: []ContentElement{
			&TextElement{Text: "bold text"},
		}}},
	}}

	// No recipes - should preserve structure
	ctx := NewRealizeContext(nil, nil)
	result, err := RealizeContent(content, ctx)

	if err != nil {
		t.Fatalf("RealizeContent error: %v", err)
	}
	if len(result.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(result.Elements))
	}
	strong, ok := result.Elements[0].(*StrongElement)
	if !ok {
		t.Fatalf("expected StrongElement, got %T", result.Elements[0])
	}
	if len(strong.Content.Elements) != 1 {
		t.Errorf("expected 1 child element, got %d", len(strong.Content.Elements))
	}
}

func TestRealizeContent_RecipeOrderMatters(t *testing.T) {
	// Later recipes take precedence (processed first in reverse order)
	sel := ElemSelector{Element: Element{Name: "text"}}

	recipe1 := &Recipe{
		Selector:  selPtr(sel),
		Transform: ContentTransformation{Content: Content{Elements: []ContentElement{&TextElement{Text: "first"}}}},
		Span:      syntax.Detached(),
	}
	recipe2 := &Recipe{
		Selector:  selPtr(sel),
		Transform: ContentTransformation{Content: Content{Elements: []ContentElement{&TextElement{Text: "second"}}}},
		Span:      syntax.Detached(),
	}

	content := Content{Elements: []ContentElement{
		&TextElement{Text: "original"},
	}}

	// recipe2 comes after recipe1, so it should be applied first
	ctx := NewRealizeContext(nil, []*Recipe{recipe1, recipe2})
	result, err := RealizeContent(content, ctx)

	if err != nil {
		t.Fatalf("RealizeContent error: %v", err)
	}
	if len(result.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(result.Elements))
	}
	// Since recipe2 is checked first (reverse order), its result "second"
	// becomes text, which then matches recipe1, producing "first"
	if text, ok := result.Elements[0].(*TextElement); !ok || text.Text != "first" {
		t.Errorf("expected 'first' (due to cascading), got %v", result.Elements[0])
	}
}

// ----------------------------------------------------------------------------
// PreparedTransform Tests
// ----------------------------------------------------------------------------

func TestPrepareTransform(t *testing.T) {
	recipe := &Recipe{
		Transform: NoneTransformation{},
		Span:      syntax.Detached(),
	}
	matchedVal := Str("test")

	prep := PrepareTransform(recipe, matchedVal)

	if prep.Recipe != recipe {
		t.Error("expected same recipe")
	}
	if prep.MatchedValue != matchedVal {
		t.Error("expected same matched value")
	}
	if _, ok := prep.Transform.(NoneTransformation); !ok {
		t.Errorf("expected NoneTransformation, got %T", prep.Transform)
	}
}

// ----------------------------------------------------------------------------
// StyledElement Tests
// ----------------------------------------------------------------------------

func TestStyledElement(t *testing.T) {
	elem := &StyledElement{
		Content: Content{Elements: []ContentElement{&TextElement{Text: "styled"}}},
		Styles:  &Styles{},
	}

	// Verify it implements ContentElement
	var _ ContentElement = elem

	if len(elem.Content.Elements) != 1 {
		t.Errorf("expected 1 element, got %d", len(elem.Content.Elements))
	}
}

// ----------------------------------------------------------------------------
// Error Tests
// ----------------------------------------------------------------------------

func TestRecursionLimitError(t *testing.T) {
	err := &RecursionLimitError{
		Message: "maximum show rule recursion depth exceeded",
		Depth:   64,
	}

	expected := "maximum show rule recursion depth exceeded"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

// Additional test with proper selector pointer conversion
func TestDetermineVerdict_ElemSelector_WithPointer(t *testing.T) {
	sel := ElemSelector{Element: Element{Name: "heading"}}
	recipe := &Recipe{
		Selector:  selPtr(sel),
		Transform: NoneTransformation{},
		Span:      syntax.Detached(),
	}

	elem := &HeadingElement{Level: 1}
	ctx := NewRealizeContext(nil, nil)

	verdict, _ := DetermineVerdict(recipe, elem, ctx)

	if verdict != VerdictAccept {
		t.Errorf("expected VerdictAccept, got %v", verdict)
	}
}
