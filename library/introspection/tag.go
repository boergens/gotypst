// Tag and location types for document introspection.
// Translated from typst-library/src/introspection/tag.rs and location.rs

package introspection

import "github.com/boergens/gotypst/library/foundations"

// Tag represents a start or end marker for content in the document.
// Tags are used to track element locations after layout for introspection.
type Tag struct {
	// Kind is "start" or "end".
	Kind TagKind
	// Location uniquely identifies the tagged element.
	Location Location
	// Elem is the element being tagged (for start tags).
	Elem foundations.ContentElement
	// Key is a hash key for end tags.
	Key uint64
	// Flags contains tag flags.
	Flags TagFlags
}

// TagKind indicates whether a tag is a start or end tag.
type TagKind int

const (
	TagStart TagKind = iota
	TagEnd
)

// TagFlags contains flags for tag behavior.
type TagFlags struct {
	// Introspectable indicates the element can be found via introspection.
	Introspectable bool
	// Tagged indicates the element is semantically tagged (for accessibility).
	Tagged bool
}

// Any returns true if any flag is set.
func (f TagFlags) Any() bool {
	return f.Introspectable || f.Tagged
}

// Location uniquely identifies an element in the document.
type Location struct {
	// Hash is the unique hash for this location.
	Hash uint64
	// Variant differentiates multiple elements with the same hash.
	Variant uint32
}

// TagElem represents a tag element in content.
type TagElem struct {
	Tag Tag
}

func (*TagElem) IsContentElement() {}

// NewStartTag creates a start tag for an element.
func NewStartTag(elem foundations.ContentElement, loc Location, flags TagFlags) *TagElem {
	return &TagElem{
		Tag: Tag{
			Kind:     TagStart,
			Location: loc,
			Elem:     elem,
			Flags:    flags,
		},
	}
}

// NewEndTag creates an end tag for an element.
func NewEndTag(loc Location, key uint64, flags TagFlags) *TagElem {
	return &TagElem{
		Tag: Tag{
			Kind:     TagEnd,
			Location: loc,
			Key:      key,
			Flags:    flags,
		},
	}
}
