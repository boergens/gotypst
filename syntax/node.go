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
func (n *SyntaxNode) Kind() SyntaxKind {
	return n.kind
}

// Text returns the text of a leaf node.
// For inner nodes, this returns an empty string.
func (n *SyntaxNode) Text() string {
	return n.text
}

// Children returns the children of an inner node.
// For leaf nodes, this returns nil.
func (n *SyntaxNode) Children() []*SyntaxNode {
	return n.children
}

// Error returns the error for an error node.
func (n *SyntaxNode) Error() *SyntaxError {
	return n.err
}

// IsLeaf returns true if this is a leaf node (token).
func (n *SyntaxNode) IsLeaf() bool {
	return n.children == nil
}

// IsInner returns true if this is an inner node with children.
func (n *SyntaxNode) IsInner() bool {
	return n.children != nil
}

// IsError returns true if this is an error node.
func (n *SyntaxNode) IsError() bool {
	return n.kind == Error
}

// Len returns the length of this node in bytes.
func (n *SyntaxNode) Len() int {
	if n.IsLeaf() {
		return len(n.text)
	}
	total := 0
	for _, child := range n.children {
		total += child.Len()
	}
	return total
}
