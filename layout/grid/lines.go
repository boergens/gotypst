package grid

import (
	"github.com/boergens/gotypst/layout"
)

// LineGenerator generates grid line segments for rendering.
type LineGenerator struct {
	// Grid is the grid structure.
	Grid *Grid
	// RCols holds resolved column widths.
	RCols []layout.Abs
	// RowHeights maps row index to height.
	RowHeights map[int]layout.Abs
	// IsRTL indicates right-to-left layout.
	IsRTL bool
}

// NewLineGenerator creates a new line generator.
func NewLineGenerator(grid *Grid, rcols []layout.Abs, rowHeights map[int]layout.Abs, isRTL bool) *LineGenerator {
	return &LineGenerator{
		Grid:       grid,
		RCols:      rcols,
		RowHeights: rowHeights,
		IsRTL:      isRTL,
	}
}

// GenerateHorizontalLines generates all horizontal line segments.
// These are lines that run between rows.
func (lg *LineGenerator) GenerateHorizontalLines() []LineSegment {
	var segments []LineSegment

	// Calculate total width.
	totalWidth := layout.Abs(0)
	for _, w := range lg.RCols {
		totalWidth += w
	}

	// Generate line at top (y=0).
	segments = append(segments, lg.generateHLine(0, totalWidth)...)

	// Generate lines between rows and at bottom.
	y := layout.Abs(0)
	for row := 0; row < lg.Grid.RowCount; row++ {
		height := lg.RowHeights[row]
		y += height

		segments = append(segments, lg.generateHLine(y, totalWidth)...)
	}

	return segments
}

// GenerateVerticalLines generates all vertical line segments.
// These are lines that run between columns.
func (lg *LineGenerator) GenerateVerticalLines() []LineSegment {
	var segments []LineSegment

	// Calculate total height.
	totalHeight := layout.Abs(0)
	for row := 0; row < lg.Grid.RowCount; row++ {
		totalHeight += lg.RowHeights[row]
	}

	// Generate line at left (x=0).
	segments = append(segments, lg.generateVLine(0, totalHeight)...)

	// Generate lines between columns and at right.
	x := layout.Abs(0)
	for col := 0; col < lg.Grid.ColCount; col++ {
		x += lg.RCols[col]
		segments = append(segments, lg.generateVLine(x, totalHeight)...)
	}

	return segments
}

// generateHLine generates horizontal line segments at the given y position.
func (lg *LineGenerator) generateHLine(y, maxLength layout.Abs) []LineSegment {
	// Determine the stroke for this line.
	stroke := lg.Grid.Stroke
	if stroke == nil {
		return nil
	}

	// For now, generate a single segment across the full width.
	// TODO: Implement proper segment generation with:
	// - Interruptions for merged cells (colspan/rowspan blocking)
	// - Stroke changes from cell overrides
	// - Priority handling

	return []LineSegment{
		{
			Stroke:   stroke,
			Offset:   y,
			Length:   maxLength,
			Priority: GridStrokePriority,
		},
	}
}

// generateVLine generates vertical line segments at the given x position.
func (lg *LineGenerator) generateVLine(x, maxLength layout.Abs) []LineSegment {
	stroke := lg.Grid.Stroke
	if stroke == nil {
		return nil
	}

	// Adjust x for RTL layout.
	if lg.IsRTL {
		totalWidth := layout.Abs(0)
		for _, w := range lg.RCols {
			totalWidth += w
		}
		x = totalWidth - x
	}

	return []LineSegment{
		{
			Stroke:   stroke,
			Offset:   x,
			Length:   maxLength,
			Priority: GridStrokePriority,
		},
	}
}

// GenerateAllLines generates all grid lines (horizontal and vertical).
func (lg *LineGenerator) GenerateAllLines() (hlines, vlines []LineSegment) {
	return lg.GenerateHorizontalLines(), lg.GenerateVerticalLines()
}

// ResolveStrokePriority determines the winning stroke when multiple strokes overlap.
func ResolveStrokePriority(segments []LineSegment) *Stroke {
	if len(segments) == 0 {
		return nil
	}

	// Find the highest priority stroke.
	best := &segments[0]
	for i := 1; i < len(segments); i++ {
		if segments[i].Priority > best.Priority {
			best = &segments[i]
		}
	}

	return best.Stroke
}

// MergeSegments merges adjacent line segments with the same stroke.
func MergeSegments(segments []LineSegment) []LineSegment {
	if len(segments) <= 1 {
		return segments
	}

	var merged []LineSegment
	current := segments[0]

	for i := 1; i < len(segments); i++ {
		seg := segments[i]

		// Check if this segment can be merged with current.
		if strokesEqual(current.Stroke, seg.Stroke) &&
			current.Priority == seg.Priority &&
			current.Offset+current.Length == seg.Offset {
			// Extend current segment.
			current.Length += seg.Length
		} else {
			// Output current and start new.
			merged = append(merged, current)
			current = seg
		}
	}

	merged = append(merged, current)
	return merged
}

