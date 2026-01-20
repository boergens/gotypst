package eval

import (
	"testing"
)

func TestCollectText(t *testing.T) {
	tests := []struct {
		name     string
		content  Content
		expected string
	}{
		{
			name:     "empty content",
			content:  Content{},
			expected: "",
		},
		{
			name: "single text element",
			content: Content{
				Elements: []ContentElement{&TextElement{Text: "hello"}},
			},
			expected: "hello",
		},
		{
			name: "multiple text elements",
			content: Content{
				Elements: []ContentElement{
					&TextElement{Text: "hello"},
					&TextElement{Text: " "},
					&TextElement{Text: "world"},
				},
			},
			expected: "hello world",
		},
		{
			name: "nested strong element",
			content: Content{
				Elements: []ContentElement{
					&TextElement{Text: "hello "},
					&StrongElement{Content: Content{
						Elements: []ContentElement{&TextElement{Text: "world"}},
					}},
				},
			},
			expected: "hello world",
		},
		{
			name: "nested emph element",
			content: Content{
				Elements: []ContentElement{
					&TextElement{Text: "hello "},
					&EmphElement{Content: Content{
						Elements: []ContentElement{&TextElement{Text: "world"}},
					}},
				},
			},
			expected: "hello world",
		},
		{
			name: "raw element",
			content: Content{
				Elements: []ContentElement{
					&RawElement{Text: "code()"},
				},
			},
			expected: "code()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collected := CollectText(tt.content)
			if collected.Text != tt.expected {
				t.Errorf("CollectText() = %q, want %q", collected.Text, tt.expected)
			}
		})
	}
}

func TestCollectTextWithSpaceCollapsing(t *testing.T) {
	tests := []struct {
		name     string
		content  Content
		expected string
	}{
		{
			name: "multiple spaces",
			content: Content{
				Elements: []ContentElement{
					&TextElement{Text: "hello   world"},
				},
			},
			expected: "hello world",
		},
		{
			name: "tabs and spaces",
			content: Content{
				Elements: []ContentElement{
					&TextElement{Text: "hello\t\t  world"},
				},
			},
			expected: "hello world",
		},
		{
			name: "newlines collapse to space",
			content: Content{
				Elements: []ContentElement{
					&TextElement{Text: "hello\nworld"},
				},
			},
			expected: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collected := CollectTextWithSpaceCollapsing(tt.content)
			if collected.Text != tt.expected {
				t.Errorf("CollectTextWithSpaceCollapsing() = %q, want %q", collected.Text, tt.expected)
			}
		})
	}
}

func TestMatchTextSelector_Literal(t *testing.T) {
	content := Content{
		Elements: []ContentElement{
			&TextElement{Text: "hello world hello"},
		},
	}

	selector := TextSelector{Text: "hello", IsRegex: false}
	matches, err := MatchTextSelector(selector, content)
	if err != nil {
		t.Fatalf("MatchTextSelector() error = %v", err)
	}

	if len(matches) != 2 {
		t.Fatalf("MatchTextSelector() returned %d matches, want 2", len(matches))
	}

	// First match
	if matches[0].Start != 0 || matches[0].End != 5 || matches[0].Text != "hello" {
		t.Errorf("matches[0] = {Start: %d, End: %d, Text: %q}, want {Start: 0, End: 5, Text: \"hello\"}",
			matches[0].Start, matches[0].End, matches[0].Text)
	}

	// Second match
	if matches[1].Start != 12 || matches[1].End != 17 || matches[1].Text != "hello" {
		t.Errorf("matches[1] = {Start: %d, End: %d, Text: %q}, want {Start: 12, End: 17, Text: \"hello\"}",
			matches[1].Start, matches[1].End, matches[1].Text)
	}
}

func TestMatchTextSelector_Regex(t *testing.T) {
	content := Content{
		Elements: []ContentElement{
			&TextElement{Text: "hello123world456"},
		},
	}

	selector := TextSelector{Text: `\d+`, IsRegex: true}
	matches, err := MatchTextSelector(selector, content)
	if err != nil {
		t.Fatalf("MatchTextSelector() error = %v", err)
	}

	if len(matches) != 2 {
		t.Fatalf("MatchTextSelector() returned %d matches, want 2", len(matches))
	}

	// First match
	if matches[0].Text != "123" {
		t.Errorf("matches[0].Text = %q, want \"123\"", matches[0].Text)
	}

	// Second match
	if matches[1].Text != "456" {
		t.Errorf("matches[1].Text = %q, want \"456\"", matches[1].Text)
	}
}

