// Package layout provides the layout engine for Typst.
//
// This package is a Go translation of typst-layout from the original Typst
// compiler. It handles converting abstract document content into positioned
// frames ready for rendering.
package layout

import (
	"math"

	"github.com/boergens/gotypst/syntax"
)

// Abs represents an absolute length in typographic points (1/72 inch).
type Abs float64

// Common Abs constructors.
func Pt(v float64) Abs       { return Abs(v) }
func Mm(v float64) Abs       { return Abs(v * 2.8346456692913) }
func Cm(v float64) Abs       { return Abs(v * 28.346456692913) }
func Inches(v float64) Abs   { return Abs(v * 72.0) }
func AbsZero() Abs           { return Abs(0) }
func AbsInf() Abs            { return Abs(math.Inf(1)) }

// ToPt returns the length in points.
func (a Abs) ToPt() float64 { return float64(a) }

// ToMm returns the length in millimeters.
func (a Abs) ToMm() float64 { return float64(a) / 2.8346456692913 }

// ToCm returns the length in centimeters.
func (a Abs) ToCm() float64 { return float64(a) / 28.346456692913 }

// ToInches returns the length in inches.
func (a Abs) ToInches() float64 { return float64(a) / 72.0 }

// IsZero returns true if the length is zero.
func (a Abs) IsZero() bool { return a == 0 }

// IsFinite returns true if the length is finite.
func (a Abs) IsFinite() bool { return !math.IsInf(float64(a), 0) && !math.IsNaN(float64(a)) }

// Min returns the smaller of two lengths.
func (a Abs) Min(b Abs) Abs {
	if a < b {
		return a
	}
	return b
}

// Max returns the larger of two lengths.
func (a Abs) Max(b Abs) Abs {
	if a > b {
		return a
	}
	return b
}

// Clamp constrains the length between min and max.
func (a Abs) Clamp(min, max Abs) Abs {
	if a < min {
		return min
	}
	if a > max {
		return max
	}
	return a
}

// Abs returns the absolute value.
func (a Abs) Abs() Abs {
	if a < 0 {
		return -a
	}
	return a
}

// Fr represents a fraction of remaining space.
type Fr float64

// FrZero returns a zero fraction.
func FrZero() Fr { return Fr(0) }

// FrOne returns one fraction unit.
func FrOne() Fr { return Fr(1) }

// IsZero returns true if the fraction is zero.
func (f Fr) IsZero() bool { return f == 0 }

// Share calculates this fraction's share of the total.
func (f Fr) Share(total Fr, available Abs) Abs {
	if total == 0 {
		return 0
	}
	return Abs(float64(f) / float64(total) * float64(available))
}

// Em represents a font-relative length unit.
type Em float64

// EmZero returns a zero em value.
func EmZero() Em { return Em(0) }

// ToAbs converts em to absolute length given a font size.
func (e Em) ToAbs(fontSize Abs) Abs {
	return Abs(float64(e) * float64(fontSize))
}

// Ratio represents a ratio (percentage) as a fraction.
type Ratio float64

// RatioZero returns a zero ratio.
func RatioZero() Ratio { return Ratio(0) }

// RatioOne returns 100%.
func RatioOne() Ratio { return Ratio(1) }

// Percent creates a ratio from a percentage value.
func Percent(v float64) Ratio { return Ratio(v / 100.0) }

// IsZero returns true if the ratio is zero.
func (r Ratio) IsZero() bool { return r == 0 }

// Of calculates the ratio of an absolute length.
func (r Ratio) Of(a Abs) Abs {
	return Abs(float64(r) * float64(a))
}

// Point represents a 2D coordinate in absolute units.
type Point struct {
	X Abs
	Y Abs
}

// PointZero returns the origin point.
func PointZero() Point { return Point{X: 0, Y: 0} }

// NewPoint creates a point from x and y coordinates.
func NewPoint(x, y Abs) Point { return Point{X: x, Y: y} }

// Add returns the sum of two points.
func (p Point) Add(other Point) Point {
	return Point{X: p.X + other.X, Y: p.Y + other.Y}
}

// Sub returns the difference of two points.
func (p Point) Sub(other Point) Point {
	return Point{X: p.X - other.X, Y: p.Y - other.Y}
}

