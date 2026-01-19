package syntax

import (
	"testing"
)

func TestSyntaxNodeLeaf(t *testing.T) {
	node := Leaf(Text, "hello")

	if node.Kind() != Text {
		t.Errorf("expected kind Text, got %v", node.Kind())
	}
	if node.Text() != "hello" {
		t.Errorf("expected text 'hello', got %q", node.Text())
	}
	if node.Len() != 5 {
		t.Errorf("expected len 5, got %d", node.Len())
	}
	if !node.IsLeaf() {
		t.Error("expected IsLeaf() to be true")
	}
	if node.Erroneous() {
		t.Error("expected Erroneous() to be false for leaf")
	}
	if node.Descendants() != 1 {
		t.Errorf("expected 1 descendant, got %d", node.Descendants())
	}
	if !node.Span().IsDetached() {
		t.Error("expected detached span")
	}
}

func TestSyntaxNodeInner(t *testing.T) {
	child1 := Leaf(Text, "foo")
	child2 := Leaf(Space, " ")
	child3 := Leaf(Text, "bar")

	node := Inner(Markup, []*SyntaxNode{child1, child2, child3})

	if node.Kind() != Markup {
		t.Errorf("expected kind Markup, got %v", node.Kind())
	}
	if node.Len() != 7 {
		t.Errorf("expected len 7, got %d", node.Len())
	}
	if node.IsLeaf() {
		t.Error("expected IsLeaf() to be false for inner node")
	}
	if node.Erroneous() {
		t.Error("expected Erroneous() to be false")
	}
	if node.Descendants() != 4 { // 1 inner + 3 leaves
		t.Errorf("expected 4 descendants, got %d", node.Descendants())
	}
	if len(node.Children()) != 3 {
		t.Errorf("expected 3 children, got %d", len(node.Children()))
	}
	if node.Text() != "" {
		t.Errorf("expected empty text for inner node, got %q", node.Text())
	}
}

func TestSyntaxNodeIntoText(t *testing.T) {
	child1 := Leaf(Text, "hello")
	child2 := Leaf(Space, " ")
	child3 := Leaf(Text, "world")

	node := Inner(Markup, []*SyntaxNode{child1, child2, child3})

	text := node.IntoText()
	if text != "hello world" {
		t.Errorf("expected 'hello world', got %q", text)
	}
}

func TestSyntaxNodeError(t *testing.T) {
	err := NewSyntaxError("unexpected token")
	node := ErrorNode(err, "???")

	if node.Kind() != Error {
		t.Errorf("expected kind Error, got %v", node.Kind())
	}
	if node.Text() != "???" {
		t.Errorf("expected text '???', got %q", node.Text())
	}
	if !node.Erroneous() {
		t.Error("expected Erroneous() to be true")
	}

	errors := node.Errors()
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
	if errors[0].Message != "unexpected token" {
		t.Errorf("expected message 'unexpected token', got %q", errors[0].Message)
	}
}

func TestSyntaxNodeInnerWithError(t *testing.T) {
	child1 := Leaf(Text, "foo")
	errNode := ErrorNode(NewSyntaxError("bad syntax"), "xxx")
	child3 := Leaf(Text, "bar")

	node := Inner(Markup, []*SyntaxNode{child1, errNode, child3})

	if !node.Erroneous() {
		t.Error("expected inner node with error child to be erroneous")
	}

	errors := node.Errors()
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
}

func TestSyntaxNodeHint(t *testing.T) {
	err := NewSyntaxError("problem")
	node := ErrorNode(err, "oops")

	node.Hint("try this instead")

	errors := node.Errors()
	if len(errors[0].Hints) != 1 {
		t.Fatalf("expected 1 hint, got %d", len(errors[0].Hints))
	}
	if errors[0].Hints[0] != "try this instead" {
		t.Errorf("expected hint 'try this instead', got %q", errors[0].Hints[0])
	}
}

func TestSyntaxNodeSpanlessEq(t *testing.T) {
	node1 := Leaf(Text, "hello")
	node2 := Leaf(Text, "hello")
	node3 := Leaf(Text, "world")
	node4 := Leaf(Space, "hello")

	if !node1.SpanlessEq(node2) {
		t.Error("expected node1 and node2 to be spanless equal")
	}
	if node1.SpanlessEq(node3) {
		t.Error("expected node1 and node3 to NOT be spanless equal (different text)")
	}
	if node1.SpanlessEq(node4) {
		t.Error("expected node1 and node4 to NOT be spanless equal (different kind)")
	}
}

