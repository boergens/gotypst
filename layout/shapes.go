package layout

import (
	"fmt"
	"math"

	"github.com/boergens/gotypst/syntax"
)

// SQRT2 is the square root of 2.
const SQRT2 = math.Sqrt2

// LineElem represents a line element configuration.
type LineElem struct {
	Start  Axes[Rel]
	End    *Axes[Rel]
	Length Abs
	Angle  float64 // Radians
	Stroke *Stroke
	Span   syntax.Span
}

// LayoutLine lays out a line element.
//
// A line can be defined either by start/end points or by start point,
// length, and angle.
func LayoutLine(
	elem *LineElem,
	region Region,
) (*Frame, error) {
	// Resolve start point relative to region.
	start := NewPoint(
		elem.Start.X.RelativeTo(region.Size.X),
		elem.Start.Y.RelativeTo(region.Size.Y),
	)

	// Calculate delta from either end point or length/angle.
	var delta Point
	if elem.End != nil {
		end := NewPoint(
			elem.End.X.RelativeTo(region.Size.X),
			elem.End.Y.RelativeTo(region.Size.Y),
		)
		delta = end.Sub(start)
	} else {
		// Calculate from length and angle.
		x := math.Cos(elem.Angle) * float64(elem.Length)
		y := math.Sin(elem.Angle) * float64(elem.Length)
		delta = NewPoint(
			Rel{Abs: Abs(x)}.RelativeTo(region.Size.X),
			Rel{Abs: Abs(y)}.RelativeTo(region.Size.Y),
		)
	}

	// Get stroke or use default.
	stroke := elem.Stroke
	if stroke == nil {
		stroke = &Stroke{
			Paint:     Paint{Color: &ColorBlack},
			Thickness: Pt(1),
		}
	}

	// Calculate size from start and endpoint.
	endPoint := start.Add(delta)
	size := NewSize(
		start.X.Max(endPoint.X).Max(0),
		start.Y.Max(endPoint.Y).Max(0),
	)

	if !size.X.IsFinite() || !size.Y.IsFinite() {
		return nil, fmt.Errorf("cannot create line with infinite length at %v", elem.Span)
	}

	frame := NewFrameSoft(size)

	// Create the line curve.
	curve := Curve{
		Segments: []CurveSegment{
			MoveTo{Point: PointZero()},
			LineTo{Point: delta},
		},
	}

	frame.Push(start, ShapeItem{
		Curve:  curve,
		Stroke: stroke,
		Span:   elem.Span,
	})

	return frame, nil
}

// PolygonElem represents a polygon element configuration.
type PolygonElem struct {
	Vertices []Axes[Rel]
	Fill     *Paint
	FillRule FillRule
	Stroke   *Stroke
	Span     syntax.Span
}

// FillRule represents how to determine the interior of a shape.
type FillRule int

const (
	FillRuleNonZero FillRule = iota
	FillRuleEvenOdd
)

// LayoutPolygon lays out a polygon element.
func LayoutPolygon(
	elem *PolygonElem,
	region Region,
) (*Frame, error) {
	// Resolve all vertices relative to region.
	points := make([]Point, len(elem.Vertices))
	for i, v := range elem.Vertices {
		points[i] = NewPoint(
			v.X.RelativeTo(region.Size.X),
			v.Y.RelativeTo(region.Size.Y),
		)
	}

	// Calculate bounding size.
	maxPoint := PointZero()
	for _, p := range points {
		if p.X > maxPoint.X {
			maxPoint.X = p.X
		}
		if p.Y > maxPoint.Y {
			maxPoint.Y = p.Y
		}
	}
	size := NewSize(maxPoint.X, maxPoint.Y)

	if !size.X.IsFinite() || !size.Y.IsFinite() {
		return nil, fmt.Errorf("cannot create polygon with infinite size at %v", elem.Span)
	}

	frame := NewFrameHard(size)

	// Only create a curve if there are points.
	if len(points) == 0 {
		return frame, nil
	}

	// Prepare fill and stroke.
	fill := elem.Fill
	stroke := elem.Stroke
	if stroke == nil && fill == nil {
		// Default stroke if no fill.
		stroke = &Stroke{
			Paint:     Paint{Color: &ColorBlack},
			Thickness: Pt(1),
		}
	}

	// Construct a closed curve from all points.
	segments := make([]CurveSegment, 0, len(points)+1)
	segments = append(segments, MoveTo{Point: points[0]})
	for _, p := range points[1:] {
		segments = append(segments, LineTo{Point: p})
	}
	segments = append(segments, ClosePath{})

	curve := Curve{Segments: segments, Closed: true}

	frame.Push(PointZero(), ShapeItem{
		Curve:  curve,
		Fill:   fill,
		Stroke: stroke,
		Span:   elem.Span,
	})

	return frame, nil
}

