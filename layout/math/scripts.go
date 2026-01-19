package math

// ScriptsItem represents scripts (sub/superscripts) for layout.
type ScriptsItem struct {
	Base        *MathItem
	TopLeft     *MathItem // Pre-superscript
	Top         *MathItem // Upper limit
	TopRight    *MathItem // Post-superscript
	BottomLeft  *MathItem // Pre-subscript
	Bottom      *MathItem // Lower limit
	BottomRight *MathItem // Post-subscript
}

// PrimesItem represents prime marks for layout.
type PrimesItem struct {
	Prime *MathItem
	Count int
}

// layoutScriptsImpl lays out scripts around a base.
func layoutScriptsImpl(
	item *ScriptsItem,
	ctx *MathContext,
	styles *StyleChain,
	props *MathProperties,
) error {
	// Helper to layout optional items
	layout := func(m *MathItem) (MathFragment, error) {
		if m == nil {
			return nil, nil
		}
		return ctx.LayoutIntoFragment(m, styles)
	}

	// Layout top and bottom early for measuring
	t, err := layout(item.Top)
	if err != nil {
		return err
	}
	b, err := layout(item.Bottom)
	if err != nil {
		return err
	}

	// TODO: Calculate relative_to_width for stretching

	// Layout base
	baseFrag, err := ctx.LayoutIntoFragment(item.Base, styles)
	if err != nil {
		return err
	}

	// Layout all scripts
	tl, err := layout(item.TopLeft)
	if err != nil {
		return err
	}
	tr, err := layout(item.TopRight)
	if err != nil {
		return err
	}
	bl, err := layout(item.BottomLeft)
	if err != nil {
		return err
	}
	br, err := layout(item.BottomRight)
	if err != nil {
		return err
	}

	return layoutAttachments(ctx, props, styles, baseFrag, [6]MathFragment{tl, t, tr, bl, b, br})
}

// layoutPrimesImpl lays out prime marks.
func layoutPrimesImpl(
	item *PrimesItem,
	ctx *MathContext,
	styles *StyleChain,
	props *MathProperties,
) error {
	primeFrag, err := ctx.LayoutIntoFragment(item.Prime, styles)
	if err != nil {
		return err
	}
	prime := primeFrag.IntoFrame()

	// Calculate width: primes overlap by half
	width := prime.Width() * Abs(float64(item.Count+1)/2.0)
	frame := NewSoftFrame(Size{X: width, Y: prime.Height()})
	frame.SetBaseline(prime.Ascent())

	// Position each prime with overlap
	for i := 0; i < item.Count; i++ {
		pos := PointWithX(prime.Width() * Abs(float64(i)/2.0))
		frame.PushFrame(pos, prime)
	}

	ff := NewFrameFragment(props, styles.ResolveTextSize(), frame)
	ff.TextLike = true
	ctx.Push(ff)
	return nil
}

