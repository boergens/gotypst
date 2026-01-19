package math

// FractionItem represents a fraction for layout.
type FractionItem struct {
	Numerator   *MathItem
	Denominator *MathItem
	Line        bool    // Whether to draw the fraction line
	Padding     Em      // Horizontal padding
}

// SkewedFractionItem represents a skewed (diagonal) fraction.
type SkewedFractionItem struct {
	Numerator   *MathItem
	Denominator *MathItem
	Slash       *MathItem
}

// layoutFractionImpl lays out a standard fraction.
func layoutFractionImpl(
	item *FractionItem,
	ctx *MathContext,
	styles *StyleChain,
	props *MathProperties,
) error {
	// Layout numerator and denominator
	numFrag, err := ctx.LayoutIntoFragment(item.Numerator, styles)
	if err != nil {
		return err
	}
	num := numFrag.IntoFrame()

	denomFrag, err := ctx.LayoutIntoFragment(item.Denominator, styles)
	if err != nil {
		return err
	}
	denom := denomFrag.IntoFrame()

	constants := ctx.Font().Math()
	fontSize := styles.ResolveTextSize()
	mathSize := props.Size

	var frame *Frame

	if item.Line {
		// Fraction with line
		axis := constants.AxisHeight.At(fontSize)
		thickness := constants.FractionRuleThickness.At(fontSize)

		var shiftUp, shiftDown, numMin, denomMin Abs
		if mathSize == MathSizeDisplay {
			shiftUp = constants.FractionNumeratorDisplayStyleShiftUp.At(fontSize)
			shiftDown = constants.FractionDenominatorDisplayStyleShiftDown.At(fontSize)
			numMin = constants.FractionNumDisplayStyleGapMin.At(fontSize)
			denomMin = constants.FractionDenomDisplayStyleGapMin.At(fontSize)
		} else {
			shiftUp = constants.FractionNumeratorShiftUp.At(fontSize)
			shiftDown = constants.FractionDenominatorShiftDown.At(fontSize)
			numMin = constants.FractionNumeratorGapMin.At(fontSize)
			denomMin = constants.FractionDenominatorGapMin.At(fontSize)
		}

		numGap := (shiftUp - (axis + thickness/2.0) - num.Descent()).Max(numMin)
		denomGap := (shiftDown + (axis - thickness/2.0) - denom.Ascent()).Max(denomMin)

		lineWidth := num.Width().Max(denom.Width())
		width := lineWidth + 2.0*item.Padding.At(fontSize)
		height := num.Height() + numGap + thickness + denomGap + denom.Height()
		size := Size{X: width, Y: height}

		numPos := PointWithX((width - num.Width()) / 2.0)
		linePos := Point{
			X: (width - lineWidth) / 2.0,
			Y: num.Height() + numGap + thickness/2.0,
		}
		denomPos := Point{
			X: (width - denom.Width()) / 2.0,
			Y: height - denom.Height(),
		}
		baseline := linePos.Y + axis

		frame = NewSoftFrame(size)
		frame.SetBaseline(baseline)
		frame.PushFrame(numPos, num)
		frame.PushFrame(denomPos, denom)

		// Draw the fraction line
		// TODO: Get text fill from styles
		line := Stroked(
			&LineGeometry{Delta: PointWithX(lineWidth)},
			StrokeFromPair(nil, thickness),
		)
		frame.PushShape(linePos, line, props.Span)
	} else {
		// Fraction without line (stack)
		var shiftUp, shiftDown, gapMin Abs
		if mathSize == MathSizeDisplay {
			shiftUp = constants.StackTopDisplayStyleShiftUp.At(fontSize)
			shiftDown = constants.StackBottomDisplayStyleShiftDown.At(fontSize)
			gapMin = constants.StackDisplayStyleGapMin.At(fontSize)
		} else {
			shiftUp = constants.StackTopShiftUp.At(fontSize)
			shiftDown = constants.StackBottomShiftDown.At(fontSize)
			gapMin = constants.StackGapMin.At(fontSize)
		}

		gap := (shiftUp - num.Descent()) + (shiftDown - denom.Ascent())

		width := num.Width().Max(denom.Width()) + 2.0*item.Padding.At(fontSize)
		actualGap := gap
		if gapMin > gap {
			actualGap = gapMin
		}
		height := num.Height() + actualGap + denom.Height()
		size := Size{X: width, Y: height}

		numPos := PointWithX((width - num.Width()) / 2.0)
		denomPos := Point{
			X: (width - denom.Width()) / 2.0,
			Y: height - denom.Height(),
		}

		baseline := num.Ascent() + shiftUp
		if gapMin > gap {
			baseline += (gapMin - gap) / 2.0
		}

		frame = NewSoftFrame(size)
		frame.SetBaseline(baseline)
		frame.PushFrame(numPos, num)
		frame.PushFrame(denomPos, denom)
	}

	ctx.Push(NewFrameFragment(props, fontSize, frame))
	return nil
}

