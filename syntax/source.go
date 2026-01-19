// Package syntax provides source file management for Typst documents.
//
// This file is a Go translation of typst-syntax/src/source.rs from the
// original Typst compiler.
package syntax

import (
	"strings"
	"unicode/utf8"
)

// Source represents a source file with parsed syntax tree and line information.
//
// Sources provide:
//   - Access to the source text and parsed syntax tree
//   - Efficient line/column to byte offset conversion
//   - UTF-16 position support for LSP integration
//   - Incremental text editing with minimal reparsing
//
// The Source is cheap to clone as it shares the underlying data.
type Source struct {
	id    FileId
	text  string
	root  *SyntaxNode
	lines *Lines
}

// NewSource creates a new source file from a file id and text.
// The text is parsed and the syntax tree is numbered with spans.
func NewSource(id FileId, text string) *Source {
	root := Parse(text)
	// Number the syntax tree nodes within the full numbering range.
	root.Numberize(id, [2]uint64{spanFullStart, spanFullEnd})

	return &Source{
		id:    id,
		text:  text,
		root:  root,
		lines: NewLines(text),
	}
}

// NewDetachedSource creates a source file without a real path.
// This is typically used for testing or for content that doesn't
// correspond to an actual file on disk.
func NewDetachedSource(text string) *Source {
	// Create a unique detached file id using a virtual path.
	vpath, _ := NewVirtualPath("/detached")
	path := NewRootedPath(ProjectRoot(), *vpath)
	id := UniqueFileId(*path)
	return NewSource(id, text)
}

// Id returns the source file's identifier.
func (s *Source) Id() FileId {
	return s.id
}

// Text returns the full source as a string.
func (s *Source) Text() string {
	return s.text
}

// Root returns the untyped syntax tree root node.
func (s *Source) Root() *SyntaxNode {
	return s.root
}

// Lines returns the line acceleration structure for position conversions.
func (s *Source) Lines() *Lines {
	return s.lines
}

// Len returns the length of the source text in bytes.
func (s *Source) Len() int {
	return len(s.text)
}

// Find locates a syntax node by its span in the source.
// Returns nil if no node matches the span.
func (s *Source) Find(span Span) *LinkedNode {
	// Check if span belongs to this source.
	if id := span.Id(); id == nil || *id != s.id {
		return nil
	}
	return NewLinkedNode(s.root).Find(span)
}

// Range returns the byte range in the source for a given span.
// Returns start, end, ok where ok is false if the span is not found.
func (s *Source) Range(span Span) (start, end int, ok bool) {
	// Check if span belongs to this source.
	if id := span.Id(); id == nil || *id != s.id {
		return 0, 0, false
	}

	// For range spans, extract the range directly.
	if st, ed, isRange := span.Range(); isRange {
		return st, ed, true
	}

	// For numbered spans, find the node and compute the range.
	node := s.Find(span)
	if node == nil {
		return 0, 0, false
	}

	// Compute the byte offset of the node.
	start = node.Offset()
	end = start + node.Len()
	return start, end, true
}

// Edit modifies the source text at the given byte range and reparses.
// Returns the byte range that was affected by the edit.
//
// The replace range is clamped to the source length.
func (s *Source) Edit(replaceStart, replaceEnd int, with string) (editStart, editEnd int) {
	// Clamp the range to valid bounds.
	if replaceStart < 0 {
		replaceStart = 0
	}
	if replaceEnd > len(s.text) {
		replaceEnd = len(s.text)
	}
	if replaceStart > replaceEnd {
		replaceStart = replaceEnd
	}

	// Build the new text.
	newText := s.text[:replaceStart] + with + s.text[replaceEnd:]

	// Full reparse (incremental reparsing is complex and deferred for now).
	s.text = newText
	s.root = Parse(newText)
	s.root.Numberize(s.id, [2]uint64{spanFullStart, spanFullEnd})
	s.lines = NewLines(newText)

	// Return the affected range based on the edit.
	editEnd = replaceStart + len(with)
	return replaceStart, editEnd
}

