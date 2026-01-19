// Package syntax provides the syntax tree types for Typst.
//
// This file is a Go translation of typst-syntax/src/node.rs from the
// original Typst compiler. It defines SyntaxNode (untyped syntax tree nodes)
// and LinkedNode (context-aware navigation).
package syntax

import (
	"fmt"
	"strings"
)

// SyntaxNode represents a node in the untyped syntax tree.
// It comes in three flavors: leaf nodes, inner nodes, and error nodes.
//
// This is the primary abstraction for working with Typst's concrete syntax tree.
// The tree is "concrete" because it preserves all source information including
// whitespace and comments.
type SyntaxNode struct {
	data nodeData
}

// nodeData is the internal representation of a syntax node.
// Implemented by leafNode, innerNode, and errorNode.
type nodeData interface {
	kind() SyntaxKind
	len() int
	span() Span
	text() string
	children() []*SyntaxNode
	erroneous() bool
	descendants() int
	spanlessEq(other nodeData) bool
	clone() nodeData
}

// leafNode is a leaf node in the untyped syntax tree.
// Contains a token kind, text content, and span.
type leafNode struct {
	nodeKind SyntaxKind
	nodeText string
	nodeSpan Span
}

func (n *leafNode) kind() SyntaxKind           { return n.nodeKind }
func (n *leafNode) len() int                   { return len(n.nodeText) }
func (n *leafNode) span() Span                 { return n.nodeSpan }
func (n *leafNode) text() string               { return n.nodeText }
func (n *leafNode) children() []*SyntaxNode    { return nil }
func (n *leafNode) erroneous() bool            { return false }
func (n *leafNode) descendants() int           { return 1 }
func (n *leafNode) spanlessEq(other nodeData) bool {
	if o, ok := other.(*leafNode); ok {
		return n.nodeKind == o.nodeKind && n.nodeText == o.nodeText
	}
	return false
}
func (n *leafNode) clone() nodeData {
	return &leafNode{nodeKind: n.nodeKind, nodeText: n.nodeText, nodeSpan: n.nodeSpan}
}

// innerNode is an inner node in the untyped syntax tree.
// Contains children nodes and aggregated metadata.
type innerNode struct {
	nodeKind       SyntaxKind
	nodeLen        int
	nodeSpan       Span
	nodeDescendants int
	nodeErroneous  bool
	upper          uint64 // upper bound of numbering range
	nodeChildren   []*SyntaxNode
}

func (n *innerNode) kind() SyntaxKind           { return n.nodeKind }
func (n *innerNode) len() int                   { return n.nodeLen }
func (n *innerNode) span() Span                 { return n.nodeSpan }
func (n *innerNode) text() string               { return "" }
func (n *innerNode) children() []*SyntaxNode    { return n.nodeChildren }
func (n *innerNode) erroneous() bool            { return n.nodeErroneous }
func (n *innerNode) descendants() int           { return n.nodeDescendants }
func (n *innerNode) spanlessEq(other nodeData) bool {
	o, ok := other.(*innerNode)
	if !ok {
		return false
	}
	if n.nodeKind != o.nodeKind || n.nodeLen != o.nodeLen ||
		n.nodeDescendants != o.nodeDescendants || n.nodeErroneous != o.nodeErroneous ||
		len(n.nodeChildren) != len(o.nodeChildren) {
		return false
	}
	for i, child := range n.nodeChildren {
		if !child.SpanlessEq(o.nodeChildren[i]) {
			return false
		}
	}
	return true
}
func (n *innerNode) clone() nodeData {
	children := make([]*SyntaxNode, len(n.nodeChildren))
	for i, c := range n.nodeChildren {
		children[i] = c.Clone()
	}
	return &innerNode{
		nodeKind:       n.nodeKind,
		nodeLen:        n.nodeLen,
		nodeSpan:       n.nodeSpan,
		nodeDescendants: n.nodeDescendants,
		nodeErroneous:  n.nodeErroneous,
		upper:          n.upper,
		nodeChildren:   children,
	}
}

// errorNode is an error node in the untyped syntax tree.
// Contains the malformed text and error details.
type errorNode struct {
	nodeText string
	error    *SyntaxError
}