// Scale multiplies the point by a scalar.
func (p Point) Scale(s float64) Point {
	return Point{X: Abs(float64(p.X) * s), Y: Abs(float64(p.Y) * s)}
}

// Size represents 2D dimensions in absolute units.
type Size struct {
	X Abs // Width
	Y Abs // Height
}

// SizeZero returns a zero size.
func SizeZero() Size { return Size{X: 0, Y: 0} }

// NewSize creates a size from width and height.
func NewSize(width, height Abs) Size { return Size{X: width, Y: height} }

// IsZero returns true if both dimensions are zero.
func (s Size) IsZero() bool { return s.X == 0 && s.Y == 0 }

// Fits returns true if this size fits within another size.
func (s Size) Fits(container Size) bool {
	return s.X <= container.X && s.Y <= container.Y
}

// Min returns the component-wise minimum.
func (s Size) Min(other Size) Size {
	return Size{X: s.X.Min(other.X), Y: s.Y.Min(other.Y)}
}

// Max returns the component-wise maximum.
func (s Size) Max(other Size) Size {
	return Size{X: s.X.Max(other.X), Y: s.Y.Max(other.Y)}
}

// Aspect returns the aspect ratio (width / height).
func (s Size) Aspect() float64 {
	if s.Y == 0 {
		return 0
	}
	return float64(s.X) / float64(s.Y)
}

// Axes represents a generic pair of values for X and Y axes.
type Axes[T any] struct {
	X T
	Y T
}

// NewAxes creates an Axes with the same value for both axes.
func NewAxes[T any](v T) Axes[T] {
	return Axes[T]{X: v, Y: v}
}

// Map applies a function to both values.
func (a Axes[T]) Map(f func(T) T) Axes[T] {
	return Axes[T]{X: f(a.X), Y: f(a.Y)}
}

// Sides represents values for four sides (left, top, right, bottom).
type Sides[T any] struct {
	Left   T
	Top    T
	Right  T
	Bottom T
}

// NewSides creates Sides with the same value for all sides.
func NewSides[T any](v T) Sides[T] {
	return Sides[T]{Left: v, Top: v, Right: v, Bottom: v}
}

// SidesLTRB creates Sides with individual values.
func SidesLTRB[T any](left, top, right, bottom T) Sides[T] {
	return Sides[T]{Left: left, Top: top, Right: right, Bottom: bottom}
}

// SidesAbs creates Sides[Abs] from individual values.
func SidesAbs(left, top, right, bottom Abs) Sides[Abs] {
	return Sides[Abs]{Left: left, Top: top, Right: right, Bottom: bottom}
}

// SidesAbsZero creates Sides[Abs] with all zeros.
func SidesAbsZero() Sides[Abs] {
	return Sides[Abs]{}
}

// SidesAbsSum returns the horizontal and vertical sums of absolute sides.
func SidesAbsSum(s Sides[Abs]) Size {
	return Size{
		X: s.Left + s.Right,
		Y: s.Top + s.Bottom,
	}
}

// At returns the value for the given side.
func (s Sides[T]) At(side Side) T {
	switch side {
	case SideLeft:
		return s.Left
	case SideTop:
		return s.Top
	case SideRight:
		return s.Right
	case SideBottom:
		return s.Bottom
	default:
		return s.Left
	}
}

// Side represents one of four sides.
type Side int

const (
	SideLeft Side = iota
	SideTop
	SideRight
	SideBottom
)

// Corners represents values for four corners.
type Corners[T any] struct {
	TopLeft     T
	TopRight    T
	BottomRight T
	BottomLeft  T
}

// NewCorners creates Corners with the same value for all corners.
func NewCorners[T any](v T) Corners[T] {
	return Corners[T]{TopLeft: v, TopRight: v, BottomRight: v, BottomLeft: v}
}

// Dir represents a reading direction.
type Dir int

const (
	DirLTR Dir = iota // Left-to-right
	DirRTL            // Right-to-left
	DirTTB            // Top-to-bottom
	DirBTT            // Bottom-to-top
)

// Axis returns the axis of this direction.
func (d Dir) Axis() Axis {
	switch d {
	case DirLTR, DirRTL:
		return AxisX
	default:
		return AxisY
	}
}

