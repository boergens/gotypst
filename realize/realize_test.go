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
			&eval.HeadingElement{Depth: 1, Content: eval.Content{
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
				&eval.HeadingElement{Depth: 1},
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

func TestStyleChainGetWithDefault(t *testing.T) {
	funcName := "text"
	propName := "size"
	args := &eval.Args{
		Items: []eval.Arg{
			{Name: &propName, Value: syntax.SpannedDetached[eval.Value](eval.FloatValue(12.0))},
		},
	}

	styles := &eval.Styles{
		Rules: []eval.StyleRule{
			{Func: &eval.Func{Name: &funcName}, Args: args},
		},
	}

	chain := NewStyleChain(styles, nil)

	// Test GetWithDefault for existing property
	val := chain.GetWithDefault("text", "size", eval.FloatValue(10.0))
	if f, ok := val.(eval.FloatValue); !ok || float64(f) != 12.0 {
		t.Errorf("expected 12.0, got %v", val)
	}

	// Test GetWithDefault for missing property
	val = chain.GetWithDefault("text", "fill", eval.StrValue("black"))
	if s, ok := val.(eval.StrValue); !ok || string(s) != "black" {
		t.Errorf("expected 'black', got %v", val)
	}
}

func TestStyleChainGetProperty(t *testing.T) {
	funcName := "text"
	prop1 := "size"
	prop2 := "size"

	parentArgs := &eval.Args{
		Items: []eval.Arg{
			{Name: &prop1, Value: syntax.SpannedDetached[eval.Value](eval.FloatValue(10.0))},
		},
	}
	childArgs := &eval.Args{
		Items: []eval.Arg{
			{Name: &prop2, Value: syntax.SpannedDetached[eval.Value](eval.FloatValue(14.0))},
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

	// GetProperty should return both values (child first)
	values := child.GetProperty("text", "size")
	if len(values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(values))
	}

	// First value should be child's (14.0)
	if f, ok := values[0].(eval.FloatValue); !ok || float64(f) != 14.0 {
		t.Errorf("expected first value 14.0, got %v", values[0])
	}

	// Second value should be parent's (10.0)
	if f, ok := values[1].(eval.FloatValue); !ok || float64(f) != 10.0 {
		t.Errorf("expected second value 10.0, got %v", values[1])
	}
}

func TestStyleChainIsEmpty(t *testing.T) {
	// Empty chain
	empty := EmptyStyleChain()
	if !empty.IsEmpty() {
		t.Error("expected empty chain to be empty")
	}

	// Non-empty chain
	funcName := "text"
	propName := "size"
	args := &eval.Args{
		Items: []eval.Arg{
			{Name: &propName, Value: syntax.SpannedDetached[eval.Value](eval.FloatValue(12.0))},
		},
	}
	styles := &eval.Styles{
		Rules: []eval.StyleRule{
			{Func: &eval.Func{Name: &funcName}, Args: args},
		},
	}
	chain := NewStyleChain(styles, nil)
	if chain.IsEmpty() {
		t.Error("expected non-empty chain to not be empty")
	}

	// Nil chain
	var nilChain *StyleChain
	if !nilChain.IsEmpty() {
		t.Error("expected nil chain to be empty")
	}
}

func TestStyleChainDepth(t *testing.T) {
	// Nil chain has depth 0
	var nilChain *StyleChain
	if nilChain.Depth() != 0 {
		t.Errorf("expected nil chain depth 0, got %d", nilChain.Depth())
	}

	// Single chain has depth 1
	chain1 := EmptyStyleChain()
	if chain1.Depth() != 1 {
		t.Errorf("expected single chain depth 1, got %d", chain1.Depth())
	}

	// Chained has depth 2
	chain2 := chain1.Chain(&eval.Styles{
		Rules: []eval.StyleRule{},
	})
	// Note: Chain returns parent if styles are empty
	// So we need actual content
	funcName := "text"
	propName := "size"
	chain3 := chain1.Chain(&eval.Styles{
		Rules: []eval.StyleRule{
			{Func: &eval.Func{Name: &funcName}, Args: &eval.Args{
				Items: []eval.Arg{
					{Name: &propName, Value: syntax.SpannedDetached[eval.Value](eval.FloatValue(12.0))},
				},
			}},
		},
	})
	if chain3.Depth() != 2 {
		t.Errorf("expected chained depth 2, got %d", chain3.Depth())
	}

	// Verify chain2 returns parent when empty
	if chain2 != chain1 {
		t.Error("expected Chain to return parent when styles are empty")
	}
}

func TestStyleChainGetAllRules(t *testing.T) {
	funcName := "text"
	otherFunc := "par"
	sizeProp := "size"
	fillProp := "fill"

	parentArgs := &eval.Args{
		Items: []eval.Arg{
			{Name: &sizeProp, Value: syntax.SpannedDetached[eval.Value](eval.FloatValue(10.0))},
		},
	}
	childArgs := &eval.Args{
		Items: []eval.Arg{
			{Name: &fillProp, Value: syntax.SpannedDetached[eval.Value](eval.StrValue("red"))},
		},
	}
	otherArgs := &eval.Args{
		Items: []eval.Arg{
			{Name: &sizeProp, Value: syntax.SpannedDetached[eval.Value](eval.FloatValue(5.0))},
		},
	}

	parentStyles := &eval.Styles{
		Rules: []eval.StyleRule{
			{Func: &eval.Func{Name: &funcName}, Args: parentArgs},
			{Func: &eval.Func{Name: &otherFunc}, Args: otherArgs},
		},
	}
	childStyles := &eval.Styles{
		Rules: []eval.StyleRule{
			{Func: &eval.Func{Name: &funcName}, Args: childArgs},
		},
	}

	parent := NewStyleChain(parentStyles, nil)
	child := parent.Chain(childStyles)

	// GetAllRules for "text" should return both rules
	rules := child.GetAllRules("text")
	if len(rules) != 2 {
		t.Fatalf("expected 2 text rules, got %d", len(rules))
	}

	// GetAllRules for "par" should return 1 rule
	rules = child.GetAllRules("par")
	if len(rules) != 1 {
		t.Fatalf("expected 1 par rule, got %d", len(rules))
	}

	// GetAllRules for unknown function should return 0
	rules = child.GetAllRules("unknown")
	if len(rules) != 0 {
		t.Fatalf("expected 0 rules for unknown, got %d", len(rules))
	}
}

func TestStyleChainFold(t *testing.T) {
	funcName := "text"
	sizeProp := "size"

	// Create a chain with multiple size values
	parent := NewStyleChain(&eval.Styles{
		Rules: []eval.StyleRule{
			{Func: &eval.Func{Name: &funcName}, Args: &eval.Args{
				Items: []eval.Arg{
					{Name: &sizeProp, Value: syntax.SpannedDetached[eval.Value](eval.FloatValue(10.0))},
				},
			}},
		},
	}, nil)

	child := parent.Chain(&eval.Styles{
		Rules: []eval.StyleRule{
			{Func: &eval.Func{Name: &funcName}, Args: &eval.Args{
				Items: []eval.Arg{
					{Name: &sizeProp, Value: syntax.SpannedDetached[eval.Value](eval.FloatValue(14.0))},
				},
			}},
		},
	})

	// Fold with override (take newest)
	override := func(acc, val eval.Value) eval.Value {
		return val
	}
	result := child.Fold("text", "size", eval.FloatValue(8.0), override)
	if f, ok := result.(eval.FloatValue); !ok || float64(f) != 14.0 {
		t.Errorf("expected fold result 14.0 (newest), got %v", result)
	}

	// Fold with addition
	addFloats := func(acc, val eval.Value) eval.Value {
		a, aok := acc.(eval.FloatValue)
		v, vok := val.(eval.FloatValue)
		if aok && vok {
			return eval.FloatValue(float64(a) + float64(v))
		}
		return acc
	}
	result = child.Fold("text", "size", eval.FloatValue(0), addFloats)
	if f, ok := result.(eval.FloatValue); !ok || float64(f) != 24.0 {
		t.Errorf("expected fold result 24.0 (10+14), got %v", result)
	}

	// Fold with no values should return initial
	result = child.Fold("text", "nonexistent", eval.FloatValue(5.0), override)
	if f, ok := result.(eval.FloatValue); !ok || float64(f) != 5.0 {
		t.Errorf("expected initial value 5.0, got %v", result)
	}
}

func TestStyleChainGetRecipesFor(t *testing.T) {
	// Create a style chain with recipes
	headingElement := eval.Element{Name: "heading"}
	textElement := eval.Element{Name: "text"}

	headingSelector := eval.ElemSelector{Element: headingElement}
	textSelector := eval.ElemSelector{Element: textElement}

	headingSel := eval.Selector(headingSelector)
	textSel := eval.Selector(textSelector)

	styles := &eval.Styles{
		Recipes: []*eval.Recipe{
			{Selector: &headingSel, Transform: eval.NoneTransformation{}},
			{Selector: &textSel, Transform: eval.NoneTransformation{}},
			{Selector: nil, Transform: eval.NoneTransformation{}}, // Global recipe
		},
	}

	chain := NewStyleChain(styles, nil)

	// Get recipes for heading
	headingRecipes := chain.GetRecipesFor(&eval.HeadingElement{})
	// Should match: heading selector + global (nil selector)
	if len(headingRecipes) != 2 {
		t.Errorf("expected 2 recipes for heading, got %d", len(headingRecipes))
	}

	// Get recipes for text
	textRecipes := chain.GetRecipesFor(&eval.TextElement{})
	// Should match: text selector + global (nil selector)
	if len(textRecipes) != 2 {
		t.Errorf("expected 2 recipes for text, got %d", len(textRecipes))
	}

	// Get recipes for nil element
	nilRecipes := chain.GetRecipesFor(nil)
	if len(nilRecipes) != 0 {
		t.Errorf("expected 0 recipes for nil element, got %d", len(nilRecipes))
	}
}