// strokesEqual compares two strokes for equality.
func strokesEqual(a, b *Stroke) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Compare stroke properties.
	return a.Thickness == b.Thickness &&
		a.Cap == b.Cap &&
		a.Join == b.Join
	// Note: Paint comparison is simplified; may need deeper comparison.
}

// CellStrokeResolver resolves strokes for cells with per-side overrides.
type CellStrokeResolver struct {
	// DefaultStroke is the grid-level stroke.
	DefaultStroke *Stroke
}

// NewCellStrokeResolver creates a new resolver.
func NewCellStrokeResolver(defaultStroke *Stroke) *CellStrokeResolver {
	return &CellStrokeResolver{DefaultStroke: defaultStroke}
}

// ResolveTop returns the stroke for the top edge of a cell.
func (r *CellStrokeResolver) ResolveTop(cell *Cell) *Stroke {
	if cell.Stroke.Top != nil {
		return cell.Stroke.Top
	}
	return r.DefaultStroke
}

// ResolveBottom returns the stroke for the bottom edge of a cell.
func (r *CellStrokeResolver) ResolveBottom(cell *Cell) *Stroke {
	if cell.Stroke.Bottom != nil {
		return cell.Stroke.Bottom
	}
	return r.DefaultStroke
}

// ResolveLeft returns the stroke for the left edge of a cell.
func (r *CellStrokeResolver) ResolveLeft(cell *Cell) *Stroke {
	if cell.Stroke.Left != nil {
		return cell.Stroke.Left
	}
	return r.DefaultStroke
}

// ResolveRight returns the stroke for the right edge of a cell.
func (r *CellStrokeResolver) ResolveRight(cell *Cell) *Stroke {
	if cell.Stroke.Right != nil {
		return cell.Stroke.Right
	}
	return r.DefaultStroke
}

// HLineSpec specifies a user-defined horizontal line.
type HLineSpec struct {
	// Y is the row index where the line appears (before this row).
	Y int
	// Stroke is the stroke style.
	Stroke *Stroke
	// Start is the starting column (0 for left edge).
	Start int
	// End is the ending column (ColCount for right edge).
	End int
}

// VLineSpec specifies a user-defined vertical line.
type VLineSpec struct {
	// X is the column index where the line appears (before this column).
	X int
	// Stroke is the stroke style.
	Stroke *Stroke
	// Start is the starting row (0 for top edge).
	Start int
	// End is the ending row (RowCount for bottom edge).
	End int
}

// ExplicitLineGenerator handles user-specified hline/vline elements.
type ExplicitLineGenerator struct {
	HLines []HLineSpec
	VLines []VLineSpec
}

// NewExplicitLineGenerator creates a new explicit line generator.
func NewExplicitLineGenerator() *ExplicitLineGenerator {
	return &ExplicitLineGenerator{}
}

// AddHLine adds a horizontal line specification.
func (g *ExplicitLineGenerator) AddHLine(spec HLineSpec) {
	g.HLines = append(g.HLines, spec)
}

// AddVLine adds a vertical line specification.
func (g *ExplicitLineGenerator) AddVLine(spec VLineSpec) {
	g.VLines = append(g.VLines, spec)
}

// GenerateExplicitSegments generates segments for explicit lines.
func (g *ExplicitLineGenerator) GenerateExplicitSegments(
	rcols []layout.Abs,
	rowHeights map[int]layout.Abs,
) (hsegs, vsegs []LineSegment) {
	// Generate horizontal line segments.
	for _, hl := range g.HLines {
		// Calculate y position.
		y := layout.Abs(0)
		for row := 0; row < hl.Y; row++ {
			y += rowHeights[row]
		}

		// Calculate x offset and length.
		x := layout.Abs(0)
		for col := 0; col < hl.Start && col < len(rcols); col++ {
			x += rcols[col]
		}

		length := layout.Abs(0)
		for col := hl.Start; col < hl.End && col < len(rcols); col++ {
			length += rcols[col]
		}

		if hl.Stroke != nil && length > 0 {
			hsegs = append(hsegs, LineSegment{
				Stroke:   hl.Stroke,
				Offset:   y,
				Length:   length,
				Priority: ExplicitLinePriority,
			})
		}
	}

	// Generate vertical line segments.
	for _, vl := range g.VLines {
		// Calculate x position.
		x := layout.Abs(0)
		for col := 0; col < vl.X && col < len(rcols); col++ {
			x += rcols[col]
		}

		// Calculate y offset and length.
		y := layout.Abs(0)
		for row := 0; row < vl.Start; row++ {
			y += rowHeights[row]
		}

		length := layout.Abs(0)
		for row := vl.Start; row < vl.End; row++ {
			length += rowHeights[row]
		}

		if vl.Stroke != nil && length > 0 {
			vsegs = append(vsegs, LineSegment{
				Stroke:   vl.Stroke,
				Offset:   x,
				Length:   length,
				Priority: ExplicitLinePriority,
			})
		}
	}

	return hsegs, vsegs
}