func (n *errorNode) kind() SyntaxKind           { return Error }
func (n *errorNode) len() int                   { return len(n.nodeText) }
func (n *errorNode) span() Span                 { return n.error.Span }
func (n *errorNode) text() string               { return n.nodeText }
func (n *errorNode) children() []*SyntaxNode    { return nil }
func (n *errorNode) erroneous() bool            { return true }
func (n *errorNode) descendants() int           { return 1 }
func (n *errorNode) spanlessEq(other nodeData) bool {
	if o, ok := other.(*errorNode); ok {
		return n.nodeText == o.nodeText && n.error.spanlessEq(o.error)
	}
	return false
}
func (n *errorNode) clone() nodeData {
	return &errorNode{nodeText: n.nodeText, error: n.error.Clone()}
}

// SyntaxError represents a syntactical error.
type SyntaxError struct {
	// Span is the source location of the error.
	Span Span
	// Message is the error message.
	Message string
	// Hints provides additional guidance to the user.
	Hints []string
}

// NewSyntaxError creates a new detached syntax error.
func NewSyntaxError(message string) *SyntaxError {
	return &SyntaxError{
		Span:    Detached(),
		Message: message,
		Hints:   nil,
	}
}

// AddHint adds a user-presentable hint to this error.
func (e *SyntaxError) AddHint(hint string) {
	e.Hints = append(e.Hints, hint)
}

// Clone creates a copy of the error.
func (e *SyntaxError) Clone() *SyntaxError {
	hints := make([]string, len(e.Hints))
	copy(hints, e.Hints)
	return &SyntaxError{
		Span:    e.Span,
		Message: e.Message,
		Hints:   hints,
	}
}

func (e *SyntaxError) spanlessEq(other *SyntaxError) bool {
	if e.Message != other.Message || len(e.Hints) != len(other.Hints) {
		return false
	}
	for i, h := range e.Hints {
		if h != other.Hints[i] {
			return false
		}
	}
	return true
}

// String implements fmt.Stringer.
func (e *SyntaxError) String() string {
	return e.Message
}

// Error implements the error interface.
func (e *SyntaxError) Error() string {
	return e.Message
}

// --- SyntaxNode constructors ---

// Leaf creates a new leaf node.
func Leaf(kind SyntaxKind, text string) *SyntaxNode {
	if kind == Error {
		panic("cannot create leaf node with Error kind; use ErrorNode instead")
	}
	return &SyntaxNode{
		data: &leafNode{
			nodeKind: kind,
			nodeText: text,
			nodeSpan: Detached(),
		},
	}
}

// Inner creates a new inner node with children.
func Inner(kind SyntaxKind, children []*SyntaxNode) *SyntaxNode {
	if kind == Error {
		panic("cannot create inner node with Error kind; use ErrorNode instead")
	}

	var totalLen int
	descendants := 1
	erroneous := false

	for _, child := range children {
		totalLen += child.Len()
		descendants += child.Descendants()
		erroneous = erroneous || child.Erroneous()
	}

	return &SyntaxNode{
		data: &innerNode{
			nodeKind:       kind,
			nodeLen:        totalLen,
			nodeSpan:       Detached(),
			nodeDescendants: descendants,
			nodeErroneous:  erroneous,
			upper:          0,
			nodeChildren:   children,
		},
	}
}

// ErrorNode creates a new error node.
func ErrorNode(err *SyntaxError, text string) *SyntaxNode {
	return &SyntaxNode{
		data: &errorNode{
			nodeText: text,
			error:    err,
		},
	}
}

// Placeholder creates a dummy node of the given kind.
// Panics if kind is Error.
func Placeholder(kind SyntaxKind) *SyntaxNode {
	if kind == Error {
		panic("cannot create error placeholder")
	}
	return &SyntaxNode{
		data: &leafNode{
			nodeKind: kind,
			nodeText: "",
			nodeSpan: Detached(),
		},
	}
}

// Default creates a default empty node (End kind with no text).
func Default() *SyntaxNode {
	return Leaf(End, "")
}

// --- SyntaxNode accessors ---

// Kind returns the type of the node.
func (n *SyntaxNode) Kind() SyntaxKind {
	return n.data.kind()
}

// IsEmpty returns true if the node has zero byte length.
func (n *SyntaxNode) IsEmpty() bool {
	return n.Len() == 0
}

// Len returns the byte length of the node in the source text.
func (n *SyntaxNode) Len() int {
	return n.data.len()
}

// Span returns the span of the node.
func (n *SyntaxNode) Span() Span {
	return n.data.span()
}

