package syntax

import (
	"unicode"

	"golang.org/x/text/unicode/runenames"
)

// IsNewline returns true if the character is a newline character.
func IsNewline(c rune) bool {
	switch c {
	// Line Feed, Vertical Tab, Form Feed, Carriage Return.
	case '\n', '\x0B', '\x0C', '\r':
		return true
	// Next Line, Line Separator, Paragraph Separator.
	case '\u0085', '\u2028', '\u2029':
		return true
	}
	return false
}

// IsSpace returns true if the character is whitespace in the given mode.
func IsSpace(c rune, mode SyntaxMode) bool {
	switch mode {
	case ModeMarkup:
		return c == ' ' || c == '\t' || IsNewline(c)
	default:
		return unicode.IsSpace(c)
	}
}

// IsIDStart returns true if the character can start an identifier.
// This uses Unicode XID_Start plus underscore.
func IsIDStart(c rune) bool {
	return unicode.Is(unicode.L, c) || // Letters
		unicode.Is(unicode.Nl, c) || // Letter numbers
		c == '_'
}

// IsIDContinue returns true if the character can continue an identifier.
// This uses Unicode XID_Continue plus underscore and hyphen.
func IsIDContinue(c rune) bool {
	return unicode.Is(unicode.L, c) || // Letters
		unicode.Is(unicode.Nl, c) || // Letter numbers
		unicode.Is(unicode.Mn, c) || // Nonspacing marks
		unicode.Is(unicode.Mc, c) || // Spacing combining marks
		unicode.Is(unicode.Nd, c) || // Decimal digits
		unicode.Is(unicode.Pc, c) || // Connector punctuation
		c == '_' || c == '-'
}

// IsMathIDStart returns true if the character can start a math identifier.
func IsMathIDStart(c rune) bool {
	// Math identifiers are more restrictive - no underscore
	return unicode.Is(unicode.L, c) || unicode.Is(unicode.Nl, c)
}

// IsMathIDContinue returns true if the character can continue a math identifier.
func IsMathIDContinue(c rune) bool {
	// Math identifiers exclude underscore
	return unicode.Is(unicode.L, c) ||
		unicode.Is(unicode.Nl, c) ||
		unicode.Is(unicode.Mn, c) ||
		unicode.Is(unicode.Mc, c) ||
		unicode.Is(unicode.Nd, c) ||
		unicode.Is(unicode.Pc, c)
}

// IsValidInLabelLiteral returns true if the character can be part of a label literal.
func IsValidInLabelLiteral(c rune) bool {
	return IsIDContinue(c) || c == ':' || c == '.'
}

// IsIdent returns true if the string is a valid Typst identifier.
func IsIdent(s string) bool {
	if len(s) == 0 {
		return false
	}
	runes := []rune(s)
	if !IsIDStart(runes[0]) {
		return false
	}
	for _, r := range runes[1:] {
		if !IsIDContinue(r) {
			return false
		}
	}
	return true
}

// IsValidLabelLiteralID returns true if the string is valid in a label literal.
func IsValidLabelLiteralID(id string) bool {
	if len(id) == 0 {
		return false
	}
	for _, r := range id {
		if !IsValidInLabelLiteral(r) {
			return false
		}
	}
	return true
}

// MathClass represents the Unicode math class of a character.
type MathClass int

const (
	MathClassNone MathClass = iota
	MathClassNormal
	MathClassAlphabetic
	MathClassBinary
	MathClassClosing
	MathClassDiacritic
	MathClassFence
	MathClassGlyphPart
	MathClassLarge
	MathClassOpening
	MathClassPunctuation
	MathClassRelation
	MathClassSpace
	MathClassUnary
	MathClassVary
	MathClassSpecial
)

// DefaultMathClass returns the default math class for a Unicode character.
// This is a simplified implementation - in production, this would use a proper
// Unicode math classification table.
func DefaultMathClass(c rune) MathClass {
	// Opening brackets
	if c == '(' || c == '[' || c == '{' ||
		c == '\u2308' || // Left ceiling
		c == '\u230A' || // Left floor
		c == '\u2329' || // Left angle bracket
		c == '\u27E8' || // Mathematical left angle bracket
		c == '\u27EA' || // Mathematical left double angle bracket
		c == '\u27EC' || // Mathematical left white tortoise shell bracket
		c == '\u27EE' || // Mathematical left flattened parenthesis
		c == '\u2983' || // Left white curly bracket
		c == '\u2985' || // Left white parenthesis
		c == '\u2987' || // Z notation left image bracket
		c == '\u2989' || // Z notation left binding bracket
		c == '\u298B' || // Left square bracket with underbar
		c == '\u298D' || // Left square bracket with tick in top corner
		c == '\u298F' || // Left square bracket with tick in bottom corner
		c == '\u2991' || // Left angle bracket with dot
		c == '\u2993' || // Left arc less-than bracket
		c == '\u2995' || // Double left arc greater-than bracket
		c == '\u2997' || // Left black tortoise shell bracket
		c == '\u29FC' { // Left-pointing curved angle bracket
		return MathClassOpening
	}

	// Closing brackets
	if c == ')' || c == ']' || c == '}' ||
		c == '\u2309' || // Right ceiling
		c == '\u230B' || // Right floor
		c == '\u232A' || // Right angle bracket
		c == '\u27E9' || // Mathematical right angle bracket
		c == '\u27EB' || // Mathematical right double angle bracket
		c == '\u27ED' || // Mathematical right white tortoise shell bracket
		c == '\u27EF' || // Mathematical right flattened parenthesis
		c == '\u2984' || // Right white curly bracket
		c == '\u2986' || // Right white parenthesis
		c == '\u2988' || // Z notation right image bracket
		c == '\u298A' || // Z notation right binding bracket
		c == '\u298C' || // Right square bracket with underbar
		c == '\u298E' || // Right square bracket with tick in top corner
		c == '\u2990' || // Right square bracket with tick in bottom corner
		c == '\u2992' || // Right angle bracket with dot
		c == '\u2994' || // Right arc greater-than bracket
		c == '\u2996' || // Double right arc less-than bracket
		c == '\u2998' || // Right black tortoise shell bracket
		c == '\u29FD' { // Right-pointing curved angle bracket
		return MathClassClosing
	}

	return MathClassNone
}

// Script represents a Unicode script.
type Script int

const (
	ScriptUnknown Script = iota
	ScriptLatin
	ScriptHan
	ScriptHiragana
	ScriptKatakana
	ScriptHangul
	// ... other scripts as needed
)

// GetScript returns the Unicode script for a character.
// This is a simplified implementation.
func GetScript(c rune) Script {
	name := runenames.Name(c)
	// Simple heuristic based on character ranges
	switch {
	case c >= 0x4E00 && c <= 0x9FFF: // CJK Unified Ideographs
		return ScriptHan
	case c >= 0x3040 && c <= 0x309F: // Hiragana
		return ScriptHiragana
	case c >= 0x30A0 && c <= 0x30FF: // Katakana
		return ScriptKatakana
	case c >= 0xAC00 && c <= 0xD7AF: // Hangul Syllables
		return ScriptHangul
	case (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z'):
		return ScriptLatin
	default:
		_ = name // suppress unused
		return ScriptUnknown
	}
}
