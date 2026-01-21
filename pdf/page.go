package pdf

import (
	"github.com/boergens/gotypst/layout/pages"
)

// NumberingStyle represents the PDF page label numbering style.
// Maps to PDF's /S entry in page label dictionaries.
type NumberingStyle int

const (
	// StyleArabic represents decimal Arabic numerals (1, 2, 3, ...).
	StyleArabic NumberingStyle = iota
	// StyleLowerRoman represents lowercase Roman numerals (i, ii, iii, ...).
	StyleLowerRoman
	// StyleUpperRoman represents uppercase Roman numerals (I, II, III, ...).
	StyleUpperRoman
	// StyleLowerAlpha represents lowercase letters (a, b, c, ...).
	StyleLowerAlpha
	// StyleUpperAlpha represents uppercase letters (A, B, C, ...).
	StyleUpperAlpha
)

// String returns the PDF name for the numbering style.
func (s NumberingStyle) String() string {
	switch s {
	case StyleArabic:
		return "D"
	case StyleLowerRoman:
		return "r"
	case StyleUpperRoman:
		return "R"
	case StyleLowerAlpha:
		return "a"
	case StyleUpperAlpha:
		return "A"
	default:
		return "D"
	}
}

// PageLabel represents a PDF page label.
// Page labels provide logical page numbering that can differ from
// physical page indices.
type PageLabel struct {
	// Style is the numbering style for this page label range.
	// If nil, no numeric portion is used (prefix only).
	Style *NumberingStyle
	// Prefix is an optional string prefix for the page label.
	Prefix *string
	// Offset is the starting number for this range.
	// If nil, defaults to 1 in PDF.
	Offset *uint32
}

// NewPageLabel creates a new PageLabel with the given parameters.
func NewPageLabel(style *NumberingStyle, prefix *string, offset *uint32) PageLabel {
	return PageLabel{
		Style:  style,
		Prefix: prefix,
		Offset: offset,
	}
}

// GeneratePageLabel creates a PageLabel from a Numbering applied to a page number.
// Returns nil if the numbering pattern cannot be converted to a PDF page label.
func GeneratePageLabel(numbering *pages.Numbering, number uint64) *PageLabel {
	if numbering == nil || numbering.Pattern == "" {
		return nil
	}

	pattern := numbering.Pattern

	// Parse the pattern to extract kind and determine style.
	// The pattern uses format specifiers: "1" (Arabic), "i"/"I" (Roman), "a"/"A" (alpha)
	prefix, kind := parseNumberingPattern(pattern)

	// Determine the style based on the kind character.
	// If the pattern has a suffix (text after the format specifier),
	// we cannot use the common style optimization since PDF does not
	// provide a suffix field.
	hasSuffix := hasSuffixAfterKind(pattern, kind)
	var style *NumberingStyle
	if !hasSuffix && kind != 0 {
		s := kindToStyle(kind, number)
		style = s
	}

	// Prefix and offset depend on the style:
	// If style is supported by PDF, we use the given prefix and an offset.
	// Otherwise, the entire formatted string goes into the prefix.
	var prefixStr *string
	var offset *uint32

	if style == nil {
		// Format the full page label as prefix
		formatted := formatWithPattern(pattern, number)
		prefixStr = &formatted
	} else {
		if prefix != "" {
			prefixStr = &prefix
		}
		if number > 0 && number <= 0xFFFFFFFF {
			n := uint32(number)
			offset = &n
		}
	}

	return &PageLabel{
		Style:  style,
		Prefix: prefixStr,
		Offset: offset,
	}
}

// ArabicPageLabel creates an Arabic page label with the specified page number.
// For example, this will display page label "11" when given page number 11.
func ArabicPageLabel(number uint64) PageLabel {
	style := StyleArabic
	var offset *uint32
	if number > 0 && number <= 0xFFFFFFFF {
		n := uint32(number)
		offset = &n
	}
	return PageLabel{
		Style:  &style,
		Prefix: nil,
		Offset: offset,
	}
}

// NumberingKind represents the type of numbering pattern.
type NumberingKind rune

const (
	KindArabic     NumberingKind = '1'
	KindLowerRoman NumberingKind = 'i'
	KindUpperRoman NumberingKind = 'I'
	KindLowerLatin NumberingKind = 'a'
	KindUpperLatin NumberingKind = 'A'
)

// parseNumberingPattern extracts the prefix and kind from a numbering pattern.
// For example, "Page 1" returns ("Page ", '1').
func parseNumberingPattern(pattern string) (prefix string, kind NumberingKind) {
	// Find the first format specifier
	for i, r := range pattern {
		switch r {
		case '1', 'i', 'I', 'a', 'A':
			return pattern[:i], NumberingKind(r)
		}
	}
	// No format specifier found
	return pattern, 0
}