// RectElem represents a rectangle element configuration.
type RectElem struct {
	Body   interface{}
	Fill   *Paint
	Stroke *Sides[*Stroke]
	Inset  Sides[Rel]
	Outset Sides[Rel]
	Radius Corners[Rel]
	Span   syntax.Span
}

// LayoutRect lays out a rectangle element.
func LayoutRect(
	elem *RectElem,
	layoutFunc LayoutFunc,
	styles StyleChain,
	region Region,
) (*Frame, error) {
	return layoutShape(
		layoutFunc,
		styles,
		region,
		shapeKindRect,
		elem.Body,
		elem.Fill,
		elem.Stroke,
		elem.Inset,
		elem.Outset,
		elem.Radius,
		elem.Span,
	)
}

// SquareElem represents a square element configuration.
type SquareElem struct {
	Body   interface{}
	Fill   *Paint
	Stroke *Sides[*Stroke]
	Inset  Sides[Rel]
	Outset Sides[Rel]
	Radius Corners[Rel]
	Span   syntax.Span
}

// LayoutSquare lays out a square element.
func LayoutSquare(
	elem *SquareElem,
	layoutFunc LayoutFunc,
	styles StyleChain,
	region Region,
) (*Frame, error) {
	return layoutShape(
		layoutFunc,
		styles,
		region,
		shapeKindSquare,
		elem.Body,
		elem.Fill,
		elem.Stroke,
		elem.Inset,
		elem.Outset,
		elem.Radius,
		elem.Span,
	)
}

// EllipseElem represents an ellipse element configuration.
type EllipseElem struct {
	Body   interface{}
	Fill   *Paint
	Stroke *Stroke
	Inset  Sides[Rel]
	Outset Sides[Rel]
	Span   syntax.Span
}

// LayoutEllipse lays out an ellipse element.
func LayoutEllipse(
	elem *EllipseElem,
	layoutFunc LayoutFunc,
	styles StyleChain,
	region Region,
) (*Frame, error) {
	var strokeSides *Sides[*Stroke]
	if elem.Stroke != nil {
		strokeSides = &Sides[*Stroke]{
			Left:   elem.Stroke,
			Top:    elem.Stroke,
			Right:  elem.Stroke,
			Bottom: elem.Stroke,
		}
	}
	return layoutShape(
		layoutFunc,
		styles,
		region,
		shapeKindEllipse,
		elem.Body,
		elem.Fill,
		strokeSides,
		elem.Inset,
		elem.Outset,
		Corners[Rel]{}, // No radius for ellipse
		elem.Span,
	)
}

// CircleElem represents a circle element configuration.
type CircleElem struct {
	Body   interface{}
	Fill   *Paint
	Stroke *Stroke
	Inset  Sides[Rel]
	Outset Sides[Rel]
	Span   syntax.Span
}

