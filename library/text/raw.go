// Raw text element for Typst.
// Translated from typst-library/src/text/raw.rs

package text

import "github.com/boergens/gotypst/library/foundations"

// RawElem represents raw text with optional syntax highlighting.
//
// Displays the text verbatim and in a monospace font. This is typically used
// to embed computer code into your document.
//
// Corresponds to Rust's RawElem in text/raw.rs.
type RawElem struct {
	// Text is the raw text content.
	// Required field.
	Text string

	// Block indicates whether the raw text is displayed as a separate block.
	// In markup mode, using one-backtick notation makes this false.
	// Using three-backtick notation makes it true if the enclosed content
	// contains at least one line break.
	// Default: false
	Block bool

	// Lang is the language to syntax-highlight in.
	// Empty string means no syntax highlighting.
	Lang string

	// Align is the horizontal alignment for each line in a raw block.
	// Default: "start"
	Align string
}

func (*RawElem) IsContentElement() {}

// RawLineElem represents a single line of raw text.
// Used for custom styling of individual lines via show rules.
// Corresponds to Rust's RawLine in text/raw.rs.
type RawLineElem struct {
	// Number is the line number (1-indexed).
	Number int

	// Count is the total number of lines.
	Count int

	// Text is the line content.
	Text string

	// Body is the styled line content.
	Body foundations.Content
}

func (*RawLineElem) IsContentElement() {}
