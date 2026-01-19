// Package layout provides the layout engine for typst documents.
//
// This package handles converting abstract document content into positioned
// frames ready for rendering. It implements block-level flow layout, inline
// text layout, grid layout, and geometric transformations.
package layout

import (
	"math"
)

// Abs represents an absolute length in points (1/72 inch).
type Abs float64

// Zero is the zero absolute length.
const Zero Abs = 0

// Pt creates an absolute length from points.
func Pt(pts float64) Abs {
	return Abs(pts)
}

// Points returns the length in points.
func (a Abs) Points() float64 {
	return float64(a)
}

// IsZero returns true if the length is approximately zero.
func (a Abs) IsZero() bool {
	return math.Abs(float64(a)) < 1e-9
}

// IsFinite returns true if the length is not infinite.
func (a Abs) IsFinite() bool {
	return !math.IsInf(float64(a), 0)
}

// Max returns the larger of two absolute lengths.
func (a Abs) Max(other Abs) Abs {
	if a > other {
		return a
	}
	return other
}

// Min returns the smaller of two absolute lengths.
func (a Abs) Min(other Abs) Abs {
	if a < other {
		return a
	}
	return other
}

// Clamp restricts the length to the given range.
func (a Abs) Clamp(min, max Abs) Abs {
	if a < min {
		return min
	}
	if a > max {
		return max
	}
	return a
}

// ApproxEq checks if two lengths are approximately equal.
func (a Abs) ApproxEq(other Abs) bool {
	return math.Abs(float64(a-other)) < 1e-9
}

// Fr represents a fractional unit for flexible sizing.
type Fr float64

// Em represents a font-relative unit.
type Em float64

// ToAbs converts Em to Abs given a font size.
func (e Em) ToAbs(fontSize Abs) Abs {
	return Abs(float64(e) * float64(fontSize))
}

// Ratio represents a normalized ratio value (typically 0-1).
type Ratio float64

// Rel represents a relative length that can contain both
// absolute and relative components.
type Rel struct {
	Abs Abs   // Absolute component
	Rel Ratio // Relative component (0-1)
}

// NewRel creates a relative length from absolute and ratio components.
func NewRel(abs Abs, rel Ratio) Rel {
	return Rel{Abs: abs, Rel: rel}
}

// RelAbs creates a purely absolute Rel value.
func RelAbs(abs Abs) Rel {
	return Rel{Abs: abs, Rel: 0}
}

// RelRatio creates a purely relative Rel value.
func RelRatio(ratio Ratio) Rel {
	return Rel{Abs: 0, Rel: ratio}
}

// Resolve converts the relative length to absolute given a base.
func (r Rel) Resolve(base Abs) Abs {
	return r.Abs + Abs(float64(r.Rel)*float64(base))
}

// Point represents a 2D coordinate.
type Point struct {
	X Abs
	Y Abs
}

// Origin is the zero point.
var Origin = Point{X: 0, Y: 0}

// Size represents 2D dimensions.
type Size struct {
	Width  Abs
	Height Abs
}

// ZeroSize is the zero size.
var ZeroSize = Size{Width: 0, Height: 0}

// NewSize creates a size from width and height.
func NewSize(width, height Abs) Size {
	return Size{Width: width, Height: height}
}

// IsZero returns true if both dimensions are zero.
func (s Size) IsZero() bool {
	return s.Width.IsZero() && s.Height.IsZero()
}

// Axes represents a generic 2D pair (X and Y components).
type Axes[T any] struct {
	X T
	Y T
}

// NewAxes creates an Axes with the given values.
func NewAxes[T any](x, y T) Axes[T] {
	return Axes[T]{X: x, Y: y}
}

// Sides represents values for all four sides (left, top, right, bottom).
type Sides[T any] struct {
	Left   T
	Top    T
	Right  T
	Bottom T
}

// NewSides creates Sides with the given values.
func NewSides[T any](left, top, right, bottom T) Sides[T] {
	return Sides[T]{Left: left, Top: top, Right: right, Bottom: bottom}
}