// Text returns the text of the node if it is a leaf or error node.
// Returns the empty string for inner nodes.
func (n *SyntaxNode) Text() string {
	return n.data.text()
}

// IntoText extracts the full text from the node.
// For inner nodes, this recursively builds the string from children.
func (n *SyntaxNode) IntoText() string {
	if inner, ok := n.data.(*innerNode); ok {
		var sb strings.Builder
		for _, child := range inner.nodeChildren {
			sb.WriteString(child.IntoText())
		}
		return sb.String()
	}
	return n.data.text()
}

// Children returns an iterator over the node's children.
func (n *SyntaxNode) Children() []*SyntaxNode {
	return n.data.children()
}

// Erroneous returns true if this node or any of its children contain an error.
func (n *SyntaxNode) Erroneous() bool {
	return n.data.erroneous()
}

// Descendants returns the number of nodes in the subtree, including this node.
func (n *SyntaxNode) Descendants() int {
	return n.data.descendants()
}

// Errors returns all error messages for this node and its descendants.
func (n *SyntaxNode) Errors() []*SyntaxError {
	if !n.Erroneous() {
		return nil
	}

	if err, ok := n.data.(*errorNode); ok {
		return []*SyntaxError{err.error}
	}

	var errors []*SyntaxError
	for _, child := range n.Children() {
		if child.Erroneous() {
			errors = append(errors, child.Errors()...)
		}
	}
	return errors
}

// Hint adds a user-presentable hint if this is an error node.
func (n *SyntaxNode) Hint(hint string) {
	if err, ok := n.data.(*errorNode); ok {
		err.error.AddHint(hint)
	}
}

// Synthesize sets a synthetic span for the node and all its descendants.
func (n *SyntaxNode) Synthesize(span Span) {
	switch d := n.data.(type) {
	case *leafNode:
		d.nodeSpan = span
	case *innerNode:
		d.nodeSpan = span
		d.upper = span.Number()
		for _, child := range d.nodeChildren {
			child.Synthesize(span)
		}
	case *errorNode:
		d.error.Span = span
	}
}

// SpanlessEq returns true if two syntax nodes are structurally equal,
// ignoring spans.
func (n *SyntaxNode) SpanlessEq(other *SyntaxNode) bool {
	return n.data.spanlessEq(other.data)
}

// Clone creates a deep copy of the syntax node.
func (n *SyntaxNode) Clone() *SyntaxNode {
	return &SyntaxNode{data: n.data.clone()}
}

// IsLeaf returns true if this is a leaf node.
func (n *SyntaxNode) IsLeaf() bool {
	_, ok := n.data.(*leafNode)
	return ok
}

// --- Parser/internal methods ---

// ConvertToKind converts the node to another kind.
// Don't use this for converting to an error.
func (n *SyntaxNode) ConvertToKind(kind SyntaxKind) {
	if kind == Error {
		panic("cannot convert to Error kind; use ConvertToError instead")
	}
	switch d := n.data.(type) {
	case *leafNode:
		d.nodeKind = kind
	case *innerNode:
		d.nodeKind = kind
	case *errorNode:
		panic("cannot convert error node")
	}
}

// ConvertToError converts the node to an error if it isn't already one.
func (n *SyntaxNode) ConvertToError(message string) {
	if n.Kind() != Error {
		text := n.IntoText()
		n.data = &errorNode{
			nodeText: text,
			error:    NewSyntaxError(message),
		}
	}
}

// Expected converts the node to an error stating that the given thing was expected.
func (n *SyntaxNode) Expected(expected string) {
	kind := n.Kind()
	n.ConvertToError(fmt.Sprintf("expected %s, found %s", expected, kind.Name()))
	if kind.IsKeyword() && (expected == "identifier" || expected == "pattern") {
		text := n.Text()
		n.Hint(fmt.Sprintf("keyword `%s` is not allowed as an identifier; try `%s_` instead", text, text))
	}
}

// Unexpected converts the node to an error stating it was unexpected.
func (n *SyntaxNode) Unexpected() {
	n.ConvertToError(fmt.Sprintf("unexpected %s", n.Kind().Name()))
}

// Upper returns the upper bound of assigned numbers in this subtree.
func (n *SyntaxNode) Upper() uint64 {
	switch d := n.data.(type) {
	case *leafNode:
		return d.nodeSpan.Number() + 1
	case *innerNode:
		return d.upper
	case *errorNode:
		return d.error.Span.Number() + 1
	}
	return 0
}