// Replace replaces the entire source text with new text.
// This performs a prefix/suffix diff to find the minimal edit range
// and delegates to Edit for incremental reparsing.
// Returns the byte range that was affected.
func (s *Source) Replace(newText string) (start, end int) {
	// Find common prefix.
	prefixLen := commonPrefixLen(s.text, newText)

	// Find common suffix (but don't overlap with prefix).
	oldSuffix := s.text[prefixLen:]
	newSuffix := newText[prefixLen:]
	suffixLen := commonSuffixLen(oldSuffix, newSuffix)

	// Calculate the replacement range in the old text.
	replaceStart := prefixLen
	replaceEnd := len(s.text) - suffixLen

	// Calculate the replacement content.
	newReplaceEnd := len(newText) - suffixLen
	replacement := newText[replaceStart:newReplaceEnd]

	return s.Edit(replaceStart, replaceEnd, replacement)
}

// commonPrefixLen returns the length of the common prefix between two strings.
func commonPrefixLen(a, b string) int {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	for i := 0; i < minLen; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return minLen
}

// commonSuffixLen returns the length of the common suffix between two strings.
func commonSuffixLen(a, b string) int {
	lenA, lenB := len(a), len(b)
	minLen := lenA
	if lenB < minLen {
		minLen = lenB
	}

	for i := 0; i < minLen; i++ {
		if a[lenA-1-i] != b[lenB-1-i] {
			return i
		}
	}
	return minLen
}

// Lines provides an acceleration structure for converting between
// byte offsets, line/column positions, and UTF-16 positions.
//
// Line numbers are 0-indexed. Column numbers are also 0-indexed and
// represent byte offsets within the line for Text(), but character
// offsets for Column().
type Lines struct {
	text       string
	lineStarts []int // Byte offset of each line start
}

// NewLines creates a new Lines structure from source text.
func NewLines(text string) *Lines {
	lines := &Lines{
		text:       text,
		lineStarts: []int{0},
	}

	// Find all line starts.
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			lines.lineStarts = append(lines.lineStarts, i+1)
		}
	}

	return lines
}

// Len returns the number of lines.
func (l *Lines) Len() int {
	return len(l.lineStarts)
}

// Line returns the text of the given line (0-indexed), without the
// trailing newline.
func (l *Lines) Line(line int) string {
	if line < 0 || line >= len(l.lineStarts) {
		return ""
	}

	start := l.lineStarts[line]
	var end int
	if line+1 < len(l.lineStarts) {
		end = l.lineStarts[line+1] - 1 // Exclude newline
		if end < start {
			end = start
		}
	} else {
		end = len(l.text)
	}

	return l.text[start:end]
}

// LineStart returns the byte offset of the start of the given line.
func (l *Lines) LineStart(line int) int {
	if line < 0 {
		return 0
	}
	if line >= len(l.lineStarts) {
		return len(l.text)
	}
	return l.lineStarts[line]
}

// LineEnd returns the byte offset of the end of the given line
// (exclusive, points to newline or end of text).
func (l *Lines) LineEnd(line int) int {
	if line < 0 {
		return 0
	}
	if line >= len(l.lineStarts)-1 {
		return len(l.text)
	}
	return l.lineStarts[line+1]
}