// LayoutCircle lays out a circle element.
func LayoutCircle(
	elem *CircleElem,
	layoutFunc LayoutFunc,
	styles StyleChain,
	region Region,
) (*Frame, error) {
	var strokeSides *Sides[*Stroke]
	if elem.Stroke != nil {
		strokeSides = &Sides[*Stroke]{
			Left:   elem.Stroke,
			Top:    elem.Stroke,
			Right:  elem.Stroke,
			Bottom: elem.Stroke,
		}
	}
	return layoutShape(
		layoutFunc,
		styles,
		region,
		shapeKindCircle,
		elem.Body,
		elem.Fill,
		strokeSides,
		elem.Inset,
		elem.Outset,
		Corners[Rel]{}, // No radius for circle
		elem.Span,
	)
}

// shapeKind is a category of shape.
type shapeKind int

const (
	shapeKindSquare  shapeKind = iota // Rectangle with equal side lengths
	shapeKindRect                     // Quadrilateral with four right angles
	shapeKindCircle                   // Ellipse with coinciding foci
	shapeKindEllipse                  // Curve around two focal points
)

// isRound returns whether this shape kind is curvy.
func (k shapeKind) isRound() bool {
	return k == shapeKindCircle || k == shapeKindEllipse
}

// isQuadratic returns whether this shape kind has equal side lengths.
func (k shapeKind) isQuadratic() bool {
	return k == shapeKindSquare || k == shapeKindCircle
}