// IsPositive returns true if the direction goes toward increasing coordinates.
func (d Dir) IsPositive() bool {
	return d == DirLTR || d == DirTTB
}

// Axis represents a layout axis.
type Axis int

const (
	AxisX Axis = iota // Horizontal
	AxisY             // Vertical
)

// Other returns the perpendicular axis.
func (a Axis) Other() Axis {
	if a == AxisX {
		return AxisY
	}
	return AxisX
}

// Alignment represents 2D alignment.
type Alignment struct {
	X FixedAlignment
	Y FixedAlignment
}

// FixedAlignment represents alignment along one axis.
type FixedAlignment int

const (
	AlignStart  FixedAlignment = iota // Start of axis
	AlignCenter                       // Center
	AlignEnd                          // End of axis
)

// Position calculates the offset needed to align content within a container.
func (a FixedAlignment) Position(content, container Abs) Abs {
	space := container - content
	switch a {
	case AlignStart:
		return 0
	case AlignCenter:
		return space / 2
	case AlignEnd:
		return space
	default:
		return 0
	}
}

// Region represents a layout region with size and expansion constraints.
type Region struct {
	Size   Size
	Expand Axes[bool]
}

// Regions represents multiple layout regions.
type Regions struct {
	// Size is the size of the first region.
	Size Size
	// Full is the full region size (for subsequent regions).
	Full Size
	// Backlog contains additional region sizes.
	Backlog []Size
	// Last indicates if there are no more regions.
	Last bool
	// Expand indicates whether to expand in each direction.
	Expand Axes[bool]
}

// First returns the first region.
func (r *Regions) First() Region {
	return Region{Size: r.Size, Expand: r.Expand}
}

// IsFinite returns true if all region sizes are finite.
func (r *Regions) IsFinite() bool {
	return r.Size.X.IsFinite() && r.Size.Y.IsFinite()
}

// Transform represents a 2D affine transformation matrix.
// The matrix is stored as [a, b, c, d, e, f] where:
//   | a c e |
//   | b d f |
//   | 0 0 1 |
type Transform struct {
	A, B, C, D, E, F float64
}

// TransformIdentity returns the identity transform.
func TransformIdentity() Transform {
	return Transform{A: 1, B: 0, C: 0, D: 1, E: 0, F: 0}
}

// TransformTranslate creates a translation transform.
func TransformTranslate(dx, dy Abs) Transform {
	return Transform{A: 1, B: 0, C: 0, D: 1, E: float64(dx), F: float64(dy)}
}

// TransformScale creates a scaling transform.
func TransformScale(sx, sy float64) Transform {
	return Transform{A: sx, B: 0, C: 0, D: sy, E: 0, F: 0}
}

// TransformRotate creates a rotation transform (angle in radians).
func TransformRotate(angle float64) Transform {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return Transform{A: cos, B: sin, C: -sin, D: cos, E: 0, F: 0}
}

// IsIdentity returns true if this is the identity transform.
func (t Transform) IsIdentity() bool {
	return t.A == 1 && t.B == 0 && t.C == 0 && t.D == 1 && t.E == 0 && t.F == 0
}

// Concat concatenates two transforms: self * other.
func (t Transform) Concat(other Transform) Transform {
	return Transform{
		A: t.A*other.A + t.C*other.B,
		B: t.B*other.A + t.D*other.B,
		C: t.A*other.C + t.C*other.D,
		D: t.B*other.C + t.D*other.D,
		E: t.A*other.E + t.C*other.F + t.E,
		F: t.B*other.E + t.D*other.F + t.F,
	}
}

// Apply transforms a point.
func (t Transform) Apply(p Point) Point {
	return Point{
		X: Abs(t.A*float64(p.X) + t.C*float64(p.Y) + t.E),
		Y: Abs(t.B*float64(p.X) + t.D*float64(p.Y) + t.F),
	}
}

// Frame represents a container for positioned layout items.
type Frame struct {
	// Size is the frame's dimensions.
	Size Size
	// Baseline is the vertical position of the text baseline.
	Baseline *Abs
	// Items contains the positioned items.
	Items []FrameItem
	// Kind indicates if this is a soft or hard frame.
	Kind FrameKind
}

