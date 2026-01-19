package math

// DefaultStrokeThickness is the default stroke thickness for augmentation lines.
const DefaultStrokeThickness Em = 0.05

// TableItem represents a math table/matrix for layout.
type TableItem struct {
	Cells      [][]*MathItem
	Gap        Axes[Rel]
	Align      FixedAlignment
	Alternator LeftRightAlternator
	Augment    *AugmentSpec
}

// AugmentSpec specifies augmentation lines for matrices.
type AugmentSpec struct {
	HLine  AugmentOffsets
	VLine  AugmentOffsets
	Stroke *FixedStroke
}

// layoutTableImpl lays out a math table/matrix.
func layoutTableImpl(
	item *TableItem,
	ctx *MathContext,
	styles *StyleChain,
	props *MathProperties,
) error {
	rows := item.Cells
	nrows := len(rows)
	if nrows == 0 {
		ctx.Push(NewFrameFragment(props, styles.ResolveTextSize(), NewSoftFrame(Size{})))
		return nil
	}

	ncols := len(rows[0])
	if ncols == 0 {
		ctx.Push(NewFrameFragment(props, styles.ResolveTextSize(), NewSoftFrame(Size{})))
		return nil
	}

	// Calculate gap
	regionSize := ctx.Region.Size
	gap := Axes[Abs]{
		X: item.Gap.X.RelativeTo(regionSize.X),
		Y: item.Gap.Y.RelativeTo(regionSize.Y),
	}
	halfGap := Axes[Abs]{X: gap.X / 2.0, Y: gap.Y / 2.0}

	// Default stroke for augmentation
	fontSize := styles.ResolveTextSize()
	defaultStrokeThickness := DefaultStrokeThickness.Resolve(fontSize)
	defaultStroke := FixedStroke{
		Thickness: defaultStrokeThickness,
		Cap:       LineCapSquare,
	}

	// Get augmentation specs
	var hline, vline AugmentOffsets
	stroke := defaultStroke
	if item.Augment != nil {
		hline = item.Augment.HLine
		vline = item.Augment.VLine
		if item.Augment.Stroke != nil {
			stroke = *item.Augment.Stroke
		}
	}

	// Transpose rows to columns
	columns := make([][]*MathItem, ncols)
	for i := range columns {
		columns[i] = make([]*MathItem, nrows)
		for j := 0; j < nrows; j++ {
			if i < len(rows[j]) {
				columns[i][j] = rows[j][i]
			}
		}
	}

	// Layout cells and track heights
	cols := make([][][]MathFragment, ncols)
	heights := make([][2]Abs, nrows) // [ascent, descent] for each row

	// Get paren for padding reference
	// TODO: Create paren glyph for height padding

	for i, column := range columns {
		cols[i] = make([][]MathFragment, nrows)
		for j, cell := range column {
			if cell == nil {
				cols[i][j] = nil
				continue
			}

			cellFrags, err := ctx.LayoutIntoFragments(cell, styles)
			if err != nil {
				return err
			}

			cellAscent := FragmentsAscent(cellFrags)
			cellDescent := FragmentsDescent(cellFrags)

			if cellAscent > heights[j][0] {
				heights[j][0] = cellAscent
			}
			if cellDescent > heights[j][1] {
				heights[j][1] = cellDescent
			}

			cols[i][j] = cellFrags
		}
	}

	// Normalize negative line indices
	normalizeOffsets := func(offsets *AugmentOffsets, max int) {
		for i := range offsets.Offsets {
			if offsets.Offsets[i] < 0 {
				offsets.Offsets[i] += max
			}
		}
	}
	normalizeOffsets(&hline, nrows)
	normalizeOffsets(&vline, ncols)

	// Calculate total height
	totalHeight := Abs(0)
	for _, h := range heights {
		totalHeight += h[0] + h[1]
	}
	totalHeight += gap.Y * Abs(nrows-1)

	// Add space for edge hlines
	containsOffset := func(offsets AugmentOffsets, val int) bool {
		for _, o := range offsets.Offsets {
			if o == val {
				return true
			}
		}
		return false
	}

	if containsOffset(hline, 0) {
		totalHeight += gap.Y
	}
	if containsOffset(hline, nrows) {
		totalHeight += gap.Y
	}

	// Build frame
	frame := NewSoftFrame(Size{X: 0, Y: totalHeight})
	x := Abs(0)

	// Add leading vline
	if containsOffset(vline, 0) {
		lineFrame := lineItem(totalHeight, true, stroke, props.Span)
		frame.PushFrame(PointWithX(x+halfGap.X), lineFrame)
		x += gap.X
	}

	// Layout columns
	for colIdx, col := range cols {
		alignResult := Alignments(col)
		rcol := alignResult.Width

		y := Abs(0)
		if containsOffset(hline, 0) {
			y = gap.Y
		}

		for rowIdx, cellFrags := range col {
			ascent, descent := heights[rowIdx][0], heights[rowIdx][1]

			cellFrame := IntoLineFrame(cellFrags, alignResult.Points, item.Alternator)
			posX := x
			if len(alignResult.Points) == 0 {
				posX += item.Align.Position(rcol - cellFrame.Width())
			}
			posY := y + ascent - cellFrame.Ascent()

			frame.PushFrame(Point{X: posX, Y: posY}, cellFrame)
			y += ascent + descent + gap.Y
		}

		x += rcol

		// Add vline after column if needed
		if containsOffset(vline, colIdx+1) {
			lineFrame := lineItem(totalHeight, true, stroke, props.Span)
			frame.PushFrame(PointWithX(x+halfGap.X), lineFrame)
		}

		x += gap.X
	}

	// Adjust total width
	totalWidth := x - gap.X
	if containsOffset(vline, ncols) {
		totalWidth = x
	}

	// Add horizontal lines
	for _, lineIdx := range hline.Offsets {
		var offset Abs
		if lineIdx == 0 {
			offset = gap.Y
		} else {
			for i := 0; i < lineIdx; i++ {
				offset += heights[i][0] + heights[i][1]
			}
			offset += gap.Y * Abs(lineIdx-1) + halfGap.Y
		}

		lineFrame := lineItem(totalWidth, false, stroke, props.Span)
		frame.PushFrame(PointWithY(offset), lineFrame)
	}

	frame.SizeMut().X = totalWidth

	// Set baseline at center on axis
	axis := ctx.Font().Math().AxisHeight.Resolve(fontSize)
	frame.SetBaseline(frame.Height()/2.0 + axis)

	ctx.Push(NewFrameFragment(props, fontSize, frame))
	return nil
}

// lineItem creates a line frame for table augmentation.
func lineItem(length Abs, vertical bool, stroke FixedStroke, span Span) *Frame {
	var delta Point
	if vertical {
		delta = PointWithY(length)
	} else {
		delta = PointWithX(length)
	}

	frame := NewSoftFrame(Size{})
	shape := Stroked(&LineGeometry{Delta: delta}, stroke)
	frame.PushShape(Point{}, shape, span)
	return frame
}
