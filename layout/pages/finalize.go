package pages

import (
	"strconv"
	"strings"

	"github.com/boergens/gotypst/layout"
)

// Finalize pieces together the inner page frame and the marginals.
// We can only do this at the very end because inside/outside margins
// require knowledge of the physical page number, which is unknown
// during parallel layout.
func Finalize(engine *Engine, counter *ManualPageCounter, tags *[]Tag, layouted LayoutedPage) (*Page, error) {
	margin := layouted.Margin

	// If two-sided, left becomes inside and right becomes outside.
	// Thus, for left-bound pages, we want to swap on even pages and
	// for right-bound pages, we want to swap on odd pages.
	if layouted.TwoSided && layouted.Binding.Swap(counter.Physical()) {
		margin.Left, margin.Right = margin.Right, margin.Left
	}

	// Create a frame for the full page
	fullSize := layout.Size{
		Width:  layouted.Inner.Size.Width + margin.Left + margin.Right,
		Height: layouted.Inner.Size.Height + margin.Top + margin.Bottom,
	}
	frame := Hard(fullSize)

	// Add tags
	for _, tag := range *tags {
		frame.Push(layout.Point{X: 0, Y: 0}, TagItem{Tag: tag})
	}
	*tags = (*tags)[:0] // Clear the tags slice

	// Add the "before" marginals. The order in which we push things here is
	// important as it affects the relative ordering of introspectable elements
	// and thus how counters resolve.
	if layouted.Background != nil {
		frame.PushFrame(layout.Point{X: 0, Y: 0}, *layouted.Background)
	}
	if layouted.Header != nil {
		frame.PushFrame(layout.Point{X: margin.Left, Y: 0}, *layouted.Header)
	}

	// Add the inner contents
	frame.PushFrame(layout.Point{X: margin.Left, Y: margin.Top}, layouted.Inner)

	// Apply counter updates from within the page to the manual page counter
	// We do this before generating page numbers so any counter.set() in the
	// content takes effect.
	if err := counter.Visit(&layouted.Inner); err != nil {
		return nil, err
	}

	// Get this page's number (before stepping to next page)
	number := counter.Logical()

	// Add the "after" marginals
	if layouted.Footer != nil {
		y := fullSize.Height - layouted.Footer.Size.Height
		frame.PushFrame(layout.Point{X: margin.Left, Y: y}, *layouted.Footer)
	} else if layouted.Numbering != nil {
		// Create page number in footer when numbering is set but no explicit footer
		numStr := formatPageNumber(number, layouted.Numbering.Pattern)
		footerSize := layout.Size{
			Width:  layouted.Inner.Size.Width,
			Height: margin.Bottom,
		}
		footerFrame := createPageNumberFrame(numStr, footerSize)
		y := fullSize.Height - footerFrame.Size.Height
		frame.PushFrame(layout.Point{X: margin.Left, Y: y}, footerFrame)
	}
	if layouted.Foreground != nil {
		frame.PushFrame(layout.Point{X: 0, Y: 0}, *layouted.Foreground)
	}

	// Step to the next page
	counter.Step()

	return &Page{
		Frame:      frame,
		Fill:       layouted.Fill,
		Numbering:  layouted.Numbering,
		Supplement: layouted.Supplement,
		Number:     number,
	}, nil
}

// formatPageNumber formats a page number according to the given pattern.
// Supported patterns:
//   - "1" or "" - Arabic numerals (1, 2, 3, ...)
//   - "i" - lowercase Roman numerals (i, ii, iii, iv, ...)
//   - "I" - uppercase Roman numerals (I, II, III, IV, ...)
//   - "a" - lowercase letters (a, b, c, ..., aa, ab, ...)
//   - "A" - uppercase letters (A, B, C, ..., AA, AB, ...)
//
// The pattern may also include surrounding text, where the numeral placeholder
// is replaced. For example, "- 1 -" becomes "- 5 -" for page 5.
func formatPageNumber(pageNum int, pattern string) string {
	if pattern == "" || pattern == "1" {
		return strconv.Itoa(pageNum)
	}

	// Check for specific format patterns
	switch pattern {
	case "i":
		return toRomanLower(pageNum)
	case "I":
		return toRomanUpper(pageNum)
	case "a":
		return toLetterLower(pageNum)
	case "A":
		return toLetterUpper(pageNum)
	}

	// For patterns with surrounding text, find and replace the format specifier
	// E.g., "Page 1 of N" -> "Page 5 of N"
	if strings.Contains(pattern, "1") {
		return strings.ReplaceAll(pattern, "1", strconv.Itoa(pageNum))
	}
	if strings.Contains(pattern, "i") {
		return strings.ReplaceAll(pattern, "i", toRomanLower(pageNum))
	}
	if strings.Contains(pattern, "I") {
		return strings.ReplaceAll(pattern, "I", toRomanUpper(pageNum))
	}
	if strings.Contains(pattern, "a") {
		return strings.ReplaceAll(pattern, "a", toLetterLower(pageNum))
	}
	if strings.Contains(pattern, "A") {
		return strings.ReplaceAll(pattern, "A", toLetterUpper(pageNum))
	}

	// Default to Arabic numerals
	return strconv.Itoa(pageNum)
}

// toRomanUpper converts a number to uppercase Roman numerals.
func toRomanUpper(n int) string {
	if n <= 0 || n >= 4000 {
		return strconv.Itoa(n) // Fallback for out of range
	}

	values := []int{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	symbols := []string{"M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"}

	var result strings.Builder
	for i, val := range values {
		for n >= val {
			result.WriteString(symbols[i])
			n -= val
		}
	}
	return result.String()
}

// toRomanLower converts a number to lowercase Roman numerals.
func toRomanLower(n int) string {
	return strings.ToLower(toRomanUpper(n))
}

// toLetterUpper converts a number to uppercase letters (A, B, ..., Z, AA, AB, ...).
func toLetterUpper(n int) string {
	if n <= 0 {
		return strconv.Itoa(n)
	}

	var result strings.Builder
	for n > 0 {
		n-- // Make 1-indexed into 0-indexed
		result.WriteByte(byte('A' + n%26))
		n /= 26
	}

	// Reverse the string
	s := result.String()
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// toLetterLower converts a number to lowercase letters.
func toLetterLower(n int) string {
	return strings.ToLower(toLetterUpper(n))
}

// createPageNumberFrame creates a frame with centered page number text.
func createPageNumberFrame(numStr string, size layout.Size) Frame {
	frame := Hard(size)

	// Default font size for page numbers
	fontSize := layout.Abs(10)
	lineHeight := fontSize * 1.2

	// Position text centered horizontally and vertically aligned to bottom
	// In a real implementation, we'd measure the text width for proper centering
	x := size.Width / 2 // Approximate center (should subtract half text width)
	y := size.Height - lineHeight

	frame.Push(layout.Point{X: x, Y: y}, TextItem{
		Text:     numStr,
		FontSize: fontSize,
	})

	return frame
}
