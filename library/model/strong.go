// Strong emphasis element for Typst.
// Translated from typst-library/src/model/strong.rs

package model

import "github.com/boergens/gotypst/library/foundations"

// StrongElem strongly emphasizes content by increasing the font weight.
//
// Increases the current font weight by a given delta.
// Corresponds to Rust's StrongElem in model/strong.rs.
type StrongElem struct {
	// Delta is the amount to apply on the font weight.
	// Default: 300
	Delta int64

	// Body is the content to strongly emphasize.
	// Required field.
	Body foundations.Content
}

func (*StrongElem) IsContentElement() {}

// DefaultStrongDelta is the default font weight increase for strong emphasis.
const DefaultStrongDelta = 300
