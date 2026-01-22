// Content type for Typst.
// Translated from foundations/content/mod.rs

package foundations

// Content represents typeset content.
type Content struct {
	// Elements contains the content elements.
	Elements []ContentElement
}

// ContentElement is a placeholder interface for content elements.
// IsContentElement is exported to allow cross-package type assertions.
type ContentElement interface {
	IsContentElement()
}

// ContentValue represents content as a Value.
type ContentValue struct {
	Content Content
}

func (ContentValue) Type() Type         { return TypeContent }
func (v ContentValue) Display() Content { return v.Content }
func (v ContentValue) Clone() Value     { return v } // TODO: deep clone
func (ContentValue) isValue()           {}

// StyledElem is content alongside styles.
// Corresponds to Rust's StyledElem in foundations/content/mod.rs.
type StyledElem struct {
	// Child is the content being styled.
	Child Content
	// Styles are the styles to apply.
	Styles *Styles
}

func (*StyledElem) IsContentElement() {}

// StyledWithMap wraps content with a style map.
// Corresponds to Rust's Content::styled_with_map.
func StyledWithMap(content Content, styles *Styles) Content {
	if styles == nil || styles.IsEmpty() {
		return content
	}
	return Content{
		Elements: []ContentElement{&StyledElem{
			Child:  content,
			Styles: styles,
		}},
	}
}

// SequenceElem is a sequence of content elements.
// Corresponds to Rust's SequenceElem in foundations/content/mod.rs.
type SequenceElem struct {
	// Children are the content elements in sequence.
	Children []Content
}

func (*SequenceElem) IsContentElement() {}

// SymbolElem represents a symbol in math mode.
// Corresponds to Rust's SymbolElem in typst-library/src/text/symbol.rs.
type SymbolElem struct {
	// Text is the symbol text/character.
	Text string
}

func (*SymbolElem) IsContentElement() {}