// hasSuffixAfterKind checks if there's text after the format specifier.
func hasSuffixAfterKind(pattern string, kind NumberingKind) bool {
	if kind == 0 {
		return false
	}
	// Find where the kind character is
	for i, r := range pattern {
		if NumberingKind(r) == kind {
			// Check if there's anything after it
			return i+1 < len(pattern)
		}
	}
	return false
}

// kindToStyle converts a NumberingKind to a PDF NumberingStyle.
// Returns nil for kinds that cannot be represented as PDF styles
// (e.g., lowercase/uppercase latin for page numbers > 26).
func kindToStyle(kind NumberingKind, number uint64) *NumberingStyle {
	var s NumberingStyle
	switch kind {
	case KindArabic:
		s = StyleArabic
	case KindLowerRoman:
		s = StyleLowerRoman
	case KindUpperRoman:
		s = StyleUpperRoman
	case KindLowerLatin:
		// PDF only supports single letters (a-z), so limit to 26
		if number > 26 {
			return nil
		}
		s = StyleLowerAlpha
	case KindUpperLatin:
		// PDF only supports single letters (A-Z), so limit to 26
		if number > 26 {
			return nil
		}
		s = StyleUpperAlpha
	default:
		return nil
	}
	return &s
}

// formatWithPattern formats a page number according to the given pattern.
// This is used when the pattern cannot be represented as a PDF style.
func formatWithPattern(pattern string, number uint64) string {
	// Use the same formatting logic as layout/pages/finalize.go
	// For simplicity, delegate to the formatPageNumber equivalent here
	switch pattern {
	case "", "1":
		return formatArabic(number)
	case "i":
		return formatRomanLower(number)
	case "I":
		return formatRomanUpper(number)
	case "a":
		return formatLetterLower(number)
	case "A":
		return formatLetterUpper(number)
	}

	// Handle patterns with surrounding text
	result := pattern
	for i, r := range pattern {
		switch r {
		case '1':
			return pattern[:i] + formatArabic(number) + pattern[i+1:]
		case 'i':
			return pattern[:i] + formatRomanLower(number) + pattern[i+1:]
		case 'I':
			return pattern[:i] + formatRomanUpper(number) + pattern[i+1:]
		case 'a':
			return pattern[:i] + formatLetterLower(number) + pattern[i+1:]
		case 'A':
			return pattern[:i] + formatLetterUpper(number) + pattern[i+1:]
		}
	}

	return result
}

// formatArabic formats a number as Arabic numerals.
func formatArabic(n uint64) string {
	if n == 0 {
		return "0"
	}
	var result []byte
	for n > 0 {
		result = append([]byte{byte('0' + n%10)}, result...)
		n /= 10
	}
	return string(result)
}

// formatRomanUpper formats a number as uppercase Roman numerals.
func formatRomanUpper(n uint64) string {
	if n <= 0 || n >= 4000 {
		return formatArabic(n)
	}

	values := []uint64{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	symbols := []string{"M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"}

	var result []byte
	for i, val := range values {
		for n >= val {
			result = append(result, symbols[i]...)
			n -= val
		}
	}
	return string(result)
}

// formatRomanLower formats a number as lowercase Roman numerals.
func formatRomanLower(n uint64) string {
	if n <= 0 || n >= 4000 {
		return formatArabic(n)
	}

	values := []uint64{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	symbols := []string{"m", "cm", "d", "cd", "c", "xc", "l", "xl", "x", "ix", "v", "iv", "i"}

	var result []byte
	for i, val := range values {
		for n >= val {
			result = append(result, symbols[i]...)
			n -= val
		}
	}
	return string(result)
}

// formatLetterUpper formats a number as uppercase letters (A, B, ..., Z, AA, AB, ...).
func formatLetterUpper(n uint64) string {
	if n == 0 {
		return "0"
	}

	var result []byte
	for n > 0 {
		n-- // Make 1-indexed into 0-indexed
		result = append([]byte{byte('A' + n%26)}, result...)
		n /= 26
	}
	return string(result)
}

// formatLetterLower formats a number as lowercase letters (a, b, ..., z, aa, ab, ...).
func formatLetterLower(n uint64) string {
	if n == 0 {
		return "0"
	}

	var result []byte
	for n > 0 {
		n-- // Make 1-indexed into 0-indexed
		result = append([]byte{byte('a' + n%26)}, result...)
		n /= 26
	}
	return string(result)
}

// ToDict converts a PageLabel to a PDF dictionary for use in the page labels tree.
func (p PageLabel) ToDict() Dict {
	dict := make(Dict)

	if p.Style != nil {
		dict[Name("S")] = Name(p.Style.String())
	}
	if p.Prefix != nil {
		dict[Name("P")] = String(*p.Prefix)
	}
	if p.Offset != nil {
		dict[Name("St")] = Int(int(*p.Offset))
	}

	return dict
}
