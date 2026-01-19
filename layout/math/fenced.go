package math

// FencedItem represents fenced content (parentheses, brackets).
type FencedItem struct {
	Open     *MathItem // Opening delimiter (optional)
	Body     *MathItem
	Close    *MathItem // Closing delimiter (optional)
	Balanced bool      // Whether to balance delimiters on axis
}

// layoutFencedImpl lays out fenced content with stretching delimiters.
func layoutFencedImpl(
	item *FencedItem,
	ctx *MathContext,
	styles *StyleChain,
	props *MathProperties,
) error {
	// Layout body first to compute delimiter sizing
	bodyFragments, err := ctx.LayoutIntoFragments(item.Body, styles)
	if err != nil {
		return err
	}

	// Calculate relative_to for delimiter sizing
	var relativeTo Abs
	if item.Balanced {
		// Balance on axis: use maximum extent from axis
		var maxExtent Abs
		for _, f := range bodyFragments {
			font, fontSize := FragmentFont(f, ctx, styles)
			axis := font.Math().AxisHeight.At(fontSize)
			extent := (f.Ascent() - axis).Max(f.Descent() + axis)
			if extent > maxExtent {
				maxExtent = extent
			}
		}
		relativeTo = 2.0 * maxExtent
	} else {
		// Use maximum height
		for _, f := range bodyFragments {
			if h := f.Height(); h > relativeTo {
				relativeTo = h
			}
		}
	}

	// Check for mid-stretched items and set stretch info
	hasMidStretched := false
	// TODO: Check body items for mid-stretched flag and set stretch info
	_ = hasMidStretched

	// Layout opening delimiter
	if item.Open != nil {
		// TODO: Set stretch relative to on open delimiter
		openFrag, err := ctx.LayoutIntoFragment(item.Open, styles)
		if err != nil {
			return err
		}
		ctx.Push(openFrag)
	}

	// Re-layout body if mid-stretched items present
	// For now, just extend the fragments
	ctx.Extend(bodyFragments)

	// Layout closing delimiter
	if item.Close != nil {
		// TODO: Set stretch relative to on close delimiter
		closeFrag, err := ctx.LayoutIntoFragment(item.Close, styles)
		if err != nil {
			return err
		}
		ctx.Push(closeFrag)
	}

	return nil
}
