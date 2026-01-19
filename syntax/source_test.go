package syntax

import (
	"testing"
)

func TestNewSource(t *testing.T) {
	vpath, _ := NewVirtualPath("/test.typ")
	path := NewRootedPath(ProjectRoot(), *vpath)
	id := NewFileId(*path)

	src := NewSource(id, "Hello World")

	if src.Id() != id {
		t.Error("expected source id to match")
	}
	if src.Text() != "Hello World" {
		t.Errorf("expected text 'Hello World', got %q", src.Text())
	}
	if src.Len() != 11 {
		t.Errorf("expected length 11, got %d", src.Len())
	}
	if src.Root() == nil {
		t.Error("expected root to be non-nil")
	}
	if src.Lines() == nil {
		t.Error("expected lines to be non-nil")
	}
}

func TestNewDetachedSource(t *testing.T) {
	src := NewDetachedSource("Test content")

	if src.Text() != "Test content" {
		t.Errorf("expected text 'Test content', got %q", src.Text())
	}
	if src.Root() == nil {
		t.Error("expected root to be non-nil")
	}
}

func TestSourceEdit(t *testing.T) {
	src := NewDetachedSource("Hello World")

	// Edit to replace "World" with "Go"
	start, end := src.Edit(6, 11, "Go")

	if src.Text() != "Hello Go" {
		t.Errorf("expected 'Hello Go', got %q", src.Text())
	}
	if start != 6 || end != 8 {
		t.Errorf("expected range [6, 8], got [%d, %d]", start, end)
	}
}

func TestSourceEditBoundaries(t *testing.T) {
	src := NewDetachedSource("Hello")

	// Test clamping of negative start
	src.Edit(-5, 2, "XX")
	if src.Text() != "XXllo" {
		t.Errorf("expected 'XXllo', got %q", src.Text())
	}

	// Test clamping of end past length
	src = NewDetachedSource("Hello")
	src.Edit(3, 100, "p")
	if src.Text() != "Help" {
		t.Errorf("expected 'Help', got %q", src.Text())
	}

	// Test start > end (start is clamped to end, so inserts at position 2)
	src = NewDetachedSource("Hello")
	src.Edit(4, 2, "X")
	if src.Text() != "HeXllo" {
		t.Errorf("expected 'HeXllo' (insert at position 2), got %q", src.Text())
	}
}

func TestSourceReplace(t *testing.T) {
	src := NewDetachedSource("Hello World")

	// Replace with text that has common prefix/suffix
	start, end := src.Replace("Hello Go")

	if src.Text() != "Hello Go" {
		t.Errorf("expected 'Hello Go', got %q", src.Text())
	}
	if start != 6 || end != 8 {
		t.Errorf("expected range [6, 8], got [%d, %d]", start, end)
	}
}

func TestSourceReplaceCompleteChange(t *testing.T) {
	src := NewDetachedSource("abc")
	src.Replace("xyz")

	if src.Text() != "xyz" {
		t.Errorf("expected 'xyz', got %q", src.Text())
	}
}

func TestSourceGetText(t *testing.T) {
	src := NewDetachedSource("Hello World")

	text := src.GetText(0, 5)
	if text != "Hello" {
		t.Errorf("expected 'Hello', got %q", text)
	}

	text = src.GetText(6, 11)
	if text != "World" {
		t.Errorf("expected 'World', got %q", text)
	}

	// Test out-of-bounds clamping
	text = src.GetText(-5, 5)
	if text != "Hello" {
		t.Errorf("expected 'Hello', got %q", text)
	}

	text = src.GetText(6, 100)
	if text != "World" {
		t.Errorf("expected 'World', got %q", text)
	}
}

func TestLinesBasic(t *testing.T) {
	lines := NewLines("Hello\nWorld\nTest")

	if lines.Len() != 3 {
		t.Errorf("expected 3 lines, got %d", lines.Len())
	}

	if lines.Line(0) != "Hello" {
		t.Errorf("expected line 0 to be 'Hello', got %q", lines.Line(0))
	}
	if lines.Line(1) != "World" {
		t.Errorf("expected line 1 to be 'World', got %q", lines.Line(1))
	}
	if lines.Line(2) != "Test" {
		t.Errorf("expected line 2 to be 'Test', got %q", lines.Line(2))
	}
}

func TestLinesSingleLine(t *testing.T) {
	lines := NewLines("No newlines here")

	if lines.Len() != 1 {
		t.Errorf("expected 1 line, got %d", lines.Len())
	}
	if lines.Line(0) != "No newlines here" {
		t.Errorf("expected line 0 to be 'No newlines here', got %q", lines.Line(0))
	}
}

func TestLinesEmpty(t *testing.T) {
	lines := NewLines("")

	if lines.Len() != 1 {
		t.Errorf("expected 1 line (empty), got %d", lines.Len())
	}
	if lines.Line(0) != "" {
		t.Errorf("expected line 0 to be empty, got %q", lines.Line(0))
	}
}

