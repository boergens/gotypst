package eval

import (
	"github.com/boergens/gotypst/library/foundations"
	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Pad Element
// ----------------------------------------------------------------------------
// Reference: typst-reference/crates/typst-library/src/layout/pad.rs

// PadElement represents a padding container element.
// It adds spacing around its content.
//
// Uses the declarative element definition system with struct tags.
type PadElement struct {
	// Left padding.
	Left *Length `typst:"left,type=length"`
	// Top padding.
	Top *Length `typst:"top,type=length"`
	// Right padding.
	Right *Length `typst:"right,type=length"`
	// Bottom padding.
	Bottom *Length `typst:"bottom,type=length"`
	// Body is the content to pad.
	Body Content `typst:"body,positional,required"`
}

func (*PadElement) IsContentElement() {}

// PadShorthands defines shorthand argument expansions for pad().
// - "rest" applies to all four sides
// - "x" applies to left and right
// - "y" applies to top and bottom
//
// Shorthand order matters: rest is processed first, then x/y, so individual
// sides can override shorthands.
var PadShorthands = map[string][]string{
	"rest": {"left", "top", "right", "bottom"},
	"x":    {"left", "right"},
	"y":    {"top", "bottom"},
}

// PadShorthandOrder defines the order in which shorthands are processed.
// Earlier shorthands can be overridden by later ones.
var PadShorthandOrder = []string{"rest", "x", "y"}

// padElementDef is the registered element definition for pad.
var padElementDef *foundations.ElementDef

func init() {
	padElementDef = foundations.RegisterElementWithOrder[PadElement](
		"pad",
		PadShorthands,
		PadShorthandOrder,
	)
}

// PadFunc creates the pad element function using the declarative system.
func PadFunc() *Func {
	name := "pad"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: padNative,
			Info: padElementDef.ToFuncInfo(),
		},
	}
}

// padNative implements the pad() function using the generic element parser.
//
// Arguments:
//   - body (positional, content): The content to pad
//   - left (named, length): Left padding
//   - top (named, length): Top padding
//   - right (named, length): Right padding
//   - bottom (named, length): Bottom padding
//   - x (named, length): Horizontal padding (sets left and right)
//   - y (named, length): Vertical padding (sets top and bottom)
//   - rest (named, length): Padding for all sides
func padNative(engine foundations.Engine, context foundations.Context, args *Args) (Value, error) {
	elem, err := foundations.ParseElement[PadElement](padElementDef, args)
	if err != nil {
		return nil, err
	}

	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
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