// layoutShape lays out a shape element.
func layoutShape(
	layoutFunc LayoutFunc,
	styles StyleChain,
	region Region,
	kind shapeKind,
	body interface{},
	fill *Paint,
	stroke *Sides[*Stroke],
	inset Sides[Rel],
	outset Sides[Rel],
	radius Corners[Rel],
	span syntax.Span,
) (*Frame, error) {
	var frame *Frame

	if body != nil {
		// Apply extra inset to round shapes.
		workingInset := inset
		if kind.isRound() {
			extraInset := Ratio(0.5 - SQRT2/4.0)
			workingInset = Sides[Rel]{
				Left:   Rel{Abs: inset.Left.Abs, Rel: inset.Left.Rel + extraInset},
				Top:    Rel{Abs: inset.Top.Abs, Rel: inset.Top.Rel + extraInset},
				Right:  Rel{Abs: inset.Right.Abs, Rel: inset.Right.Rel + extraInset},
				Bottom: Rel{Abs: inset.Bottom.Abs, Rel: inset.Bottom.Rel + extraInset},
			}
		}

		hasInset := !workingInset.Left.IsZero() || !workingInset.Top.IsZero() ||
			!workingInset.Right.IsZero() || !workingInset.Bottom.IsZero()

		// Take the inset into account.
		pod := region
		if hasInset {
			sidesRel := NewSidesRel(workingInset.Left, workingInset.Top, workingInset.Right, workingInset.Bottom)
			pod.Size = Shrink(region.Size, sidesRel)
		}

		// If the shape is quadratic, measure first then layout with full expansion.
		if kind.isQuadratic() {
			length := quadraticSize(pod)
			if length != nil {
				pod = Region{
					Size:   NewSize(*length, *length),
					Expand: Axes[bool]{X: true, Y: true},
				}
			} else {
				// Measure the child to determine size.
				measured, err := layoutFunc(body, styles, &Regions{
					Size:   pod.Size,
					Full:   pod.Size,
					Expand: pod.Expand,
				})
				if err != nil {
					return nil, err
				}
				if len(measured) > 0 {
					maxSide := measured[0].Size.X.Max(measured[0].Size.Y)
					minAvail := pod.Size.X.Min(pod.Size.Y)
					length := maxSide.Min(minAvail)
					pod = Region{
						Size:   NewSize(length, length),
						Expand: Axes[bool]{X: true, Y: true},
					}
				}
			}
		}

		// Layout the child.
		fragment, err := layoutFunc(body, styles, &Regions{
			Size:   pod.Size,
			Full:   pod.Size,
			Expand: pod.Expand,
		})
		if err != nil {
			return nil, err
		}

		if len(fragment) > 0 {
			frameCopy := fragment[0]
			frame = &frameCopy
		} else {
			frame = NewFrameSoft(pod.Size)
		}

		// Apply the inset.
		if hasInset {
			sidesRel := NewSidesRel(workingInset.Left, workingInset.Top, workingInset.Right, workingInset.Bottom)
			Grow(frame, sidesRel)
		}
	} else {
		// Default size for shapes without children.
		defaultSize := NewSize(Pt(45.0), Pt(30.0)).Min(region.Size)

		var size Size
		if kind.isQuadratic() {
			length := quadraticSize(region)
			if length != nil {
				size = NewSize(*length, *length)
			} else {
				minDefault := defaultSize.X.Min(defaultSize.Y)
				size = NewSize(minDefault, minDefault)
			}
		} else {
			// For each dimension, pick the region size if forced,
			// otherwise use the default size.
			if region.Expand.X {
				size.X = region.Size.X
			} else {
				size.X = defaultSize.X
			}
			if region.Expand.Y {
				size.Y = region.Size.Y
			} else {
				size.Y = defaultSize.Y
			}
		}

		frame = NewFrameSoft(size)
	}

	// Prepare stroke.
	var resolvedStroke Sides[*Stroke]
	if stroke != nil {
		resolvedStroke = *stroke
	} else if fill == nil {
		// Default stroke if no fill and no stroke specified.
		defaultStroke := &Stroke{
			Paint:     Paint{Color: &ColorBlack},
			Thickness: Pt(1),
		}
		resolvedStroke = Sides[*Stroke]{
			Left:   defaultStroke,
			Top:    defaultStroke,
			Right:  defaultStroke,
			Bottom: defaultStroke,
		}
	}

	// Add fill and/or stroke.
	hasStroke := resolvedStroke.Left != nil || resolvedStroke.Top != nil ||
		resolvedStroke.Right != nil || resolvedStroke.Bottom != nil
	if fill != nil || hasStroke {
		if kind.isRound() {
			resolvedOutset := SidesRelToAbs(outset, frame.Size)
			outsetSize := NewSize(
				frame.Size.X+resolvedOutset.Left+resolvedOutset.Right,
				frame.Size.Y+resolvedOutset.Top+resolvedOutset.Bottom,
			)
			pos := NewPoint(-resolvedOutset.Left, -resolvedOutset.Top)

			// Create ellipse curve.
			ellipseCurve := CurveEllipse(outsetSize)

			shape := ShapeItem{
				Curve:  ellipseCurve,
				Fill:   fill,
				Stroke: resolvedStroke.Left,
				Span:   span,
			}
			frame.Items = append([]FrameItem{PositionedItem{Position: pos, Item: shape}}, frame.Items...)
		} else {
			FillAndStroke(frame, fill, &resolvedStroke, &outset, &radius, span)
		}
	}

	return frame, nil
}

// quadraticSize determines the forced size of a quadratic shape based on the region.
func quadraticSize(region Region) *Abs {
	if region.Expand.X && region.Expand.Y {
		// If both width and height are specified, choose the smaller one.
		length := region.Size.X.Min(region.Size.Y)
		return &length
	} else if region.Expand.X {
		return &region.Size.X
	} else if region.Expand.Y {
		return &region.Size.Y
	}
	return nil
}

// SidesRelToAbs resolves sides relative to a given size.
func SidesRelToAbs(s Sides[Rel], size Size) Sides[Abs] {
	return Sides[Abs]{
		Left:   s.Left.RelativeTo(size.X),
		Top:    s.Top.RelativeTo(size.Y),
		Right:  s.Right.RelativeTo(size.X),
		Bottom: s.Bottom.RelativeTo(size.Y),
	}
}