func TestLinesTrailingNewline(t *testing.T) {
	lines := NewLines("Hello\n")

	if lines.Len() != 2 {
		t.Errorf("expected 2 lines, got %d", lines.Len())
	}
	if lines.Line(0) != "Hello" {
		t.Errorf("expected line 0 to be 'Hello', got %q", lines.Line(0))
	}
	if lines.Line(1) != "" {
		t.Errorf("expected line 1 to be empty, got %q", lines.Line(1))
	}
}

func TestLinesLineStart(t *testing.T) {
	lines := NewLines("Hello\nWorld\nTest")

	if lines.LineStart(0) != 0 {
		t.Errorf("expected line 0 start at 0, got %d", lines.LineStart(0))
	}
	if lines.LineStart(1) != 6 {
		t.Errorf("expected line 1 start at 6, got %d", lines.LineStart(1))
	}
	if lines.LineStart(2) != 12 {
		t.Errorf("expected line 2 start at 12, got %d", lines.LineStart(2))
	}
}

func TestLinesLineEnd(t *testing.T) {
	lines := NewLines("Hello\nWorld\nTest")

	// Line 0 ends at position 6 (exclusive, pointing to newline)
	if lines.LineEnd(0) != 6 {
		t.Errorf("expected line 0 end at 6, got %d", lines.LineEnd(0))
	}
	// Line 1 ends at position 12
	if lines.LineEnd(1) != 12 {
		t.Errorf("expected line 1 end at 12, got %d", lines.LineEnd(1))
	}
	// Line 2 ends at position 16 (end of text)
	if lines.LineEnd(2) != 16 {
		t.Errorf("expected line 2 end at 16, got %d", lines.LineEnd(2))
	}
}

func TestLinesByteToLine(t *testing.T) {
	lines := NewLines("Hello\nWorld\nTest")

	tests := []struct {
		offset   int
		expected int
	}{
		{0, 0},   // 'H'
		{5, 0},   // 'o' (before newline)
		{6, 1},   // 'W'
		{11, 1},  // 'd'
		{12, 2},  // 'T'
		{15, 2},  // 't'
		{-1, 0},  // negative clamps to 0
		{100, 2}, // beyond end clamps to last line
	}

	for _, tt := range tests {
		result := lines.ByteToLine(tt.offset)
		if result != tt.expected {
			t.Errorf("ByteToLine(%d) = %d, expected %d", tt.offset, result, tt.expected)
		}
	}
}

func TestLinesByteToColumn(t *testing.T) {
	lines := NewLines("Hello\nWorld")

	// Test basic ASCII
	if col := lines.ByteToColumn(0); col != 0 {
		t.Errorf("expected column 0 at offset 0, got %d", col)
	}
	if col := lines.ByteToColumn(3); col != 3 {
		t.Errorf("expected column 3 at offset 3, got %d", col)
	}
	if col := lines.ByteToColumn(6); col != 0 {
		t.Errorf("expected column 0 at offset 6 (start of line 1), got %d", col)
	}
	if col := lines.ByteToColumn(8); col != 2 {
		t.Errorf("expected column 2 at offset 8, got %d", col)
	}
}

func TestLinesByteToColumnUnicode(t *testing.T) {
	// "日本語" is 9 bytes (3 bytes per character), 3 characters
	lines := NewLines("日本語\nHello")

	// Column is character count, not byte count
	if col := lines.ByteToColumn(0); col != 0 {
		t.Errorf("expected column 0 at offset 0, got %d", col)
	}
	if col := lines.ByteToColumn(3); col != 1 {
		t.Errorf("expected column 1 at offset 3, got %d", col)
	}
	if col := lines.ByteToColumn(6); col != 2 {
		t.Errorf("expected column 2 at offset 6, got %d", col)
	}
	if col := lines.ByteToColumn(9); col != 3 {
		t.Errorf("expected column 3 at offset 9, got %d", col)
	}
}

func TestLinesByteToLineColumn(t *testing.T) {
	lines := NewLines("Hello\nWorld")

	line, col := lines.ByteToLineColumn(0)
	if line != 0 || col != 0 {
		t.Errorf("expected (0, 0), got (%d, %d)", line, col)
	}

	line, col = lines.ByteToLineColumn(5)
	if line != 0 || col != 5 {
		t.Errorf("expected (0, 5), got (%d, %d)", line, col)
	}

	line, col = lines.ByteToLineColumn(6)
	if line != 1 || col != 0 {
		t.Errorf("expected (1, 0), got (%d, %d)", line, col)
	}

	line, col = lines.ByteToLineColumn(9)
	if line != 1 || col != 3 {
		t.Errorf("expected (1, 3), got (%d, %d)", line, col)
	}
}

