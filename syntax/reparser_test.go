package syntax

import (
	"testing"
)

// testReparse tests that reparsing produces the same result as full parsing.
// If incremental is true, verifies that reparsing was actually incremental.
func testReparse(t *testing.T, prev string, rangeStart, rangeEnd int, with string, incremental bool) {
	t.Helper()

	// Create a test file ID for span numbering
	vpath, _ := NewVirtualPath("/test.typ")
	testPath := NewRootedPath(ProjectRoot(), *vpath)
	fileId := NewFileId(*testPath)

	// Create source with previous text
	prevRoot := Parse(prev)
	// Use proper span numbering so ReplaceChildren can renumber
	prevRoot.Numberize(fileId, [2]uint64{2, 1 << 47})

	// Cap rangeEnd to the string length
	if rangeEnd > len(prev) {
		rangeEnd = len(prev)
	}

	// Simulate the edit
	newText := prev[:rangeStart] + with + prev[rangeEnd:]

	// Clone the root for reparsing
	root := prevRoot.Clone()

	// Reparse
	reparsedStart, reparsedEnd := Reparse(root, newText, rangeStart, rangeEnd, len(with))

	// Synthesize spans for comparison (use detached for comparison)
	root.Synthesize(Detached())

	// Parse the new text from scratch for comparison
	expected := Parse(newText)
	expected.Synthesize(Detached())

	// Compare
	if !root.SpanlessEq(expected) {
		t.Errorf("reparsing mismatch for %q -> %q\nexpected: %s\ngot: %s",
			prev, newText, debugNode(expected), debugNode(root))
	}

	// Check incrementality
	if incremental {
		if reparsedEnd-reparsedStart == len(newText) {
			t.Errorf("expected incremental reparse, but got full reparse for %q -> %q", prev, newText)
		}
	} else {
		if reparsedEnd-reparsedStart != len(newText) {
			t.Errorf("expected full reparse, but got incremental reparse for %q -> %q (reparsed %d..%d of %d)",
				prev, newText, reparsedStart, reparsedEnd, len(newText))
		}
	}
}

// debugNode returns a debug string representation of a node.
func debugNode(n *SyntaxNode) string {
	return debugNodeIndent(n, 0)
}

func debugNodeIndent(n *SyntaxNode, indent int) string {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	result := prefix + n.Kind().Name()
	if n.IsLeaf() || n.Kind().IsError() {
		result += ": " + n.Text()
	}
	result += "\n"

	for _, child := range n.Children() {
		result += debugNodeIndent(child, indent+1)
	}

	return result
}

func TestReparseMarkup(t *testing.T) {
	// Basic markup reparsing tests
	testReparse(t, "abc~def~gh~", 5, 6, "+", true)
	testReparse(t, "~~~~~~~", 3, 4, "A", true)
	testReparse(t, "abc~~", 1, 2, "", true)

	// Tests that should trigger full reparse
	testReparse(t, "#var. hello", 5, 6, " ", false)
	testReparse(t, "#var;hello", 9, 10, "a", false)
	testReparse(t, "https:/world", 7, 7, "/", false)
	testReparse(t, "hello world", 7, 12, "walkers", false)
	testReparse(t, "some content", 0, 12, "", false)
	testReparse(t, "", 0, 0, "do it", false)
	testReparse(t, "a d e", 1, 3, " b c d", false)
	testReparse(t, "~*~*~", 2, 2, "*", false)

	// Code block incremental tests
	testReparse(t, "* #{1+2} *", 6, 7, "3", true)
	testReparse(t, "#{(0, 1, 2)}", 6, 7, "11pt", true)

	// Heading changes
	testReparse(t, "\n= A heading", 4, 4, "n evocative", false)

	// Function calls
	testReparse(t, "#call() abc~d", 7, 7, "[]", true)
	testReparse(t, "a your thing a", 6, 7, "a", false)
	testReparse(t, "#grid(columns: (auto, 1fr, 40%))", 16, 20, "4pt", false)

	// Show rules and keywords
	testReparse(t, "#show f: a => b..", 16, 16, "c", false)
	testReparse(t, "#for", 4, 4, "//", false)
	testReparse(t, "a\n#let \nb", 7, 7, "i", true)

	// Raw blocks
	testReparse(t, "a ```typst hello```", 16, 17, "", false)

	// Hash handling
	testReparse(t, "a{b}c", 1, 1, "#", false)
	testReparse(t, "a#{b}c", 1, 2, "", false)
}

