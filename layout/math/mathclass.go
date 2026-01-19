package math

// Class represents the mathematical class of a symbol.
// This is based on the Unicode Math Class property.
type Class int

const (
	// Normal represents ordinary symbols.
	Normal Class = iota
	// Alphabetic represents alphabetic symbols.
	Alphabetic
	// Binary represents binary operators (+, -, ×).
	Binary
	// Closing represents closing delimiters.
	Closing
	// Diacritic represents combining diacritics.
	Diacritic
	// Fence represents fence characters.
	Fence
	// GlyphPart represents glyph parts.
	GlyphPart
	// Large represents large operators (∑, ∫).
	Large
	// Opening represents opening delimiters.
	Opening
	// Punctuation represents punctuation.
	Punctuation
	// Relation represents relation operators (=, <, >).
	Relation
	// Space represents space characters.
	Space
	// Special represents special characters.
	Special
	// Unary represents unary operators.
	Unary
	// Vary represents characters that vary by context.
	Vary
)

// String returns a string representation of the class.
func (c Class) String() string {
	switch c {
	case Normal:
		return "Normal"
	case Alphabetic:
		return "Alphabetic"
	case Binary:
		return "Binary"
	case Closing:
		return "Closing"
	case Diacritic:
		return "Diacritic"
	case Fence:
		return "Fence"
	case GlyphPart:
		return "GlyphPart"
	case Large:
		return "Large"
	case Opening:
		return "Opening"
	case Punctuation:
		return "Punctuation"
	case Relation:
		return "Relation"
	case Space:
		return "Space"
	case Special:
		return "Special"
	case Unary:
		return "Unary"
	case Vary:
		return "Vary"
	default:
		return "Unknown"
	}
}

// GetMathClass returns the math class for a Unicode character.
func GetMathClass(r rune) Class {
	// Basic ASCII classification
	switch {
	case r >= '0' && r <= '9':
		return Normal
	case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
		return Alphabetic
	case r == '+', r == '-', r == '*', r == '/', r == '·', r == '×', r == '÷':
		return Binary
	case r == '=', r == '<', r == '>', r == '≤', r == '≥', r == '≠', r == '≈':
		return Relation
	case r == '(', r == '[', r == '{', r == '⟨':
		return Opening
	case r == ')', r == ']', r == '}', r == '⟩':
		return Closing
	case r == ',', r == '.', r == ';', r == ':':
		return Punctuation
	case r == ' ', r == '\t', r == '\n':
		return Space
	case r == '∑', r == '∏', r == '∫', r == '∮':
		return Large
	}

	// Greek letters
	if (r >= 'α' && r <= 'ω') || (r >= 'Α' && r <= 'Ω') {
		return Alphabetic
	}

	// Extended math classification
	switch {
	// Arrows
	case r >= 0x2190 && r <= 0x21FF:
		return Relation
	// Mathematical operators
	case r >= 0x2200 && r <= 0x22FF:
		return classifyMathOperator(r)
	// Miscellaneous technical
	case r >= 0x2300 && r <= 0x23FF:
		return Normal
	// Supplemental mathematical operators
	case r >= 0x2A00 && r <= 0x2AFF:
		return classifySupplementalMathOp(r)
	}

	return Normal
}

// classifyMathOperator classifies characters in the Mathematical Operators block.
func classifyMathOperator(r rune) Class {
	switch r {
	// Quantifiers and logical symbols
	case '∀', '∃', '∄', '∴', '∵', '∎':
		return Normal
	// Set operations
	case '∈', '∉', '∋', '∌', '⊂', '⊃', '⊆', '⊇', '⊈', '⊉':
		return Relation
	case '∩', '∪', '∖', '△':
		return Binary
	// Relations
	case '∼', '≃', '≅', '≈', '≊', '≋':
		return Relation
	case '≪', '≫', '≺', '≻', '≼', '≽':
		return Relation
	// Large operators
	case '∑', '∏', '∐', '∫', '∬', '∭', '∮', '∯', '∰':
		return Large
	// Binary operators
	case '∓', '∔', '⊕', '⊖', '⊗', '⊘', '⊙', '⊚', '⊛', '⊜', '⊝':
		return Binary
	case '⊞', '⊟', '⊠', '⊡':
		return Binary
	case '⋅', '⋆', '⋇', '⋈':
		return Binary
	}
	return Normal
}

// classifySupplementalMathOp classifies supplemental mathematical operators.
func classifySupplementalMathOp(r rune) Class {
	// Most are binary operators
	return Binary
}
