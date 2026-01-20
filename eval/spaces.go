package eval

// ----------------------------------------------------------------------------
// Space Element and Space State
// ----------------------------------------------------------------------------

// SpaceElement represents a space in content.
// Unlike TextElement{Text: " "}, SpaceElement participates in space collapsing.
type SpaceElement struct {
	// Weak indicates if this is a weak space that collapses more readily.
	// Weak spaces are discarded when adjacent to destructive elements.
	Weak bool
}

func (*SpaceElement) isContentElement() {}

// SpaceState categorizes content elements for the space collapsing algorithm.
// The state determines how an element interacts with adjacent spaces.
type SpaceState int

const (
	// Invisible elements don't affect space collapsing.
	// Examples: tags, metadata markers, labels.
	// Spaces pass through invisible elements as if they weren't there.
	Invisible SpaceState = iota

	// Destructive elements discard adjacent spaces.
	// Examples: parbreaks, headings, block-level elements.
	// When a space is adjacent to a destructive element, the space is removed.
	Destructive

	// Supportive elements require spaces on both sides to be preserved.
	// Examples: text, raw content, inline elements.
	// Spaces adjacent to supportive elements are kept (unless collapsed with other spaces).
	Supportive

	// Space elements can collapse with adjacent spaces.
	// Multiple consecutive spaces collapse into a single space.
	// Examples: SpaceElement, certain whitespace in math mode.
	Space
)

// GetSpaceState returns the SpaceState for a content element.
// This determines how the element participates in space collapsing.
func GetSpaceState(elem ContentElement) SpaceState {
	switch elem.(type) {
	// Space elements
	case *SpaceElement:
		return Space

	// Destructive elements - consume adjacent spaces
	case *ParbreakElement:
		return Destructive
	case *HeadingElement:
		return Destructive
	case *ListItemElement:
		return Destructive
	case *EnumItemElement:
		return Destructive
	case *TermItemElement:
		return Destructive
	case *ParagraphElement:
		return Destructive
	case *LinebreakElement:
		// Linebreaks are destructive - spaces before/after line breaks are removed
		return Destructive

	// Invisible elements - don't affect spacing
	// These are metadata/introspection elements that don't produce visible output
	// Currently we don't have many of these in gotypst, but the category exists
	// for tags and similar elements.

	// Supportive elements - everything else that produces visible content
	case *TextElement:
		return Supportive
	case *RawElement:
		return Supportive
	case *StrongElement:
		return Supportive
	case *EmphElement:
		return Supportive
	case *LinkElement:
		return Supportive
	case *RefElement:
		return Supportive
	case *SmartQuoteElement:
		return Supportive
	case *EquationElement:
		return Supportive
	case *MathFracElement:
		return Supportive
	case *MathRootElement:
		return Supportive
	case *MathAttachElement:
		return Supportive
	case *MathDelimitedElement:
		return Supportive
	case *MathAlignElement:
		return Supportive
	case *MathSymbolElement:
		return Supportive

	default:
		// Default to supportive - unknown elements should preserve spaces
		return Supportive
	}
}

// CollapseSpaces performs space collapsing on a slice of content elements.
// It modifies the slice in-place and returns the new length.
//
// The algorithm:
// 1. Removes spaces at the start and end of content
// 2. Collapses consecutive spaces into a single space
// 3. Removes spaces adjacent to destructive elements
// 4. Preserves spaces between supportive elements
//
// This uses a cursor-based in-place algorithm for efficiency.
func CollapseSpaces(elements []ContentElement) int {
	if len(elements) == 0 {
		return 0
	}

	// Write cursor - where we write kept elements
	write := 0
	// Previous non-invisible element's state
	prevState := Destructive // Start as if there was a destructive element before
	// Track if we have a pending space to potentially write
	pendingSpace := false

	for _, elem := range elements {
		state := GetSpaceState(elem)

		switch state {
		case Invisible:
			// Copy invisible elements as-is, they don't affect spacing
			elements[write] = elem
			write++

		case Space:
			// Only mark as pending if previous was supportive
			// (spaces after destructive elements are discarded)
			if prevState == Supportive {
				pendingSpace = true
			}
			// Don't update prevState for spaces - we'll handle them when we see the next element

		case Destructive:
			// Destructive elements discard pending spaces
			pendingSpace = false
			elements[write] = elem
			write++
			prevState = Destructive

		case Supportive:
			// Write pending space if we have one
			if pendingSpace {
				elements[write] = &SpaceElement{}
				write++
				pendingSpace = false
			}
			elements[write] = elem
			write++
			prevState = Supportive
		}
	}

	// No trailing spaces - pendingSpace is discarded at end

	// Clear remaining elements to allow GC
	for i := write; i < len(elements); i++ {
		elements[i] = nil
	}

	return write
}

// CollapseSpacesContent performs space collapsing on Content and returns new Content.
// This is a convenience wrapper around CollapseSpaces.
func CollapseSpacesContent(c Content) Content {
	if len(c.Elements) == 0 {
		return c
	}

	// Make a copy to avoid modifying the original
	elements := make([]ContentElement, len(c.Elements))
	copy(elements, c.Elements)

	newLen := CollapseSpaces(elements)
	return Content{Elements: elements[:newLen]}
}
