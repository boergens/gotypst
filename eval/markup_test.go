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