// FrameKind distinguishes between soft and hard frames.
type FrameKind int

const (
	FrameSoft FrameKind = iota // Soft frame (can be merged)
	FrameHard                  // Hard frame (boundary)
)

// NewFrame creates a new soft frame with the given size.
func NewFrame(size Size) *Frame {
	return &Frame{Size: size, Kind: FrameSoft}
}

// NewFrameSoft creates a new soft frame.
func NewFrameSoft(size Size) *Frame {
	return &Frame{Size: size, Kind: FrameSoft}
}

// NewFrameHard creates a new hard frame.
func NewFrameHard(size Size) *Frame {
	return &Frame{Size: size, Kind: FrameHard}
}

// Push adds an item at the given position.
func (f *Frame) Push(pos Point, item FrameItem) {
	f.Items = append(f.Items, PositionedItem{Position: pos, Item: item})
}

// PushFrame adds a sub-frame at the given position.
func (f *Frame) PushFrame(pos Point, frame *Frame) {
	f.Push(pos, SubFrame{Frame: frame})
}

// SetBaseline sets the baseline position.
func (f *Frame) SetBaseline(b Abs) {
	f.Baseline = &b
}

// Resize changes the frame size with alignment.
func (f *Frame) Resize(target Size, align Axes[FixedAlignment]) {
	offset := Point{
		X: align.X.Position(f.Size.X, target.X),
		Y: align.Y.Position(f.Size.Y, target.Y),
	}
	if !offset.X.IsZero() || !offset.Y.IsZero() {
		f.Translate(offset)
	}
	f.Size = target
}

// Translate moves all items by the given offset.
func (f *Frame) Translate(offset Point) {
	for i := range f.Items {
		if pi, ok := f.Items[i].(PositionedItem); ok {
			pi.Position = pi.Position.Add(offset)
			f.Items[i] = pi
		}
	}
}

// Clip clips the frame to a curve.
func (f *Frame) Clip(curve Curve) {
	// Wrap all items in a clip group
	oldItems := f.Items
	f.Items = []FrameItem{ClipItem{Curve: curve, Items: oldItems}}
}

// Width returns the frame width.
func (f *Frame) Width() Abs { return f.Size.X }

// Height returns the frame height.
func (f *Frame) Height() Abs { return f.Size.Y }

// FrameItem represents an item that can be placed in a frame.
type FrameItem interface {
	isFrameItem()
}

// PositionedItem wraps an item with a position.
type PositionedItem struct {
	Position Point
	Item     FrameItem
}

func (PositionedItem) isFrameItem() {}

// SubFrame represents a nested frame.
type SubFrame struct {
	Frame *Frame
}

func (SubFrame) isFrameItem() {}

// TextItem represents shaped text.
type TextItem struct {
	// Text content (placeholder for full text shaping)
	Text string
	Size Abs
}

func (TextItem) isFrameItem() {}

// ImageItem represents an image.
type ImageItem struct {
	Image Image
	Size  Size
	Span  syntax.Span
}

func (ImageItem) isFrameItem() {}

// ShapeItem represents a geometric shape.
type ShapeItem struct {
	Curve  Curve
	Fill   *Paint
	Stroke *Stroke
	Span   syntax.Span
}

func (ShapeItem) isFrameItem() {}

// ClipItem represents clipped content.
type ClipItem struct {
	Curve Curve
	Items []FrameItem
}

func (ClipItem) isFrameItem() {}

// LinkItem represents a hyperlink.
type LinkItem struct {
	Dest string
	Size Size
}

func (LinkItem) isFrameItem() {}

// TagItem represents a document tag.
type TagItem struct {
	Tag Tag
}

func (TagItem) isFrameItem() {}

// Fragment represents a collection of frames resulting from layout.
type Fragment []Frame

// Image represents an image (placeholder for full image support).
type Image struct {
	Data   []byte
	Format ImageFormat
	Width  float64
	Height float64
	DPI    *float64
}

// DefaultDPI is the default image DPI.
const DefaultDPI = 72.0

// ImageFormat represents an image format.
type ImageFormat int

const (
	ImagePNG ImageFormat = iota
	ImageJPEG
	ImageGIF
	ImageSVG
)

// ImageFit represents how an image should fit its container.
type ImageFit int

