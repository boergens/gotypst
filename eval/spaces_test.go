package eval

import (
	"reflect"
	"testing"
)

func TestGetSpaceState(t *testing.T) {
	tests := []struct {
		name     string
		elem     ContentElement
		expected SpaceState
	}{
		// Space elements
		{"SpaceElement", &SpaceElement{}, Space},
		{"WeakSpaceElement", &SpaceElement{Weak: true}, Space},

		// Destructive elements
		{"ParbreakElement", &ParbreakElement{}, Destructive},
		{"HeadingElement", &HeadingElement{Level: 1}, Destructive},
		{"ListItemElement", &ListItemElement{}, Destructive},
		{"EnumItemElement", &EnumItemElement{}, Destructive},
		{"TermItemElement", &TermItemElement{}, Destructive},
		{"ParagraphElement", &ParagraphElement{}, Destructive},
		{"LinebreakElement", &LinebreakElement{}, Destructive},

		// Supportive elements
		{"TextElement", &TextElement{Text: "hello"}, Supportive},
		{"RawElement", &RawElement{Text: "code"}, Supportive},
		{"StrongElement", &StrongElement{}, Supportive},
		{"EmphElement", &EmphElement{}, Supportive},
		{"LinkElement", &LinkElement{URL: "https://example.com"}, Supportive},
		{"RefElement", &RefElement{Target: "label"}, Supportive},
		{"SmartQuoteElement", &SmartQuoteElement{Double: true}, Supportive},
		{"EquationElement", &EquationElement{}, Supportive},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := GetSpaceState(tc.elem)
			if got != tc.expected {
				t.Errorf("GetSpaceState(%T) = %v, want %v", tc.elem, got, tc.expected)
			}
		})
	}
}

func TestCollapseSpaces_EmptyContent(t *testing.T) {
	elements := []ContentElement{}
	n := CollapseSpaces(elements)
	if n != 0 {
		t.Errorf("CollapseSpaces(empty) returned %d, want 0", n)
	}
}

func TestCollapseSpaces_NoSpaces(t *testing.T) {
	elements := []ContentElement{
		&TextElement{Text: "hello"},
		&TextElement{Text: "world"},
	}
	n := CollapseSpaces(elements)
	if n != 2 {
		t.Errorf("CollapseSpaces returned %d, want 2", n)
	}
	// Elements should be unchanged
	if _, ok := elements[0].(*TextElement); !ok {
		t.Error("elements[0] should be TextElement")
	}
	if _, ok := elements[1].(*TextElement); !ok {
		t.Error("elements[1] should be TextElement")
	}
}

func TestCollapseSpaces_SingleSpace(t *testing.T) {
	elements := []ContentElement{
		&TextElement{Text: "hello"},
		&SpaceElement{},
		&TextElement{Text: "world"},
	}
	n := CollapseSpaces(elements)
	if n != 3 {
		t.Errorf("CollapseSpaces returned %d, want 3", n)
	}
	// Space between two text elements should be preserved
	if _, ok := elements[1].(*SpaceElement); !ok {
		t.Errorf("elements[1] should be SpaceElement, got %T", elements[1])
	}
}

func TestCollapseSpaces_ConsecutiveSpaces(t *testing.T) {
	// Multiple consecutive spaces should collapse into one
	elements := []ContentElement{
		&TextElement{Text: "hello"},
		&SpaceElement{},
		&SpaceElement{},
		&SpaceElement{},
		&TextElement{Text: "world"},
	}
	n := CollapseSpaces(elements)
	if n != 3 {
		t.Errorf("CollapseSpaces returned %d, want 3 (text + space + text)", n)
	}
	if te, ok := elements[0].(*TextElement); !ok || te.Text != "hello" {
		t.Error("elements[0] should be 'hello'")
	}
	if _, ok := elements[1].(*SpaceElement); !ok {
		t.Error("elements[1] should be SpaceElement")
	}
	if te, ok := elements[2].(*TextElement); !ok || te.Text != "world" {
		t.Error("elements[2] should be 'world'")
	}
}