func TestReparseBlock(t *testing.T) {
	// Code block reparsing
	testReparse(t, "Hello #{ x + 1 }!", 9, 10, "abc", true)
	testReparse(t, "#{ [= x] }!", 5, 5, "=", true)

	// Unbalanced braces should trigger full reparse
	testReparse(t, "#{}}", 2, 2, "{", false)
	testReparse(t, "A#{}!", 3, 3, "\"", false)

	// Nested blocks
	testReparse(t, "A: #[BC]", 6, 6, "{", true)
	testReparse(t, "A: #[BC]", 6, 6, "#{", true)
	testReparse(t, "A: #[BC]", 6, 6, "#{}", true)

	// Code block with function call
	testReparse(t, "a#{call(); abc}b", 8, 8, "[]", true)
	testReparse(t, "a #while x {\n g(x) \n} b", 12, 12, "//", true)
}

func TestReparseEdgeCases(t *testing.T) {
	// Empty document
	testReparse(t, "", 0, 0, "hello", false)

	// Single character changes
	testReparse(t, "a", 0, 1, "b", false)
	testReparse(t, "abc", 1, 2, "x", false)

	// Deletion at start
	testReparse(t, "hello", 0, 1, "", false)

	// Deletion at end
	testReparse(t, "hello", 4, 5, "", false)

	// Middle insertion
	testReparse(t, "hello world", 5, 5, " big", false)
}

func TestReparseBlockTypes(t *testing.T) {
	// Test ReparseBlock directly
	text := "{ x + 1 }"
	node := ReparseBlock(text, 0, len(text))
	if node == nil {
		t.Error("ReparseBlock returned nil for valid code block")
	} else if node.Kind() != CodeBlock {
		t.Errorf("Expected CodeBlock, got %s", node.Kind().Name())
	}

	text = "[hello world]"
	node = ReparseBlock(text, 0, len(text))
	if node == nil {
		t.Error("ReparseBlock returned nil for valid content block")
	} else if node.Kind() != ContentBlock {
		t.Errorf("Expected ContentBlock, got %s", node.Kind().Name())
	}

	// Invalid blocks
	text = "not a block"
	node = ReparseBlock(text, 0, len(text))
	if node != nil {
		t.Error("ReparseBlock should return nil for non-block text")
	}

	// Unbalanced block
	text = "{ x + 1"
	node = ReparseBlock(text, 0, len(text))
	if node != nil {
		t.Error("ReparseBlock should return nil for unbalanced block")
	}
}

func TestReparseHelpers(t *testing.T) {
	// Test includes
	if !includes([2]int{0, 10}, [2]int{2, 8}) {
		t.Error("includes should return true for [0,10] containing [2,8]")
	}
	if includes([2]int{0, 10}, [2]int{0, 5}) {
		t.Error("includes should return false for touching at start")
	}
	if includes([2]int{0, 10}, [2]int{5, 10}) {
		t.Error("includes should return false for touching at end")
	}

	// Test overlaps
	if !overlaps([2]int{0, 5}, [2]int{3, 8}) {
		t.Error("overlaps should return true for overlapping ranges")
	}
	if !overlaps([2]int{0, 5}, [2]int{5, 10}) {
		t.Error("overlaps should return true for touching ranges")
	}
	if overlaps([2]int{0, 5}, [2]int{6, 10}) {
		t.Error("overlaps should return false for non-touching ranges")
	}
}
