package math

import (
	"math"
)

// Abs represents an absolute length in points (1/72 inch).
type Abs float64

// Zero returns zero absolute length.
func (a Abs) Zero() bool { return a == 0 }

// Max returns the maximum of two absolute lengths.
func (a Abs) Max(b Abs) Abs {
	if a > b {
		return a
	}
	return b
}

// Min returns the minimum of two absolute lengths.
func (a Abs) Min(b Abs) Abs {
	if a < b {
		return a
	}
	return b
}

// SetMax updates the value to the maximum of itself and the given value.
func (a *Abs) SetMax(b Abs) {
	if b > *a {
		*a = b
	}
}

// Fits returns true if this length fits within the given space.
func (a Abs) Fits(space Abs) bool {
	return a <= space
}

// Clamp restricts the value to the given range.
func (a Abs) Clamp(min, max Abs) Abs {
	if a < min {
		return min
	}
	if a > max {
		return max
	}
	return a
}

// Em represents a font-relative unit (1em = font size).
type Em float64

// At converts em to absolute length at the given font size.
func (e Em) At(fontSize Abs) Abs {
	return Abs(float64(e) * float64(fontSize))
}

// Resolve converts em to absolute length at the resolved style size.
func (e Em) Resolve(fontSize Abs) Abs {
	return e.At(fontSize)
}

// FromAbs creates an Em value from an absolute length and font size.
func FromAbs(abs Abs, fontSize Abs) Em {
	if fontSize == 0 {
		return 0
	}
	return Em(float64(abs) / float64(fontSize))
}

// Size represents 2D dimensions.
type Size struct {
	X, Y Abs
}

// Zero returns a zero size.
func SizeZero() Size {
	return Size{}
}

// WithX returns a size with only the X component set.
func SizeWithX(x Abs) Size {
	return Size{X: x}
}

// WithY returns a size with only the Y component set.
func SizeWithY(y Abs) Size {
	return Size{Y: y}
}

// ToPoint converts a Size to a Point.
func (s Size) ToPoint() Point {
	return Point{X: s.X, Y: s.Y}
}

// Get returns the component for the given axis.
func (s Size) Get(axis Axis) Abs {
	if axis == AxisX {
		return s.X
	}
	return s.Y
}

// Point represents a 2D coordinate.
type Point struct {
	X, Y Abs
}

// PointZero returns the origin point.
func PointZero() Point {
	return Point{}
}

// WithX returns a point with only the X component set.
func PointWithX(x Abs) Point {
	return Point{X: x}
}

// WithY returns a point with only the Y component set.
func PointWithY(y Abs) Point {
	return Point{Y: y}
}

// Hypot returns the Euclidean distance from the origin.
func (p Point) Hypot() Abs {
	return Abs(math.Hypot(float64(p.X), float64(p.Y)))
}

// Axis represents a coordinate axis.
type Axis int

const (
	AxisX Axis = iota
	AxisY
)

// Corner represents a corner of a rectangle.
type Corner int

const (
	CornerTopLeft Corner = iota
	CornerTopRight
	CornerBottomRight
	CornerBottomLeft
)

// Inv returns the opposite corner.
func (c Corner) Inv() Corner {
	switch c {
	case CornerTopLeft:
		return CornerBottomRight
	case CornerTopRight:
		return CornerBottomLeft
	case CornerBottomRight:
		return CornerTopLeft
	case CornerBottomLeft:
		return CornerTopRight
	default:
		return c
	}
}

// VAlignment represents vertical alignment.
type VAlignment int

const (
	VAlignTop VAlignment = iota
	VAlignHorizon
	VAlignBottom
)

// Inv returns the inverse alignment.
func (v VAlignment) Inv() VAlignment {
	switch v {
	case VAlignTop:
		return VAlignBottom
	case VAlignBottom:
		return VAlignTop
	default:
		return v
	}
}

// Position calculates the position for the given size.
func (v VAlignment) Position(size Abs) Abs {
	switch v {
	case VAlignTop:
		return 0
	case VAlignHorizon:
		return size / 2.0
	case VAlignBottom:
		return size
	default:
		return 0
	}
}

// FixedAlignment represents a fixed alignment value.
type FixedAlignment int

const (
	FixedAlignStart FixedAlignment = iota
	FixedAlignCenter
	FixedAlignEnd
)

// Position calculates the position for the given size.
func (a FixedAlignment) Position(size Abs) Abs {
	switch a {
	case FixedAlignStart:
		return 0
	case FixedAlignCenter:
		return size / 2.0
	case FixedAlignEnd:
		return size
	default:
		return 0
	}
}