// layoutAttachments lays out all attachments around a base.
func layoutAttachments(
	ctx *MathContext,
	props *MathProperties,
	styles *StyleChain,
	base MathFragment,
	attachments [6]MathFragment, // tl, t, tr, bl, b, br
) error {
	tl, t, tr, bl, b, br := attachments[0], attachments[1], attachments[2], attachments[3], attachments[4], attachments[5]

	font, fontSize := FragmentFont(base, ctx, styles)
	constants := font.Math()
	cramped := false // TODO: Get from styles

	// Calculate script shifts
	txShift, bxShift := Abs(0), Abs(0)
	if tl != nil || tr != nil || bl != nil || br != nil {
		txShift, bxShift = computeScriptShifts(font, fontSize, cramped, base, [4]MathFragment{tl, tr, bl, br})
	}

	// Calculate limit shifts
	tShift, bShift := computeLimitShifts(font, fontSize, base, [2]MathFragment{t, b})

	// Helper to get fragment dimension or zero
	measure := func(f MathFragment, fn func(MathFragment) Abs) Abs {
		if f == nil {
			return 0
		}
		return fn(f)
	}

	// Calculate final frame height
	ascent := base.Ascent()
	if v := txShift + measure(tr, MathFragment.Ascent); v > ascent {
		ascent = v
	}
	if v := txShift + measure(tl, MathFragment.Ascent); v > ascent {
		ascent = v
	}
	if v := tShift + measure(t, MathFragment.Ascent); v > ascent {
		ascent = v
	}

	descent := base.Descent()
	if v := bxShift + measure(br, MathFragment.Descent); v > descent {
		descent = v
	}
	if v := bxShift + measure(bl, MathFragment.Descent); v > descent {
		descent = v
	}
	if v := bShift + measure(b, MathFragment.Descent); v > descent {
		descent = v
	}

	height := ascent + descent

	// Calculate Y positions
	baseY := ascent - base.Ascent()
	txY := func(f MathFragment) Abs { return ascent - txShift - f.Ascent() }
	bxY := func(f MathFragment) Abs { return ascent + bxShift - f.Ascent() }
	tY := func(f MathFragment) Abs { return ascent - tShift - f.Ascent() }
	bY := func(f MathFragment) Abs { return ascent + bShift - f.Ascent() }

	// Calculate limit widths
	tPreWidth, tPostWidth := computeLimitWidth(base, t)
	bPreWidth, bPostWidth := computeLimitWidth(base, b)

	// Space after script
	spaceAfterScript := constants.SpaceAfterScript.At(fontSize)

	// Calculate pre-script widths
	tlPreWidth, blPreWidth := computePreScriptWidths(base, [2]MathFragment{tl, bl}, txShift, bxShift, spaceAfterScript)

	// Calculate post-script widths
	trPostWidth, trKern := computePostScriptWidth(base, tr, txShift, spaceAfterScript)
	brPostWidth, brKern := computePostScriptWidth(base, br, bxShift, spaceAfterScript)
	// Adjust subscript for italics correction
	if br != nil {
		brKern -= base.ItalicsCorrection()
	}

	// Calculate final frame width
	preWidth := tPreWidth.Max(bPreWidth).Max(tlPreWidth).Max(blPreWidth)
	baseWidth := base.Width()
	postWidth := tPostWidth.Max(bPostWidth).Max(trPostWidth).Max(brPostWidth)
	width := preWidth + baseWidth + postWidth

	// Calculate X positions
	baseX := preWidth
	tlX := preWidth - tlPreWidth + spaceAfterScript
	blX := preWidth - blPreWidth + spaceAfterScript
	trX := preWidth + baseWidth + trKern
	brX := preWidth + baseWidth + brKern
	tX := preWidth - tPreWidth
	bX := preWidth - bPreWidth

	// Create final frame
	frame := NewSoftFrame(Size{X: width, Y: height})
	frame.SetBaseline(ascent)
	frame.PushFrame(Point{X: baseX, Y: baseY}, base.IntoFrame())

	// Push scripts
	if tl != nil {
		frame.PushFrame(Point{X: tlX, Y: txY(tl)}, tl.IntoFrame())
	}
	if bl != nil {
		frame.PushFrame(Point{X: blX, Y: bxY(bl)}, bl.IntoFrame())
	}
	if tr != nil {
		frame.PushFrame(Point{X: trX, Y: txY(tr)}, tr.IntoFrame())
	}
	if br != nil {
		frame.PushFrame(Point{X: brX, Y: bxY(br)}, br.IntoFrame())
	}
	if t != nil {
		frame.PushFrame(Point{X: tX, Y: tY(t)}, t.IntoFrame())
	}
	if b != nil {
		frame.PushFrame(Point{X: bX, Y: bY(b)}, b.IntoFrame())
	}

	ctx.Push(NewFrameFragment(props, fontSize, frame))
	return nil
}

// computeScriptShifts calculates baseline shifts for scripts.
func computeScriptShifts(
	font *Font,
	fontSize Abs,
	cramped bool,
	base MathFragment,
	scripts [4]MathFragment, // tl, tr, bl, br
) (txShift, bxShift Abs) {
	constants := font.Math()

	supShiftUp := constants.SuperscriptShiftUp
	if cramped {
		supShiftUp = constants.SuperscriptShiftUpCramped
	}

	supBottomMin := constants.SuperscriptBottomMin.At(fontSize)
	supBottomMaxWithSub := constants.SuperscriptBottomMaxWithSubscript.At(fontSize)
	supDropMax := constants.SuperscriptBaselineDropMax.At(fontSize)
	gapMin := constants.SubSuperscriptGapMin.At(fontSize)
	subShiftDown := constants.SubscriptShiftDown.At(fontSize)
	subTopMax := constants.SubscriptTopMax.At(fontSize)
	subDropMin := constants.SubscriptBaselineDropMin.At(fontSize)

	tl, tr, bl, br := scripts[0], scripts[1], scripts[2], scripts[3]
	isTextLike := base.IsTextLike()

	// Calculate superscript shift
	if tl != nil || tr != nil {
		baseAscent := BaseAscent(base)
		txShift = supShiftUp.At(fontSize)
		if !isTextLike {
			if v := baseAscent - supDropMax; v > txShift {
				txShift = v
			}
		}
		if tl != nil {
			if v := supBottomMin + tl.Descent(); v > txShift {
				txShift = v
			}
		}
		if tr != nil {
			if v := supBottomMin + tr.Descent(); v > txShift {
				txShift = v
			}
		}
	}

	// Calculate subscript shift
	if bl != nil || br != nil {
		baseDescent := BaseDescent(base)
		bxShift = subShiftDown
		if !isTextLike {
			if v := baseDescent + subDropMin; v > bxShift {
				bxShift = v
			}
		}
		if bl != nil {
			if v := bl.Ascent() - subTopMax; v > bxShift {
				bxShift = v
			}
		}
		if br != nil {
			if v := br.Ascent() - subTopMax; v > bxShift {
				bxShift = v
			}
		}
	}

	// Adjust for sub-superscript gap
	pairs := [][2]MathFragment{{tl, bl}, {tr, br}}
	for _, pair := range pairs {
		sup, sub := pair[0], pair[1]
		if sup != nil && sub != nil {
			supBottom := txShift - sup.Descent()
			subTop := sub.Ascent() - bxShift
			gap := supBottom - subTop
			if gap < gapMin {
				increase := gapMin - gap
				supOnly := (supBottomMaxWithSub - supBottom).Clamp(0, increase)
				rest := (increase - supOnly) / 2.0
				txShift += supOnly + rest
				bxShift += rest
			}
		}
	}

	return txShift, bxShift
}

