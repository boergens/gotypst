package realize

import (
	"github.com/boergens/gotypst/eval"
)

// ----------------------------------------------------------------------------
// Space Collapsing
// ----------------------------------------------------------------------------
// Space collapsing is applied during realization to handle whitespace properly.
// This matches Rust's typst-realize/src/spaces.rs implementation.

// SpaceState categorizes elements for space collapsing.
// Matches Rust: SpaceState enum in typst-realize/src/spaces.rs
type SpaceState int

const (
	// StateInvisible - Elements that don't affect space collapsing (tags, metadata).
	StateInvisible SpaceState = iota
	// StateDestructive - Elements that discard adjacent spaces (block elements).
	StateDestructive
	// StateSupportive - Normal elements requiring spaces on both sides.
	StateSupportive
	// StateSpace - Space elements that can collapse with adjacent spaces.
	StateSpace
)

// getSpaceState returns the space state for an element.
// Matches Rust: impl SpaceState for various element types
func getSpaceState(elem eval.ContentElement) SpaceState {
	if elem == nil {
		return StateInvisible
	}

	switch e := elem.(type) {
	// Space elements
	case *eval.SpaceElement:
		return StateSpace

	// Text elements: check if whitespace-only
	case *eval.TextElement:
		if isWhitespaceOnly(e.Text) {
			return StateSpace
		}
		return StateSupportive

	// Destructive: block-level and paragraph boundary elements
	case *eval.ParbreakElement, *eval.LinebreakElement:
		return StateDestructive

	case *eval.ParagraphElement, *eval.HeadingElement:
		return StateDestructive

	case *eval.ListItemElement, *eval.EnumItemElement, *eval.TermItemElement:
		return StateDestructive

	case *eval.ListElement, *eval.EnumElement, *eval.TermsElement:
		return StateDestructive

	case *eval.BlockElement:
		return StateDestructive

	case *eval.VElem:
		return StateDestructive

	// Tags are invisible (pass through)
	case *eval.TagElem:
		return StateInvisible

	// Sequences need inspection (but treat as supportive for now)
	case *eval.SequenceElem:
		return StateSupportive

	// Styled content is supportive
	case *eval.StyledElement:
		return StateSupportive

	// Inline elements are supportive
	case *eval.StrongElement, *eval.EmphElement, *eval.RawElement:
		return StateSupportive

	case *eval.LinkElement, *eval.RefElement, *eval.SmartQuoteElement:
		return StateSupportive

	case *eval.HElem, *eval.BoxElement, *eval.InlineElem:
		return StateSupportive

	case *eval.EquationElement:
		return StateSupportive

	default:
		return StateSupportive
	}
}

// isWhitespaceOnly checks if a string contains only whitespace.
func isWhitespaceOnly(s string) bool {
	for _, r := range s {
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
			return false
		}
	}
	return len(s) > 0
}

// collapseSpaces collapses spaces within a slice of pairs starting from an offset.
// This modifies the slice in-place.
// Matches Rust: collapse_spaces() in typst-realize/src/spaces.rs
func collapseSpaces(pairs []Pair, start int) {
	if len(pairs) <= start {
		return
	}

	// Work on the slice from start onwards
	work := pairs[start:]
	if len(work) == 0 {
		return
	}

	write := 0
	lastState := StateDestructive // Treat start as destructive (no leading spaces)
	pendingSpace := -1            // Index of pending space in work slice

	for i := 0; i < len(work); i++ {
		state := getSpaceState(work[i].Content)

		switch state {
		case StateInvisible:
			// Always keep invisible elements, copy to write position
			if write != i {
				work[write] = work[i]
			}
			write++

		case StateSpace:
			// Space handling: collapse multiple spaces
			if lastState == StateDestructive || lastState == StateSpace {
				// Skip: adjacent to destructive or another space
				continue
			}
			// Mark as pending (may be removed if followed by destructive)
			pendingSpace = write
			if write != i {
				work[write] = work[i]
			}
			write++
			lastState = StateSpace

		case StateDestructive:
			// Remove pending space if any
			if pendingSpace >= 0 && pendingSpace < write {
				// Remove the space at pendingSpace by shifting
				copy(work[pendingSpace:], work[pendingSpace+1:write])
				write--
			}
			pendingSpace = -1

			if write != i {
				work[write] = work[i]
			}
			write++
			lastState = StateDestructive

		case StateSupportive:
			pendingSpace = -1
			if write != i {
				work[write] = work[i]
			}
			write++
			lastState = StateSupportive
		}
	}

	// Remove trailing space
	if write > 0 && getSpaceState(work[write-1].Content) == StateSpace {
		write--
	}

	// Truncate by zeroing out the unused portion (can't actually resize)
	// The caller is responsible for respecting the logical length
	for i := write; i < len(work); i++ {
		work[i] = Pair{}
	}
}

// normalizeSpaces normalizes whitespace within text elements.
// This collapses multiple spaces/tabs/newlines into single spaces.
func normalizeSpaces(text string) string {
	if text == "" {
		return text
	}

	result := make([]byte, 0, len(text))
	prevWasSpace := true // Treat start as having space to trim leading

	for i := 0; i < len(text); i++ {
		c := text[i]
		isSpace := c == ' ' || c == '\t' || c == '\n' || c == '\r'

		if isSpace {
			if !prevWasSpace {
				result = append(result, ' ')
				prevWasSpace = true
			}
		} else {
			result = append(result, c)
			prevWasSpace = false
		}
	}

	// Trim trailing space
	if len(result) > 0 && result[len(result)-1] == ' ' {
		result = result[:len(result)-1]
	}

	return string(result)
}

// trimLeadingSpace trims leading whitespace from the first text element.
func trimLeadingSpace(pairs []Pair) []Pair {
	if len(pairs) == 0 {
		return pairs
	}

	for i := range pairs {
		state := getSpaceState(pairs[i].Content)
		if state == StateInvisible {
			continue
		}
		if state == StateSpace {
			// Remove this space element
			return append(pairs[:i], pairs[i+1:]...)
		}
		if text, ok := pairs[i].Content.(*eval.TextElement); ok {
			// Trim leading space from text
			trimmed := trimLeadingWhitespace(text.Text)
			if trimmed != text.Text {
				pairs[i].Content = &eval.TextElement{Text: trimmed}
			}
		}
		break
	}
	return pairs
}

// trimTrailingSpace trims trailing whitespace from the last text element.
func trimTrailingSpace(pairs []Pair) []Pair {
	if len(pairs) == 0 {
		return pairs
	}

	for i := len(pairs) - 1; i >= 0; i-- {
		state := getSpaceState(pairs[i].Content)
		if state == StateInvisible {
			continue
		}
		if state == StateSpace {
			// Remove this space element
			return pairs[:i]
		}
		if text, ok := pairs[i].Content.(*eval.TextElement); ok {
			// Trim trailing space from text
			trimmed := trimTrailingWhitespace(text.Text)
			if trimmed != text.Text {
				pairs[i].Content = &eval.TextElement{Text: trimmed}
			}
		}
		break
	}
	return pairs
}

// trimLeadingWhitespace removes leading whitespace from a string.
func trimLeadingWhitespace(s string) string {
	for i, r := range s {
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
			return s[i:]
		}
	}
	return ""
}

// trimTrailingWhitespace removes trailing whitespace from a string.
func trimTrailingWhitespace(s string) string {
	for i := len(s) - 1; i >= 0; i-- {
		r := s[i]
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
			return s[:i+1]
		}
	}
	return ""
}