// MathSize represents the math size context.
type MathSize int

const (
	MathSizeDisplay MathSize = iota
	MathSizeText
	MathSizeScript
	MathSizeScriptScript
)

// LeftRightAlternator alternates between left and right alignment.
type LeftRightAlternator int

const (
	LeftRightAlternatorLeft LeftRightAlternator = iota
	LeftRightAlternatorRight
)

// Next returns the next alternation value.
func (a *LeftRightAlternator) Next() LeftRightAlternator {
	current := *a
	if *a == LeftRightAlternatorLeft {
		*a = LeftRightAlternatorRight
	} else {
		*a = LeftRightAlternatorLeft
	}
	return current
}

// Position represents a position (above or below).
type Position int

const (
	PositionAbove Position = iota
	PositionBelow
)

// Rel represents a relative length (percentage of some reference).
type Rel struct {
	Abs Abs
	Rel float64 // Ratio (0.5 = 50%)
}

// RelativeTo computes the absolute length relative to a reference.
func (r Rel) RelativeTo(reference Abs) Abs {
	return r.Abs + Abs(r.Rel*float64(reference))
}

// Angle represents an angle in radians.
type Angle float64

// Rad creates an angle from radians.
func Rad(radians float64) Angle {
	return Angle(radians)
}

// Deg creates an angle from degrees.
func Deg(degrees float64) Angle {
	return Angle(degrees * math.Pi / 180.0)
}

// Transform represents a 2D affine transformation matrix.
type Transform struct {
	A, B, C, D, E, F float64
}

// Identity returns the identity transform.
func TransformIdentity() Transform {
	return Transform{A: 1, D: 1}
}

// Rotate returns a rotation transform.
func TransformRotate(angle Angle) Transform {
	sin := math.Sin(float64(angle))
	cos := math.Cos(float64(angle))
	return Transform{A: cos, B: sin, C: -sin, D: cos}
}

// GlyphID represents a font glyph identifier.
type GlyphID uint16

// Glyph represents a shaped glyph.
type Glyph struct {
	ID       uint16
	XAdvance Em
	XOffset  Em
	YAdvance Em
	YOffset  Em
	Range    Range
	Span     SpanInfo
}

// Range represents a byte range in text.
type Range struct {
	Start, End int
}

// SpanInfo represents source span information.
type SpanInfo struct {
	Span   Span
	Offset int
}

// Span represents a source location.
type Span uint64

// Detached returns a detached span (no source location).
func SpanDetached() Span {
	return 0
}

// Tag represents an introspection tag.
type Tag struct {
	// Implementation will be filled in when needed
}

// TextItem represents a text item for rendering.
type TextItem struct {
	Text   string
	Font   *Font
	Size   Abs
	Fill   Paint
	Stroke *FixedStroke
	Lang   string
	Region string
	Glyphs []Glyph
}

// Width returns the total width of the text item.
func (t *TextItem) Width() Abs {
	var width Em
	for _, g := range t.Glyphs {
		width += g.XAdvance
	}
	return width.At(t.Size)
}

// Paint represents a fill color or gradient.
type Paint interface {
	isPaint()
}

// SolidPaint represents a solid color.
type SolidPaint struct {
	R, G, B, A uint8
}

func (*SolidPaint) isPaint() {}

// FixedStroke represents a resolved stroke.
type FixedStroke struct {
	Thickness Abs
	Paint     Paint
	Cap       LineCap
	Join      LineJoin
	Dash      *DashPattern
}

// FromPair creates a stroke from paint and thickness.
func StrokeFromPair(paint Paint, thickness Abs) FixedStroke {
	return FixedStroke{
		Thickness: thickness,
		Paint:     paint,
		Cap:       LineCapButt,
		Join:      LineJoinMiter,
	}
}

// LineCap represents line cap styles.
type LineCap int

const (
	LineCapButt LineCap = iota
	LineCapRound
	LineCapSquare
)

// LineJoin represents line join styles.
type LineJoin int

const (
	LineJoinMiter LineJoin = iota
	LineJoinRound
	LineJoinBevel
)

// DashPattern represents a dash pattern for strokes.
type DashPattern struct {
	Array  []Abs
	Phase  Abs
}

// Font represents a font with metrics.
type Font struct {
	// Font implementation will be expanded
	math *MathConstants
}

// Math returns the math constants for this font.
func (f *Font) Math() *MathConstants {
	if f.math == nil {
		f.math = &MathConstants{}
	}
	return f.math
}