// SetSpan sets the span on the node directly (for leaf/error nodes).
func (n *SyntaxNode) SetSpan(span Span) {
	switch d := n.data.(type) {
	case *leafNode:
		d.nodeSpan = span
	case *errorNode:
		d.error.Span = span
	}
}

// ChildrenMut returns a mutable reference to the children slice.
func (n *SyntaxNode) ChildrenMut() []*SyntaxNode {
	if inner, ok := n.data.(*innerNode); ok {
		return inner.nodeChildren
	}
	return nil
}

// String implements fmt.Stringer for debugging.
func (n *SyntaxNode) String() string {
	switch d := n.data.(type) {
	case *leafNode:
		return fmt.Sprintf("%s: %q", d.nodeKind, d.nodeText)
	case *innerNode:
		return fmt.Sprintf("%s: %d", d.nodeKind, d.nodeLen)
	case *errorNode:
		return fmt.Sprintf("Error: %q (%s)", d.nodeText, d.error.Message)
	}
	return "unknown"
}

// --- Numbering ---

// Unnumberable indicates that a node cannot be numbered within a given interval.
type Unnumberable struct{}

func (Unnumberable) Error() string {
	return "cannot number within this interval"
}

// NumberingResult is the result type for span assignment operations.
type NumberingResult error

// Numberize assigns spans to each node within the given interval.
func (n *SyntaxNode) Numberize(id FileId, within [2]uint64) NumberingResult {
	if within[0] >= within[1] {
		return Unnumberable{}
	}

	mid := (within[0] + within[1]) / 2
	midSpan, ok := SpanFromNumber(id, mid)
	if !ok {
		return Unnumberable{}
	}

	switch d := n.data.(type) {
	case *leafNode:
		d.nodeSpan = midSpan
	case *innerNode:
		return d.numberize(id, nil, within)
	case *errorNode:
		d.error.Span = midSpan
	}

	return nil
}

// numberize assigns span numbers within an interval to this inner node's subtree.
func (inner *innerNode) numberize(id FileId, rangeIdx *[2]int, within [2]uint64) NumberingResult {
	// Determine how many nodes we will number.
	var descendants int
	if rangeIdx != nil {
		if rangeIdx[0] >= rangeIdx[1] {
			return nil
		}
		for _, child := range inner.nodeChildren[rangeIdx[0]:rangeIdx[1]] {
			descendants += child.Descendants()
		}
	} else {
		descendants = inner.nodeDescendants
	}

	// Determine the distance between two neighbouring assigned numbers.
	space := within[1] - within[0]
	stride := space / (2 * uint64(descendants))
	if stride == 0 {
		stride = space / uint64(inner.nodeDescendants)
		if stride == 0 {
			return Unnumberable{}
		}
	}

	// Number the node itself.
	start := within[0]
	if rangeIdx == nil {
		end := start + stride
		midSpan, _ := SpanFromNumber(id, (start+end)/2)
		inner.nodeSpan = midSpan
		inner.upper = within[1]
		start = end
	}

	// Number the children.
	childStart := 0
	childEnd := len(inner.nodeChildren)
	if rangeIdx != nil {
		childStart = rangeIdx[0]
		childEnd = rangeIdx[1]
	}

	for _, child := range inner.nodeChildren[childStart:childEnd] {
		end := start + uint64(child.Descendants())*stride
		if err := child.Numberize(id, [2]uint64{start, end}); err != nil {
			return err
		}
		start = end
	}

	return nil
}

// ReplaceChildren replaces a range of children with a replacement.
// May have mutated the children if it returns an error.
func (n *SyntaxNode) ReplaceChildren(rangeStart, rangeEnd int, replacement []*SyntaxNode) NumberingResult {
	inner, ok := n.data.(*innerNode)
	if !ok {
		return nil
	}
	return inner.replaceChildren(rangeStart, rangeEnd, replacement)
}

