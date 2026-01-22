package layout

import (
	"github.com/boergens/gotypst/library/foundations"
)

// BoxElement represents an inline box container element.
// It can size its content, apply fills/strokes, and clip overflow.
//
// Reference: typst-reference/crates/typst-library/src/layout/container.rs
type BoxElement struct {
	// Width of the box (in points). If nil, auto-sizes to content.
	Width *float64
	// Height of the box (in points). If nil, auto-sizes to content.
	Height *float64
	// Baseline position (in points from bottom). If nil, uses content baseline.
	Baseline *float64
	// Fill color for the background. If nil, no fill.
	Fill foundations.Value
	// Stroke for the border. Can be length, color, or stroke dict. If nil, no stroke.
	Stroke foundations.Value
	// Radius for rounded corners. Can be single value or dictionary.
	Radius foundations.Value
	// Inset padding inside the box.
	Inset foundations.Value
	// Outset expansion outside the box.
	Outset foundations.Value
	// Whether to clip content that overflows the box.
	Clip bool
	// Body is the content inside the box.
	Body foundations.Content
}

func (*BoxElement) IsContentElement() {}

// BlockElement represents a block-level container element.
// It creates a new block in the document flow with optional sizing and styling.
//
// Reference: typst-reference/crates/typst-library/src/layout/container.rs
type BlockElement struct {
	// Width of the block (in points). If nil, auto-sizes.
	Width *float64
	// Height of the block (in points). If nil, auto-sizes.
	Height *float64
	// Whether the block can break across pages.
	Breakable *bool
	// Fill color for the background.
	Fill foundations.Value
	// Stroke for the border.
	Stroke foundations.Value
	// Radius for rounded corners.
	Radius foundations.Value
	// Inset padding inside the block.
	Inset foundations.Value
	// Outset expansion outside the block.
	Outset foundations.Value
	// Spacing between adjacent blocks.
	Spacing *float64
	// Spacing above this block (overrides Spacing).
	Above *float64
	// Spacing below this block (overrides Spacing).
	Below *float64
	// Whether to clip content that overflows.
	Clip bool
	// Whether the block sticks to the next block.
	Sticky bool
	// Body is the content inside the block.
	Body foundations.Content
}

func (*BlockElement) IsContentElement() {}