// XAdvance returns the x advance for a glyph.
func (f *Font) XAdvance(id uint16) Em {
	// TODO: Implement actual font lookup
	return 0.5 // Placeholder
}

// YAdvance returns the y advance for a glyph.
func (f *Font) YAdvance(id uint16) Em {
	// TODO: Implement actual font lookup
	return 0
}

// ToEm converts a font unit value to em.
func (f *Font) ToEm(value int16) Em {
	// TODO: Implement actual conversion based on units per em
	return Em(float64(value) / 1000.0) // Placeholder
}

// MathConstants holds OpenType MATH table constants.
type MathConstants struct {
	AxisHeight                          Em
	SpaceWidth                          Em
	FractionRuleThickness               Em
	FractionNumeratorDisplayStyleShiftUp Em
	FractionNumeratorShiftUp            Em
	FractionDenominatorDisplayStyleShiftDown Em
	FractionDenominatorShiftDown        Em
	FractionNumDisplayStyleGapMin       Em
	FractionNumeratorGapMin             Em
	FractionDenomDisplayStyleGapMin     Em
	FractionDenominatorGapMin           Em
	StackTopDisplayStyleShiftUp         Em
	StackTopShiftUp                     Em
	StackBottomDisplayStyleShiftDown    Em
	StackBottomShiftDown                Em
	StackDisplayStyleGapMin             Em
	StackGapMin                         Em
	SkewedFractionVerticalGap           Em
	SkewedFractionHorizontalGap         Em
	RadicalRuleThickness                Em
	RadicalDisplayStyleVerticalGap      Em
	RadicalVerticalGap                  Em
	RadicalExtraAscender                Em
	RadicalKernBeforeDegree             Em
	RadicalKernAfterDegree              Em
	RadicalDegreeBottomRaisePercent     float64
	AccentBaseHeight                    Em
	FlattenedAccentBaseHeight           Em
	OverbarExtraAscender                Em
	OverbarRuleThickness                Em
	OverbarVerticalGap                  Em
	UnderbarExtraDescender              Em
	UnderbarRuleThickness               Em
	UnderbarVerticalGap                 Em
	SuperscriptShiftUp                  Em
	SuperscriptShiftUpCramped           Em
	SuperscriptBottomMin                Em
	SuperscriptBottomMaxWithSubscript   Em
	SuperscriptBaselineDropMax          Em
	SubSuperscriptGapMin                Em
	SubscriptShiftDown                  Em
	SubscriptTopMax                     Em
	SubscriptBaselineDropMin            Em
	UpperLimitGapMin                    Em
	UpperLimitBaselineRiseMin           Em
	LowerLimitGapMin                    Em
	LowerLimitBaselineDropMin           Em
	SpaceAfterScript                    Em
	DisplayOperatorMinHeight            Em
}

// MathProperties holds math layout properties.
type MathProperties struct {
	Class    Class
	Size     MathSize
	Span     Span
	Ignorant bool
	LSpace   *Rel
	RSpace   *Rel
}

// DefaultMathProperties returns default math properties.
func DefaultMathProperties() *MathProperties {
	return &MathProperties{
		Class: Normal,
		Size:  MathSizeText,
	}
}

// Geometry represents a geometric shape.
type Geometry interface {
	isGeometry()
}

// LineGeometry represents a line.
type LineGeometry struct {
	Delta Point
}

func (*LineGeometry) isGeometry() {}

// RectGeometry represents a rectangle.
type RectGeometry struct {
	Size Size
}

func (*RectGeometry) isGeometry() {}

// Stroked creates a stroked shape from geometry.
func Stroked(geom Geometry, stroke FixedStroke) *Shape {
	return &Shape{
		Geometry: geom,
		Stroke:   &stroke,
	}
}

// FilledAndStroked creates a filled and stroked shape.
func FilledAndStroked(geom Geometry, fill Paint, stroke FixedStroke) *Shape {
	return &Shape{
		Geometry: geom,
		Fill:     fill,
		Stroke:   &stroke,
	}
}

// Shape represents a drawable shape.
type Shape struct {
	Geometry Geometry
	Fill     Paint
	Stroke   *FixedStroke
}

// FillRule represents the fill rule for shapes.
type FillRule int

const (
	FillRuleNonZero FillRule = iota
	FillRuleEvenOdd
)

// kernAtHeightFromFont looks up the kerning value at a specific corner and height.
func kernAtHeightFromFont(font *Font, id GlyphID, corner Corner, height Em) Em {
	// TODO: Implement actual font kerning lookup
	return 0
}

// AugmentOffsets represents augmentation line offsets for matrices.
type AugmentOffsets struct {
	Offsets []int
}
