package math

// RadicalItem represents a radical (root) for layout.
type RadicalItem struct {
	Radicand *MathItem
	Index    *MathItem // Optional root index
	Sqrt     *MathItem // Square root symbol
}

// layoutRadicalImpl lays out a radical symbol with radicand.
func layoutRadicalImpl(
	item *RadicalItem,
	ctx *MathContext,
	styles *StyleChain,
	props *MathProperties,
) error {
	// Layout radicand
	radicandFrag, err := ctx.LayoutIntoFragment(item.Radicand, styles)
	if err != nil {
		return err
	}

	// Check if radicand is multiline
	multiline := false // TODO: Check if item.Radicand.IsMultiline()
	var radicand *Frame
	if multiline {
		// Align frame center with math axis
		font, fontSize := FragmentFont(radicandFrag, ctx, styles)
		axis := font.Math().AxisHeight.At(fontSize)
		radicand = radicandFrag.IntoFrame()
		radicand.SetBaseline(radicand.Height()/2.0 + axis)
	} else {
		radicand = radicandFrag.IntoFrame()
	}

	// Layout sqrt symbol to determine target height
	sqrtFrag, err := ctx.LayoutIntoFragment(item.Sqrt, styles)
	if err != nil {
		return err
	}
	sqrtFont, sqrtFontSize := FragmentFont(sqrtFrag, ctx, styles)
	constants := sqrtFont.Math()
	thickness := constants.RadicalRuleThickness.At(sqrtFontSize)

	mathSize := props.Size
	var gap Abs
	if mathSize == MathSizeDisplay {
		gap = constants.RadicalDisplayStyleVerticalGap.At(sqrtFontSize)
	} else {
		gap = constants.RadicalVerticalGap.At(sqrtFontSize)
	}
	_ = radicand.Height() + thickness + gap // Target for sqrt stretching

	// TODO: Set stretch relative to target on sqrt

	// Re-layout sqrt at target size
	sqrtFrag2, err := ctx.LayoutIntoFragment(item.Sqrt, styles)
	if err != nil {
		return err
	}
	sqrtFont2, sqrtFontSize2 := FragmentFont(sqrtFrag2, ctx, styles)
	constants2 := sqrtFont2.Math()

	thickness2 := constants2.RadicalRuleThickness.At(sqrtFontSize2)
	extraAscender := constants2.RadicalExtraAscender.At(sqrtFontSize2)
	kernBefore := constants2.RadicalKernBeforeDegree.At(sqrtFontSize2)
	kernAfter := constants2.RadicalKernAfterDegree.At(sqrtFontSize2)
	raiseFactor := constants2.RadicalDegreeBottomRaisePercent

	mathSize2 := props.Size
	if mathSize2 == MathSizeDisplay {
		gap = constants2.RadicalDisplayStyleVerticalGap.At(sqrtFontSize2)
	} else {
		gap = constants2.RadicalVerticalGap.At(sqrtFontSize2)
	}

	sqrt := sqrtFrag2.IntoFrame()

	// Layout optional index
	var index *Frame
	if item.Index != nil {
		indexFrag, err := ctx.LayoutIntoFragment(item.Index, styles)
		if err != nil {
			return err
		}
		index = indexFrag.IntoFrame()
	}

	// TeXbook page 443, item 11: distribute remaining space
	if freeSpace := sqrt.Height() - thickness2 - radicand.Height(); freeSpace > gap {
		gap = (gap + freeSpace) / 2.0
	}

	sqrtAscent := radicand.Ascent() + gap + thickness2
	descent := sqrt.Height() - sqrtAscent
	innerAscent := sqrtAscent + extraAscender

	sqrtOffset := Abs(0)
	shiftUp := Abs(0)
	ascent := innerAscent

	if index != nil {
		sqrtOffset = kernBefore + index.Width() + kernAfter
		// TeXbook formula for raising the index
		shiftUp = Abs(raiseFactor*float64(innerAscent-descent)) + index.Descent()
		if v := shiftUp + index.Ascent(); v > ascent {
			ascent = v
		}
	}

	sqrtX := sqrtOffset.Max(0)
	radicandX := sqrtX + sqrt.Width()
	radicandY := ascent - radicand.Ascent()
	lineWidth := radicand.Width()
	size := Size{X: radicandX + lineWidth, Y: ascent + descent}

	sqrtPos := Point{X: sqrtX, Y: radicandY - gap - thickness2}
	linePos := Point{X: radicandX, Y: radicandY - gap - thickness2/2.0}
	radicandPos := Point{X: radicandX, Y: radicandY}

	frame := NewSoftFrame(size)
	frame.SetBaseline(ascent)

	if index != nil {
		indexX := (-sqrtOffset).Max(0) + kernBefore
		indexPos := Point{X: indexX, Y: ascent - index.Ascent() - shiftUp}
		frame.PushFrame(indexPos, index)
	}

	frame.PushFrame(sqrtPos, sqrt)

	// Draw horizontal line of root symbol
	// TODO: Get text fill from styles
	line := Stroked(
		&LineGeometry{Delta: PointWithX(lineWidth)},
		StrokeFromPair(nil, thickness2),
	)
	frame.PushShape(linePos, line, props.Span)

	frame.PushFrame(radicandPos, radicand)

	ctx.Push(NewFrameFragment(props, sqrtFontSize2, frame))
	return nil
}
