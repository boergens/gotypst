package math

// TightLeading is the leading between rows in script and scriptscript size.
const TightLeading Em = 0.25

// Rows splits fragments by linebreaks and returns rows.
func Rows(fragments []MathFragment) [][]MathFragment {
	var rows [][]MathFragment
	var current []MathFragment

	for _, f := range fragments {
		if _, ok := f.(*LinebreakFragment); ok {
			rows = append(rows, current)
			current = nil
		} else {
			current = append(current, f)
		}
	}
	rows = append(rows, current)
	return rows
}

// FragmentsAscent calculates the maximum ascent of fragments.
func FragmentsAscent(fragments []MathFragment) Abs {
	var maxAscent Abs
	for _, f := range fragments {
		if affectsRowHeight(f) {
			if a := f.Ascent(); a > maxAscent {
				maxAscent = a
			}
		}
	}
	return maxAscent
}

// FragmentsDescent calculates the maximum descent of fragments.
func FragmentsDescent(fragments []MathFragment) Abs {
	var maxDescent Abs
	for _, f := range fragments {
		if affectsRowHeight(f) {
			if d := f.Descent(); d > maxDescent {
				maxDescent = d
			}
		}
	}
	return maxDescent
}

// affectsRowHeight returns whether a fragment affects row height calculations.
func affectsRowHeight(f MathFragment) bool {
	switch f.(type) {
	case *AlignFragment, *LinebreakFragment, *TagFragment:
		return false
	default:
		return true
	}
}

// IsMultiline returns whether fragments contain linebreaks.
func IsMultiline(fragments []MathFragment) bool {
	for _, f := range fragments {
		if _, ok := f.(*LinebreakFragment); ok {
			return true
		}
	}
	return false
}

// MathRunFrameBuilder builds a multi-row math frame.
type MathRunFrameBuilder struct {
	Size   Size
	Frames []struct {
		Frame *Frame
		Pos   Point
	}
}

// Build constructs the final frame from the builder.
func (b *MathRunFrameBuilder) Build() *Frame {
	frame := NewSoftFrame(b.Size)
	for _, sub := range b.Frames {
		frame.PushFrame(sub.Pos, sub.Frame)
	}
	return frame
}

// MultilineFrameBuilder creates a builder for multi-row layout.
func MultilineFrameBuilder(fragments []MathFragment, styles *StyleChain) *MathRunFrameBuilder {
	rows := Rows(fragments)
	rowCount := len(rows)
	alignments := Alignments(rows)

	// Calculate leading
	var leading Abs
	mathSize := MathSizeText // TODO: Get from styles
	if mathSize >= MathSizeText {
		leading = 14.0 // TODO: Get ParElem::leading from styles
	} else {
		leading = TightLeading.Resolve(styles.ResolveTextSize())
	}

	// Get alignment
	align := FixedAlignStart // TODO: Get from styles

	var frames []struct {
		Frame *Frame
		Pos   Point
	}
	size := Size{}

	for i, row := range rows {
		// Skip empty last row
		if i == rowCount-1 && len(row) == 0 {
			continue
		}

		sub := IntoLineFrame(row, alignments.Points, LeftRightAlternatorRight)

		if i > 0 {
			size.Y += leading
		}

		pos := PointWithY(size.Y)
		if len(alignments.Points) == 0 {
			pos.X = align.Position(alignments.Width - sub.Width())
		}

		if sub.Width() > size.X {
			size.X = sub.Width()
		}
		size.Y += sub.Height()

		frames = append(frames, struct {
			Frame *Frame
			Pos   Point
		}{Frame: sub, Pos: pos})
	}

	return &MathRunFrameBuilder{Size: size, Frames: frames}
}

// IntoLineFrame lays out fragments into a single-row frame.
func IntoLineFrame(fragments []MathFragment, points []Abs, alternator LeftRightAlternator) *Frame {
	ascent := FragmentsAscent(fragments)
	descent := FragmentsDescent(fragments)

	frame := NewSoftFrame(Size{X: 0, Y: ascent + descent})
	frame.SetBaseline(ascent)

	// Calculate chunk widths for alignment
	var widths []Abs
	if len(points) > 0 {
		var current Abs
		for _, f := range fragments {
			if _, ok := f.(*AlignFragment); ok {
				widths = append(widths, current)
				current = 0
			} else {
				current += f.Width()
			}
		}
		widths = append(widths, current)
	}

	// Calculate starting X positions
	nextX := func() func() *Abs {
		prevPoints := append([]Abs{0}, points...)
		pointWidths := make([]struct{ point, width Abs }, len(points))
		for i := 0; i < len(points) && i < len(widths); i++ {
			pointWidths[i] = struct{ point, width Abs }{points[i], widths[i]}
		}
		idx := 0
		prevIdx := 0
		alt := alternator

		return func() *Abs {
			if idx >= len(pointWidths) || prevIdx >= len(prevPoints) {
				return nil
			}

			pw := pointWidths[idx]
			prevPoint := prevPoints[prevIdx]
			idx++
			prevIdx++

			currentAlt := alt.Next()
			var result Abs
			if currentAlt == LeftRightAlternatorRight {
				result = pw.point - pw.width
			} else {
				result = prevPoint
			}
			return &result
		}
	}()

	xPtr := nextX()
	var x Abs
	if xPtr != nil {
		x = *xPtr
	}

	for _, f := range fragments {
		if _, ok := f.(*AlignFragment); ok {
			if xPtr := nextX(); xPtr != nil {
				x = *xPtr
			}
			continue
		}

		y := ascent - f.Ascent()
		frame.PushFrame(Point{X: x, Y: y}, f.IntoFrame())
		x += f.Width()
	}

	frame.SizeMut().X = x
	return frame
}