func TestLinesLineColumnToByte(t *testing.T) {
	lines := NewLines("Hello\nWorld")

	tests := []struct {
		line, col int
		expected  int
	}{
		{0, 0, 0},
		{0, 5, 5},
		{1, 0, 6},
		{1, 3, 9},
		{-1, 0, -1}, // invalid line
		{5, 0, -1},  // invalid line
	}

	for _, tt := range tests {
		result := lines.LineColumnToByte(tt.line, tt.col)
		if result != tt.expected {
			t.Errorf("LineColumnToByte(%d, %d) = %d, expected %d", tt.line, tt.col, result, tt.expected)
		}
	}
}

func TestLinesLineColumnToByteUnicode(t *testing.T) {
	// "日本語" is 9 bytes (3 bytes per character), 3 characters
	lines := NewLines("日本語\nHello")

	// Column is character count
	result := lines.LineColumnToByte(0, 1)
	if result != 3 {
		t.Errorf("expected byte offset 3 for (0, 1), got %d", result)
	}

	result = lines.LineColumnToByte(0, 2)
	if result != 6 {
		t.Errorf("expected byte offset 6 for (0, 2), got %d", result)
	}
}

func TestLinesUTF16Len(t *testing.T) {
	// ASCII only: UTF-16 length equals byte length
	lines := NewLines("Hello")
	if utf16len := lines.UTF16Len(5); utf16len != 5 {
		t.Errorf("expected UTF-16 length 5, got %d", utf16len)
	}

	// Unicode with 2-byte chars: still 1 UTF-16 unit each
	lines = NewLines("日本語")
	if utf16len := lines.UTF16Len(9); utf16len != 3 {
		t.Errorf("expected UTF-16 length 3, got %d", utf16len)
	}

	// Emoji (surrogate pair): 1 character = 2 UTF-16 units
	lines = NewLines("😀") // This emoji is 4 bytes in UTF-8
	if utf16len := lines.UTF16Len(4); utf16len != 2 {
		t.Errorf("expected UTF-16 length 2 (surrogate pair), got %d", utf16len)
	}
}

func TestLinesUTF16ToByteOffset(t *testing.T) {
	// ASCII
	lines := NewLines("Hello")
	if offset := lines.UTF16ToByteOffset(3); offset != 3 {
		t.Errorf("expected byte offset 3, got %d", offset)
	}

	// Unicode (3-byte UTF-8 chars, 1 UTF-16 unit each)
	lines = NewLines("日本語")
	if offset := lines.UTF16ToByteOffset(2); offset != 6 {
		t.Errorf("expected byte offset 6, got %d", offset)
	}

	// Emoji (4-byte UTF-8, 2 UTF-16 units)
	lines = NewLines("😀World")
	// After the emoji (2 UTF-16 units), we're at byte offset 4
	if offset := lines.UTF16ToByteOffset(2); offset != 4 {
		t.Errorf("expected byte offset 4 after emoji, got %d", offset)
	}
}

func TestLinesByteToUTF16LineColumn(t *testing.T) {
	lines := NewLines("Hello\n日本語")

	// ASCII position
	line, col := lines.ByteToUTF16LineColumn(3)
	if line != 0 || col != 3 {
		t.Errorf("expected (0, 3), got (%d, %d)", line, col)
	}

	// Unicode position: byte 9 is after "日本" (6 bytes), so 2 UTF-16 units
	line, col = lines.ByteToUTF16LineColumn(12) // After "日本" on line 1
	if line != 1 || col != 2 {
		t.Errorf("expected (1, 2), got (%d, %d)", line, col)
	}
}

func TestLinesUTF16LineColumnToByte(t *testing.T) {
	lines := NewLines("Hello\n日本語")

	// ASCII position
	offset := lines.UTF16LineColumnToByte(0, 3)
	if offset != 3 {
		t.Errorf("expected byte offset 3, got %d", offset)
	}

	// Unicode: UTF-16 column 2 on line 1 is byte offset 12 (6 bytes for "日本")
	offset = lines.UTF16LineColumnToByte(1, 2)
	if offset != 12 {
		t.Errorf("expected byte offset 12, got %d", offset)
	}
}

func TestPositionFromByte(t *testing.T) {
	lines := NewLines("Hello\nWorld")

	pos := PositionFromByte(lines, 8)
	if pos.Line != 1 || pos.Column != 2 {
		t.Errorf("expected Position{1, 2}, got Position{%d, %d}", pos.Line, pos.Column)
	}
}

func TestPositionToByte(t *testing.T) {
	lines := NewLines("Hello\nWorld")

	pos := Position{Line: 1, Column: 2}
	offset := pos.ToByte(lines)
	if offset != 8 {
		t.Errorf("expected byte offset 8, got %d", offset)
	}
}

