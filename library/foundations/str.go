package foundations

import (
	"unicode/utf8"

	"github.com/rivo/uniseg"
)

// Str inspection methods
// Translated from foundations/str.rs

// graphemeClusters returns all grapheme clusters in a string.
// This is the fundamental unit of string indexing in Typst.
func graphemeClusters(s string) []string {
	var clusters []string
	gr := uniseg.NewGraphemes(s)
	for gr.Next() {
		clusters = append(clusters, gr.Str())
	}
	return clusters
}

// StrLen returns the number of grapheme clusters in the string.
// This is what Typst considers the "length" of a string.
func StrLen(s Str) Int {
	count := 0
	gr := uniseg.NewGraphemes(string(s))
	for gr.Next() {
		count++
	}
	return Int(count)
}

// StrIsEmpty returns true if the string contains no grapheme clusters.
func StrIsEmpty(s Str) Bool {
	gr := uniseg.NewGraphemes(string(s))
	return Bool(!gr.Next())
}

// StrFirst returns the first grapheme cluster of the string.
// Returns None if the string is empty.
func StrFirst(s Str) Value {
	gr := uniseg.NewGraphemes(string(s))
	if !gr.Next() {
		return None
	}
	return Str(gr.Str())
}

// StrLast returns the last grapheme cluster of the string.
// Returns None if the string is empty.
func StrLast(s Str) Value {
	clusters := graphemeClusters(string(s))
	if len(clusters) == 0 {
		return None
	}
	return Str(clusters[len(clusters)-1])
}

// normalizeIndex converts a possibly-negative index to a positive index.
// Returns the normalized index and whether it's valid.
func normalizeIndex(index int64, length int) (int, bool) {
	idx := int(index)
	if idx < 0 {
		idx = length + idx
	}
	if idx < 0 || idx >= length {
		return 0, false
	}
	return idx, true
}

// StrAt returns the grapheme cluster at the given index.
// Supports negative indices (counting from the end).
// Returns an error if the index is out of bounds.
func StrAt(s Str, index Int) (Value, error) {
	clusters := graphemeClusters(string(s))
	idx, ok := normalizeIndex(int64(index), len(clusters))
	if !ok {
		return nil, &OpError{
			Message: "string index out of bounds",
			Hint:    "index is " + Int(index).String() + ", but string has " + Int(len(clusters)).String() + " grapheme cluster(s)",
		}
	}
	return Str(clusters[idx]), nil
}

// StrSlice returns a substring from start to end.
// Both indices support negative values (counting from the end).
// If end is nil, slices to the end of the string.
// Returns an error if indices are out of bounds.
func StrSlice(s Str, start Int, end *Int) (Value, error) {
	clusters := graphemeClusters(string(s))
	length := len(clusters)

	// Normalize start index
	startIdx := int(start)
	if startIdx < 0 {
		startIdx = length + startIdx
	}

	// Normalize end index
	var endIdx int
	if end == nil {
		endIdx = length
	} else {
		endIdx = int(*end)
		if endIdx < 0 {
			endIdx = length + endIdx
		}
	}

	// Clamp indices to valid range
	if startIdx < 0 {
		startIdx = 0
	}
	if endIdx > length {
		endIdx = length
	}

	// Handle inverted range
	if startIdx > endIdx {
		return Str(""), nil
	}

	// Build result string
	result := ""
	for i := startIdx; i < endIdx; i++ {
		result += clusters[i]
	}
	return Str(result), nil
}

// StrClusters returns an array of all grapheme clusters in the string.
func StrClusters(s Str) *Array {
	clusters := graphemeClusters(string(s))
	items := make([]Value, len(clusters))
	for i, c := range clusters {
		items[i] = Str(c)
	}
	return NewArray(items...)
}

// StrCodepoints returns an array of all Unicode codepoints in the string.
// Each codepoint is returned as a single-character string.
func StrCodepoints(s Str) *Array {
	str := string(s)
	items := make([]Value, 0, utf8.RuneCountInString(str))
	for _, r := range str {
		items = append(items, Str(string(r)))
	}
	return NewArray(items...)
}
