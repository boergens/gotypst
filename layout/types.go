// Package layout provides the layout engine for GoTypst.
//
// This package is a Go translation of typst-layout from the original Typst
// compiler. It handles converting abstract document content into positioned
// frames ready for rendering.
package layout

import (
	"fmt"
	"math"
)

// Abs represents an absolute length in typographic points (1/72 inch).
// This is the fundamental unit for all layout calculations.
type Abs float64

// Common length constants.
const (
	// Pt is one typographic point.
	Pt Abs = 1.0
	// Mm is one millimeter.
	Mm Abs = 2.8346456692913
	// Cm is one centimeter.
	Cm Abs = 28.346456692913
	// In is one inch.
	In Abs = 72.0
)

// Zero returns the zero length.
func (a Abs) Zero() Abs {
	return 0
}

// IsZero returns true if the length is zero.
func (a Abs) IsZero() bool {
	return a == 0
}

// Abs returns the absolute value.
func (a Abs) Abs() Abs {
	if a < 0 {
		return -a
	}
	return a
}

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

// Clamp clamps the length to the given range.
func (a Abs) Clamp(min, max Abs) Abs {
	if a < min {
		return min
	}
	if a > max {
		return max
	}
	return a
}

// Points returns the length in points.
func (a Abs) Points() float64 {
	return float64(a)
}

// Fr represents a fractional unit for distributing remaining space.
// In Typst, "1fr" means one fraction of the remaining space.
type Fr float64

// Point represents a 2D point in layout coordinates.
type Point struct {
	X Abs
	Y Abs
}

// Zero returns the origin point.
func (Point) Zero() Point {
	return Point{}
}

// IsZero returns true if this is the origin.
func (p Point) IsZero() bool {
	return p.X == 0 && p.Y == 0
}

// Add adds two points.
func (p Point) Add(q Point) Point {
	return Point{X: p.X + q.X, Y: p.Y + q.Y}
}

// Sub subtracts two points.
func (p Point) Sub(q Point) Point {
	return Point{X: p.X - q.X, Y: p.Y - q.Y}
}

// Scale multiplies the point by a scalar.
func (p Point) Scale(s float64) Point {
	return Point{X: Abs(float64(p.X) * s), Y: Abs(float64(p.Y) * s)}
}

// Size represents 2D dimensions (width and height).
type Size struct {
	Width  Abs
	Height Abs
}

// Zero returns zero size.
func (Size) Zero() Size {
	return Size{}
}

// IsZero returns true if both dimensions are zero.
func (s Size) IsZero() bool {
	return s.Width == 0 && s.Height == 0
}

// Splat creates a size with equal width and height.
func SizeSplat(v Abs) Size {
	return Size{Width: v, Height: v}
}

// Contains returns true if the point is within the size bounds.
func (s Size) Contains(p Point) bool {
	return p.X >= 0 && p.X <= s.Width && p.Y >= 0 && p.Y <= s.Height
}

// AspectRatio returns width/height ratio.
func (s Size) AspectRatio() float64 {
	if s.Height == 0 {
		return math.Inf(1)
	}
	return float64(s.Width) / float64(s.Height)
}

// Axes represents a generic 2D pair for axes (horizontal/vertical).
type Axes[T any] struct {
	X T
	Y T
}

// Sides represents values for four sides (padding, margins, etc.).
type Sides[T any] struct {
	Left   T
	Top    T
	Right  T
	Bottom T
}

// Splat creates Sides with the same value on all sides.
func SidesSplat[T any](v T) Sides[T] {
	return Sides[T]{Left: v, Top: v, Right: v, Bottom: v}
}

// Corners represents values for four corners (border radii, etc.).
type Corners[T any] struct {
	TopLeft     T
	TopRight    T
	BottomRight T
	BottomLeft  T
}

// CornersSplat creates Corners with the same value on all corners.
func CornersSplat[T any](v T) Corners[T] {
	return Corners[T]{TopLeft: v, TopRight: v, BottomRight: v, BottomLeft: v}
}

// Ratio represents a ratio/percentage value (0.5 = 50%).
type Ratio float64

// Resolve resolves the ratio against a given whole.
func (r Ratio) Resolve(whole Abs) Abs {
	return Abs(float64(r) * float64(whole))
}

// Relative represents a combination of absolute and relative length.
// For example, "50% + 10pt" would be Relative{Abs: 10, Rel: 0.5}.
type Relative struct {
	Abs Abs
	Rel Ratio
}

// IsZero returns true if both components are zero.
func (r Relative) IsZero() bool {
	return r.Abs == 0 && r.Rel == 0
}

// Resolve resolves the relative length against a given whole.
func (r Relative) Resolve(whole Abs) Abs {
	return r.Abs + r.Rel.Resolve(whole)
}

// Alignment represents 2D alignment (horizontal and vertical).
type Alignment struct {
	X HAlign
	Y VAlign
}

// HAlign represents horizontal alignment.
type HAlign int

const (
	HAlignStart  HAlign = iota // Left in LTR, Right in RTL
	HAlignCenter               // Center
	HAlignEnd                  // Right in LTR, Left in RTL
	HAlignLeft                 // Always left
	HAlignRight                // Always right
)

// VAlign represents vertical alignment.
type VAlign int

const (
	VAlignTop    VAlign = iota // Top
	VAlignHorizon              // Baseline/middle
	VAlignBottom               // Bottom
)

// Dir represents text direction.
type Dir int

