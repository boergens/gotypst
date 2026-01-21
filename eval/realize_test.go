package eval

import (
	"testing"

	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// ElementName Tests
// ----------------------------------------------------------------------------

func TestElementName(t *testing.T) {
	tests := []struct {
		name     string
		elem     ContentElement
		expected string
	}{
		{"TextElement", &TextElement{Text: "hello"}, "text"},
		{"ParagraphElement", &ParagraphElement{}, "par"},
		{"StrongElement", &StrongElement{}, "strong"},
		{"EmphElement", &EmphElement{}, "emph"},
		{"RawElement", &RawElement{}, "raw"},
		{"LinkElement", &LinkElement{}, "link"},
		{"RefElement", &RefElement{}, "ref"},
		{"HeadingElement", &HeadingElement{Depth: 1}, "heading"},
		{"ListItemElement", &ListItemElement{}, "list.item"},
		{"EnumItemElement", &EnumItemElement{}, "enum.item"},
		{"TermItemElement", &TermItemElement{}, "terms.item"},
		{"LinebreakElement", &LinebreakElement{}, "linebreak"},
		{"ParbreakElement", &ParbreakElement{}, "parbreak"},
		{"SmartQuoteElement", &SmartQuoteElement{}, "smartquote"},
		{"EquationElement", &EquationElement{}, "equation"},
		{"MathFracElement", &MathFracElement{}, "math.frac"},
		{"MathRootElement", &MathRootElement{}, "math.root"},
		{"MathAttachElement", &MathAttachElement{}, "math.attach"},
		{"MathDelimitedElement", &MathDelimitedElement{}, "math.lr"},
		{"MathAlignElement", &MathAlignElement{}, "math.align-point"},
		{"MathSymbolElement", &MathSymbolElement{}, "math.symbol"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ElementName(tt.elem)
			if got != tt.expected {
				t.Errorf("ElementName(%T) = %q, want %q", tt.elem, got, tt.expected)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// MatchSelector Tests
// ----------------------------------------------------------------------------

func TestMatchElemSelector(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Create a heading element
	heading := &HeadingElement{Depth: 1, Content: Content{}}

	// Selector that matches heading
	sel := ElemSelector{Element: Element{Name: "heading"}}

	result, err := MatchSelector(sel, heading, vm)
	if err != nil {
		t.Fatalf("MatchSelector error: %v", err)
	}
	if !result.Matched {
		t.Error("expected heading to match heading selector")
	}

	// Selector that doesn't match
	sel2 := ElemSelector{Element: Element{Name: "text"}}
	result2, err := MatchSelector(sel2, heading, vm)
	if err != nil {
		t.Fatalf("MatchSelector error: %v", err)
	}
	if result2.Matched {
		t.Error("expected heading not to match text selector")
	}
}

func TestMatchTextSelector(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	text := &TextElement{Text: "Hello, World!"}

	// Plain string match
	sel := TextSelector{Text: "World", IsRegex: false}
	result, err := MatchSelector(sel, text, vm)
	if err != nil {
		t.Fatalf("MatchSelector error: %v", err)
	}
	if !result.Matched {
		t.Error("expected 'World' to match in 'Hello, World!'")
	}
	if result.MatchedText != "World" {
		t.Errorf("MatchedText = %q, want 'World'", result.MatchedText)
	}

	// No match
	sel2 := TextSelector{Text: "foo", IsRegex: false}
	result2, err := MatchSelector(sel2, text, vm)
	if err != nil {
		t.Fatalf("MatchSelector error: %v", err)
	}
	if result2.Matched {
		t.Error("expected 'foo' not to match in 'Hello, World!'")
	}
}

func TestMatchTextSelectorRegex(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	text := &TextElement{Text: "Error code: 404"}

	// Regex match
	sel := TextSelector{Text: `\d+`, IsRegex: true}
	result, err := MatchSelector(sel, text, vm)
	if err != nil {
		t.Fatalf("MatchSelector error: %v", err)
	}
	if !result.Matched {
		t.Error("expected regex to match digits")
	}
	if result.MatchedText != "404" {
		t.Errorf("MatchedText = %q, want '404'", result.MatchedText)
	}
}

func TestMatchOrSelector(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	heading := &HeadingElement{Depth: 1}

	// Or selector with heading or text
	sel := OrSelector{
		Selectors: []Selector{
			ElemSelector{Element: Element{Name: "text"}},
			ElemSelector{Element: Element{Name: "heading"}},
		},
	}

	result, err := MatchSelector(sel, heading, vm)
	if err != nil {
		t.Fatalf("MatchSelector error: %v", err)
	}
	if !result.Matched {
		t.Error("expected OrSelector to match heading")
	}
}

func TestMatchAndSelector(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	heading := &HeadingElement{Depth: 1}

	// And selector (both must match - impossible for different element types)
	sel := AndSelector{
		Selectors: []Selector{
			ElemSelector{Element: Element{Name: "heading"}},
			ElemSelector{Element: Element{Name: "text"}},
		},
	}

	result, err := MatchSelector(sel, heading, vm)
	if err != nil {
		t.Fatalf("MatchSelector error: %v", err)
	}
	if result.Matched {
		t.Error("expected AndSelector with conflicting types not to match")
	}

	// And selector (same type twice - should match)
	sel2 := AndSelector{
		Selectors: []Selector{
			ElemSelector{Element: Element{Name: "heading"}},
			ElemSelector{Element: Element{Name: "heading"}},
		},
	}

	result2, err := MatchSelector(sel2, heading, vm)
	if err != nil {
		t.Fatalf("MatchSelector error: %v", err)
	}
	if !result2.Matched {
		t.Error("expected AndSelector with same type to match")
	}
}

// ----------------------------------------------------------------------------
// ApplyTransformation Tests
// ----------------------------------------------------------------------------

func TestApplyNoneTransformation(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	text := &TextElement{Text: "hello"}
	trans := NoneTransformation{}

	result, err := ApplyTransformation(trans, text, vm)
	if err != nil {
		t.Fatalf("ApplyTransformation error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected NoneTransformation to return no elements, got %d", len(result))
	}
}

func TestApplyContentTransformation(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	text := &TextElement{Text: "hello"}
	replacement := Content{
		Elements: []ContentElement{
			&StrongElement{Content: Content{Elements: []ContentElement{&TextElement{Text: "HELLO"}}}},
		},
	}
	trans := ContentTransformation{Content: replacement}

	result, err := ApplyTransformation(trans, text, vm)
	if err != nil {
		t.Fatalf("ApplyTransformation error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 element, got %d", len(result))
	}
	if _, ok := result[0].(*StrongElement); !ok {
		t.Errorf("expected StrongElement, got %T", result[0])
	}
}

func TestApplyStyleTransformation(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	text := &TextElement{Text: "hello"}
	trans := StyleTransformation{Styles: &Styles{}}

	result, err := ApplyTransformation(trans, text, vm)
	if err != nil {
		t.Fatalf("ApplyTransformation error: %v", err)
	}
	// Style transformations don't change the element structure
	if len(result) != 1 {
		t.Errorf("expected 1 element, got %d", len(result))
	}
	if result[0] != text {
		t.Error("expected element to be unchanged")
	}
}

func TestApplyFuncTransformation(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	text := &TextElement{Text: "hello"}

	// Create a transformation function that wraps in strong
	name := "transform"
	transformFunc := &Func{
		Name: &name,
		Repr: NativeFunc{
			Func: func(_ *Vm, args *Args) (Value, error) {
				bodyArg := args.Eat()
				if bodyArg == nil {
					return None, nil
				}
				body, ok := bodyArg.V.(ContentValue)
				if !ok {
					return None, nil
				}
				return ContentValue{Content: Content{
					Elements: []ContentElement{&StrongElement{Content: body.Content}},
				}}, nil
			},
		},
	}
	trans := FuncTransformation{Func: transformFunc}

	result, err := ApplyTransformation(trans, text, vm)
	if err != nil {
		t.Fatalf("ApplyTransformation error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 element, got %d", len(result))
	}
	strong, ok := result[0].(*StrongElement)
	if !ok {
		t.Fatalf("expected StrongElement, got %T", result[0])
	}
	if len(strong.Content.Elements) != 1 {
		t.Errorf("expected 1 inner element, got %d", len(strong.Content.Elements))
	}
}

// ----------------------------------------------------------------------------
// Realize Tests
// ----------------------------------------------------------------------------

func TestRealizeNoRules(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	content := Content{
		Elements: []ContentElement{
			&TextElement{Text: "hello"},
			&TextElement{Text: " "},
			&TextElement{Text: "world"},
		},
	}

	// No styles
	result, err := Realize(content, nil, vm)
	if err != nil {
		t.Fatalf("Realize error: %v", err)
	}
	if len(result.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(result.Elements))
	}
}

func TestRealizeWithNoneTransform(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	content := Content{
		Elements: []ContentElement{
			&TextElement{Text: "hello"},
			&HeadingElement{Depth: 1, Content: Content{Elements: []ContentElement{&TextElement{Text: "Title"}}}},
			&TextElement{Text: "world"},
		},
	}

	// Hide all headings
	headingSel := ElemSelector{Element: Element{Name: "heading"}}
	sel := Selector(headingSel)
	styles := &Styles{
		Recipes: []*Recipe{
			{
				Selector:  &sel,
				Transform: NoneTransformation{},
			},
		},
	}

	result, err := Realize(content, styles, vm)
	if err != nil {
		t.Fatalf("Realize error: %v", err)
	}
	// Heading should be removed
	if len(result.Elements) != 2 {
		t.Errorf("expected 2 elements (heading hidden), got %d", len(result.Elements))
	}
	for _, elem := range result.Elements {
		if _, ok := elem.(*HeadingElement); ok {
			t.Error("heading should have been hidden")
		}
	}
}

func TestRealizeWithContentReplacement(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	content := Content{
		Elements: []ContentElement{
			&HeadingElement{Depth: 1, Content: Content{Elements: []ContentElement{&TextElement{Text: "Title"}}}},
		},
	}

	// Replace heading with emphasized content
	headingSel := ElemSelector{Element: Element{Name: "heading"}}
	sel := Selector(headingSel)
	replacement := Content{
		Elements: []ContentElement{
			&EmphElement{Content: Content{Elements: []ContentElement{&TextElement{Text: "Replaced Title"}}}},
		},
	}
	styles := &Styles{
		Recipes: []*Recipe{
			{
				Selector:  &sel,
				Transform: ContentTransformation{Content: replacement},
			},
		},
	}

	result, err := Realize(content, styles, vm)
	if err != nil {
		t.Fatalf("Realize error: %v", err)
	}
	if len(result.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(result.Elements))
	}
	emph, ok := result.Elements[0].(*EmphElement)
	if !ok {
		t.Fatalf("expected EmphElement, got %T", result.Elements[0])
	}
	if len(emph.Content.Elements) != 1 {
		t.Errorf("expected 1 inner element, got %d", len(emph.Content.Elements))
	}
}

func TestRealizeRecursiveChildren(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	// Content with nested structure
	content := Content{
		Elements: []ContentElement{
			&StrongElement{
				Content: Content{
					Elements: []ContentElement{
						&HeadingElement{
							Depth: 1,
							Content: Content{
								Elements: []ContentElement{
									&TextElement{Text: "Nested Title"},
								},
							},
						},
					},
				},
			},
		},
	}

	// Hide all headings (should apply inside strong)
	headingSel := ElemSelector{Element: Element{Name: "heading"}}
	sel := Selector(headingSel)
	styles := &Styles{
		Recipes: []*Recipe{
			{
				Selector:  &sel,
				Transform: NoneTransformation{},
			},
		},
	}

	result, err := Realize(content, styles, vm)
	if err != nil {
		t.Fatalf("Realize error: %v", err)
	}
	if len(result.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(result.Elements))
	}
	strong, ok := result.Elements[0].(*StrongElement)
	if !ok {
		t.Fatalf("expected StrongElement, got %T", result.Elements[0])
	}
	// The heading inside should be hidden
	if len(strong.Content.Elements) != 0 {
		t.Errorf("expected 0 inner elements (heading hidden), got %d", len(strong.Content.Elements))
	}
}

func TestRealizeWithUniversalRule(t *testing.T) {
	scopes := NewScopes(nil)
	vm := NewVm(nil, NewContext(), scopes, syntax.Detached())

	content := Content{
		Elements: []ContentElement{
			&TextElement{Text: "hello"},
		},
	}

	// Rule with no selector (matches everything)
	replacement := Content{
		Elements: []ContentElement{
			&StrongElement{Content: Content{Elements: []ContentElement{&TextElement{Text: "replaced"}}}},
		},
	}
	styles := &Styles{
		Recipes: []*Recipe{
			{
				Selector:  nil, // No selector = matches all
				Transform: ContentTransformation{Content: replacement},
			},
		},
	}

	result, err := Realize(content, styles, vm)
	if err != nil {
		t.Fatalf("Realize error: %v", err)
	}
	if len(result.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(result.Elements))
	}
	_, ok := result.Elements[0].(*StrongElement)
	if !ok {
		t.Fatalf("expected StrongElement, got %T", result.Elements[0])
	}
}

// ----------------------------------------------------------------------------
// MergeStyles Tests
// ----------------------------------------------------------------------------

func TestMergeStyles(t *testing.T) {
	s1 := &Styles{
		Rules: []StyleRule{{Func: &Func{Name: realizeStrPtr("a")}}},
		Recipes: []*Recipe{
			{Transform: NoneTransformation{}},
		},
	}
	s2 := &Styles{
		Rules: []StyleRule{{Func: &Func{Name: realizeStrPtr("b")}}},
		Recipes: []*Recipe{
			{Transform: NoneTransformation{}},
		},
	}

	merged := MergeStyles(s1, s2)
	if len(merged.Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(merged.Rules))
	}
	if len(merged.Recipes) != 2 {
		t.Errorf("expected 2 recipes, got %d", len(merged.Recipes))
	}
}

