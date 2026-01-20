package realize

import (
	"github.com/boergens/gotypst/eval"
)

// SpaceState categorizes elements for space collapsing.
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
func getSpaceState(elem eval.ContentElement) SpaceState {
	if elem == nil {
		return StateInvisible
	}

	switch e := elem.(type) {
	case *eval.TextElement:
		// Check if it's whitespace-only
		if isWhitespaceOnly(e.Text) {
			return StateSpace
		}
		return StateSupportive

	case *eval.ParbreakElement, *eval.LinebreakElement:
		// Breaks are destructive to surrounding spaces
		return StateDestructive

	case *eval.ParagraphElement, *eval.HeadingElement,
		*eval.ListItemElement, *eval.EnumItemElement, *eval.TermItemElement:
		// Block elements are destructive
		return StateDestructive

	case *eval.StrongElement, *eval.EmphElement, *eval.RawElement,
		*eval.LinkElement, *eval.RefElement, *eval.SmartQuoteElement:
		// Inline elements are supportive
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

// collapseSpaces removes unnecessary spaces from realized pairs.
//
// The algorithm handles:
//  1. Removing spaces at content boundaries (start/end)
//  2. Collapsing adjacent spaces into single spaces
//  3. Removing spaces adjacent to destructive elements
//
// It operates in-place for efficiency (no allocations when possible).
func collapseSpaces(pairs []Pair) []Pair {
	if len(pairs) == 0 {
		return pairs
	}

	result := make([]Pair, 0, len(pairs))
	var lastState SpaceState = StateDestructive // Treat start as destructive

	for i, pair := range pairs {
		state := getSpaceState(pair.Element)

		switch state {
		case StateInvisible:
			// Always keep invisible elements
			result = append(result, pair)

		case StateSpace:
			// Space handling
			if lastState == StateDestructive || lastState == StateSpace {
				// Skip: adjacent to destructive or another space
				continue
			}
			// Look ahead to see if next non-invisible is destructive
			if nextIsDestructive(pairs[i+1:]) {
				continue
			}
			result = append(result, pair)
			lastState = StateSpace

		case StateDestructive:
			// Remove trailing space if present
			if len(result) > 0 && getSpaceState(result[len(result)-1].Element) == StateSpace {
				result = result[:len(result)-1]
			}
			result = append(result, pair)
			lastState = StateDestructive

		case StateSupportive:
			result = append(result, pair)
			lastState = StateSupportive
		}
	}

	// Remove trailing spaces
	for len(result) > 0 && getSpaceState(result[len(result)-1].Element) == StateSpace {
		result = result[:len(result)-1]
	}

	return result
}

// nextIsDestructive checks if the next visible element is destructive.
func nextIsDestructive(pairs []Pair) bool {
	for _, pair := range pairs {
		state := getSpaceState(pair.Element)
		if state == StateInvisible {
			continue
		}
		return state == StateDestructive
	}
	// End of content is destructive
	return true
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
		state := getSpaceState(pairs[i].Element)
		if state == StateInvisible {
			continue
		}
		if state == StateSpace {
			// Remove this space element
			return append(pairs[:i], pairs[i+1:]...)
		}
		if text, ok := pairs[i].Element.(*eval.TextElement); ok {
			// Trim leading space from text
			trimmed := trimLeadingWhitespace(text.Text)
			if trimmed != text.Text {
				pairs[i].Element = &eval.TextElement{Text: trimmed}
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
		state := getSpaceState(pairs[i].Element)
		if state == StateInvisible {
			continue
		}
		if state == StateSpace {
			// Remove this space element
			return pairs[:i]
		}
		if text, ok := pairs[i].Element.(*eval.TextElement); ok {
			// Trim trailing space from text
			trimmed := trimTrailingWhitespace(text.Text)
			if trimmed != text.Text {
				pairs[i].Element = &eval.TextElement{Text: trimmed}
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
