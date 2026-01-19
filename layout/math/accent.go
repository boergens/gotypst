package math

// AccentItem represents an accent for layout.
type AccentItem struct {
	Base            *MathItem
	Accent          *MathItem
	Position        Position
	ExactFrameWidth bool
}

// layoutAccentImpl lays out an accent above or below a base.
func layoutAccentImpl(
	item *AccentItem,
	ctx *MathContext,
	styles *StyleChain,
	props *MathProperties,
) error {
	topAccent := item.Position == PositionAbove

	// Layout base
	baseFrag, err := ctx.LayoutIntoFragment(item.Base, styles)
	if err != nil {
		return err
	}

	font, fontSize := FragmentFont(baseFrag, ctx, styles)
	baseAttachTop, baseAttachBottom := baseFrag.AccentAttach()

	// Check for flattened accent variant
	flattenedBaseHeight := font.Math().FlattenedAccentBaseHeight.At(fontSize)
	if topAccent && baseFrag.Ascent() > flattenedBaseHeight {
		// TODO: Set flac feature on accent
	}

	// TODO: Set stretch relative to base width

	// Layout accent
	accentFrag, err := ctx.LayoutIntoFragment(item.Accent, styles)
	if err != nil {
		return err
	}
	accentAttachTop, _ := accentFrag.AccentAttach()
	accent := accentFrag.IntoFrame()

	// Calculate width and positions
	var width, baseX, accentX Abs
	if topAccent {
		baseAttachPos := baseAttachTop
		if !item.ExactFrameWidth {
			width = baseFrag.Width()
			baseX = 0
			accentX = baseAttachPos - accentAttachTop
		} else {
			preWidth := accentAttachTop - baseAttachPos
			postWidth := (accent.Width() - accentAttachTop) - (baseFrag.Width() - baseAttachPos)
			width = preWidth.Max(0) + baseFrag.Width() + postWidth.Max(0)
			if preWidth < 0 {
				baseX = 0
				accentX = -preWidth
			} else {
				baseX = preWidth
				accentX = 0
			}
		}
	} else {
		baseAttachPos := baseAttachBottom
		if !item.ExactFrameWidth {
			width = baseFrag.Width()
			baseX = 0
			accentX = baseAttachPos - accentAttachTop
		} else {
			preWidth := accentAttachTop - baseAttachPos
			postWidth := (accent.Width() - accentAttachTop) - (baseFrag.Width() - baseAttachPos)
			width = preWidth.Max(0) + baseFrag.Width() + postWidth.Max(0)
			if preWidth < 0 {
				baseX = 0
				accentX = -preWidth
			} else {
				baseX = preWidth
				accentX = 0
			}
		}
	}

	var gap, baseline Abs
	var accentPos, basePos Point

	if topAccent {
		// Top accent positioning
		accentBaseHeight := font.Math().AccentBaseHeight.At(fontSize)
		gap = -accent.Descent() - baseFrag.Ascent().Min(accentBaseHeight)
		accentPos = PointWithX(accentX)
		basePos = Point{X: baseX, Y: accent.Height() + gap}
		baseline = basePos.Y + baseFrag.Ascent()
	} else {
		// Bottom accent positioning
		gap = -accent.Ascent()
		accentPos = Point{X: accentX, Y: baseFrag.Height() + gap}
		basePos = PointWithX(baseX)
		baseline = baseFrag.Ascent()
	}

	size := Size{X: width, Y: accent.Height() + gap + baseFrag.Height()}

	baseTextLike := !item.ExactFrameWidth && baseFrag.IsTextLike()
	baseItalicsCorrection := baseFrag.ItalicsCorrection()
	baseAscent := BaseAscent(baseFrag)
	baseDescent := BaseDescent(baseFrag)

	frame := NewSoftFrame(size)
	frame.SetBaseline(baseline)
	frame.PushFrame(accentPos, accent)
	frame.PushFrame(basePos, baseFrag.IntoFrame())

	ff := NewFrameFragment(props, fontSize, frame)
	ff.BaseAscent = baseAscent
	ff.BaseDescent = baseDescent
	ff.ItalicsCorr = baseItalicsCorrection
	ff.TextLike = baseTextLike
	ff.AccentAttachTop, ff.AccentAttachBottom = baseAttachTop, baseAttachBottom

	ctx.Push(ff)
	return nil
}
