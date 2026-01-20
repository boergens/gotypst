package eval

import (
	"testing"
)

func TestShorthandToSymbol(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"~", "\u00A0"},         // Non-breaking space
		{"---", "\u2014"},       // Em dash
		{"--", "\u2013"},        // En dash
		{"-?", "\u00AD"},        // Soft hyphen
		{"...", "\u2026"},       // Horizontal ellipsis
		{"-1", "\u22121"},       // Minus sign + digit
		{"-42", "\u221242"},     // Minus sign + digits
		{"other", "other"},      // Unknown shorthand passes through
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := shorthandToSymbol(tt.input)
			if result != tt.expected {
				t.Errorf("shorthandToSymbol(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseEscapeSequence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`\n`, "n"},                  // Simple escape
		{`\*`, "*"},                  // Escaped asterisk
		{`\\`, "\\"},                 // Escaped backslash
		{`\u{0041}`, "A"},            // Unicode escape for 'A'
		{`\u{00A0}`, "\u00A0"},       // Unicode escape for non-breaking space
		{`\u{2014}`, "\u2014"},       // Unicode escape for em dash
		{`\u{1F600}`, "\U0001F600"},  // Unicode escape for emoji
		{"nobackslash", "nobackslash"}, // Pass through if no backslash
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseEscapeSequence(tt.input)
			if result != tt.expected {
				t.Errorf("parseEscapeSequence(%q) = %q (len=%d), want %q (len=%d)",
					tt.input, result, len(result), tt.expected, len(tt.expected))
			}
		})
	}
}

func TestSmartQuoteElement(t *testing.T) {
	// Test that SmartQuoteElement properly tracks quote type
	doubleQuote := &SmartQuoteElement{Double: true}
	singleQuote := &SmartQuoteElement{Double: false}

	if !doubleQuote.Double {
		t.Error("Double quote element should have Double=true")
	}
	if singleQuote.Double {
		t.Error("Single quote element should have Double=false")
	}

	// Verify they satisfy ContentElement interface
	var _ ContentElement = doubleQuote
	var _ ContentElement = singleQuote
}

func TestRefElement(t *testing.T) {
	// Test RefElement with no supplement
	ref := &RefElement{Target: "my-label", Supplement: nil}
	if ref.Target != "my-label" {
		t.Errorf("RefElement target = %q, want %q", ref.Target, "my-label")
	}
	if ref.Supplement != nil {
		t.Error("RefElement supplement should be nil")
	}

	// Test RefElement with supplement
	supp := &Content{
		Elements: []ContentElement{&TextElement{Text: "supplement text"}},
	}
	refWithSupp := &RefElement{Target: "another-label", Supplement: supp}
	if refWithSupp.Supplement == nil {
		t.Error("RefElement supplement should not be nil")
	}
	if len(refWithSupp.Supplement.Elements) != 1 {
		t.Errorf("RefElement supplement should have 1 element, got %d", len(refWithSupp.Supplement.Elements))
	}

	// Verify it satisfies ContentElement interface
	var _ ContentElement = ref
}

func TestRawElement(t *testing.T) {
	// Test basic RawElement creation
	raw := &RawElement{
		Text:  "print('hello')",
		Lang:  "python",
		Block: true,
	}
	if raw.Text != "print('hello')" {
		t.Errorf("RawElement text = %q, want %q", raw.Text, "print('hello')")
	}
	if raw.Lang != "python" {
		t.Errorf("RawElement lang = %q, want %q", raw.Lang, "python")
	}
	if !raw.Block {
		t.Error("RawElement block should be true")
	}

	// Test inline raw
	inline := &RawElement{
		Text:  "code",
		Lang:  "",
		Block: false,
	}
	if inline.Lang != "" {
		t.Errorf("RawElement inline lang = %q, want empty", inline.Lang)
	}
	if inline.Block {
		t.Error("RawElement inline should not be block")
	}

	// Test with multiline text
	multiline := &RawElement{
		Text:  "line1\nline2\nline3",
		Lang:  "go",
		Block: true,
	}
	if multiline.Text != "line1\nline2\nline3" {
		t.Errorf("RawElement multiline text = %q, want %q", multiline.Text, "line1\nline2\nline3")
	}

	// Verify it satisfies ContentElement interface
	var _ ContentElement = raw
	var _ ContentElement = inline
	var _ ContentElement = multiline
}

func TestRawElementInContent(t *testing.T) {
	// Test RawElement as part of Content
	content := Content{
		Elements: []ContentElement{
			&TextElement{Text: "Here is some code: "},
			&RawElement{Text: "fn main() {}", Lang: "rust", Block: false},
			&TextElement{Text: " and more text."},
		},
	}

	if len(content.Elements) != 3 {
		t.Fatalf("Content should have 3 elements, got %d", len(content.Elements))
	}

	// Verify the middle element is a RawElement
	raw, ok := content.Elements[1].(*RawElement)
	if !ok {
		t.Fatalf("Middle element should be *RawElement, got %T", content.Elements[1])
	}
	if raw.Text != "fn main() {}" {
		t.Errorf("RawElement text = %q, want %q", raw.Text, "fn main() {}")
	}
	if raw.Lang != "rust" {
		t.Errorf("RawElement lang = %q, want %q", raw.Lang, "rust")
	}
}