func TestCollapseSpaces_LeadingSpaces(t *testing.T) {
	// Spaces at the start should be removed
	elements := []ContentElement{
		&SpaceElement{},
		&SpaceElement{},
		&TextElement{Text: "hello"},
	}
	n := CollapseSpaces(elements)
	if n != 1 {
		t.Errorf("CollapseSpaces returned %d, want 1", n)
	}
	if te, ok := elements[0].(*TextElement); !ok || te.Text != "hello" {
		t.Errorf("elements[0] should be 'hello', got %T", elements[0])
	}
}

func TestCollapseSpaces_TrailingSpaces(t *testing.T) {
	// Spaces at the end should be removed
	elements := []ContentElement{
		&TextElement{Text: "hello"},
		&SpaceElement{},
		&SpaceElement{},
	}
	n := CollapseSpaces(elements)
	if n != 1 {
		t.Errorf("CollapseSpaces returned %d, want 1", n)
	}
	if te, ok := elements[0].(*TextElement); !ok || te.Text != "hello" {
		t.Errorf("elements[0] should be 'hello', got %T", elements[0])
	}
}

func TestCollapseSpaces_SpacesBeforeDestructive(t *testing.T) {
	// Spaces before destructive elements should be removed
	elements := []ContentElement{
		&TextElement{Text: "hello"},
		&SpaceElement{},
		&ParbreakElement{},
		&TextElement{Text: "world"},
	}
	n := CollapseSpaces(elements)
	if n != 3 {
		t.Errorf("CollapseSpaces returned %d, want 3", n)
	}
	if _, ok := elements[0].(*TextElement); !ok {
		t.Error("elements[0] should be TextElement")
	}
	if _, ok := elements[1].(*ParbreakElement); !ok {
		t.Error("elements[1] should be ParbreakElement")
	}
	if _, ok := elements[2].(*TextElement); !ok {
		t.Error("elements[2] should be TextElement")
	}
}

func TestCollapseSpaces_SpacesAfterDestructive(t *testing.T) {
	// Spaces after destructive elements should be removed
	elements := []ContentElement{
		&ParbreakElement{},
		&SpaceElement{},
		&TextElement{Text: "hello"},
	}
	n := CollapseSpaces(elements)
	if n != 2 {
		t.Errorf("CollapseSpaces returned %d, want 2", n)
	}
	if _, ok := elements[0].(*ParbreakElement); !ok {
		t.Error("elements[0] should be ParbreakElement")
	}
	if _, ok := elements[1].(*TextElement); !ok {
		t.Error("elements[1] should be TextElement")
	}
}

func TestCollapseSpaces_MultipleDestructive(t *testing.T) {
	// Multiple destructive elements with spaces between
	elements := []ContentElement{
		&HeadingElement{Level: 1},
		&SpaceElement{},
		&SpaceElement{},
		&ParbreakElement{},
		&SpaceElement{},
		&TextElement{Text: "content"},
	}
	n := CollapseSpaces(elements)
	if n != 3 {
		t.Errorf("CollapseSpaces returned %d, want 3", n)
	}
	if _, ok := elements[0].(*HeadingElement); !ok {
		t.Error("elements[0] should be HeadingElement")
	}
	if _, ok := elements[1].(*ParbreakElement); !ok {
		t.Error("elements[1] should be ParbreakElement")
	}
	if _, ok := elements[2].(*TextElement); !ok {
		t.Error("elements[2] should be TextElement")
	}
}

func TestCollapseSpaces_OnlySpaces(t *testing.T) {
	// Content with only spaces should collapse to nothing
	elements := []ContentElement{
		&SpaceElement{},
		&SpaceElement{},
		&SpaceElement{},
	}
	n := CollapseSpaces(elements)
	if n != 0 {
		t.Errorf("CollapseSpaces returned %d, want 0", n)
	}
}

