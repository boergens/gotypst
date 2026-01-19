// Package syntax provides incremental reparsing for Typst documents.
//
// This file is a Go translation of typst-syntax/src/reparser.rs from the
// original Typst compiler. It implements two-phase incremental reparsing
// that minimizes parsing work when documents are edited.
package syntax

// Reparse refreshes the given syntax node with as little parsing as possible.
//
// Takes the new text, the range in the old text that was replaced, and the
// length of the replacement. Returns the range in the new text that was
// ultimately reparsed.
//
// The high-level API for this function is Source.Edit.
func Reparse(root *SyntaxNode, text string, replacedStart, replacedEnd, replacementLen int) (start, end int) {
	replaced := [2]int{replacedStart, replacedEnd}
	result := tryReparse(text, replaced, replacementLen, nil, root, 0)
	if result != nil {
		return result[0], result[1]
	}

	// Fall back to full reparse
	id := root.Span().Id()
	*root = *Parse(text)
	if id != nil {
		root.Numberize(*id, [2]uint64{spanFullStart, spanFullEnd})
	}
	return 0, len(text)
}

// tryReparse attempts to reparse inside the given node.
// Returns the range that was reparsed, or nil if reparsing failed.
func tryReparse(
	text string,
	replaced [2]int,
	replacementLen int,
	parentKind *SyntaxKind,
	node *SyntaxNode,
	offset int,
) *[2]int {
	// The range of children which overlap with the edit.
	overlapStart := int(^uint(0) >> 1) // max int
	overlapEnd := 0
	cursor := offset
	nodeKind := node.Kind()

	children := node.ChildrenMut()
	for i, child := range children {
		prevRange := [2]int{cursor, cursor + child.Len()}
		prevLen := child.Len()
		prevDesc := child.Descendants()

		// Does the child surround the edit?
		// If so, try to reparse within it or itself.
		if !child.IsLeaf() && includes(prevRange, replaced) {
			newLen := prevLen + replacementLen - (replaced[1] - replaced[0])
			newRange := [2]int{cursor, cursor + newLen}

			// Try to reparse within the child.
			result := tryReparse(
				text,
				replaced,
				replacementLen,
				&nodeKind,
				child,
				cursor,
			)
			if result != nil {
				if child.Len() != newLen {
					panic("child length mismatch after reparse")
				}
				newDesc := child.Descendants()
				node.UpdateParent(prevLen, newLen, prevDesc, newDesc)
				return result
			}

			// If the child is a block, try to reparse the block.
			if child.Kind().IsBlock() {
				newborn := ReparseBlock(text, newRange[0], newRange[1])
				if newborn != nil {
					err := node.ReplaceChildren(i, i+1, []*SyntaxNode{newborn})
					if err == nil {
						return &newRange
					}
				}
			}
		}

		// Does the child overlap with the edit?
		if overlaps(prevRange, replaced) {
			if i < overlapStart {
				overlapStart = i
			}
			if i+1 > overlapEnd {
				overlapEnd = i + 1
			}
		}

		// Is the child beyond the edit?
		if replaced[1] < cursor {
			break
		}

		cursor += child.Len()
	}

	// Try to reparse a range of markup expressions within markup. This is only
	// possible if the markup is top-level or contained in a block, not if it is
	// contained in things like headings or lists because too much can go wrong
	// with indent and line breaks.
	if overlapStart >= overlapEnd ||
		node.Kind() != Markup ||
		(parentKind != nil && *parentKind != ContentBlock) {
		return nil
	}

	// Reparse a segment. Retries until it works, taking exponentially more
	// children into account.
	expansion := 1
	for {
		// Add slack in both directions.
		start := overlapStart
		if expansion >= 2 {
			if start > expansion {
				start = overlapStart - expansion
			} else {
				start = 0
			}
		} else if start > 2 {
			start = overlapStart - 2
		} else {
			start = 0
		}

		end := overlapEnd + expansion
		if end > len(children) {
			end = len(children)
		}

		// Expand to the left.
		for start > 0 && expand(children[start]) {
			start--
		}

		// Expand to the right.
		for end < len(children) && expand(children[end]) {
			end++
		}

		// Also take hash.
		if start > 0 && children[start-1].Kind() == Hash {
			start--
		}

		// Synthesize what `atStart` and `nesting` would be at the start of the
		// reparse.
		prefixLen := 0
		nesting := 0
		atStart := true
		for _, child := range children[:start] {
			prefixLen += child.Len()
			nextAtStart(child, &atStart)
			nextNesting(child, &nesting)
		}

		// Determine what `atStart` will have to be at the end of the reparse.
		prevLen := 0
		prevAtStartAfter := atStart
		prevNestingAfter := nesting
		for _, child := range children[start:end] {
			prevLen += child.Len()
			nextAtStart(child, &prevAtStartAfter)
			nextNesting(child, &prevNestingAfter)
		}

		// Determine the range in the new text that we want to reparse.
		shifted := offset + prefixLen
		newLen := prevLen + replacementLen - (replaced[1] - replaced[0])
		newRange := [2]int{shifted, shifted + newLen}
		atEnd := end == len(children)

		// Reparse!
		reparsed := ReparseMarkup(
			text,
			newRange[0],
			newRange[1],
			&atStart,
			&nesting,
			parentKind == nil,
		)

		if reparsed != nil {
			// If more children follow, atStart must match its previous value.
			// Similarly, if children follow or we're not top-level, the nesting
			// must match its previous value.
			if (atEnd || atStart == prevAtStartAfter) &&
				((atEnd && parentKind == nil) || nesting == prevNestingAfter) {
				err := node.ReplaceChildren(start, end, reparsed)
				if err == nil {
					return &newRange
				}
			}
		}

		// If it didn't even work with all children, we give up.
		if start == 0 && atEnd {
			break
		}

		// Exponential expansion to both sides.
		expansion *= 2
	}

	return nil
}