// layoutSkewedFractionImpl lays out a skewed fraction.
func layoutSkewedFractionImpl(
	item *SkewedFractionItem,
	ctx *MathContext,
	styles *StyleChain,
	props *MathProperties,
) error {
	constants := ctx.Font().Math()
	fontSize := styles.ResolveTextSize()
	vgap := constants.SkewedFractionVerticalGap.At(fontSize)
	hgap := constants.SkewedFractionHorizontalGap.At(fontSize)
	axis := constants.AxisHeight.At(fontSize)

	// Layout numerator
	numFrag, err := ctx.LayoutIntoFragment(item.Numerator, styles)
	if err != nil {
		return err
	}
	numFrame := numFrag.IntoFrame()
	numSize := numFrame.Size()

	// Layout denominator
	denomFrag, err := ctx.LayoutIntoFragment(item.Denominator, styles)
	if err != nil {
		return err
	}
	denomFrame := denomFrag.IntoFrame()
	denomSize := denomFrame.Size()

	// Initial height calculation
	fractionHeight := numSize.Y + denomSize.Y + vgap

	// Layout slash glyph
	// TODO: Set stretch relative to height
	slashFrag, err := ctx.LayoutIntoFragment(item.Slash, styles)
	if err != nil {
		return err
	}
	slashFrame := slashFrag.IntoFrame()
	slashSize := slashFrame.Size()

	// Adjust height if slash overflows
	verticalOffset := Abs(0)
	if overflow := slashSize.Y - fractionHeight; overflow > 0 {
		verticalOffset = overflow / 2.0
	}
	if slashSize.Y > fractionHeight {
		fractionHeight = slashSize.Y
	}

	// Calculate reference points
	slashUpLeft := Point{
		X: numSize.X + hgap/2.0 - slashSize.X/2.0,
		Y: fractionHeight/2.0 - slashSize.Y/2.0,
	}
	numUpLeft := PointWithY(verticalOffset)
	denomUpLeft := Point{
		X: numUpLeft.X + numSize.X + hgap,
		Y: numUpLeft.Y + numSize.Y + vgap,
	}

	// Calculate width
	fractionWidth := (denomUpLeft.X + denomSize.X).Max(slashUpLeft.X + slashSize.X)
	if slashUpLeft.X < 0 {
		fractionWidth += -slashUpLeft.X
	}

	// Adjust for negative slash position
	horizontalOffset := PointWithX(Abs(0).Max(-slashUpLeft.X))
	slashUpLeft.X += horizontalOffset.X
	slashUpLeft.Y += horizontalOffset.Y
	numUpLeft.X += horizontalOffset.X
	numUpLeft.Y += horizontalOffset.Y
	denomUpLeft.X += horizontalOffset.X
	denomUpLeft.Y += horizontalOffset.Y

	// Build final frame
	fractionFrame := NewSoftFrame(Size{X: fractionWidth, Y: fractionHeight})
	fractionFrame.SetBaseline(fractionHeight/2.0 + axis)

	fractionFrame.PushFrame(numUpLeft, numFrame)
	fractionFrame.PushFrame(denomUpLeft, denomFrame)
	fractionFrame.PushFrame(slashUpLeft, slashFrame)

	ctx.Push(NewFrameFragment(props, fontSize, fractionFrame))
	return nil
}
