package layout

import "math"

// Abs represents an absolute length in typographic points (1/72 inch).
type Abs float64

// Zero returns true if the value is zero (within floating point tolerance).
func (a Abs) Zero() bool {
	return a == 0
}

// ApproxEq checks if two Abs values are approximately equal.
func (a Abs) ApproxEq(other Abs) bool {
	const epsilon = 1e-9
	return math.Abs(float64(a-other)) < epsilon
}

// Fits returns true if other fits within this Abs value.
func (a Abs) Fits(other Abs) bool {
	return other <= a || a.ApproxEq(other)
}

// At converts an Em value to Abs at the given font size.
func (e Em) At(fontSize Abs) Abs {
	return Abs(float64(e) * float64(fontSize))
}

// Em represents a font-relative length (em units).
type Em float64

// NewEm creates a new Em value.
func NewEm(v float64) Em {
	return Em(v)
}

// Fr represents a fractional unit for flexible spacing.
type Fr float64

// Point represents a 2D coordinate.
type Point struct {
	X, Y Abs
}

// Size represents 2D dimensions.
type Size struct {
	Width, Height Abs
}

// Dir represents text direction.
type Dir int

const (
	DirLTR Dir = iota
	DirRTL
)

// Start returns the start alignment for this direction.
func (d Dir) Start() Alignment {
	if d == DirRTL {
		return AlignEnd
	}
	return AlignStart
}

// Alignment represents horizontal alignment.
type Alignment int

const (
	AlignStart Alignment = iota
	AlignCenter
	AlignEnd
)

// Linebreaks represents the line breaking algorithm mode.
type Linebreaks int

const (
	LinebreaksSimple Linebreaks = iota
	LinebreaksOptimized
)

// Lang represents a language tag.
type Lang string

// Common language constants.
const (
	LangChinese  Lang = "zh"
	LangJapanese Lang = "ja"
)
