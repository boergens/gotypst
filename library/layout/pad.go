// Package layout provides layout elements for Typst documents.
// This corresponds to typst-library/src/layout/ in the Rust implementation.
package layout

import (
	"github.com/boergens/gotypst/library/foundations"
)

// PadElement represents a padding container element.
// It adds spacing around its content.
//
// Reference: typst-reference/crates/typst-library/src/layout/pad.rs
type PadElement struct {
	// Left padding.
	Left *foundations.Length `typst:"left,type=length"`
	// Top padding.
	Top *foundations.Length `typst:"top,type=length"`
	// Right padding.
	Right *foundations.Length `typst:"right,type=length"`
	// Bottom padding.
	Bottom *foundations.Length `typst:"bottom,type=length"`
	// Body is the content to pad.
	Body foundations.Content `typst:"body,positional,required"`
}

func (*PadElement) IsContentElement() {}

// PadShorthands defines shorthand argument expansions for pad().
// - "rest" applies to all four sides
// - "x" applies to left and right
// - "y" applies to top and bottom
var PadShorthands = map[string][]string{
	"rest": {"left", "top", "right", "bottom"},
	"x":    {"left", "right"},
	"y":    {"top", "bottom"},
}

// PadShorthandOrder defines the order in which shorthands are processed.
// Earlier shorthands can be overridden by later ones.
var PadShorthandOrder = []string{"rest", "x", "y"}

// PadDef is the registered element definition for pad.
var PadDef *foundations.ElementDef

func init() {
	PadDef = foundations.RegisterElementWithOrder[PadElement](
		"pad",
		PadShorthands,
		PadShorthandOrder,
	)
}

// LeftPts returns the left padding in points, or 0 if not set.
func (p *PadElement) LeftPts() float64 {
	if p.Left == nil {
		return 0
	}
	return p.Left.Points
}

// TopPts returns the top padding in points, or 0 if not set.
func (p *PadElement) TopPts() float64 {
	if p.Top == nil {
		return 0
	}
	return p.Top.Points
}

// RightPts returns the right padding in points, or 0 if not set.
func (p *PadElement) RightPts() float64 {
	if p.Right == nil {
		return 0
	}
	return p.Right.Points
}

// BottomPts returns the bottom padding in points, or 0 if not set.
func (p *PadElement) BottomPts() float64 {
	if p.Bottom == nil {
		return 0
	}
	return p.Bottom.Points
}