func TestCollapseSpaces_PreservesSpaceBetweenSupportive(t *testing.T) {
	// Complex case: spaces should be preserved between supportive elements
	elements := []ContentElement{
		&TextElement{Text: "a"},
		&SpaceElement{},
		&StrongElement{},
		&SpaceElement{},
		&TextElement{Text: "b"},
	}
	n := CollapseSpaces(elements)
	if n != 5 {
		t.Errorf("CollapseSpaces returned %d, want 5", n)
	}
}

func TestCollapseSpacesContent(t *testing.T) {
	// Test the Content wrapper
	original := Content{
		Elements: []ContentElement{
			&SpaceElement{},
			&TextElement{Text: "hello"},
			&SpaceElement{},
			&SpaceElement{},
			&TextElement{Text: "world"},
			&SpaceElement{},
		},
	}

	result := CollapseSpacesContent(original)

	// Original should be unchanged
	if len(original.Elements) != 6 {
		t.Error("Original content was modified")
	}

	// Result should have collapsed spaces
	if len(result.Elements) != 3 {
		t.Errorf("CollapseSpacesContent returned %d elements, want 3", len(result.Elements))
	}

	// Check the elements
	if te, ok := result.Elements[0].(*TextElement); !ok || te.Text != "hello" {
		t.Error("result.Elements[0] should be 'hello'")
	}
	if _, ok := result.Elements[1].(*SpaceElement); !ok {
		t.Error("result.Elements[1] should be SpaceElement")
	}
	if te, ok := result.Elements[2].(*TextElement); !ok || te.Text != "world" {
		t.Error("result.Elements[2] should be 'world'")
	}
}

func TestSpaceStateString(t *testing.T) {
	// Verify we handle all SpaceState values in GetSpaceState
	states := []SpaceState{Invisible, Destructive, Supportive, Space}
	for _, s := range states {
		if s < 0 || s > Space {
			t.Errorf("Unexpected SpaceState value: %v", s)
		}
	}
}

func TestCollapseSpaces_RealWorldExample(t *testing.T) {
	// Simulate: "Hello world\n\nNew paragraph"
	// This would be: Text("Hello") Space Text("world") Parbreak Text("New") Space Text("paragraph")
	elements := []ContentElement{
		&TextElement{Text: "Hello"},
		&SpaceElement{},
		&TextElement{Text: "world"},
		&ParbreakElement{},
		&TextElement{Text: "New"},
		&SpaceElement{},
		&TextElement{Text: "paragraph"},
	}

	n := CollapseSpaces(elements)

	// Expected: Text Space Text Parbreak Text Space Text (7 elements)
	if n != 7 {
		t.Errorf("CollapseSpaces returned %d, want 7", n)
	}

	expected := []string{"TextElement", "SpaceElement", "TextElement", "ParbreakElement", "TextElement", "SpaceElement", "TextElement"}
	for i := 0; i < n; i++ {
		typeName := reflect.TypeOf(elements[i]).Elem().Name()
		if typeName != expected[i] {
			t.Errorf("elements[%d] type = %s, want %s", i, typeName, expected[i])
		}
	}
}

func TestCollapseSpaces_HeadingWithSpaces(t *testing.T) {
	// Heading with spaces around it
	elements := []ContentElement{
		&SpaceElement{},
		&HeadingElement{Level: 1, Content: Content{Elements: []ContentElement{&TextElement{Text: "Title"}}}},
		&SpaceElement{},
		&TextElement{Text: "Body text"},
	}

	n := CollapseSpaces(elements)

	// Spaces before and after heading should be removed
	if n != 2 {
		t.Errorf("CollapseSpaces returned %d, want 2", n)
	}
	if _, ok := elements[0].(*HeadingElement); !ok {
		t.Error("elements[0] should be HeadingElement")
	}
	if _, ok := elements[1].(*TextElement); !ok {
		t.Error("elements[1] should be TextElement")
	}
}
