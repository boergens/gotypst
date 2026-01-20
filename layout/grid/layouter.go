package grid

import (
	"github.com/boergens/gotypst/layout"
	"github.com/boergens/gotypst/layout/flow"
)

// GridLayouter handles the layout of grid and table elements.
type GridLayouter struct {
	// Grid is the grid structure being laid out.
	Grid *Grid
	// Regions is the available layout regions.
	Regions *flow.Regions
	// RCols contains resolved column widths.
	RCols []layout.Abs
	// Width is the total grid width.
	Width layout.Abs

	// RRows contains resolved row states by region.
	RRows []RowState
	// Current holds the current region state.
	Current Current
	// Finished contains completed region frames.
	Finished []flow.Frame
	// UnbreakableRowsLeft tracks remaining unbreakable rows in a group.
	UnbreakableRowsLeft int

	// HeaderManager manages header and footer state.
	HeaderManager *HeaderManager
	// Rowspans tracks active rowspans.
	Rowspans []Rowspan
	// IsRTL indicates if layout is right-to-left.
	IsRTL bool
	// CurrentRow tracks the current row being processed.
	CurrentRow int
	// RegionIndex tracks which region we're in.
	RegionIndex int
	// HasContentRows indicates if any non-header/footer rows have been placed.
	HasContentRows bool
}

// Rowspan tracks a cell that spans multiple rows.
type Rowspan struct {
	// X is the column index.
	X int
	// Y is the starting row index.
	Y int
	// Count is the number of rows spanned.
	Count int
	// DX is the horizontal offset within the cell.
	DX layout.Abs
	// DY is the vertical offset accumulated so far.
	DY layout.Abs
	// FirstRegion is the region index where this rowspan started.
	FirstRegion int
	// RegionFull is the full available height in the first region.
	RegionFull layout.Abs
	// Heights tracks available height per region.
	Heights []layout.Abs
	// MaxResolvedRow tracks the furthest row resolved for this rowspan.
	MaxResolvedRow *int
	// IsBeingRepeated indicates if this rowspan is in a repeating header.
	IsBeingRepeated bool
	// Cell is the rowspan cell.
	Cell *Cell
	// Frame is the laid out cell content (once resolved).
	Frame *flow.Frame
}

// NewGridLayouter creates a new grid layouter.
func NewGridLayouter(grid *Grid, regions *flow.Regions, isRTL bool) *GridLayouter {
	return &GridLayouter{
		Grid:          grid,
		Regions:       regions,
		IsRTL:         isRTL,
		HeaderManager: NewHeaderManager(),
	}
}

// Layout performs the complete grid layout.
func (g *GridLayouter) Layout() ([]flow.Frame, error) {
	// Phase 1: Measure columns
	if err := g.measureColumns(); err != nil {
		return nil, err
	}

	// Phase 2: Layout rows
	for g.CurrentRow < len(g.Grid.Rows) {
		if err := g.layoutRow(g.CurrentRow); err != nil {
			return nil, err
		}
		g.CurrentRow++
	}

	// Phase 3: Finish the last region
	if err := g.finishRegion(true); err != nil {
		return nil, err
	}

	return g.Finished, nil
}

// measureColumns resolves column widths.
func (g *GridLayouter) measureColumns() error {
	availableWidth := g.Regions.Size.Width

	// First pass: resolve fixed and relative columns
	var remainingWidth layout.Abs = availableWidth
	var autoCount int
	var frTotal layout.Fr

	g.RCols = make([]layout.Abs, len(g.Grid.Cols))

	for i, col := range g.Grid.Cols {
		switch s := col.Sizing.(type) {
		case SizingFixed:
			g.RCols[i] = s.Value
			remainingWidth -= s.Value
		case SizingRelative:
			width := layout.Abs(s.Ratio * float64(availableWidth))
			g.RCols[i] = width
			remainingWidth -= width
		case SizingAuto:
			autoCount++
		case SizingFractional:
			frTotal += s.Fr
		}
	}

	// Second pass: measure auto columns (simplified - just use equal share)
	if autoCount > 0 && remainingWidth > 0 {
		autoWidth := remainingWidth / layout.Abs(autoCount+int(frTotal))
		for i, col := range g.Grid.Cols {
			if _, ok := col.Sizing.(SizingAuto); ok {
				g.RCols[i] = autoWidth
				remainingWidth -= autoWidth
			}
		}
	}

	// Third pass: distribute fractional columns
	if frTotal > 0 && remainingWidth > 0 {
		for i, col := range g.Grid.Cols {
			if fr, ok := col.Sizing.(SizingFractional); ok {
				g.RCols[i] = layout.Abs(float64(fr.Fr) / float64(frTotal) * float64(remainingWidth))
			}
		}
	}

	// Calculate total width
	for _, w := range g.RCols {
		g.Width += w
	}

	return nil
}