const (
	DirLTR Dir = iota // Left-to-right
	DirRTL            // Right-to-left
	DirTTB            // Top-to-bottom
	DirBTT            // Bottom-to-top
)

// IsHorizontal returns true for horizontal directions.
func (d Dir) IsHorizontal() bool {
	return d == DirLTR || d == DirRTL
}

// IsPositive returns true if the direction is positive (LTR or TTB).
func (d Dir) IsPositive() bool {
	return d == DirLTR || d == DirTTB
}

// Transform represents a 2D affine transformation matrix.
// The matrix is stored in row-major order:
//
//	| a  b  e |
//	| c  d  f |
//	| 0  0  1 |
type Transform struct {
	A, B, C, D float64 // Scale and rotation
	E, F       float64 // Translation
}

// Identity returns the identity transform.
func Identity() Transform {
	return Transform{A: 1, B: 0, C: 0, D: 1, E: 0, F: 0}
}

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

// Then composes two transforms (this then other).
func (t Transform) Then(o Transform) Transform {
	return Transform{
		A: t.A*o.A + t.B*o.C,
		B: t.A*o.B + t.B*o.D,
		C: t.C*o.A + t.D*o.C,
		D: t.C*o.B + t.D*o.D,
		E: t.E*o.A + t.F*o.C + o.E,
		F: t.E*o.B + t.F*o.D + o.F,
	}
}

// Apply applies the transform to a point.
func (t Transform) Apply(p Point) Point {
	return Point{
		X: Abs(t.A*float64(p.X) + t.B*float64(p.Y) + t.E),
		Y: Abs(t.C*float64(p.X) + t.D*float64(p.Y) + t.F),
	}
}

// IsIdentity returns true if this is the identity transform.
func (t Transform) IsIdentity() bool {
	return t.A == 1 && t.B == 0 && t.C == 0 && t.D == 1 && t.E == 0 && t.F == 0
}

// Frame represents a container for positioned layout items.
// This is the fundamental output of layout operations.
type Frame struct {
	// Size is the dimensions of the frame.
	Size Size
	// Baseline is the vertical position of the text baseline.
	Baseline Abs
	// Items contains the positioned items in this frame.
	Items []FrameItem
}

// NewFrame creates a new empty frame with the given size.
func NewFrame(size Size) *Frame {
	return &Frame{Size: size}
}

// Push adds an item to the frame.
func (f *Frame) Push(pos Point, item FrameItem) {
	// Store position in the item for now
	if positioned, ok := item.(*PositionedItem); ok {
		positioned.Pos = pos
		f.Items = append(f.Items, positioned)
	} else {
		f.Items = append(f.Items, &PositionedItem{Pos: pos, Item: item})
	}
}

// PushFrame adds a subframe at the given position.
func (f *Frame) PushFrame(pos Point, sub *Frame) {
	f.Items = append(f.Items, &PositionedItem{Pos: pos, Item: sub})
}

// Translate shifts all items by the given offset.
func (f *Frame) Translate(offset Point) {
	for _, item := range f.Items {
		if pos, ok := item.(*PositionedItem); ok {
			pos.Pos = pos.Pos.Add(offset)
		}
	}
}

// FrameItem represents an item that can be placed in a frame.
type FrameItem interface {
	isFrameItem()
}

// PositionedItem wraps a FrameItem with its position.
type PositionedItem struct {
	Pos  Point
	Item FrameItem
}

func (*PositionedItem) isFrameItem() {}
func (*Frame) isFrameItem()          {}

// TextItem represents rendered text in a frame.
type TextItem struct {
	Text string
	// Font information would go here
}

func (*TextItem) isFrameItem() {}

// ShapeItem represents a geometric shape in a frame.
type ShapeItem struct {
	Shape Shape
	Fill  *Color
	Stroke *Stroke
}

func (*ShapeItem) isFrameItem() {}

// Shape represents a geometric shape.
type Shape interface {
	isShape()
}

// RectShape represents a rectangle shape.
type RectShape struct {
	Size   Size
	Radius Corners[Abs]
}

func (*RectShape) isShape() {}

// LineShape represents a line shape.
type LineShape struct {
	Start Point
	End   Point
}

func (*LineShape) isShape() {}

// Color represents an RGBA color.
type Color struct {
	R, G, B, A uint8
}

// Stroke represents stroke/line styling.
type Stroke struct {
	Paint     *Color
	Thickness Abs
	LineCap   LineCap
	LineJoin  LineJoin
	DashArray []Abs
	DashPhase Abs
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

// Fragment represents a multi-frame result from layout.
// When content spans multiple regions (pages), layout returns a Fragment.
type Fragment []*Frame

// String returns a debug representation.
func (f Fragment) String() string {
	return fmt.Sprintf("Fragment(%d frames)", len(f))
}

// Region represents a layout region with available space.
type Region struct {
	// Size is the available size in this region.
	Size Size
	// Full indicates whether this is the full region or a partial.
	Full bool
	// Backlog contains sizes of subsequent regions if known.
	Backlog []Size
	// Last indicates this is the last region.
	Last bool
}

// Regions represents a sequence of layout regions.
type Regions struct {
	// Current is the current region.
	Current Region
	// Backlog contains subsequent regions.
	Backlog []Region
}

// First returns the first region.
func (r *Regions) First() Region {
	return r.Current
}

// IsEmpty returns true if there are no regions.
func (r *Regions) IsEmpty() bool {
	return r.Current.Size.IsZero()
}