// CurveEllipse creates an ellipse curve.
func CurveEllipse(size Size) Curve {
	// Approximate ellipse with 4 cubic bezier curves.
	// Control point distance for quarter circle: (4/3) * tan(π/8) ≈ 0.5522847498
	k := 0.5522847498

	rx := float64(size.X) / 2
	ry := float64(size.Y) / 2
	cx := rx
	cy := ry

	kx := Abs(k * rx)
	ky := Abs(k * ry)

	return Curve{
		Segments: []CurveSegment{
			MoveTo{Point: NewPoint(Abs(cx+rx), Abs(cy))},
			CubicTo{
				Control1: NewPoint(Abs(cx+rx), Abs(cy)+ky),
				Control2: NewPoint(Abs(cx)+kx, Abs(cy+ry)),
				End:      NewPoint(Abs(cx), Abs(cy+ry)),
			},
			CubicTo{
				Control1: NewPoint(Abs(cx)-kx, Abs(cy+ry)),
				Control2: NewPoint(Abs(cx-rx), Abs(cy)+ky),
				End:      NewPoint(Abs(cx-rx), Abs(cy)),
			},
			CubicTo{
				Control1: NewPoint(Abs(cx-rx), Abs(cy)-ky),
				Control2: NewPoint(Abs(cx)-kx, Abs(cy-ry)),
				End:      NewPoint(Abs(cx), Abs(cy-ry)),
			},
			CubicTo{
				Control1: NewPoint(Abs(cx)+kx, Abs(cy-ry)),
				Control2: NewPoint(Abs(cx+rx), Abs(cy)-ky),
				End:      NewPoint(Abs(cx+rx), Abs(cy)),
			},
			ClosePath{},
		},
		Closed: true,
	}
}

// FillAndStroke adds fill and stroke with optional radius and outset to the frame.
func FillAndStroke(
	frame *Frame,
	fill *Paint,
	stroke *Sides[*Stroke],
	outset *Sides[Rel],
	radius *Corners[Rel],
	span syntax.Span,
) {
	resolvedOutset := SidesRelToAbs(*outset, frame.Size)
	size := NewSize(
		frame.Size.X+resolvedOutset.Left+resolvedOutset.Right,
		frame.Size.Y+resolvedOutset.Top+resolvedOutset.Bottom,
	)
	pos := NewPoint(-resolvedOutset.Left, -resolvedOutset.Top)

	// Create the styled rectangle shapes.
	shapes := StyledRect(size, radius, fill, stroke)

	// Prepend shapes to frame.
	newItems := make([]FrameItem, 0, len(shapes)+len(frame.Items))
	for _, shape := range shapes {
		newItems = append(newItems, PositionedItem{Position: pos, Item: shape})
	}
	newItems = append(newItems, frame.Items...)
	frame.Items = newItems
}

// StyledRect creates a styled rectangle with shapes.
func StyledRect(
	size Size,
	radius *Corners[Rel],
	fill *Paint,
	stroke *Sides[*Stroke],
) []ShapeItem {
	// Check if stroke is uniform and radius is zero.
	strokeUniform := stroke.Left == stroke.Top && stroke.Top == stroke.Right && stroke.Right == stroke.Bottom
	radiusZero := radius.TopLeft.IsZero() && radius.TopRight.IsZero() &&
		radius.BottomRight.IsZero() && radius.BottomLeft.IsZero()

	if strokeUniform && radiusZero {
		return SimpleRect(size, fill, stroke.Left)
	}
	return SegmentedRect(size, radius, fill, stroke)
}

// SimpleRect creates a simple rectangle.
func SimpleRect(size Size, fill *Paint, stroke *Stroke) []ShapeItem {
	return []ShapeItem{{
		Curve:  CurveRect(size),
		Fill:   fill,
		Stroke: stroke,
	}}
}