func TestRangePositionFromBytes(t *testing.T) {
	lines := NewLines("Hello\nWorld")

	rp := RangePositionFromBytes(lines, 0, 11)
	if rp.Start.Line != 0 || rp.Start.Column != 0 {
		t.Errorf("expected start Position{0, 0}, got Position{%d, %d}", rp.Start.Line, rp.Start.Column)
	}
	if rp.End.Line != 1 || rp.End.Column != 5 {
		t.Errorf("expected end Position{1, 5}, got Position{%d, %d}", rp.End.Line, rp.End.Column)
	}
}

func TestSourceLineCount(t *testing.T) {
	src := NewDetachedSource("Line1\nLine2\nLine3")
	if src.LineCount() != 3 {
		t.Errorf("expected 3 lines, got %d", src.LineCount())
	}
}

func TestSourceGetLine(t *testing.T) {
	src := NewDetachedSource("Line1\nLine2\nLine3")

	if src.GetLine(0) != "Line1" {
		t.Errorf("expected 'Line1', got %q", src.GetLine(0))
	}
	if src.GetLine(1) != "Line2" {
		t.Errorf("expected 'Line2', got %q", src.GetLine(1))
	}
	if src.GetLine(2) != "Line3" {
		t.Errorf("expected 'Line3', got %q", src.GetLine(2))
	}
}

func TestCommonPrefixLen(t *testing.T) {
	tests := []struct {
		a, b     string
		expected int
	}{
		{"hello", "hello", 5},
		{"hello", "help", 3},
		{"abc", "xyz", 0},
		{"", "hello", 0},
		{"hello", "", 0},
		{"", "", 0},
	}

	for _, tt := range tests {
		result := commonPrefixLen(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("commonPrefixLen(%q, %q) = %d, expected %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestCommonSuffixLen(t *testing.T) {
	tests := []struct {
		a, b     string
		expected int
	}{
		{"hello", "hello", 5},
		{"hello", "cello", 4},
		{"abc", "xyz", 0},
		{"", "hello", 0},
		{"hello", "", 0},
		{"", "", 0},
	}

	for _, tt := range tests {
		result := commonSuffixLen(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("commonSuffixLen(%q, %q) = %d, expected %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestSourceFind(t *testing.T) {
	vpath, _ := NewVirtualPath("/find-test.typ")
	path := NewRootedPath(ProjectRoot(), *vpath)
	id := NewFileId(*path)
	src := NewSource(id, "= Hello World")

	// Find using the root's span
	rootSpan := src.Root().Span()
	found := src.Find(rootSpan)
	if found == nil {
		t.Fatal("expected to find root node")
	}
	if found.Kind() != src.Root().Kind() {
		t.Errorf("expected to find node with kind %v, got %v", src.Root().Kind(), found.Kind())
	}
}

func TestSourceRange(t *testing.T) {
	vpath, _ := NewVirtualPath("/range-test.typ")
	path := NewRootedPath(ProjectRoot(), *vpath)
	id := NewFileId(*path)
	src := NewSource(id, "Hello World")

	// Get range for root span
	rootSpan := src.Root().Span()
	start, end, ok := src.Range(rootSpan)
	if !ok {
		t.Fatal("expected to find range for root span")
	}
	// The root should span the entire text
	if start != 0 {
		t.Errorf("expected start 0, got %d", start)
	}
	if end != 11 {
		t.Errorf("expected end 11, got %d", end)
	}
}

func TestSourceRangeDetached(t *testing.T) {
	src := NewDetachedSource("Hello")

	// Detached span should not be found
	_, _, ok := src.Range(Detached())
	if ok {
		t.Error("expected detached span to not be found")
	}
}

func TestLinesOutOfBounds(t *testing.T) {
	lines := NewLines("Hello")

	// Test out of bounds line access
	if lines.Line(-1) != "" {
		t.Error("expected empty string for negative line")
	}
	if lines.Line(100) != "" {
		t.Error("expected empty string for line beyond end")
	}

	// Test line start bounds
	if lines.LineStart(-1) != 0 {
		t.Error("expected 0 for negative line start")
	}
	if lines.LineStart(100) != 5 {
		t.Errorf("expected %d for line start beyond end, got %d", 5, lines.LineStart(100))
	}
}

func TestUtf16Len(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"", 0},
		{"hello", 5},
		{"日本語", 3},        // 3 characters, each fits in BMP
		{"😀😁", 4},          // 2 emoji, each is a surrogate pair (2 UTF-16 units)
		{"a😀b", 4},          // 1 + 2 + 1 = 4 UTF-16 units
	}

	for _, tt := range tests {
		result := utf16Len(tt.input)
		if result != tt.expected {
			t.Errorf("utf16Len(%q) = %d, expected %d", tt.input, result, tt.expected)
		}
	}
}