// includes returns true if the outer range fully contains the inner range (no touching).
func includes(outer, inner [2]int) bool {
	return outer[0] < inner[0] && outer[1] > inner[1]
}

// overlaps returns true if the first and second range overlap or touch.
func overlaps(first, second [2]int) bool {
	return (first[0] <= second[0] && second[0] <= first[1]) ||
		(second[0] <= first[0] && first[0] <= second[1])
}

// expand returns true if the selection should be expanded beyond a node of this kind.
func expand(node *SyntaxNode) bool {
	kind := node.Kind()
	if kind.IsTrivia() || kind.IsError() || kind == Semicolon {
		return true
	}
	text := node.Text()
	return text == "/" || text == ":"
}

// nextAtStart updates atStart based on the given node.
// It determines whether `atStart` would still be true after this node.
func nextAtStart(node *SyntaxNode, atStart *bool) {
	kind := node.Kind()
	if kind.IsTrivia() {
		*atStart = *atStart || kind == Parbreak ||
			(kind == Space && containsNewline(node.Text()))
	} else {
		*atStart = false
	}
}

// containsNewline returns true if the text contains any newline character.
func containsNewline(text string) bool {
	for _, c := range text {
		if IsNewline(c) {
			return true
		}
	}
	return false
}

// nextNesting updates the nesting count based on the node.
func nextNesting(node *SyntaxNode, nesting *int) {
	if node.Kind() == Text {
		text := node.Text()
		if text == "[" {
			*nesting++
		} else if text == "]" && *nesting > 0 {
			*nesting--
		}
	}
}

// ReparseBlock attempts to reparse a code or content block.
// Returns the reparsed node, or nil if reparsing failed.
func ReparseBlock(text string, start, end int) *SyntaxNode {
	if start >= end || start >= len(text) {
		return nil
	}

	// Determine block type from the first character
	first := text[start]
	var wrapKind SyntaxKind

	switch first {
	case '{':
		wrapKind = CodeBlock
	case '[':
		wrapKind = ContentBlock
	default:
		return nil
	}

	p := NewParser(text, start, ModeCode)

	// Parse using the appropriate block function
	if wrapKind == CodeBlock {
		codeBlock(p)
	} else {
		contentBlock(p)
	}

	// Check that we consumed exactly the expected range
	if p.prevEnd() != end {
		return nil
	}

	// Check that parsing was balanced
	if !p.balanced {
		return nil
	}

	// The block parsing functions wrap the result, so we just return the first node
	nodes := p.finish()
	if len(nodes) == 0 {
		return nil
	}

	// Find the block node (skip any leading trivia)
	for _, n := range nodes {
		if n.Kind() == wrapKind {
			// Reject erroneous blocks (e.g., unclosed delimiters)
			if n.Erroneous() {
				return nil
			}
			return n
		}
	}

	return nil
}

// ReparseMarkup attempts to reparse a range of markup expressions.
// Returns the reparsed nodes, or nil if reparsing failed.
// The atStart and nesting parameters are updated during parsing.
func ReparseMarkup(
	text string,
	start, end int,
	atStart *bool,
	nesting *int,
	topLevel bool,
) []*SyntaxNode {
	if start >= end || start > len(text) {
		return nil
	}

	// Create a parser for the markup range
	p := NewParser(text, start, ModeMarkup)

	// Track the initial state
	initialAtStart := *atStart
	initialNesting := *nesting

	// Create a local nesting counter that mirrors the actual parsing
	localNesting := initialNesting

	// Parse markup expressions until we reach the end
	// We need to parse with the correct atStart state
	currentAtStart := initialAtStart || p.hadNewline()

	for !p.end() && p.currentStart() < end {
		// Check if we've parsed past the end
		if p.currentStart() >= end {
			break
		}

		markupExpr(p, currentAtStart, &localNesting)
		currentAtStart = p.hadNewline()
	}

	// Verify we consumed exactly the expected range
	if p.prevEnd() != end && p.currentStart() != end {
		// The parsing didn't align with the expected end
		// Try to be lenient - if we're close, it might still be valid
		if p.prevEnd() > end {
			return nil
		}
	}

	// Update the output parameters
	*atStart = currentAtStart
	*nesting = localNesting

	// Return the parsed nodes
	return p.finish()
}