// computeLimitShifts calculates baseline shifts for limits.
func computeLimitShifts(
	font *Font,
	fontSize Abs,
	base MathFragment,
	limits [2]MathFragment, // t, b
) (tShift, bShift Abs) {
	constants := font.Math()
	t, b := limits[0], limits[1]

	if t != nil {
		upperGapMin := constants.UpperLimitGapMin.At(fontSize)
		upperRiseMin := constants.UpperLimitBaselineRiseMin.At(fontSize)
		tShift = base.Ascent() + upperRiseMin.Max(upperGapMin+t.Descent())
	}

	if b != nil {
		lowerGapMin := constants.LowerLimitGapMin.At(fontSize)
		lowerDropMin := constants.LowerLimitBaselineDropMin.At(fontSize)
		bShift = base.Descent() + lowerDropMin.Max(lowerGapMin+b.Ascent())
	}

	return tShift, bShift
}

// computeLimitWidth calculates how far a limit extends beyond base width.
func computeLimitWidth(base MathFragment, limit MathFragment) (preWidth, postWidth Abs) {
	if limit == nil {
		return 0, 0
	}

	delta := base.ItalicsCorrection() / 2.0
	half := (limit.Width() - base.Width()) / 2.0

	// Upper limit shifts right, lower limit shifts left
	return half - delta, half + delta
}

// computePreScriptWidths calculates pre-script widths.
func computePreScriptWidths(
	base MathFragment,
	scripts [2]MathFragment, // tl, bl
	tlShift, blShift Abs,
	spaceBeforePreScript Abs,
) (tlPreWidth, blPreWidth Abs) {
	tl, bl := scripts[0], scripts[1]

	if tl != nil {
		kern := mathKern(base, tl, tlShift, CornerTopLeft)
		tlPreWidth = spaceBeforePreScript + tl.Width() + kern
	}

	if bl != nil {
		kern := mathKern(base, bl, blShift, CornerBottomLeft)
		blPreWidth = spaceBeforePreScript + bl.Width() + kern
	}

	return tlPreWidth, blPreWidth
}

// computePostScriptWidth calculates post-script width and kerning.
func computePostScriptWidth(
	base MathFragment,
	script MathFragment,
	shift Abs,
	spaceAfterScript Abs,
) (postWidth, kern Abs) {
	if script == nil {
		return 0, 0
	}

	kern = mathKern(base, script, shift, CornerTopRight)
	postWidth = spaceAfterScript + script.Width() + kern
	return postWidth, kern
}

// mathKern calculates the kerning value for a script relative to base.
func mathKern(base, script MathFragment, shift Abs, pos Corner) Abs {
	var corrHeightTop, corrHeightBot Abs

	switch pos {
	case CornerTopLeft, CornerTopRight:
		// Superscript corrections
		corrHeightTop = base.Ascent() - shift
		corrHeightBot = shift - script.Descent()
	case CornerBottomLeft, CornerBottomRight:
		// Subscript corrections
		corrHeightTop = script.Ascent() - shift
		corrHeightBot = shift - base.Descent()
	}

	// Calculate summed kern for each correction height
	summedKern := func(height Abs) Abs {
		baseKern := KernAtHeight(base, pos, height)
		attachKern := KernAtHeight(script, pos.Inv(), height)
		return baseKern + attachKern
	}

	// Take the smaller kern amount (larger value)
	k1 := summedKern(corrHeightTop)
	k2 := summedKern(corrHeightBot)
	if k1 > k2 {
		return k1
	}
	return k2
}