// Uniform creates Sides with the same value on all sides.
func Uniform[T any](value T) Sides[T] {
	return Sides[T]{Left: value, Top: value, Right: value, Bottom: value}
}

// SidesAbs is a specialized Sides type for absolute values.
type SidesAbs = Sides[Abs]

// SumHorizontal returns the sum of left and right.
func SumHorizontal(s Sides[Abs]) Abs {
	return s.Left + s.Right
}

// SumVertical returns the sum of top and bottom.
func SumVertical(s Sides[Abs]) Abs {
	return s.Top + s.Bottom
}

// Corners represents values for all four corners.
type Corners[T any] struct {
	TopLeft     T
	TopRight    T
	BottomRight T
	BottomLeft  T
}

// Sizing represents how a dimension should be sized.
type Sizing interface {
	isSizing()
}

// SizingAuto represents automatic sizing based on content.
type SizingAuto struct{}

func (SizingAuto) isSizing() {}

// SizingFr represents fractional sizing for flexible layouts.
type SizingFr struct {
	Value Fr
}

func (SizingFr) isSizing() {}

// SizingRel represents relative sizing with absolute and relative components.
type SizingRel struct {
	Value Rel
}

func (SizingRel) isSizing() {}

// Auto is the singleton auto sizing value.
var Auto Sizing = SizingAuto{}

// NewFr creates a fractional sizing value.
func NewFr(value Fr) Sizing {
	return SizingFr{Value: value}
}

// NewRelSizing creates a relative sizing value.
func NewRelSizing(value Rel) Sizing {
	return SizingRel{Value: value}
}

// IsAuto returns true if the sizing is automatic.
func IsAuto(s Sizing) bool {
	_, ok := s.(SizingAuto)
	return ok
}

// IsFr returns true if the sizing is fractional.
func IsFr(s Sizing) bool {
	_, ok := s.(SizingFr)
	return ok
}

// IsAutoOrFr returns true if the sizing is automatic or fractional.
func IsAutoOrFr(s Sizing) bool {
	return IsAuto(s) || IsFr(s)
}

// ResolveSizing resolves a Sizing to an absolute value given a base.
// Returns the full base for Auto and Fr, resolves Rel against the base.
func ResolveSizing(sizing Sizing, base Abs) Abs {
	switch s := sizing.(type) {
	case SizingAuto, SizingFr:
		return base
	case SizingRel:
		return s.Value.Resolve(base)
	default:
		return base
	}
}

// Transform represents a 2D affine transformation matrix.
// The matrix is stored in row-major order:
// | a b e |
// | c d f |
// | 0 0 1 |
type Transform struct {
	A, B, C, D float64 // 2x2 rotation/scale/shear
	E, F       float64 // Translation
}

// Identity is the identity transformation.
var Identity = Transform{A: 1, B: 0, C: 0, D: 1, E: 0, F: 0}

// Translate creates a translation transform.
func Translate(dx, dy Abs) Transform {
	return Transform{A: 1, B: 0, C: 0, D: 1, E: float64(dx), F: float64(dy)}
}

// Scale creates a scaling transform.
func Scale(sx, sy float64) Transform {
	return Transform{A: sx, B: 0, C: 0, D: sy, E: 0, F: 0}
}

// Rotate creates a rotation transform (angle in radians).
func Rotate(angle float64) Transform {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return Transform{A: cos, B: -sin, C: sin, D: cos, E: 0, F: 0}
}

// Then composes this transform with another (this * other).
func (t Transform) Then(other Transform) Transform {
	return Transform{
		A: t.A*other.A + t.B*other.C,
		B: t.A*other.B + t.B*other.D,
		C: t.C*other.A + t.D*other.C,
		D: t.C*other.B + t.D*other.D,
		E: t.A*other.E + t.B*other.F + t.E,
		F: t.C*other.E + t.D*other.F + t.F,
	}
}

// Apply transforms a point.
func (t Transform) Apply(p Point) Point {
	return Point{
		X: Abs(t.A*float64(p.X) + t.B*float64(p.Y) + t.E),
		Y: Abs(t.C*float64(p.X) + t.D*float64(p.Y) + t.F),
	}
}
