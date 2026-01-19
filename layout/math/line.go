package math

// LineItem represents an overline/underline for layout.
type LineItem struct {
	Base     *MathItem
	Position Position
}

// layoutLineImpl lays out an overline or underline.
func layoutLineImpl(
	item *LineItem,
	ctx *MathContext,
	styles *StyleChain,
	props *MathProperties,
) error {
	var extraHeight Abs
	var linePos, contentPos Point
	var baseline, thickness, lineAdjust Abs

	// Layout content
	contentFrag, err := ctx.LayoutIntoFragment(item.Base, styles)
	if err != nil {
		return err
	}

	font, fontSize := FragmentFont(contentFrag, ctx, styles)
	constants := font.Math()

	switch item.Position {
	case PositionBelow:
		// Underline
		sep := constants.UnderbarExtraDescender.At(fontSize)
		thickness = constants.UnderbarRuleThickness.At(fontSize)
		gap := constants.UnderbarVerticalGap.At(fontSize)
		extraHeight = sep + thickness + gap

		linePos = PointWithY(contentFrag.Height() + gap + thickness/2.0)
		contentPos = Point{}
		baseline = contentFrag.Ascent()
		lineAdjust = -contentFrag.ItalicsCorrection()

	case PositionAbove:
		// Overline
		sep := constants.OverbarExtraAscender.At(fontSize)
		thickness = constants.OverbarRuleThickness.At(fontSize)
		gap := constants.OverbarVerticalGap.At(fontSize)
		extraHeight = sep + thickness + gap

		linePos = PointWithY(sep + thickness/2.0)
		contentPos = PointWithY(extraHeight)
		baseline = contentFrag.Ascent() + extraHeight
		lineAdjust = 0
	}

	width := contentFrag.Width()
	height := contentFrag.Height() + extraHeight
	size := Size{X: width, Y: height}
	lineWidth := width + lineAdjust

	contentTextLike := contentFrag.IsTextLike()
	contentItalicsCorrection := contentFrag.ItalicsCorrection()

	frame := NewSoftFrame(size)
	frame.SetBaseline(baseline)
	frame.PushFrame(contentPos, contentFrag.IntoFrame())

	// Draw the line
	// TODO: Get text fill from styles
	line := Stroked(
		&LineGeometry{Delta: PointWithX(lineWidth)},
		StrokeFromPair(nil, thickness),
	)
	frame.PushShape(linePos, line, props.Span)

	ff := NewFrameFragment(props, fontSize, frame)
	ff.ItalicsCorr = contentItalicsCorrection
	ff.TextLike = contentTextLike

	ctx.Push(ff)
	return nil
}