// SegmentedRect creates a rectangle with potentially different strokes per side.
func SegmentedRect(
	size Size,
	radius *Corners[Rel],
	fill *Paint,
	stroke *Sides[*Stroke],
) []ShapeItem {
	// Resolve radius relative to size.
	baseRadius := size.X.Min(size.Y) / 2
	resolvedRadius := Corners[Abs]{
		TopLeft:     radius.TopLeft.RelativeTo(baseRadius * 2).Min(baseRadius),
		TopRight:    radius.TopRight.RelativeTo(baseRadius * 2).Min(baseRadius),
		BottomRight: radius.BottomRight.RelativeTo(baseRadius * 2).Min(baseRadius),
		BottomLeft:  radius.BottomLeft.RelativeTo(baseRadius * 2).Min(baseRadius),
	}

	// Create the rectangle curve with rounded corners.
	curve := CurveRoundedRect(size, resolvedRadius)

	shapes := make([]ShapeItem, 0, 5)

	// Add fill shape.
	if fill != nil {
		shapes = append(shapes, ShapeItem{
			Curve: curve,
			Fill:  fill,
		})
	}

	// Add stroke shapes for each side if different.
	// For simplicity, if strokes differ, we draw the whole outline with each stroke.
	// A more sophisticated implementation would stroke each side separately.
	if stroke.Left != nil {
		shapes = append(shapes, ShapeItem{
			Curve:  curve,
			Stroke: stroke.Left,
		})
	}

	return shapes
}

// CurveRoundedRect creates a rounded rectangle curve.
func CurveRoundedRect(size Size, radius Corners[Abs]) Curve {
	// Control point factor for quarter circle approximation.
	k := 0.5522847498

	segments := make([]CurveSegment, 0, 12)

	// Start at top-left, after the corner.
	startX := radius.TopLeft
	segments = append(segments, MoveTo{Point: NewPoint(startX, 0)})

	// Top edge to top-right corner.
	segments = append(segments, LineTo{Point: NewPoint(size.X-radius.TopRight, 0)})

	// Top-right corner.
	if radius.TopRight > 0 {
		segments = append(segments, CubicTo{
			Control1: NewPoint(size.X-radius.TopRight+Abs(k*float64(radius.TopRight)), 0),
			Control2: NewPoint(size.X, radius.TopRight-Abs(k*float64(radius.TopRight))),
			End:      NewPoint(size.X, radius.TopRight),
		})
	}

	// Right edge to bottom-right corner.
	segments = append(segments, LineTo{Point: NewPoint(size.X, size.Y-radius.BottomRight)})

	// Bottom-right corner.
	if radius.BottomRight > 0 {
		segments = append(segments, CubicTo{
			Control1: NewPoint(size.X, size.Y-radius.BottomRight+Abs(k*float64(radius.BottomRight))),
			Control2: NewPoint(size.X-radius.BottomRight+Abs(k*float64(radius.BottomRight)), size.Y),
			End:      NewPoint(size.X-radius.BottomRight, size.Y),
		})
	}

	// Bottom edge to bottom-left corner.
	segments = append(segments, LineTo{Point: NewPoint(radius.BottomLeft, size.Y)})

	// Bottom-left corner.
	if radius.BottomLeft > 0 {
		segments = append(segments, CubicTo{
			Control1: NewPoint(radius.BottomLeft-Abs(k*float64(radius.BottomLeft)), size.Y),
			Control2: NewPoint(0, size.Y-radius.BottomLeft+Abs(k*float64(radius.BottomLeft))),
			End:      NewPoint(0, size.Y-radius.BottomLeft),
		})
	}

	// Left edge to top-left corner.
	segments = append(segments, LineTo{Point: NewPoint(0, radius.TopLeft)})

	// Top-left corner.
	if radius.TopLeft > 0 {
		segments = append(segments, CubicTo{
			Control1: NewPoint(0, radius.TopLeft-Abs(k*float64(radius.TopLeft))),
			Control2: NewPoint(radius.TopLeft-Abs(k*float64(radius.TopLeft)), 0),
			End:      NewPoint(radius.TopLeft, 0),
		})
	}

	segments = append(segments, ClosePath{})

	return Curve{Segments: segments, Closed: true}
}