func TestMergeStylesNil(t *testing.T) {
	s := &Styles{Rules: []StyleRule{{Func: &Func{Name: realizeStrPtr("a")}}}}

	if MergeStyles(nil, s) != s {
		t.Error("MergeStyles(nil, s) should return s")
	}
	if MergeStyles(s, nil) != s {
		t.Error("MergeStyles(s, nil) should return s")
	}
}

// ----------------------------------------------------------------------------
// GetStyleProperty Tests
// ----------------------------------------------------------------------------

func TestGetStyleProperty(t *testing.T) {
	funcName := "text"
	styles := &Styles{
		Rules: []StyleRule{
			{
				Func: &Func{Name: &funcName},
				Args: &Args{
					Items: []Arg{
						{Name: realizeStrPtr("fill"), Value: syntax.Spanned[Value]{V: Str("red")}},
					},
				},
			},
		},
	}

	val := GetStyleProperty(styles, "text", "fill")
	if val == nil {
		t.Fatal("expected to find 'fill' property")
	}
	if s, ok := val.(StrValue); !ok || string(s) != "red" {
		t.Errorf("expected 'red', got %v", val)
	}

	// Property not found
	val2 := GetStyleProperty(styles, "text", "size")
	if val2 != nil {
		t.Errorf("expected nil for missing property, got %v", val2)
	}

	// Function not found
	val3 := GetStyleProperty(styles, "heading", "fill")
	if val3 != nil {
		t.Errorf("expected nil for missing function, got %v", val3)
	}
}

// Helper function for realize tests
func realizeStrPtr(s string) *string {
	return &s
}