// layoutRow processes a single row.
func (g *GridLayouter) layoutRow(rowIndex int) error {
	row := g.Grid.Rows[rowIndex]

	// Check if this is a header row
	if header := g.findHeaderContaining(rowIndex); header != nil {
		return g.layoutHeaderRow(rowIndex, header)
	}

	// Check if this is a footer row
	if footer := g.findFooterContaining(rowIndex); footer != nil {
		return g.layoutFooterRow(rowIndex, footer)
	}

	// This is a content row - promote any pending headers first
	if len(g.HeaderManager.PendingHeaders) > 0 {
		g.HeaderManager.PromoteAllPendingHeaders()
	}

	// Check if we need to break to a new region
	rowHeight, err := g.measureRowHeight(rowIndex)
	if err != nil {
		return err
	}

	availableHeight := g.Current.AvailableHeight(g.Regions.Size.Height)
	if rowHeight > availableHeight && g.HasContentRows {
		// Need to break - finish current region and start new one
		if err := g.finishRegion(false); err != nil {
			return err
		}
		if err := g.startNewRegion(); err != nil {
			return err
		}
	}

	// Layout the row
	frame, err := g.layoutRowContent(rowIndex, row)
	if err != nil {
		return err
	}

	// Track the row
	g.RRows = append(g.RRows, RowState{
		Height:   rowHeight,
		Y:        g.Current.UsedHeight,
		IsGutter: row.IsGutter,
	})
	g.Current.UsedHeight += rowHeight
	g.HasContentRows = true

	// Add to current region output
	_ = frame // Will be added to finished frame in finishRegion

	return nil
}

// layoutHeaderRow processes a header row.
func (g *GridLayouter) layoutHeaderRow(rowIndex int, header *Header) error {
	row := g.Grid.Rows[rowIndex]

	// If this is a new header (first row), handle level conflicts
	if rowIndex == header.StartRow {
		g.HeaderManager.HandleConflictingHeaders(header.Level)
	}

	// Layout the header row
	frame, err := g.layoutRowContent(rowIndex, row)
	if err != nil {
		return err
	}

	rowHeight, err := g.measureRowHeight(rowIndex)
	if err != nil {
		return err
	}

	// If this is the first row of this header, create pending header entry
	if rowIndex == header.StartRow {
		g.HeaderManager.AddPendingHeader(header, []flow.Frame{frame}, []layout.Abs{rowHeight}, g.Current.UsedHeight)
		g.Current.PendingHeaderHeight += rowHeight
	} else {
		// Add to existing pending header
		for _, pending := range g.HeaderManager.PendingHeaders {
			if pending.Header == header {
				pending.Frames = append(pending.Frames, frame)
				pending.Heights = append(pending.Heights, rowHeight)
				break
			}
		}
		g.Current.PendingHeaderHeight += rowHeight
	}

	// Track the row
	g.RRows = append(g.RRows, RowState{
		Height:   rowHeight,
		Y:        g.Current.UsedHeight,
		IsGutter: row.IsGutter,
	})
	g.Current.UsedHeight += rowHeight

	// Record for orphan detection
	g.HeaderManager.RecordFinishedHeaderRow(header.Level, g.Current.UsedHeight)

	return nil
}

// layoutFooterRow processes a footer row.
func (g *GridLayouter) layoutFooterRow(rowIndex int, footer *Footer) error {
	row := g.Grid.Rows[rowIndex]

	// Layout the footer row
	frame, err := g.layoutRowContent(rowIndex, row)
	if err != nil {
		return err
	}

	rowHeight, err := g.measureRowHeight(rowIndex)
	if err != nil {
		return err
	}

	// Store footer information
	if rowIndex == footer.StartRow {
		g.HeaderManager.SetFooter(footer, []flow.Frame{frame}, []layout.Abs{rowHeight})
	} else {
		g.HeaderManager.FooterFrames = append(g.HeaderManager.FooterFrames, frame)
		g.HeaderManager.FooterHeights = append(g.HeaderManager.FooterHeights, rowHeight)
	}

	g.Current.FooterHeight += rowHeight

	// Track the row
	g.RRows = append(g.RRows, RowState{
		Height:   rowHeight,
		Y:        g.Current.UsedHeight,
		IsGutter: row.IsGutter,
	})

	return nil
}