func TestSyntaxNodeClone(t *testing.T) {
	original := Inner(Markup, []*SyntaxNode{
		Leaf(Text, "foo"),
		Leaf(Space, " "),
		Leaf(Text, "bar"),
	})

	cloned := original.Clone()

	if !original.SpanlessEq(cloned) {
		t.Error("cloned node should be spanless equal to original")
	}

	// Modify original and ensure clone is not affected
	original.Children()[0].ConvertToKind(Ident)
	if original.SpanlessEq(cloned) {
		t.Error("modifying original should not affect clone")
	}
}

func TestSyntaxNodePlaceholder(t *testing.T) {
	node := Placeholder(Ident)

	if node.Kind() != Ident {
		t.Errorf("expected kind Ident, got %v", node.Kind())
	}
	if node.Text() != "" {
		t.Errorf("expected empty text, got %q", node.Text())
	}
	if node.Len() != 0 {
		t.Errorf("expected len 0, got %d", node.Len())
	}
}

func TestSyntaxNodePlaceholderPanicOnError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when creating error placeholder")
		}
	}()
	Placeholder(Error)
}

func TestSyntaxNodeConvertToKind(t *testing.T) {
	node := Leaf(Text, "foo")
	node.ConvertToKind(Ident)

	if node.Kind() != Ident {
		t.Errorf("expected kind Ident after conversion, got %v", node.Kind())
	}
	if node.Text() != "foo" {
		t.Errorf("expected text preserved, got %q", node.Text())
	}
}

func TestSyntaxNodeConvertToError(t *testing.T) {
	node := Leaf(Text, "foo")
	node.ConvertToError("bad token")

	if node.Kind() != Error {
		t.Errorf("expected kind Error after conversion, got %v", node.Kind())
	}
	if !node.Erroneous() {
		t.Error("expected node to be erroneous after conversion")
	}

	errors := node.Errors()
	if len(errors) != 1 || errors[0].Message != "bad token" {
		t.Error("expected error message 'bad token'")
	}
}

func TestSyntaxNodeSynthesize(t *testing.T) {
	child1 := Leaf(Text, "foo")
	child2 := Leaf(Text, "bar")
	node := Inner(Markup, []*SyntaxNode{child1, child2})

	span, _ := SpanFromNumber(FileId(1), 100)
	node.Synthesize(span)

	if node.Span() != span {
		t.Error("expected node span to be synthesized")
	}
	if node.Children()[0].Span() != span {
		t.Error("expected child span to be synthesized")
	}
	if node.Children()[1].Span() != span {
		t.Error("expected child span to be synthesized")
	}
}

// --- LinkedNode tests ---

func TestLinkedNodeNew(t *testing.T) {
	root := Inner(Markup, []*SyntaxNode{
		Leaf(Text, "hello"),
		Leaf(Space, " "),
		Leaf(Text, "world"),
	})

	ln := NewLinkedNode(root)

	if ln.Get() != root {
		t.Error("expected Get() to return root node")
	}
	if ln.Offset() != 0 {
		t.Errorf("expected offset 0, got %d", ln.Offset())
	}
	if ln.Index() != 0 {
		t.Errorf("expected index 0, got %d", ln.Index())
	}
	if ln.Parent() != nil {
		t.Error("expected root to have no parent")
	}
}

func TestLinkedNodeChildren(t *testing.T) {
	root := Inner(Markup, []*SyntaxNode{
		Leaf(Text, "foo"),  // offset 0, len 3
		Leaf(Space, " "),   // offset 3, len 1
		Leaf(Text, "bar"),  // offset 4, len 3
	})

	ln := NewLinkedNode(root)
	children := ln.Children()

	if len(children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(children))
	}

	// Check first child
	if children[0].Offset() != 0 {
		t.Errorf("child 0: expected offset 0, got %d", children[0].Offset())
	}
	if children[0].Index() != 0 {
		t.Errorf("child 0: expected index 0, got %d", children[0].Index())
	}
	if children[0].Text() != "foo" {
		t.Errorf("child 0: expected text 'foo', got %q", children[0].Text())
	}

	// Check second child
	if children[1].Offset() != 3 {
		t.Errorf("child 1: expected offset 3, got %d", children[1].Offset())
	}

	// Check third child
	if children[2].Offset() != 4 {
		t.Errorf("child 2: expected offset 4, got %d", children[2].Offset())
	}
}

func TestLinkedNodeRange(t *testing.T) {
	root := Inner(Markup, []*SyntaxNode{
		Leaf(Text, "hello"),
	})

	ln := NewLinkedNode(root)
	children := ln.Children()

	rng := children[0].Range()
	if rng[0] != 0 || rng[1] != 5 {
		t.Errorf("expected range [0, 5], got [%d, %d]", rng[0], rng[1])
	}
}

func TestLinkedNodeParent(t *testing.T) {
	root := Inner(Markup, []*SyntaxNode{
		Leaf(Text, "hello"),
	})

	ln := NewLinkedNode(root)
	child := ln.Children()[0]

	if child.Parent() != ln {
		t.Error("expected child's parent to be root")
	}
}