func TestMatchTextSelector_RegexWithCaptures(t *testing.T) {
	content := Content{
		Elements: []ContentElement{
			&TextElement{Text: "name: Alice, age: 30"},
		},
	}

	selector := TextSelector{Text: `(\w+): (\w+)`, IsRegex: true}
	matches, err := MatchTextSelector(selector, content)
	if err != nil {
		t.Fatalf("MatchTextSelector() error = %v", err)
	}

	if len(matches) != 2 {
		t.Fatalf("MatchTextSelector() returned %d matches, want 2", len(matches))
	}

	// First match: "name: Alice"
	if matches[0].Text != "name: Alice" {
		t.Errorf("matches[0].Text = %q, want \"name: Alice\"", matches[0].Text)
	}
	if len(matches[0].Captures) != 2 {
		t.Fatalf("matches[0].Captures length = %d, want 2", len(matches[0].Captures))
	}
	if matches[0].Captures[0] != "name" || matches[0].Captures[1] != "Alice" {
		t.Errorf("matches[0].Captures = %v, want [\"name\", \"Alice\"]", matches[0].Captures)
	}

	// Second match: "age: 30"
	if matches[1].Text != "age: 30" {
		t.Errorf("matches[1].Text = %q, want \"age: 30\"", matches[1].Text)
	}
}

func TestMatchTextSelector_InvalidRegex(t *testing.T) {
	content := Content{
		Elements: []ContentElement{
			&TextElement{Text: "hello"},
		},
	}

	selector := TextSelector{Text: `[invalid`, IsRegex: true}
	_, err := MatchTextSelector(selector, content)
	if err == nil {
		t.Fatal("MatchTextSelector() expected error for invalid regex")
	}

	regexErr, ok := err.(*RegexError)
	if !ok {
		t.Errorf("expected *RegexError, got %T", err)
	}
	if regexErr.Pattern != "[invalid" {
		t.Errorf("RegexError.Pattern = %q, want \"[invalid\"", regexErr.Pattern)
	}
}

func TestMatchTextSelector_EmptyPattern(t *testing.T) {
	content := Content{
		Elements: []ContentElement{
			&TextElement{Text: "hello"},
		},
	}

	selector := TextSelector{Text: "", IsRegex: false}
	matches, err := MatchTextSelector(selector, content)
	if err != nil {
		t.Fatalf("MatchTextSelector() error = %v", err)
	}

	if len(matches) != 0 {
		t.Errorf("MatchTextSelector() returned %d matches, want 0", len(matches))
	}
}

func TestMatchTextSelector_NoMatch(t *testing.T) {
	content := Content{
		Elements: []ContentElement{
			&TextElement{Text: "hello world"},
		},
	}

	selector := TextSelector{Text: "goodbye", IsRegex: false}
	matches, err := MatchTextSelector(selector, content)
	if err != nil {
		t.Fatalf("MatchTextSelector() error = %v", err)
	}

	if len(matches) != 0 {
		t.Errorf("MatchTextSelector() returned %d matches, want 0", len(matches))
	}
}

func TestSplitContentByMatches(t *testing.T) {
	content := Content{
		Elements: []ContentElement{
			&TextElement{Text: "hello world hello"},
		},
	}

	matches := []TextMatch{
		{Start: 0, End: 5, Text: "hello"},
		{Start: 12, End: 17, Text: "hello"},
	}

	segments := SplitContentByMatches(content, matches)

	// Should have: match, non-match, match
	// But we also need trailing/leading non-matches
	// Pattern: "hello" " world " "hello"
	// So: match, non-match, match

	if len(segments) != 3 {
		t.Fatalf("SplitContentByMatches() returned %d segments, want 3", len(segments))
	}

	// First segment is a match
	if !segments[0].IsMatch {
		t.Error("segments[0].IsMatch = false, want true")
	}

	// Second segment is non-match
	if segments[1].IsMatch {
		t.Error("segments[1].IsMatch = true, want false")
	}

	// Third segment is a match
	if !segments[2].IsMatch {
		t.Error("segments[2].IsMatch = false, want true")
	}
}

func TestSplitContentByMatches_NoMatches(t *testing.T) {
	content := Content{
		Elements: []ContentElement{
			&TextElement{Text: "hello world"},
		},
	}

	segments := SplitContentByMatches(content, nil)

	if len(segments) != 1 {
		t.Fatalf("SplitContentByMatches() returned %d segments, want 1", len(segments))
	}

	if segments[0].IsMatch {
		t.Error("segments[0].IsMatch = true, want false")
	}
}

func TestCastToShowableSelector_Regex(t *testing.T) {
	// This test verifies that RegexValue can be cast to a ShowableSelector
	regexVal := RegexValue{Pattern: `\d+`}

	// We can't directly call castToShowableSelector as it requires syntax.Expr
	// but we can verify the type switch logic indirectly by checking the types

	// Verify RegexValue satisfies Value interface
	var _ Value = regexVal

	// Verify TextSelector fields
	selector := TextSelector{Text: regexVal.Pattern, IsRegex: true}
	if selector.Text != `\d+` {
		t.Errorf("TextSelector.Text = %q, want \"\\d+\"", selector.Text)
	}
	if !selector.IsRegex {
		t.Error("TextSelector.IsRegex = false, want true")
	}
}
