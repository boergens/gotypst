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
		{"parbreak", &eval.ParbreakElement{}, StateDestructive},
		{"linebreak", &eval.LinebreakElement{}, StateDestructive},
		{"heading", &eval.HeadingElement{Level: 1}, StateDestructive},
		{"paragraph", &eval.ParagraphElement{}, StateDestructive},
		{"strong", &eval.StrongElement{}, StateSupportive},
		{"emph", &eval.EmphElement{}, StateSupportive},
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
	tests := []struct {
		name     string
		pairs    []Pair
		expected int // expected count of elements
	}{
		{
			name:     "empty",
			pairs:    nil,
			expected: 0,
		},
		{
			name: "no spaces",
			pairs: []Pair{
				{Element: &eval.TextElement{Text: "hello"}},
				{Element: &eval.TextElement{Text: "world"}},
			},
			expected: 2,
		},
		{
			name: "leading space",
			pairs: []Pair{
				{Element: &eval.TextElement{Text: " "}},
				{Element: &eval.TextElement{Text: "hello"}},
			},
			expected: 1, // leading space removed
		},
		{
			name: "trailing space",
			pairs: []Pair{
				{Element: &eval.TextElement{Text: "hello"}},
				{Element: &eval.TextElement{Text: " "}},
			},
			expected: 1, // trailing space removed
		},
		{
			name: "collapse multiple spaces",
			pairs: []Pair{
				{Element: &eval.TextElement{Text: "hello"}},
				{Element: &eval.TextElement{Text: " "}},
				{Element: &eval.TextElement{Text: " "}},
				{Element: &eval.TextElement{Text: "world"}},
			},
			expected: 3, // collapsed to: hello, single space, world
		},
		{
			name: "space before destructive",
			pairs: []Pair{
				{Element: &eval.TextElement{Text: "hello"}},
				{Element: &eval.TextElement{Text: " "}},
				{Element: &eval.ParbreakElement{}},
			},
			expected: 2, // space removed before parbreak
		},
		{
			name: "space after destructive",
			pairs: []Pair{
				{Element: &eval.ParbreakElement{}},
				{Element: &eval.TextElement{Text: " "}},
				{Element: &eval.TextElement{Text: "hello"}},
			},
			expected: 2, // space removed after parbreak
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collapseSpaces(tt.pairs)
			if len(result) != tt.expected {
				t.Errorf("collapseSpaces() returned %d elements, want %d", len(result), tt.expected)
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
