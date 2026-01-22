// Emphasis element for Typst.
// Translated from typst-library/src/model/emph.rs

package model

import "github.com/boergens/gotypst/library/foundations"

// EmphElem emphasizes content by toggling italics.
//
// - If the current text style is "normal", this turns it into "italic".
// - If it is already "italic" or "oblique", it turns it back to "normal".
//
// Corresponds to Rust's EmphElem in model/emph.rs.
type EmphElem struct {
	// Body is the content to emphasize.
	// Required field.
	Body foundations.Content
}

func (*EmphElem) IsContentElement() {}