func (inner *innerNode) replaceChildren(rangeStart, rangeEnd int, replacement []*SyntaxNode) NumberingResult {
	id := inner.nodeSpan.Id()
	if id == NoFile {
		return Unnumberable{}
	}

	replacementStart := 0
	replacementEnd := len(replacement)

	// Trim off common prefix.
	for rangeStart < rangeEnd && replacementStart < replacementEnd &&
		inner.nodeChildren[rangeStart].SpanlessEq(replacement[replacementStart]) {
		rangeStart++
		replacementStart++
	}

	// Trim off common suffix.
	for rangeStart < rangeEnd && replacementStart < replacementEnd &&
		inner.nodeChildren[rangeEnd-1].SpanlessEq(replacement[replacementEnd-1]) {
		rangeEnd--
		replacementEnd--
	}

	actualReplacement := replacement[replacementStart:replacementEnd]
	superseded := inner.nodeChildren[rangeStart:rangeEnd]

	// Compute the new byte length.
	var replacementLen int
	for _, r := range actualReplacement {
		replacementLen += r.Len()
	}
	var supersededLen int
	for _, s := range superseded {
		supersededLen += s.Len()
	}
	inner.nodeLen = inner.nodeLen + replacementLen - supersededLen

	// Compute the new number of descendants.
	var replacementDesc int
	for _, r := range actualReplacement {
		replacementDesc += r.Descendants()
	}
	var supersededDesc int
	for _, s := range superseded {
		supersededDesc += s.Descendants()
	}
	inner.nodeDescendants = inner.nodeDescendants + replacementDesc - supersededDesc

	// Determine whether we're still erroneous after the replacement.
	erroneous := false
	for _, r := range actualReplacement {
		if r.Erroneous() {
			erroneous = true
			break
		}
	}
	if !erroneous && inner.nodeErroneous {
		for _, c := range inner.nodeChildren[:rangeStart] {
			if c.Erroneous() {
				erroneous = true
				break
			}
		}
		if !erroneous {
			for _, c := range inner.nodeChildren[rangeEnd:] {
				if c.Erroneous() {
					erroneous = true
					break
				}
			}
		}
	}
	inner.nodeErroneous = erroneous

	// Perform the replacement.
	newChildren := make([]*SyntaxNode, 0, len(inner.nodeChildren)-len(superseded)+len(actualReplacement))
	newChildren = append(newChildren, inner.nodeChildren[:rangeStart]...)
	newChildren = append(newChildren, actualReplacement...)
	newChildren = append(newChildren, inner.nodeChildren[rangeEnd:]...)
	inner.nodeChildren = newChildren

	rangeEnd = rangeStart + len(actualReplacement)

	// Renumber the new children with exponential backtracking.
	maxLeft := rangeStart
	maxRight := len(inner.nodeChildren) - rangeEnd
	left := 0
	right := 0

	for {
		renumberStart := rangeStart - left
		renumberEnd := rangeEnd + right

		// The minimum assignable number.
		var startNumber uint64
		if renumberStart > 0 {
			startNumber = inner.nodeChildren[renumberStart-1].Upper()
		} else {
			startNumber = inner.nodeSpan.Number() + 1
		}

		// The upper bound for renumbering.
		var endNumber uint64
		if renumberEnd < len(inner.nodeChildren) {
			endNumber = inner.nodeChildren[renumberEnd].Span().Number()
		} else {
			endNumber = inner.upper
		}

		// Try to renumber.
		rangeIdxVal := [2]int{renumberStart, renumberEnd}
		if err := inner.numberize(id, &rangeIdxVal, [2]uint64{startNumber, endNumber}); err == nil {
			return nil
		}

		// If it didn't even work with all children, we give up.
		if left == maxLeft && right == maxRight {
			return Unnumberable{}
		}

		// Exponential expansion to both sides.
		left = min((left+1)*2, maxLeft)
		right = min((right+1)*2, maxRight)
	}
}

// UpdateParent updates this node after changes were made to one of its children.
func (n *SyntaxNode) UpdateParent(prevLen, newLen, prevDescendants, newDescendants int) {
	if inner, ok := n.data.(*innerNode); ok {
		inner.nodeLen = inner.nodeLen + newLen - prevLen
		inner.nodeDescendants = inner.nodeDescendants + newDescendants - prevDescendants
		inner.nodeErroneous = false
		for _, child := range inner.nodeChildren {
			if child.Erroneous() {
				inner.nodeErroneous = true
				break
			}
		}
	}
}

// --- LinkedNode ---

// Side indicates whether the cursor is before or after the related byte index.
type Side int

const (
	// Before means the cursor is before the byte index.
	Before Side = iota
	// After means the cursor is after the byte index.
	After
)