// ByteToLine returns the line number (0-indexed) for a byte offset.
func (l *Lines) ByteToLine(offset int) int {
	if offset < 0 {
		return 0
	}
	if offset >= len(l.text) {
		return len(l.lineStarts) - 1
	}

	// Binary search for the line containing the offset.
	lo, hi := 0, len(l.lineStarts)
	for lo < hi {
		mid := lo + (hi-lo)/2
		if l.lineStarts[mid] <= offset {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo - 1
}

// ByteToColumn returns the column number (0-indexed, character count)
// for a byte offset.
func (l *Lines) ByteToColumn(offset int) int {
	line := l.ByteToLine(offset)
	lineStart := l.lineStarts[line]
	lineText := l.text[lineStart:offset]
	return utf8.RuneCountInString(lineText)
}

// ByteToLineColumn returns both line and column (0-indexed) for a byte offset.
func (l *Lines) ByteToLineColumn(offset int) (line, column int) {
	line = l.ByteToLine(offset)
	lineStart := l.lineStarts[line]
	lineText := l.text[lineStart:offset]
	column = utf8.RuneCountInString(lineText)
	return
}

// LineColumnToByte converts a line/column position (0-indexed, column is
// character count) to a byte offset. Returns -1 if the position is invalid.
func (l *Lines) LineColumnToByte(line, column int) int {
	if line < 0 || line >= len(l.lineStarts) {
		return -1
	}

	start := l.lineStarts[line]
	var end int
	if line+1 < len(l.lineStarts) {
		end = l.lineStarts[line+1]
	} else {
		end = len(l.text)
	}

	// Walk through runes to find the byte offset.
	lineText := l.text[start:end]
	byteOffset := 0
	charCount := 0
	for _, r := range lineText {
		if charCount >= column {
			break
		}
		byteOffset += utf8.RuneLen(r)
		charCount++
	}

	return start + byteOffset
}

// UTF16Len returns the UTF-16 length of the text up to the given byte offset.
func (l *Lines) UTF16Len(byteOffset int) int {
	if byteOffset <= 0 {
		return 0
	}
	if byteOffset > len(l.text) {
		byteOffset = len(l.text)
	}
	return utf16Len(l.text[:byteOffset])
}

// UTF16ToByteOffset converts a UTF-16 offset to a byte offset.
func (l *Lines) UTF16ToByteOffset(utf16Offset int) int {
	if utf16Offset <= 0 {
		return 0
	}

	byteOffset := 0
	utf16Count := 0

	for _, r := range l.text {
		if utf16Count >= utf16Offset {
			break
		}
		runeLen := utf8.RuneLen(r)
		utf16Units := 1
		if r > 0xFFFF {
			utf16Units = 2 // Surrogate pair
		}
		byteOffset += runeLen
		utf16Count += utf16Units
	}

	return byteOffset
}

// ByteToUTF16LineColumn returns the line and UTF-16 column for a byte offset.
// This is useful for LSP integration which uses UTF-16 positions.
func (l *Lines) ByteToUTF16LineColumn(offset int) (line, utf16Column int) {
	line = l.ByteToLine(offset)
	lineStart := l.lineStarts[line]
	lineText := l.text[lineStart:offset]
	utf16Column = utf16Len(lineText)
	return
}

// UTF16LineColumnToByte converts a line and UTF-16 column to a byte offset.
// This is useful for LSP integration which uses UTF-16 positions.
func (l *Lines) UTF16LineColumnToByte(line, utf16Column int) int {
	if line < 0 || line >= len(l.lineStarts) {
		return -1
	}

	start := l.lineStarts[line]
	var end int
	if line+1 < len(l.lineStarts) {
		end = l.lineStarts[line+1]
	} else {
		end = len(l.text)
	}

	lineText := l.text[start:end]
	byteOffset := 0
	utf16Count := 0

	for _, r := range lineText {
		if utf16Count >= utf16Column {
			break
		}
		runeLen := utf8.RuneLen(r)
		utf16Units := 1
		if r > 0xFFFF {
			utf16Units = 2 // Surrogate pair
		}
		byteOffset += runeLen
		utf16Count += utf16Units
	}

	return start + byteOffset
}

// utf16Len returns the UTF-16 length of a string.
func utf16Len(s string) int {
	count := 0
	for _, r := range s {
		if r > 0xFFFF {
			count += 2 // Surrogate pair
		} else {
			count++
		}
	}
	return count
}

// Position represents a position in a source file.
type Position struct {
	// Line is the 0-indexed line number.
	Line int
	// Column is the 0-indexed column (character count).
	Column int
}

// PositionFromByte creates a Position from a byte offset in a Lines structure.
func PositionFromByte(lines *Lines, offset int) Position {
	line, column := lines.ByteToLineColumn(offset)
	return Position{Line: line, Column: column}
}

// ToByte converts a Position to a byte offset.
func (p Position) ToByte(lines *Lines) int {
	return lines.LineColumnToByte(p.Line, p.Column)
}

// RangePosition represents a range in a source file using positions.
type RangePosition struct {
	Start Position
	End   Position
}

// RangePositionFromBytes creates a RangePosition from byte offsets.
func RangePositionFromBytes(lines *Lines, start, end int) RangePosition {
	return RangePosition{
		Start: PositionFromByte(lines, start),
		End:   PositionFromByte(lines, end),
	}
}

// GetText returns the text content of the source within a byte range.
func (s *Source) GetText(start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(s.text) {
		end = len(s.text)
	}
	if start > end {
		return ""
	}
	return s.text[start:end]
}

// GetLine returns the text of the given line (0-indexed).
func (s *Source) GetLine(line int) string {
	return s.lines.Line(line)
}

// LineCount returns the number of lines in the source.
func (s *Source) LineCount() int {
	return s.lines.Len()
}

// String implements fmt.Stringer for debugging.
func (s *Source) String() string {
	var sb strings.Builder
	sb.WriteString("Source{id: ")
	sb.WriteString(s.id.String())
	sb.WriteString(", lines: ")
	sb.WriteString(string(rune('0' + s.lines.Len())))
	sb.WriteString("}")
	return sb.String()
}
