package realize

import (
	"testing"

	"github.com/boergens/gotypst/eval"
)

func TestGetSpaceState(t *testing.T) {
	tests := []struct {
		name     string
		element  eval.ContentElement
		expected SpaceState
	}{
		{"nil", nil, StateInvisible},
		{"text", &eval.TextElement{Text: "hello"}, StateSupportive},
		{"whitespace text", &eval.TextElement{Text: "   "}, StateSpace},
		{"space element", &eval.SpaceElement{}, StateSpace},
		{"parbreak", &eval.ParbreakElement{}, StateDestructive},
		{"linebreak", &eval.LinebreakElement{}, StateDestructive},
		{"heading", &eval.HeadingElement{Depth: 1}, StateDestructive},
		{"paragraph", &eval.ParagraphElement{}, StateDestructive},
		{"list", &eval.ListElement{}, StateDestructive},
		{"block", &eval.BlockElement{}, StateDestructive},
		{"v element", &eval.VElem{}, StateDestructive},
		{"tag", &eval.TagElem{}, StateInvisible},
		{"strong", &eval.StrongElement{}, StateSupportive},
		{"emph", &eval.EmphElement{}, StateSupportive},
		{"link", &eval.LinkElement{}, StateSupportive},
		{"h element", &eval.HElem{}, StateSupportive},
		{"box", &eval.BoxElement{}, StateSupportive},
		{"equation", &eval.EquationElement{}, StateSupportive},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSpaceState(tt.element)
			if got != tt.expected {
				t.Errorf("getSpaceState() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsWhitespaceOnly(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{" ", true},
		{"  ", true},
		{"\t", true},
		{"\n", true},
		{" \t\n\r", true},
		{"hello", false},
		{" hello", false},
		{"hello ", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isWhitespaceOnly(tt.input)
			if got != tt.expected {
				t.Errorf("isWhitespaceOnly(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCollapseSpaces(t *testing.T) {
	// Note: collapseSpaces modifies in-place and takes a start index
	// It's used internally by the grouping system

	tests := []struct {
		name   string
		pairs  []Pair
		start  int
		verify func([]Pair) bool
	}{
		{
			name:  "empty",
			pairs: nil,
			start: 0,
			verify: func(pairs []Pair) bool {
				return len(pairs) == 0
			},
		},
		{
			name: "no spaces",
			pairs: []Pair{
				{Content: &eval.TextElement{Text: "hello"}},
				{Content: &eval.TextElement{Text: "world"}},
			},
			start: 0,
			verify: func(pairs []Pair) bool {
				// Should have both elements
				count := 0
				for _, p := range pairs {
					if p.Content != nil {
						count++
					}
				}
				return count == 2
			},
		},
		{
			name: "collapse leading space",
			pairs: []Pair{
				{Content: &eval.SpaceElement{}},
				{Content: &eval.TextElement{Text: "hello"}},
			},
			start: 0,
			verify: func(pairs []Pair) bool {
				// Leading space should be collapsed
				// First non-nil should be TextElement
				for _, p := range pairs {
					if p.Content != nil {
						_, ok := p.Content.(*eval.TextElement)
						return ok
					}
				}
				return false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collapseSpaces(tt.pairs, tt.start)
			if !tt.verify(tt.pairs) {
				t.Errorf("collapseSpaces() verification failed")
			}
		})
	}
}

func TestNormalizeSpaces(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"hello", "hello"},
		{"hello world", "hello world"},
		{"hello  world", "hello world"},
		{"hello   world", "hello world"},
		{" hello", "hello"},
		{"hello ", "hello"},
		{" hello ", "hello"},
		{"hello\nworld", "hello world"},
		{"hello\t\n\rworld", "hello world"},
		{"  hello  world  ", "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeSpaces(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeSpaces(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTrimLeadingWhitespace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"hello", "hello"},
		{" hello", "hello"},
		{"  hello", "hello"},
		{"\thello", "hello"},
		{" \t\nhello", "hello"},
		{"   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := trimLeadingWhitespace(tt.input)
			if got != tt.expected {
				t.Errorf("trimLeadingWhitespace(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTrimTrailingWhitespace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"hello", "hello"},
		{"hello ", "hello"},
		{"hello  ", "hello"},
		{"hello\t", "hello"},
		{"hello \t\n", "hello"},
		{"   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := trimTrailingWhitespace(tt.input)
			if got != tt.expected {
				t.Errorf("trimTrailingWhitespace(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTrimLeadingSpace(t *testing.T) {
	tests := []struct {
		name     string
		pairs    []Pair
		expected int // expected count after trimming
	}{
		{
			name:     "empty",
			pairs:    []Pair{},
			expected: 0,
		},
		{
			name: "no leading space",
			pairs: []Pair{
				{Content: &eval.TextElement{Text: "hello"}},
			},
			expected: 1,
		},
		{
			name: "with leading space element",
			pairs: []Pair{
				{Content: &eval.SpaceElement{}},
				{Content: &eval.TextElement{Text: "hello"}},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trimLeadingSpace(tt.pairs)
			if len(result) != tt.expected {
				t.Errorf("trimLeadingSpace() returned %d elements, want %d", len(result), tt.expected)
			}
		})
	}
}

func TestTrimTrailingSpace(t *testing.T) {
	tests := []struct {
		name     string
		pairs    []Pair
		expected int // expected count after trimming
	}{
		{
			name:     "empty",
			pairs:    []Pair{},
			expected: 0,
		},
		{
			name: "no trailing space",
			pairs: []Pair{
				{Content: &eval.TextElement{Text: "hello"}},
			},
			expected: 1,
		},
		{
			name: "with trailing space element",
			pairs: []Pair{
				{Content: &eval.TextElement{Text: "hello"}},
				{Content: &eval.SpaceElement{}},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trimTrailingSpace(tt.pairs)
			if len(result) != tt.expected {
				t.Errorf("trimTrailingSpace() returned %d elements, want %d", len(result), tt.expected)
			}
		})
	}
}
