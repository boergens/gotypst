package flow

import (
	"github.com/boergens/gotypst/layout"
)

// Insertions is an additive collection of out-of-flow insertions.
// It tracks floats (top and bottom) and footnotes that need to be
// placed around the main content.
type Insertions struct {
	// topFloats contains floats positioned at the top.
	topFloats []floatPair
	// bottomFloats contains floats positioned at the bottom.
	bottomFloats []floatPair
	// footnotes contains footnote frames.
	footnotes []*Frame
	// footnoteSeparator is the optional separator frame.
	footnoteSeparator *Frame
	// topSize is the total height of top floats.
	topSize layout.Abs
	// bottomSize is the total height of bottom floats and footnotes.
	bottomSize layout.Abs
	// width tracks the maximum width needed.
	width layout.Abs
	// skips tracks locations that have been processed.
	skips []Location
}

// floatPair associates a placed child with its laid out frame.
type floatPair struct {
	placed *PlacedChild
	frame  *Frame
}

// NewInsertions creates an empty Insertions.
func NewInsertions() *Insertions {
	return &Insertions{
		topFloats:    nil,
		bottomFloats: nil,
		footnotes:    nil,
		skips:        nil,
	}
}

// PushFloat adds a float to the top or bottom area based on alignment.
func (ins *Insertions) PushFloat(placed *PlacedChild, frame *Frame, alignY FixedAlignment) {
	ins.width = maxAbs(ins.width, frame.Width())

	amount := frame.Height() + placed.Clearance
	pair := floatPair{placed: placed, frame: frame}

	if alignY == FixedAlignmentStart {
		ins.topSize += amount
		ins.topFloats = append(ins.topFloats, pair)
	} else {
		ins.bottomSize += amount
		ins.bottomFloats = append(ins.bottomFloats, pair)
	}
}

// PushFootnote adds a footnote frame to the bottom area.
func (ins *Insertions) PushFootnote(config *Config, frame *Frame) {
	ins.width = maxAbs(ins.width, frame.Width())
	ins.bottomSize += config.Footnote.Gap + frame.Height()
	ins.footnotes = append(ins.footnotes, frame)
}

// PushFootnoteSeparator adds the footnote separator to the bottom area.
func (ins *Insertions) PushFootnoteSeparator(config *Config, frame *Frame) {
	ins.width = maxAbs(ins.width, frame.Width())
	ins.bottomSize += config.Footnote.Clearance + frame.Height()
	ins.footnoteSeparator = frame
}

// Height returns the combined height of top and bottom areas.
func (ins *Insertions) Height() layout.Abs {
	return ins.topSize + ins.bottomSize
}

// Width returns the maximum width needed by insertions.
func (ins *Insertions) Width() layout.Abs {
	return ins.width
}

// Skips returns the list of processed locations.
func (ins *Insertions) Skips() []Location {
	return ins.skips
}

// AddSkip adds a location to the skip list.
func (ins *Insertions) AddSkip(loc Location) {
	ins.skips = append(ins.skips, loc)
}

// IsEmpty returns true if there are no insertions.
func (ins *Insertions) IsEmpty() bool {
	return len(ins.topFloats) == 0 &&
		len(ins.bottomFloats) == 0 &&
		ins.footnoteSeparator == nil &&
		len(ins.footnotes) == 0
}

// Finalize produces the final frame by combining inner content with insertions.
// The inner frame is positioned between top floats and bottom floats/footnotes.
func (ins *Insertions) Finalize(work *Work, config *Config, inner *Frame) *Frame {
	// Record skipped locations in work
	work.ExtendSkips(ins.skips)

	// If no insertions, return inner frame unchanged
	if ins.IsEmpty() {
		return inner
	}

	// Calculate output size
	size := layout.Size{
		Width:  inner.Width(),
		Height: inner.Height() + ins.Height(),
	}

	output := NewSoftFrame(size)
	offsetTop := layout.Abs(0)
	offsetBottom := size.Height - ins.bottomSize

	// Place top floats
	for _, pair := range ins.topFloats {
		x := pair.placed.AlignX.Position(size.Width - pair.frame.Width())
		y := offsetTop
		delta := pair.placed.Delta.ZipMap(size)
		offsetTop += pair.frame.Height() + pair.placed.Clearance
		output.PushFrame(layout.Point{X: x + delta.X, Y: y + delta.Y}, pair.frame)
	}

	// Place inner content
	output.PushFrame(layout.Point{X: 0, Y: ins.topSize}, inner)

	// Place bottom floats
	for _, pair := range ins.bottomFloats {
		offsetBottom += pair.placed.Clearance
		x := pair.placed.AlignX.Position(size.Width - pair.frame.Width())
		y := offsetBottom
		delta := pair.placed.Delta.ZipMap(size)
		offsetBottom += pair.frame.Height()
		output.PushFrame(layout.Point{X: x + delta.X, Y: y + delta.Y}, pair.frame)
	}

	// Place footnote separator
	if ins.footnoteSeparator != nil {
		offsetBottom += config.Footnote.Clearance
		y := offsetBottom
		offsetBottom += ins.footnoteSeparator.Height()
		output.PushFrame(layout.Point{X: 0, Y: y}, ins.footnoteSeparator)
	}

	// Place footnotes
	for _, frame := range ins.footnotes {
		offsetBottom += config.Footnote.Gap
		y := offsetBottom
		offsetBottom += frame.Height()
		output.PushFrame(layout.Point{X: 0, Y: y}, frame)
	}

	return output
}

// Reset clears all insertions for reuse.
func (ins *Insertions) Reset() {
	ins.topFloats = ins.topFloats[:0]
	ins.bottomFloats = ins.bottomFloats[:0]
	ins.footnotes = ins.footnotes[:0]
	ins.footnoteSeparator = nil
	ins.topSize = 0
	ins.bottomSize = 0
	ins.width = 0
	ins.skips = ins.skips[:0]
}

// maxAbs returns the larger of two Abs values.
func maxAbs(a, b layout.Abs) layout.Abs {
	if a > b {
		return a
	}
	return b
}

// setMaxAbs updates a if b is larger.
func setMaxAbs(a *layout.Abs, b layout.Abs) {
	if b > *a {
		*a = b
	}
}
