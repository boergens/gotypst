package layout

import (
	"github.com/boergens/gotypst/library/foundations"
)

// Alignment2D represents a 2D alignment value (horizontal and vertical).
// Reference: typst-reference/crates/typst-library/src/layout/align.rs
type Alignment2D struct {
	// Horizontal alignment (left, center, right, start, end, or nil for not specified).
	Horizontal *HAlignment
	// Vertical alignment (top, horizon, bottom, or nil for not specified).
	Vertical *VAlignment
}

// HAlignment represents horizontal alignment values.
type HAlignment string

const (
	HAlignStart  HAlignment = "start"
	HAlignLeft   HAlignment = "left"
	HAlignCenter HAlignment = "center"
	HAlignRight  HAlignment = "right"
	HAlignEnd    HAlignment = "end"
)

// VAlignment represents vertical alignment values.
type VAlignment string

const (
	VAlignTop     VAlignment = "top"
	VAlignHorizon VAlignment = "horizon"
	VAlignBottom  VAlignment = "bottom"
)

// AlignElement represents an alignment container element.
// It positions its content according to the specified alignment.
//
// Reference: typst-reference/crates/typst-library/src/layout/align.rs
type AlignElement struct {
	// Alignment is the 2D alignment specification.
	Alignment Alignment2D
	// Body is the content to align.
	Body foundations.Content
}

func (*AlignElement) IsContentElement() {}