const (
	ImageFitContain ImageFit = iota // Fit within bounds
	ImageFitCover                   // Cover bounds (may crop)
	ImageFitStretch                 // Stretch to fill
)

// Curve represents a 2D path.
type Curve struct {
	Segments []CurveSegment
	Closed   bool
}

// CurveSegment represents a segment of a curve.
type CurveSegment interface {
	isCurveSegment()
}

// MoveTo represents a move-to segment.
type MoveTo struct {
	Point Point
}

func (MoveTo) isCurveSegment() {}

// LineTo represents a line-to segment.
type LineTo struct {
	Point Point
}

func (LineTo) isCurveSegment() {}

// QuadTo represents a quadratic bezier segment.
type QuadTo struct {
	Control Point
	End     Point
}

func (QuadTo) isCurveSegment() {}

// CubicTo represents a cubic bezier segment.
type CubicTo struct {
	Control1 Point
	Control2 Point
	End      Point
}

func (CubicTo) isCurveSegment() {}

// ClosePath represents closing the current subpath.
type ClosePath struct{}

func (ClosePath) isCurveSegment() {}

// CurveRect creates a rectangular curve.
func CurveRect(size Size) Curve {
	return Curve{
		Segments: []CurveSegment{
			MoveTo{Point: PointZero()},
			LineTo{Point: NewPoint(size.X, 0)},
			LineTo{Point: NewPoint(size.X, size.Y)},
			LineTo{Point: NewPoint(0, size.Y)},
			ClosePath{},
		},
		Closed: true,
	}
}

// Paint represents a fill or stroke paint.
type Paint struct {
	Color *Color
	// Gradient, pattern, etc. would go here
}

// NewPaintColor creates a solid color paint.
func NewPaintColor(c Color) *Paint {
	return &Paint{Color: &c}
}

// Color represents an RGBA color.
type Color struct {
	R, G, B, A uint8
}

// ColorBlack is black.
var ColorBlack = Color{R: 0, G: 0, B: 0, A: 255}

// ColorWhite is white.
var ColorWhite = Color{R: 255, G: 255, B: 255, A: 255}

// ColorTransparent is fully transparent.
var ColorTransparent = Color{R: 0, G: 0, B: 0, A: 0}

// Stroke represents stroke properties.
type Stroke struct {
	Paint     Paint
	Thickness Abs
	LineCap   LineCap
	LineJoin  LineJoin
	MiterLimit float64
	DashArray  []Abs
	DashOffset Abs
}

// LineCap represents stroke line caps.
type LineCap int

const (
	LineCapButt LineCap = iota
	LineCapRound
	LineCapSquare
)

// LineJoin represents stroke line joins.
type LineJoin int

const (
	LineJoinMiter LineJoin = iota
	LineJoinRound
	LineJoinBevel
)

// Tag represents a document tag for introspection.
type Tag struct {
	// Location is the tag's source location.
	Location syntax.Span
}

// GenericSize abstracts over main/cross axis dimensions.
type GenericSize[T any] struct {
	Main  T
	Cross T
}

// GenericSizeAbsToSize converts a GenericSize[Abs] to Size based on axis direction.
func GenericSizeAbsToSize(g GenericSize[Abs], dir Dir) Size {
	if dir.Axis() == AxisX {
		return Size{X: g.Main, Y: g.Cross}
	}
	return Size{X: g.Cross, Y: g.Main}
}

// FromSize converts a Size to GenericSize based on axis direction.
func GenericSizeFromSize(size Size, dir Dir) GenericSize[Abs] {
	if dir.Axis() == AxisX {
		return GenericSize[Abs]{Main: size.X, Cross: size.Y}
	}
	return GenericSize[Abs]{Main: size.Y, Cross: size.X}
}

// GenericAxes abstracts over main/cross axis pairs.
type GenericAxes[T any] struct {
	Main  T
	Cross T
}

// ToAxes converts a GenericAxes to Axes based on axis direction.
func (g GenericAxes[T]) ToAxes(dir Dir) Axes[T] {
	if dir.Axis() == AxisX {
		return Axes[T]{X: g.Main, Y: g.Cross}
	}
	return Axes[T]{X: g.Cross, Y: g.Main}
}