func TestLinkedNodePrevNextSibling(t *testing.T) {
	// Create tree with trivia (space) between non-trivia nodes
	root := Inner(Markup, []*SyntaxNode{
		Leaf(Text, "foo"),
		Leaf(Space, " "),
		Leaf(Text, "bar"),
	})

	ln := NewLinkedNode(root)
	children := ln.Children()

	// bar's prev sibling should be foo (skipping space)
	barNode := children[2]
	prevSib := barNode.PrevSibling()
	if prevSib == nil {
		t.Fatal("expected bar to have prev sibling")
	}
	if prevSib.Text() != "foo" {
		t.Errorf("expected prev sibling 'foo', got %q", prevSib.Text())
	}

	// foo's next sibling should be bar (skipping space)
	fooNode := children[0]
	nextSib := fooNode.NextSibling()
	if nextSib == nil {
		t.Fatal("expected foo to have next sibling")
	}
	if nextSib.Text() != "bar" {
		t.Errorf("expected next sibling 'bar', got %q", nextSib.Text())
	}

	// foo should have no prev sibling
	if fooNode.PrevSibling() != nil {
		t.Error("expected foo to have no prev sibling")
	}

	// bar should have no next sibling
	if barNode.NextSibling() != nil {
		t.Error("expected bar to have no next sibling")
	}
}

func TestLinkedNodeSiblingKind(t *testing.T) {
	root := Inner(Markup, []*SyntaxNode{
		Leaf(Text, "foo"),
		Leaf(Space, " "),
		Leaf(Ident, "bar"),
	})

	ln := NewLinkedNode(root)
	children := ln.Children()

	kind, ok := children[2].PrevSiblingKind()
	if !ok || kind != Text {
		t.Errorf("expected prev sibling kind Text, got %v", kind)
	}

	kind, ok = children[0].NextSiblingKind()
	if !ok || kind != Ident {
		t.Errorf("expected next sibling kind Ident, got %v", kind)
	}
}

func TestLinkedNodeLeafAt(t *testing.T) {
	// Build a simple tree:
	// Markup["#set text(12pt)"]
	//   Hash[0..1] "#"
	//   Code[1..16]
	//     Ident[1..4] "set"
	//     Space[4..5] " "
	//     Ident[5..9] "text"
	//     ...
	hash := Leaf(Hash, "#")
	setIdent := Leaf(Ident, "set")
	space := Leaf(Space, " ")
	textIdent := Leaf(Ident, "text")
	code := Inner(Code, []*SyntaxNode{setIdent, space, textIdent})
	root := Inner(Markup, []*SyntaxNode{hash, code})

	ln := NewLinkedNode(root)

	// Find leaf at offset 2 (inside "set") with Before
	leaf := ln.LeafAt(2, Before)
	if leaf == nil {
		t.Fatal("expected to find leaf at offset 2")
	}
	if leaf.Text() != "set" {
		t.Errorf("expected 'set' at offset 2, got %q", leaf.Text())
	}

	// Find leaf at offset 5 (start of "text") with After
	leaf = ln.LeafAt(5, After)
	if leaf == nil {
		t.Fatal("expected to find leaf at offset 5")
	}
	if leaf.Text() != "text" {
		t.Errorf("expected 'text' at offset 5, got %q", leaf.Text())
	}
}

func TestLinkedNodeLeftmostRightmostLeaf(t *testing.T) {
	// Tree with trivia at edges
	root := Inner(Markup, []*SyntaxNode{
		Leaf(Space, " "),   // trivia
		Leaf(Text, "foo"),
		Leaf(Space, " "),   // trivia
		Leaf(Text, "bar"),
		Leaf(Space, " "),   // trivia
	})

	ln := NewLinkedNode(root)

	leftmost := ln.LeftmostLeaf()
	if leftmost == nil {
		t.Fatal("expected to find leftmost leaf")
	}
	if leftmost.Text() != "foo" {
		t.Errorf("expected leftmost non-trivia leaf 'foo', got %q", leftmost.Text())
	}

	rightmost := ln.RightmostLeaf()
	if rightmost == nil {
		t.Fatal("expected to find rightmost leaf")
	}
	if rightmost.Text() != "bar" {
		t.Errorf("expected rightmost non-trivia leaf 'bar', got %q", rightmost.Text())
	}
}

