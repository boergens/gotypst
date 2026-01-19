// Package inline provides inline/paragraph layout including text shaping.
package inline

import (
	"math"
	"sort"
)

// DecoLine represents a decoration line style.
type DecoLine interface {
	isDecoLine()
}

// HighlightDeco represents a highlight decoration.
type HighlightDeco struct {
	Fill       interface{} // Fill paint (color or gradient)
	Stroke     *FixedStroke
	TopEdge    TopEdge
	BottomEdge BottomEdge
	Radius     Abs
}

func (*HighlightDeco) isDecoLine() {}

// StrikethroughDeco represents a strikethrough decoration.
type StrikethroughDeco struct {
	Stroke     *FixedStroke
	Offset     *Abs // Optional offset override
	Background bool
}

func (*StrikethroughDeco) isDecoLine() {}

// OverlineDeco represents an overline decoration.
type OverlineDeco struct {
	Stroke     *FixedStroke
	Offset     *Abs // Optional offset override
	Evade      bool // Whether to evade glyph outlines
	Background bool
}

func (*OverlineDeco) isDecoLine() {}

// UnderlineDeco represents an underline decoration.
type UnderlineDeco struct {
	Stroke     *FixedStroke
	Offset     *Abs // Optional offset override
	Evade      bool // Whether to evade glyph outlines
	Background bool
}

func (*UnderlineDeco) isDecoLine() {}

// TopEdge represents the top edge metric for text bounds.
type TopEdge int

const (
	TopEdgeAscender TopEdge = iota
	TopEdgeCapHeight
	TopEdgeXHeight
	TopEdgeBounds
)

// BottomEdge represents the bottom edge metric for text bounds.
type BottomEdge int

const (
	BottomEdgeDescender BottomEdge = iota
	BottomEdgeBaseline
	BottomEdgeBounds
)

// FixedStroke represents a stroke with fixed width.
type FixedStroke struct {
	Paint     interface{} // Stroke paint (color or gradient)
	Thickness Abs
	LineCap   LineCap
	LineJoin  LineJoin
	DashArray []Abs
	DashPhase Abs
}

