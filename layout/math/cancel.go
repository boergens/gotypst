package math

import (
	"math"
)

// CancelItem represents a cancel mark for layout.
type CancelItem struct {
	Base            *MathItem
	Length          Rel
	Stroke          FixedStroke
	InvertFirstLine bool
	Angle           *CancelAngle
	Cross           bool
}

// CancelAngle specifies the angle for cancel lines.
type CancelAngle struct {
	Angle *Angle // Absolute angle
	// Func could be a function for computing angle, but omitted for simplicity
}

// layoutCancelImpl lays out a cancel mark.
func layoutCancelImpl(
	item *CancelItem,
	ctx *MathContext,
	styles *StyleChain,
	props *MathProperties,
) error {
	// Layout body
	bodyFrag, err := ctx.LayoutIntoFragment(item.Base, styles)
	if err != nil {
		return err
	}

	// Preserve body properties
	bodyTextLike := bodyFrag.IsTextLike()
	bodyItalics := bodyFrag.ItalicsCorrection()
	bodyAttachTop, bodyAttachBottom := bodyFrag.AccentAttach()

	body := bodyFrag.IntoFrame()
	bodySize := body.Size()

	// Draw first line
	firstLine := drawCancelLine(
		item.Length,
		item.Stroke,
		item.InvertFirstLine,
		item.Angle,
		bodySize,
		props.Span,
	)

	// Push line at center
	center := Point{X: bodySize.X / 2.0, Y: bodySize.Y / 2.0}
	body.PushFrame(center, firstLine)

	// Draw second line if cross
	if item.Cross {
		secondLine := drawCancelLine(
			item.Length,
			item.Stroke,
			true, // Invert
			item.Angle,
			bodySize,
			props.Span,
		)
		body.PushFrame(center, secondLine)
	}

	ff := NewFrameFragment(props, styles.ResolveTextSize(), body)
	ff.ItalicsCorr = bodyItalics
	ff.TextLike = bodyTextLike
	ff.AccentAttachTop, ff.AccentAttachBottom = bodyAttachTop, bodyAttachBottom

	ctx.Push(ff)
	return nil
}

// drawCancelLine draws a cancel line.
func drawCancelLine(
	lengthScale Rel,
	stroke FixedStroke,
	invert bool,
	angle *CancelAngle,
	bodySize Size,
	span Span,
) *Frame {
	// Default angle is diagonal
	defaultAngle := defaultCancelAngle(bodySize)

	var lineAngle Angle
	if angle == nil {
		lineAngle = defaultAngle
	} else if angle.Angle != nil {
		lineAngle = *angle.Angle
	} else {
		lineAngle = defaultAngle
	}

	// Invert means flipping along y-axis
	if invert {
		lineAngle = -lineAngle
	}

	// Default length is diagonal of body box
	defaultLength := bodySize.ToPoint().Hypot()
	length := lengthScale.RelativeTo(defaultLength)

	// Draw vertical line and rotate
	start := Point{X: 0, Y: length / 2.0}
	delta := Point{X: 0, Y: -length}

	frame := NewSoftFrame(bodySize)
	line := Stroked(&LineGeometry{Delta: delta}, stroke)
	frame.PushShape(start, line, span)

	// Apply rotation
	frame.Transform(TransformRotate(lineAngle))

	return frame
}

// defaultCancelAngle calculates the default cancel line angle.
func defaultCancelAngle(body Size) Angle {
	// Default is the diagonal angle w.r.t y-axis
	return Rad(math.Atan(float64(body.X) / float64(body.Y)))
}