// LinkedNode is a syntax node in a context.
// It knows its exact offset in the file and provides access to its
// children, parent, and siblings.
//
// Note that all sibling and leaf accessors skip over trivia!
type LinkedNode struct {
	// node is the underlying syntax node.
	node *SyntaxNode
	// parent is the parent of this node (nil for root).
	parent *LinkedNode
	// index is the index of this node in its parent's children array.
	index int
	// offset is this node's byte offset in the source file.
	offset int
}

// NewLinkedNode starts a new traversal at a root node.
func NewLinkedNode(root *SyntaxNode) *LinkedNode {
	return &LinkedNode{
		node:   root,
		parent: nil,
		index:  0,
		offset: 0,
	}
}

// Get returns the contained syntax node.
func (ln *LinkedNode) Get() *SyntaxNode {
	return ln.node
}

// Index returns the index of this node in its parent's children list.
func (ln *LinkedNode) Index() int {
	return ln.index
}

// Offset returns the absolute byte offset of this node in the source file.
func (ln *LinkedNode) Offset() int {
	return ln.offset
}

// Range returns the byte range of this node in the source file.
func (ln *LinkedNode) Range() [2]int {
	return [2]int{ln.offset, ln.offset + ln.node.Len()}
}

// Kind returns the kind of the underlying syntax node.
func (ln *LinkedNode) Kind() SyntaxKind {
	return ln.node.Kind()
}

// Len returns the byte length of the underlying syntax node.
func (ln *LinkedNode) Len() int {
	return ln.node.Len()
}

// Span returns the span of the underlying syntax node.
func (ln *LinkedNode) Span() Span {
	return ln.node.Span()
}

// Text returns the text of the underlying syntax node.
func (ln *LinkedNode) Text() string {
	return ln.node.Text()
}

// Erroneous returns true if this node or any of its children contain an error.
func (ln *LinkedNode) Erroneous() bool {
	return ln.node.Erroneous()
}

// IsLeaf returns true if this is a leaf node.
func (ln *LinkedNode) IsLeaf() bool {
	return ln.node.IsLeaf()
}

// Children returns an iterator over this node's children as LinkedNodes.
func (ln *LinkedNode) Children() []*LinkedNode {
	children := ln.node.Children()
	if len(children) == 0 {
		return nil
	}

	result := make([]*LinkedNode, len(children))
	offset := ln.offset
	for i, child := range children {
		result[i] = &LinkedNode{
			node:   child,
			parent: ln,
			index:  i,
			offset: offset,
		}
		offset += child.Len()
	}
	return result
}

// Find finds a descendant with the given span.
func (ln *LinkedNode) Find(span Span) *LinkedNode {
	if ln.Span() == span {
		return ln
	}

	inner, ok := ln.node.data.(*innerNode)
	if !ok {
		return nil
	}

	// The parent of a subtree has a smaller span number than all of its
	// descendants. Therefore, we can bail out early if the target span's
	// number is smaller than our number.
	if span.Number() < inner.nodeSpan.Number() {
		return nil
	}

	children := ln.Children()
	for i, child := range children {
		// Every node in this child's subtree has a smaller span number than
		// the next sibling. Therefore we only need to recurse if the next
		// sibling's span number is larger than the target span's number.
		var nextSpanNumber uint64
		if i+1 < len(children) {
			nextSpanNumber = children[i+1].Span().Number()
		} else {
			nextSpanNumber = ^uint64(0) // max value
		}

		if nextSpanNumber > span.Number() {
			if found := child.Find(span); found != nil {
				return found
			}
		}
	}

	return nil
}

// --- Parent and sibling access ---

// Parent returns this node's parent.
func (ln *LinkedNode) Parent() *LinkedNode {
	return ln.parent
}

// PrevSibling returns the first previous non-trivia sibling node.
func (ln *LinkedNode) PrevSibling() *LinkedNode {
	if ln.parent == nil {
		return nil
	}

	children := ln.parent.node.Children()
	offset := ln.offset

	for i := ln.index - 1; i >= 0; i-- {
		offset -= children[i].Len()
		if !children[i].Kind().IsTrivia() {
			return &LinkedNode{
				node:   children[i],
				parent: ln.parent,
				index:  i,
				offset: offset,
			}
		}
	}
	return nil
}

