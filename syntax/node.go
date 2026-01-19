package syntax

// SyntaxNode represents a node in the syntax tree.
// It can be either a leaf (token) or an inner node with children.
type SyntaxNode struct {
	kind     SyntaxKind
	text     string
	children []*SyntaxNode
	err      *SyntaxError
}

// Leaf creates a leaf node (token) with the given kind and text.
func Leaf(kind SyntaxKind, text string) *SyntaxNode {
	return &SyntaxNode{
		kind: kind,
		text: text,
	}
}

// Inner creates an inner node with the given kind and children.
func Inner(kind SyntaxKind, children []*SyntaxNode) *SyntaxNode {
	return &SyntaxNode{
		kind:     kind,
		children: children,
	}
}

// ErrorNode creates an error node with the given error and text.
func ErrorNode(err *SyntaxError, text string) *SyntaxNode {
	return &SyntaxNode{
		kind: Error,
		text: text,
		err:  err,
	}
}

// Kind returns the syntax kind of this node.
// Returns Error for nil nodes.
func (n *SyntaxNode) Kind() SyntaxKind {
	if n == nil {
		return Error
	}
	return n.kind
}

// Text returns the text of a leaf node.
// For inner nodes and nil nodes, this returns an empty string.
func (n *SyntaxNode) Text() string {
	if n == nil {
		return ""
	}
	return n.text
}

// Children returns the children of an inner node.
// For leaf nodes and nil nodes, this returns nil.
func (n *SyntaxNode) Children() []*SyntaxNode {
	if n == nil {
		return nil
	}
	return n.children
}

// Error returns the error for an error node.
func (n *SyntaxNode) Error() *SyntaxError {
	if n == nil {
		return nil
	}
	return n.err
}

// IsLeaf returns true if this is a leaf node (token).
func (n *SyntaxNode) IsLeaf() bool {
	if n == nil {
		return true
	}
	return n.children == nil
}

// IsInner returns true if this is an inner node with children.
func (n *SyntaxNode) IsInner() bool {
	if n == nil {
		return false
	}
	return n.children != nil
}

// IsError returns true if this is an error node.
func (n *SyntaxNode) IsError() bool {
	if n == nil {
		return true
	}
	return n.kind == Error
}

// Len returns the length of this node in bytes.
func (n *SyntaxNode) Len() int {
	if n == nil {
		return 0
	}
	if n.IsLeaf() {
		return len(n.text)
	}
	total := 0
	for _, child := range n.children {
		total += child.Len()
	}
	return total
}

// IsEmpty returns true if the node has no content.
func (n *SyntaxNode) IsEmpty() bool {
	if n == nil {
		return true
	}
	return n.Len() == 0
}

// Cast attempts to cast this node to a typed AST node.
// Returns nil if the cast is not valid.
func (n *SyntaxNode) Cast(kind SyntaxKind) *SyntaxNode {
	if n == nil || n.kind != kind {
		return nil
	}
	return n
}

// CastFirst returns the first child that can be cast to the given kind.
func (n *SyntaxNode) CastFirst(kind SyntaxKind) *SyntaxNode {
	if n == nil {
		return nil
	}
	for _, child := range n.children {
		if child.kind == kind {
			return child
		}
	}
	return nil
}

// CastAll returns all children that can be cast to the given kind.
func (n *SyntaxNode) CastAll(kind SyntaxKind) []*SyntaxNode {
	if n == nil {
		return nil
	}
	var result []*SyntaxNode
	for _, child := range n.children {
		if child.kind == kind {
			result = append(result, child)
		}
	}
	return result
}

// CastFirstInSet returns the first child with a kind in the given set.
func (n *SyntaxNode) CastFirstInSet(set SyntaxSet) *SyntaxNode {
	if n == nil {
		return nil
	}
	for _, child := range n.children {
		if set.Contains(child.kind) {
			return child
		}
	}
	return nil
}

// ChildrenMut returns a mutable slice of children for in-place mutation.
func (n *SyntaxNode) ChildrenMut() []*SyntaxNode {
	return n.children
}

// ConvertToKind changes this node's kind in place.
func (n *SyntaxNode) ConvertToKind(kind SyntaxKind) {
	n.kind = kind
}

// ConvertToError converts this node to an error node with the given message.
func (n *SyntaxNode) ConvertToError(message string) {
	n.kind = Error
	n.err = NewSyntaxError(message)
}

// Expected marks this node as having expected something else.
func (n *SyntaxNode) Expected(thing string) {
	n.ConvertToError("expected " + thing)
}

// Unexpected marks this node as unexpected.
func (n *SyntaxNode) Unexpected() {
	msg := "unexpected " + n.kind.Name()
	n.ConvertToError(msg)
}

// Hint adds a hint to this node's error if it is an error node.
func (n *SyntaxNode) Hint(hint string) {
	if n.err != nil {
		n.err.AddHint(hint)
	}
}

// IntoText returns the text of this node.
// For inner nodes, recursively concatenates all children's text.
func (n *SyntaxNode) IntoText() string {
	if n.IsLeaf() {
		return n.text
	}
	var result string
	for _, child := range n.children {
		result += child.IntoText()
	}
	return result
}

// NewSyntaxNode creates a new syntax node.
func NewSyntaxNode(kind SyntaxKind, text string, children []*SyntaxNode) *SyntaxNode {
	return &SyntaxNode{
		kind:     kind,
		text:     text,
		children: children,
	}
}