func TestLinkedNodePrevNextLeaf(t *testing.T) {
	// Create a tree structure
	root := Inner(Markup, []*SyntaxNode{
		Leaf(Text, "aaa"),
		Leaf(Space, " "),
		Leaf(Text, "bbb"),
	})

	ln := NewLinkedNode(root)
	children := ln.Children()

	// Get "bbb" leaf
	bbbNode := children[2]

	// Prev leaf of "bbb" should be "aaa" (skipping space trivia)
	prevLeaf := bbbNode.PrevLeaf()
	if prevLeaf == nil {
		t.Fatal("expected bbb to have prev leaf")
	}
	if prevLeaf.Text() != "aaa" {
		t.Errorf("expected prev leaf 'aaa', got %q", prevLeaf.Text())
	}

	// Get "aaa" leaf
	aaaNode := children[0]

	// Next leaf of "aaa" should be "bbb" (skipping space trivia)
	nextLeaf := aaaNode.NextLeaf()
	if nextLeaf == nil {
		t.Fatal("expected aaa to have next leaf")
	}
	if nextLeaf.Text() != "bbb" {
		t.Errorf("expected next leaf 'bbb', got %q", nextLeaf.Text())
	}
}

func TestLinkedNodeFind(t *testing.T) {
	// Create a tree and numberize it
	child1 := Leaf(Text, "foo")
	child2 := Leaf(Text, "bar")
	root := Inner(Markup, []*SyntaxNode{child1, child2})

	// Assign spans
	root.Numberize(FileId(1), [2]uint64{2, 1000})

	ln := NewLinkedNode(root)

	// Find by the root's span
	found := ln.Find(root.Span())
	if found == nil {
		t.Fatal("expected to find node by root span")
	}
	if found.Get() != root {
		t.Error("expected to find root node")
	}

	// Find by child's span
	found = ln.Find(child1.Span())
	if found == nil {
		t.Fatal("expected to find node by child1 span")
	}
	if found.Get() != child1 {
		t.Error("expected to find child1 node")
	}

	// Find by non-existent span
	nonexistent, _ := SpanFromNumber(FileId(1), 999999)
	found = ln.Find(nonexistent)
	if found != nil {
		t.Error("expected not to find non-existent span")
	}
}

// --- SyntaxError tests ---

func TestSyntaxError(t *testing.T) {
	err := NewSyntaxError("test error")

	if err.Message != "test error" {
		t.Errorf("expected message 'test error', got %q", err.Message)
	}
	if !err.Span.IsDetached() {
		t.Error("expected detached span for new error")
	}
	if len(err.Hints) != 0 {
		t.Errorf("expected no hints, got %d", len(err.Hints))
	}

	err.AddHint("hint 1")
	err.AddHint("hint 2")

	if len(err.Hints) != 2 {
		t.Errorf("expected 2 hints, got %d", len(err.Hints))
	}
	if err.Hints[0] != "hint 1" || err.Hints[1] != "hint 2" {
		t.Error("hints not added correctly")
	}
}

func TestSyntaxErrorClone(t *testing.T) {
	err := NewSyntaxError("original")
	err.AddHint("hint")

	cloned := err.Clone()

	if cloned.Message != err.Message {
		t.Error("cloned message should match")
	}
	if len(cloned.Hints) != len(err.Hints) {
		t.Error("cloned hints count should match")
	}

	// Modify original
	err.Message = "modified"
	if cloned.Message == err.Message {
		t.Error("modifying original should not affect clone")
	}
}

// --- Numberize tests ---

func TestNumberize(t *testing.T) {
	child1 := Leaf(Text, "foo")
	child2 := Leaf(Text, "bar")
	root := Inner(Markup, []*SyntaxNode{child1, child2})

	err := root.Numberize(FileId(1), [2]uint64{2, 1000})
	if err != nil {
		t.Fatalf("Numberize failed: %v", err)
	}

	// All spans should now be non-detached
	if root.Span().IsDetached() {
		t.Error("root span should not be detached after numberize")
	}
	if child1.Span().IsDetached() {
		t.Error("child1 span should not be detached after numberize")
	}
	if child2.Span().IsDetached() {
		t.Error("child2 span should not be detached after numberize")
	}

	// Spans should be in order (parent has smaller number than children in subtree)
	// But each node has its own span number
	if root.Span().Number() == child1.Span().Number() {
		t.Error("root and child1 should have different span numbers")
	}
	if child1.Span().Number() == child2.Span().Number() {
		t.Error("child1 and child2 should have different span numbers")
	}

	// All should have the same file ID
	if root.Span().Id() != FileId(1) {
		t.Error("expected file ID 1")
	}
	if child1.Span().Id() != FileId(1) {
		t.Error("expected file ID 1")
	}
}

func TestNumberizeEmptyRange(t *testing.T) {
	node := Leaf(Text, "foo")
	err := node.Numberize(FileId(1), [2]uint64{10, 10})
	if err == nil {
		t.Error("expected error for empty range")
	}
	if _, ok := err.(Unnumberable); !ok {
		t.Error("expected Unnumberable error")
	}
}