// NextSibling returns the next non-trivia sibling node.
func (ln *LinkedNode) NextSibling() *LinkedNode {
	if ln.parent == nil {
		return nil
	}

	children := ln.parent.node.Children()
	offset := ln.offset + ln.node.Len()

	for i := ln.index + 1; i < len(children); i++ {
		if !children[i].Kind().IsTrivia() {
			return &LinkedNode{
				node:   children[i],
				parent: ln.parent,
				index:  i,
				offset: offset,
			}
		}
		offset += children[i].Len()
	}
	return nil
}

// ParentKind returns the kind of this node's parent.
func (ln *LinkedNode) ParentKind() (SyntaxKind, bool) {
	if ln.parent == nil {
		return 0, false
	}
	return ln.parent.Kind(), true
}

// PrevSiblingKind returns the kind of this node's first previous non-trivia sibling.
func (ln *LinkedNode) PrevSiblingKind() (SyntaxKind, bool) {
	prev := ln.PrevSibling()
	if prev == nil {
		return 0, false
	}
	return prev.Kind(), true
}

// NextSiblingKind returns the kind of this node's next non-trivia sibling.
func (ln *LinkedNode) NextSiblingKind() (SyntaxKind, bool) {
	next := ln.NextSibling()
	if next == nil {
		return 0, false
	}
	return next.Kind(), true
}

// --- Leaf access ---

// PrevLeaf returns the rightmost non-trivia leaf before this node.
func (ln *LinkedNode) PrevLeaf() *LinkedNode {
	node := ln
	for {
		prev := node.PrevSibling()
		if prev == nil {
			if node.parent == nil {
				return nil
			}
			return node.parent.PrevLeaf()
		}
		if leaf := prev.RightmostLeaf(); leaf != nil {
			return leaf
		}
		node = prev
	}
}

// LeftmostLeaf finds the leftmost contained non-trivia leaf.
func (ln *LinkedNode) LeftmostLeaf() *LinkedNode {
	if ln.node.IsLeaf() && !ln.Kind().IsTrivia() && !ln.Kind().IsError() {
		return ln
	}

	for _, child := range ln.Children() {
		if leaf := child.LeftmostLeaf(); leaf != nil {
			return leaf
		}
	}

	return nil
}

// leafBefore gets the leaf immediately before the specified byte offset.
func (ln *LinkedNode) leafBefore(cursor int) *LinkedNode {
	children := ln.node.Children()
	if len(children) == 0 && cursor <= ln.offset+ln.node.Len() {
		return ln
	}

	offset := ln.offset
	count := len(children)
	linkedChildren := ln.Children()

	for i, child := range linkedChildren {
		length := child.Len()
		if (offset < cursor && cursor <= offset+length) ||
			(offset == cursor && i+1 == count) {
			return child.leafBefore(cursor)
		}
		offset += length
	}

	return nil
}

// leafAfter gets the leaf after the specified byte offset.
func (ln *LinkedNode) leafAfter(cursor int) *LinkedNode {
	children := ln.node.Children()
	if len(children) == 0 && cursor < ln.offset+ln.node.Len() {
		return ln
	}

	offset := ln.offset
	for _, child := range ln.Children() {
		length := child.Len()
		if offset <= cursor && cursor < offset+length {
			return child.leafAfter(cursor)
		}
		offset += length
	}

	return nil
}

// LeafAt gets the leaf at the specified byte offset.
func (ln *LinkedNode) LeafAt(cursor int, side Side) *LinkedNode {
	switch side {
	case Before:
		return ln.leafBefore(cursor)
	case After:
		return ln.leafAfter(cursor)
	}
	return nil
}

// RightmostLeaf finds the rightmost contained non-trivia leaf.
func (ln *LinkedNode) RightmostLeaf() *LinkedNode {
	if ln.node.IsLeaf() && !ln.Kind().IsTrivia() {
		return ln
	}

	children := ln.Children()
	for i := len(children) - 1; i >= 0; i-- {
		if leaf := children[i].RightmostLeaf(); leaf != nil {
			return leaf
		}
	}

	return nil
}

// NextLeaf returns the leftmost non-trivia leaf after this node.
func (ln *LinkedNode) NextLeaf() *LinkedNode {
	node := ln
	for {
		next := node.NextSibling()
		if next == nil {
			if node.parent == nil {
				return nil
			}
			return node.parent.NextLeaf()
		}
		if leaf := next.LeftmostLeaf(); leaf != nil {
			return leaf
		}
		node = next
	}
}

// String implements fmt.Stringer for debugging.
func (ln *LinkedNode) String() string {
	return ln.node.String()
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