// IntoParItems converts fragments into inline items for paragraph layout.
func IntoParItems(fragments []MathFragment) []InlineItem {
	var items []InlineItem

	var x Abs
	var ascent, descent Abs
	frame := NewSoftFrame(Size{})
	empty := true

	finalizeFrame := func() {
		frame.SetSize(Size{X: x, Y: ascent + descent})
		frame.SetBaseline(0)
		frame.Translate(PointWithY(ascent))
	}

	spaceIsVisible := false

	isSpace := func(f MathFragment) bool {
		_, ok := f.(*SpaceFragment)
		return ok
	}

	isLineBreakOpportunity := func(class Class, nextClass *Class) bool {
		switch class {
		case Binary:
			return nextClass == nil || *nextClass != Closing
		case Relation:
			if nextClass == nil {
				return true
			}
			return *nextClass != Relation && *nextClass != Closing
		default:
			return false
		}
	}

	for i, f := range fragments {
		if spaceIsVisible && isSpace(f) {
			items = append(items, &InlineSpace{Width: f.Width(), Flexible: true})
			continue
		}

		class := f.Class()
		y := f.Ascent()

		if y > ascent {
			ascent = y
		}
		if d := f.Descent(); d > descent {
			descent = d
		}

		frame.PushFrame(Point{X: x, Y: -y}, f.IntoFrame())
		x += f.Width()
		empty = false

		// Check for line break opportunity
		var nextClass *Class
		if i+1 < len(fragments) {
			c := fragments[i+1].Class()
			nextClass = &c
		}

		if isLineBreakOpportunity(class, nextClass) {
			finalizeFrame()
			items = append(items, &InlineFrame{Frame: frame})
			empty = true

			frame = NewSoftFrame(Size{})
			x = 0
			ascent = 0
			descent = 0

			spaceIsVisible = true
			if i+1 < len(fragments) && !isSpace(fragments[i+1]) {
				items = append(items, &InlineSpace{Width: 0, Flexible: true})
			}
		} else {
			spaceIsVisible = false
		}
	}

	if !empty {
		finalizeFrame()
		items = append(items, &InlineFrame{Frame: frame})
	}

	return items
}

// AlignmentResult holds alignment calculation results.
type AlignmentResult struct {
	Points []Abs
	Width  Abs
}

// Alignments calculates alignment points for rows.
func Alignments(rows [][]MathFragment) AlignmentResult {
	var widths []Abs
	var pendingWidth Abs

	for _, row := range rows {
		var width Abs
		alignmentIndex := 0

		for _, f := range row {
			if _, ok := f.(*AlignFragment); ok {
				if alignmentIndex < len(widths) {
					if width > widths[alignmentIndex] {
						widths[alignmentIndex] = width
					}
				} else {
					w := width
					if pendingWidth > w {
						w = pendingWidth
					}
					widths = append(widths, w)
				}
				width = 0
				alignmentIndex++
			} else {
				width += f.Width()
			}
		}

		if len(widths) == 0 {
			if width > pendingWidth {
				pendingWidth = width
			}
		} else if alignmentIndex < len(widths) {
			if width > widths[alignmentIndex] {
				widths[alignmentIndex] = width
			}
		} else {
			w := width
			if pendingWidth > w {
				w = pendingWidth
			}
			widths = append(widths, w)
		}
	}

	// Convert widths to cumulative points
	points := make([]Abs, len(widths))
	copy(points, widths)
	for i := 1; i < len(points); i++ {
		points[i] += points[i-1]
	}

	var totalWidth Abs
	if len(points) > 0 {
		totalWidth = points[len(points)-1]
	} else {
		totalWidth = pendingWidth
	}

	return AlignmentResult{
		Points: points,
		Width:  totalWidth,
	}
}