// layoutRowContent creates the frame for a row's content.
func (g *GridLayouter) layoutRowContent(rowIndex int, row Track) (flow.Frame, error) {
	rowHeight, err := g.measureRowHeight(rowIndex)
	if err != nil {
		return flow.Frame{}, err
	}

	frame := flow.NewFrame(layout.Size{Width: g.Width, Height: rowHeight})

	// Layout each cell in this row
	x := layout.Abs(0)
	for colIndex := range g.Grid.Cols {
		cell := g.findCell(colIndex, rowIndex)
		if cell == nil {
			x += g.RCols[colIndex]
			continue
		}

		// Skip cells that started in a previous row (rowspan continuation)
		if cell.Y != rowIndex {
			x += g.RCols[colIndex]
			continue
		}

		cellWidth := g.cellWidth(cell)
		cellFrame := flow.NewFrame(layout.Size{Width: cellWidth, Height: rowHeight})

		// Position based on RTL
		pos := layout.Point{Y: 0}
		if g.IsRTL {
			pos.X = g.Width - x - cellWidth
		} else {
			pos.X = x
		}

		frame.PushFrame(pos, cellFrame)
		x += cellWidth
	}

	return frame, nil
}

// finishRegion completes the current region.
func (g *GridLayouter) finishRegion(isFinal bool) error {
	// Check for orphan headers (headers with no content following)
	if g.HeaderManager.CheckOrphanHeaders(g.HasContentRows) {
		// Remove orphan headers by rolling back to before them
		g.HeaderManager.ClearPendingHeaders()
		// Note: In a full implementation, we'd also need to remove the
		// rows from RRows and adjust Current.UsedHeight
	}

	// Create the output frame for this region
	frameHeight := g.Current.UsedHeight
	if g.HeaderManager.ShouldRepeatFooter(isFinal) {
		frameHeight += g.HeaderManager.FooterHeight()
	}

	output := flow.NewFrame(layout.Size{Width: g.Width, Height: frameHeight})

	// Add all content rows
	for _, rrow := range g.RRows {
		// The actual row frames would be placed here
		_ = rrow
	}

	// Add footer if appropriate
	if g.HeaderManager.ShouldRepeatFooter(isFinal) {
		g.HeaderManager.PlaceFooter(&output, g.Current.UsedHeight, g.Width)
	}

	g.Finished = append(g.Finished, output)

	// Reset for next region
	g.RRows = nil
	g.HeaderManager.ClearFinishedHeaderRows()

	return nil
}

// startNewRegion begins layout in a new region.
func (g *GridLayouter) startNewRegion() error {
	g.RegionIndex++
	g.HasContentRows = false

	// Update regions
	if len(g.Regions.Backlog) > 0 {
		g.Regions.Size.Height = g.Regions.Backlog[0]
		g.Regions.Backlog = g.Regions.Backlog[1:]
	} else if g.Regions.Last != nil {
		g.Regions.Size = *g.Regions.Last
		g.Regions.Last = nil
	}

	// Reset current state
	g.Current = Current{}

	// Place repeating headers
	frames, heights := g.HeaderManager.PrepareHeadersForNewRegion()
	for i, h := range heights {
		g.Current.RepeatingHeaderHeight += h
		g.Current.UsedHeight += h
		// Place the header frame
		_ = frames[i]
	}

	// Reserve space for footer if it repeats
	if g.HeaderManager.ShouldRepeatFooter(false) {
		g.Current.FooterHeight = g.HeaderManager.FooterHeight()
	}

	return nil
}

// measureRowHeight determines the height of a row.
func (g *GridLayouter) measureRowHeight(rowIndex int) (layout.Abs, error) {
	row := g.Grid.Rows[rowIndex]

	switch s := row.Sizing.(type) {
	case SizingFixed:
		return s.Value, nil
	case SizingRelative:
		return layout.Abs(s.Ratio * float64(g.Regions.Full.Height)), nil
	case SizingAuto:
		// Measure content height (simplified - would need actual layout)
		return layout.Abs(20), nil // Default row height
	case SizingFractional:
		// Fractional rows are resolved at region end
		return layout.Abs(20), nil
	default:
		return layout.Abs(20), nil
	}
}

// findCell finds the cell at the given column and row.
func (g *GridLayouter) findCell(x, y int) *Cell {
	for i := range g.Grid.Cells {
		c := &g.Grid.Cells[i]
		if c.X <= x && x < c.X+c.Colspan &&
			c.Y <= y && y < c.Y+c.Rowspan {
			return c
		}
	}
	return nil
}

// cellWidth calculates the total width of a cell including colspans.
func (g *GridLayouter) cellWidth(cell *Cell) layout.Abs {
	var width layout.Abs
	for i := cell.X; i < cell.X+cell.Colspan && i < len(g.RCols); i++ {
		width += g.RCols[i]
	}
	return width
}

// findHeaderContaining finds the header that contains the given row, if any.
func (g *GridLayouter) findHeaderContaining(rowIndex int) *Header {
	// This would be populated from the Grid's header definitions
	// For now, return nil (no headers defined)
	return nil
}

// findFooterContaining finds the footer that contains the given row, if any.
func (g *GridLayouter) findFooterContaining(rowIndex int) *Footer {
	// This would be populated from the Grid's footer definitions
	// For now, return nil (no footers defined)
	return nil
}
