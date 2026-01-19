package layout

import (
	"github.com/boergens/gotypst/syntax"
)

// PadElem represents the padding element configuration.
type PadElem struct {
	Left   Rel
	Top    Rel
	Right  Rel
	Bottom Rel
	Body   interface{}
	Span   syntax.Span
}

// Rel represents a relative length (combination of absolute and ratio).
type Rel struct {
	Abs Abs
	Rel Ratio
}

// RelZero returns a zero relative length.
func RelZero() Rel { return Rel{} }

// RelAbs creates a Rel from an absolute length.
func RelAbs(a Abs) Rel { return Rel{Abs: a} }

// RelRatio creates a Rel from a ratio.
func RelRatio(r Ratio) Rel { return Rel{Rel: r} }

// RelativeTo resolves the relative length given a parent size.
func (r Rel) RelativeTo(parent Abs) Abs {
	return r.Abs + r.Rel.Of(parent)
}

// IsZero returns true if the relative length is zero.
func (r Rel) IsZero() bool {
	return r.Abs.IsZero() && r.Rel.IsZero()
}

// SidesRel represents padding/margin sides with relative lengths.
type SidesRel struct {
	Left   Rel
	Top    Rel
	Right  Rel
	Bottom Rel
}

// NewSidesRel creates SidesRel with relative values for each side.
func NewSidesRel(left, top, right, bottom Rel) *SidesRel {
	return &SidesRel{Left: left, Top: top, Right: right, Bottom: bottom}
}

// SumByAxis returns the sum of opposite sides as X (horizontal) and Y (vertical).
func (s *SidesRel) SumByAxis() Axes[Rel] {
	return Axes[Rel]{
		X: Rel{Abs: s.Left.Abs + s.Right.Abs, Rel: s.Left.Rel + s.Right.Rel},
		Y: Rel{Abs: s.Top.Abs + s.Bottom.Abs, Rel: s.Top.Rel + s.Bottom.Rel},
	}
}

// RelativeTo resolves all sides relative to a given size.
func (s *SidesRel) RelativeTo(size Size) Sides[Abs] {
	return Sides[Abs]{
		Left:   s.Left.RelativeTo(size.X),
		Top:    s.Top.RelativeTo(size.Y),
		Right:  s.Right.RelativeTo(size.X),
		Bottom: s.Bottom.RelativeTo(size.Y),
	}
}

// LayoutPad lays out padded content.
//
// This function applies padding by:
// 1. Shrinking the available regions by the padding amount
// 2. Laying out the child content in the reduced space
// 3. Growing the resulting frames back outward to accommodate the padding
func LayoutPad(
	elem *PadElem,
	layoutFunc LayoutFunc,
	styles StyleChain,
	regions *Regions,
) (Fragment, error) {
	padding := NewSidesRel(elem.Left, elem.Top, elem.Right, elem.Bottom)

	// Create shrunk regions for child layout.
	shrunkRegions := shrinkRegions(regions, padding)

	// Layout child into padded regions.
	fragment, err := layoutFunc(elem.Body, styles, shrunkRegions)
	if err != nil {
		return nil, err
	}

	// Grow each frame to include the padding.
	result := make(Fragment, len(fragment))
	for i := range fragment {
		frame := fragment[i]
		Grow(&frame, padding)
		result[i] = frame
	}

	return result, nil
}

// shrinkRegions creates new regions shrunk by the given inset.
func shrinkRegions(regions *Regions, inset *SidesRel) *Regions {
	result := &Regions{
		Size:    Shrink(regions.Size, inset),
		Full:    Shrink(regions.Full, inset),
		Backlog: make([]Size, len(regions.Backlog)),
		Last:    regions.Last,
		Expand:  regions.Expand,
	}

	summed := inset.SumByAxis()
	for i, sz := range regions.Backlog {
		result.Backlog[i] = Size{
			X: sz.X - summed.X.RelativeTo(sz.X),
			Y: sz.Y - summed.Y.RelativeTo(sz.Y),
		}
	}

	return result
}

// Shrink shrinks a region size by an inset relative to the size itself.
func Shrink(size Size, inset *SidesRel) Size {
	summed := inset.SumByAxis()
	return Size{
		X: size.X - summed.X.RelativeTo(size.X),
		Y: size.Y - summed.Y.RelativeTo(size.Y),
	}
}

// ShrinkMultiple shrinks the components of possibly multiple Regions by an inset
// relative to the regions themselves.
func ShrinkMultiple(
	size *Size,
	full *Abs,
	backlog []Abs,
	last *Abs,
	inset *SidesRel,
) {
	summed := inset.SumByAxis()

	size.X -= summed.X.RelativeTo(size.X)
	size.Y -= summed.Y.RelativeTo(size.Y)

	*full -= summed.Y.RelativeTo(*full)

	for i := range backlog {
		backlog[i] -= summed.Y.RelativeTo(backlog[i])
	}

	if last != nil {
		*last -= summed.Y.RelativeTo(*last)
	}
}

// Grow grows a frame's size by an inset relative to the grown size.
// This is the inverse operation to Shrink.
//
// For the horizontal axis the derivation looks as follows.
// (Vertical axis is analogous.)
//
// Let w be the grown target width,
//
//	s be the given width,
//	l be the left inset,
//	r be the right inset,
//	p = l + r.
//
// We want that: w - l.resolve(w) - r.resolve(w) = s
//
// Thus: w - l.resolve(w) - r.resolve(w) = s
//
//	<=> w - p.resolve(w) = s
//	<=> w - p.rel * w - p.abs = s
//	<=> (1 - p.rel) * w = s + p.abs
//	<=> w = (s + p.abs) / (1 - p.rel)
func Grow(frame *Frame, inset *SidesRel) {
	// Apply the padding inversely such that the grown size padded
	// yields the frame's size.
	summed := inset.SumByAxis()
	padded := Size{
		X: growDimension(frame.Size.X, summed.X),
		Y: growDimension(frame.Size.Y, summed.Y),
	}

	// Resolve inset relative to the padded size.
	resolved := inset.RelativeTo(padded)
	offset := NewPoint(resolved.Left, resolved.Top)

	// Grow the frame and translate everything in the frame inwards.
	frame.Size = padded
	frame.Translate(offset)
}

// growDimension calculates the grown dimension from the shrunken dimension.
// Implements: w = (s + p.abs) / (1 - p.rel)
func growDimension(s Abs, p Rel) Abs {
	divisor := 1.0 - float64(p.Rel)
	if divisor == 0 {
		return s + p.Abs // Avoid division by zero
	}
	return Abs((float64(s) + float64(p.Abs)) / divisor)
}