// StrokeFromPair creates a FixedStroke from paint and thickness.
func StrokeFromPair(paint interface{}, thickness Abs) FixedStroke {
	return FixedStroke{
		Paint:     paint,
		Thickness: thickness,
		LineCap:   LineCapButt,
		LineJoin:  LineJoinMiter,
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

// Decoration represents a text decoration.
type Decoration struct {
	Line   DecoLine
	Extent Abs // Extra extent beyond text bounds
}

// DecoTextItem represents text item data needed for decoration.
type DecoTextItem struct {
	Font   DecoFont
	Size   Abs
	Fill   interface{} // Text fill paint for deriving decoration color
	Glyphs []DecoGlyph
}

// DecoFont provides font metrics for decoration.
type DecoFont interface {
	Metrics() FontMetrics
	ToEm(units int16) Em
	OutlineGlyph(glyphID uint16) ([]PathSegment, *GlyphBBox, bool)
}

// FontMetrics contains decoration-related font metrics.
type FontMetrics struct {
	UnitsPerEm    float64
	Underline     DecoMetrics
	Strikethrough DecoMetrics
	Overline      DecoMetrics
}

// DecoMetrics contains position and thickness for a decoration.
type DecoMetrics struct {
	Position  Em
	Thickness Em
}

// DecoGlyph represents a glyph for decoration intersection testing.
type DecoGlyph struct {
	ID       uint16
	XAdvance Em
	XOffset  Em
}

// GlyphBBox represents a glyph bounding box in font units.
type GlyphBBox struct {
	XMin, YMin, XMax, YMax int16
}

// PathSegment represents a segment of a glyph outline.
type PathSegment interface {
	isPathSegment()
	// IntersectLine returns the x-coordinates where this segment intersects
	// the horizontal line at the given y-coordinate.
	IntersectLine(y float64) []float64
}

// LineSegment represents a line segment.
type LineSegment struct {
	X0, Y0, X1, Y1 float64
}

func (*LineSegment) isPathSegment() {}

// IntersectLine finds intersection with horizontal line at y.
func (l *LineSegment) IntersectLine(y float64) []float64 {
	// Check if y is within segment's y range
	yMin, yMax := l.Y0, l.Y1
	if yMin > yMax {
		yMin, yMax = yMax, yMin
	}
	if y < yMin || y > yMax {
		return nil
	}
	// Horizontal line case
	if l.Y1 == l.Y0 {
		return nil
	}
	// Calculate x at intersection
	t := (y - l.Y0) / (l.Y1 - l.Y0)
	x := l.X0 + t*(l.X1-l.X0)
	return []float64{x}
}

// QuadSegment represents a quadratic Bezier segment.
type QuadSegment struct {
	X0, Y0, X1, Y1, X2, Y2 float64
}

func (*QuadSegment) isPathSegment() {}

// IntersectLine finds intersections with horizontal line at y.
func (q *QuadSegment) IntersectLine(y float64) []float64 {
	// Quadratic: y = (1-t)²y0 + 2(1-t)t·y1 + t²y2
	// Rearranged: (y0 - 2y1 + y2)t² + 2(y1 - y0)t + (y0 - y) = 0
	a := q.Y0 - 2*q.Y1 + q.Y2
	b := 2 * (q.Y1 - q.Y0)
	c := q.Y0 - y

	var results []float64

	if math.Abs(a) < 1e-10 {
		// Linear case
		if math.Abs(b) > 1e-10 {
			t := -c / b
			if t >= 0 && t <= 1 {
				x := (1-t)*(1-t)*q.X0 + 2*(1-t)*t*q.X1 + t*t*q.X2
				results = append(results, x)
			}
		}
		return results
	}

	disc := b*b - 4*a*c
	if disc < 0 {
		return nil
	}

	sqrtDisc := math.Sqrt(disc)
	for _, t := range []float64{(-b + sqrtDisc) / (2 * a), (-b - sqrtDisc) / (2 * a)} {
		if t >= 0 && t <= 1 {
			x := (1-t)*(1-t)*q.X0 + 2*(1-t)*t*q.X1 + t*t*q.X2
			results = append(results, x)
		}
	}
	return results
}

// CubicSegment represents a cubic Bezier segment.
type CubicSegment struct {
	X0, Y0, X1, Y1, X2, Y2, X3, Y3 float64
}

func (*CubicSegment) isPathSegment() {}

// IntersectLine finds intersections with horizontal line at y.
// Uses subdivision for robustness.
func (c *CubicSegment) IntersectLine(y float64) []float64 {
	return cubicIntersectLine(c.X0, c.Y0, c.X1, c.Y1, c.X2, c.Y2, c.X3, c.Y3, y, 0)
}

func cubicIntersectLine(x0, y0, x1, y1, x2, y2, x3, y3, y float64, depth int) []float64 {
	// Check bounding box
	yMin := math.Min(math.Min(y0, y1), math.Min(y2, y3))
	yMax := math.Max(math.Max(y0, y1), math.Max(y2, y3))
	if y < yMin || y > yMax {
		return nil
	}

	// If curve is flat enough or max depth reached, use linear approximation
	if depth > 10 || isFlatEnough(x0, y0, x1, y1, x2, y2, x3, y3) {
		line := &LineSegment{X0: x0, Y0: y0, X1: x3, Y1: y3}
		return line.IntersectLine(y)
	}

	// Subdivide at t=0.5
	mx0 := (x0 + x1) / 2
	my0 := (y0 + y1) / 2
	mx1 := (x1 + x2) / 2
	my1 := (y1 + y2) / 2
	mx2 := (x2 + x3) / 2
	my2 := (y2 + y3) / 2

	mx3 := (mx0 + mx1) / 2
	my3 := (my0 + my1) / 2
	mx4 := (mx1 + mx2) / 2
	my4 := (my1 + my2) / 2

	mx5 := (mx3 + mx4) / 2
	my5 := (my3 + my4) / 2

	var results []float64
	results = append(results, cubicIntersectLine(x0, y0, mx0, my0, mx3, my3, mx5, my5, y, depth+1)...)
	results = append(results, cubicIntersectLine(mx5, my5, mx4, my4, mx2, my2, x3, y3, y, depth+1)...)
	return results
}

func isFlatEnough(x0, y0, x1, y1, x2, y2, x3, y3 float64) bool {
	// Check if control points are close to the line from start to end
	const tolerance = 0.5
	dx := x3 - x0
	dy := y3 - y0
	d := math.Sqrt(dx*dx + dy*dy)
	if d < 1e-10 {
		return true
	}
	// Distance from control points to line
	d1 := math.Abs((x1-x0)*dy-(y1-y0)*dx) / d
	d2 := math.Abs((x2-x0)*dy-(y2-y0)*dx) / d
	return d1 < tolerance && d2 < tolerance
}

// DecoPoint represents a 2D coordinate for decoration.
type DecoPoint struct {
	X, Y Abs
}

// DecoSize represents 2D dimensions for decoration.
type DecoSize struct {
	Width, Height Abs
}

// Decorate adds line decorations to a frame for a text item.
func Decorate(
	frame *DecoFrame,
	deco *Decoration,
	text *DecoTextItem,
	width Abs,
	shift Abs,
	pos DecoPoint,
) {
	fontMetrics := text.Font.Metrics()

	// Handle highlight decoration
	if highlight, ok := deco.Line.(*HighlightDeco); ok {
		top, bottom := determineEdges(text, highlight.TopEdge, highlight.BottomEdge)
		size := DecoSize{Width: width + 2.0*deco.Extent, Height: top + bottom}

		shapes := styledRect(size, highlight.Radius, highlight.Fill, highlight.Stroke)
		origin := DecoPoint{X: pos.X - deco.Extent, Y: pos.Y - top - shift}

		// Prepend shapes (they go behind text)
		for _, shape := range shapes {
			frame.Prepend(origin, DecoShapeItem{Shape: shape})
		}
		return
	}

	// Handle line decorations (strikethrough, overline, underline)
	var stroke *FixedStroke
	var metrics DecoMetrics
	var offset Abs
	var evade bool
	var background bool

	switch d := deco.Line.(type) {
	case *StrikethroughDeco:
		stroke = d.Stroke
		metrics = fontMetrics.Strikethrough
		if d.Offset != nil {
			offset = -*d.Offset - shift
		} else {
			offset = -metrics.Position.At(text.Size) - shift
		}
		evade = false
		background = d.Background
	case *OverlineDeco:
		stroke = d.Stroke
		metrics = fontMetrics.Overline
		if d.Offset != nil {
			offset = -*d.Offset - shift
		} else {
			offset = -metrics.Position.At(text.Size) - shift
		}
		evade = d.Evade
		background = d.Background
	case *UnderlineDeco:
		stroke = d.Stroke
		metrics = fontMetrics.Underline
		if d.Offset != nil {
			offset = -*d.Offset - shift
		} else {
			offset = -metrics.Position.At(text.Size) - shift
		}
		evade = d.Evade
		background = d.Background
	default:
		return
	}

	// Use default stroke if not specified
	if stroke == nil {
		s := StrokeFromPair(text.Fill, metrics.Thickness.At(text.Size))
		stroke = &s
	}

	gapPadding := 0.08 * text.Size
	minWidth := 0.162 * text.Size

	start := pos.X - deco.Extent
	end := pos.X + width + deco.Extent

	pushSegment := func(from, to Abs, prepend bool) {
		origin := DecoPoint{X: from, Y: pos.Y + offset}
		target := DecoPoint{X: to - from, Y: 0}

		if target.X >= minWidth || !evade {
			shape := DecoLineShape{
				Target: target,
				Stroke: *stroke,
			}

			if prepend {
				frame.Prepend(origin, DecoShapeItem{Shape: shape})
			} else {
				frame.Push(origin, DecoShapeItem{Shape: shape})
			}
		}
	}

	if !evade {
		pushSegment(start, end, background)
		return
	}

	// Find intersections with glyph outlines
	lineY := float64(offset)
	var x Abs
	var intersections []Abs

	for _, glyph := range text.Glyphs {
		dx := glyph.XOffset.At(text.Size) + x

		segments, bbox, ok := text.Font.OutlineGlyph(glyph.ID)

		x += glyph.XAdvance.At(text.Size)

		// Only do intersection test if line intersects bounding box
		if !ok || bbox == nil {
			continue
		}

		yMin := float64(-text.Font.ToEm(bbox.YMax).At(text.Size))
		yMax := float64(-text.Font.ToEm(bbox.YMin).At(text.Size))
		if lineY < yMin || lineY > yMax {
			continue
		}

		// Find intersections with each segment
		for _, seg := range segments {
			for _, ix := range seg.IntersectLine(lineY) {
				intersections = append(intersections, Abs(ix)+dx+pos.X)
			}
		}
	}

	// Add start and end points with padding
	intersections = append(intersections, start-gapPadding)
	intersections = append(intersections, end+gapPadding)

	// Sort intersections left to right
	sort.Slice(intersections, func(i, j int) bool {
		return intersections[i] < intersections[j]
	})

	// Draw segments between intersections
	for i := 0; i+1 < len(intersections); i += 2 {
		l := intersections[i]
		r := intersections[i+1]

		// Skip if too close
		if r-l < gapPadding {
			continue
		}
		pushSegment(l+gapPadding, r-gapPadding, background)
	}
}

// determineEdges calculates top and bottom edges based on glyph bounds.
func determineEdges(text *DecoTextItem, topEdge TopEdge, bottomEdge BottomEdge) (top, bottom Abs) {
	fontMetrics := text.Font.Metrics()

	// Default to font metrics-based values
	// In a complete implementation, this would calculate actual glyph bounds
	switch topEdge {
	case TopEdgeAscender:
		top = Em(0.8).At(text.Size) // Typical ascender
	case TopEdgeCapHeight:
		top = Em(0.7).At(text.Size) // Typical cap height
	case TopEdgeXHeight:
		top = Em(0.5).At(text.Size) // Typical x-height
	case TopEdgeBounds:
		// Would calculate actual bounds from glyphs
		top = Em(0.8).At(text.Size)
	}

	switch bottomEdge {
	case BottomEdgeDescender:
		bottom = Em(0.2).At(text.Size) // Typical descender
	case BottomEdgeBaseline:
		bottom = 0
	case BottomEdgeBounds:
		// Would calculate actual bounds from glyphs
		bottom = Em(0.2).At(text.Size)
	}

	// Iterate glyphs to get actual bounds if using Bounds edge
	if topEdge == TopEdgeBounds || bottomEdge == BottomEdge(TopEdgeBounds) {
		for _, g := range text.Glyphs {
			_, bbox, ok := text.Font.OutlineGlyph(g.ID)
			if !ok || bbox == nil {
				continue
			}
			t := -text.Font.ToEm(bbox.YMin).At(text.Size)
			b := -text.Font.ToEm(bbox.YMax).At(text.Size)
			if t > top {
				top = t
			}
			if b > bottom {
				bottom = b
			}
		}
	}

	_ = fontMetrics // Silence unused warning; would use in complete impl
	return top, bottom
}

// DecoFrame represents a layout frame that can hold positioned items.
type DecoFrame struct {
	Size     DecoSize
	Baseline Abs
	Items    []DecoFrameEntry
}

// DecoFrameEntry is a positioned item in a frame.
type DecoFrameEntry struct {
	Pos  DecoPoint
	Item DecoFrameItem
}

// DecoFrameItem represents an item that can be placed in a frame.
type DecoFrameItem interface {
	isDecoFrameItem()
}

// DecoShapeItem represents a shape in a frame.
type DecoShapeItem struct {
	Shape interface{}
}

func (DecoShapeItem) isDecoFrameItem() {}

// DecoTextFrameItem represents text in a frame.
type DecoTextFrameItem struct {
	Text *ShapedText
}

func (DecoTextFrameItem) isDecoFrameItem() {}

// Push adds an item to the frame.
func (f *DecoFrame) Push(pos DecoPoint, item DecoFrameItem) {
	f.Items = append(f.Items, DecoFrameEntry{Pos: pos, Item: item})
}

// Prepend adds an item to the front of the frame.
func (f *DecoFrame) Prepend(pos DecoPoint, item DecoFrameItem) {
	f.Items = append([]DecoFrameEntry{{Pos: pos, Item: item}}, f.Items...)
}

// PrependMultiple adds multiple items to the front of the frame.
func (f *DecoFrame) PrependMultiple(entries []DecoFrameEntry) {
	f.Items = append(entries, f.Items...)
}

// DecoLineShape represents a line to be stroked.
type DecoLineShape struct {
	Target DecoPoint
	Stroke FixedStroke
}

// DecoRectShape represents a rectangle.
type DecoRectShape struct {
	Size   DecoSize
	Radius Abs
	Fill   interface{}
	Stroke *FixedStroke
}

// styledRect creates shape(s) for a styled rectangle.
func styledRect(size DecoSize, radius Abs, fill interface{}, stroke *FixedStroke) []interface{} {
	var shapes []interface{}

	if fill != nil {
		shapes = append(shapes, DecoRectShape{
			Size:   size,
			Radius: radius,
			Fill:   fill,
		})
	}

	if stroke != nil {
		shapes = append(shapes, DecoRectShape{
			Size:   size,
			Radius: radius,
			Stroke: stroke,
		})
	}

	return shapes
}
