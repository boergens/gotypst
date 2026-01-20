package grid

import (
	"github.com/boergens/gotypst/layout"
)

// StrokePriority determines which stroke takes precedence when multiple
// strokes could apply to the same position. Higher values take priority.
type StrokePriority int

const (
	// GridStroke is the lowest priority - global grid styling.
	GridStroke StrokePriority = iota
	// CellStroke is medium priority - per-cell stroke overrides.
	CellStroke
	// ExplicitLine is highest priority - user-placed hline/vline.
	ExplicitLine
)

// LineSegment represents a drawable grid line segment.
// Grid lines are drawn between cells as borders and separators.
type LineSegment struct {
	// Stroke is the stroke style for this segment.
	// May be shared with other segments; do not mutate.
	Stroke *Stroke
	// Offset is the position along the axis (perpendicular to the line direction).
	Offset layout.Abs
	// Length is the length of the segment along the line direction.
	Length layout.Abs
	// Priority determines which stroke wins in overlapping positions.
	Priority StrokePriority
}

// Stroke represents stroke styling for grid lines.
// This is a simplified version focused on grid line rendering.
type Stroke struct {
	// Paint is the stroke color or gradient.
	Paint Paint
	// Thickness is the stroke width in points.
	Thickness layout.Abs
	// Cap is the line cap style.
	Cap LineCap
	// Join is the line join style.
	Join LineJoin
	// Dash is the dash pattern (nil for solid lines).
	Dash *StrokeDash
	// MiterLimit is the miter limit for miter joins.
	MiterLimit float64
}

// NewStroke creates a new stroke with default settings.
func NewStroke(paint Paint, thickness layout.Abs) *Stroke {
	return &Stroke{
		Paint:      paint,
		Thickness:  thickness,
		Cap:        LineCapButt,
		Join:       LineJoinMiter,
		MiterLimit: 4.0,
	}
}

// WithCap returns a copy of the stroke with the given cap style.
func (s *Stroke) WithCap(cap LineCap) *Stroke {
	copy := *s
	copy.Cap = cap
	return &copy
}

// WithJoin returns a copy of the stroke with the given join style.
func (s *Stroke) WithJoin(join LineJoin) *Stroke {
	copy := *s
	copy.Join = join
	return &copy
}

// WithDash returns a copy of the stroke with the given dash pattern.
func (s *Stroke) WithDash(dash *StrokeDash) *Stroke {
	copy := *s
	copy.Dash = dash
	return &copy
}

// StrokeDash represents a dash pattern for stroked lines.
type StrokeDash struct {
	// Array contains the dash lengths (alternating on/off).
	Array []layout.Abs
	// Phase is the starting offset into the pattern.
	Phase layout.Abs
}

// NewDash creates a simple dash pattern with on/off lengths.
func NewDash(on, off layout.Abs) *StrokeDash {
	return &StrokeDash{
		Array: []layout.Abs{on, off},
		Phase: 0,
	}
}

// LineCap represents line cap styles.
type LineCap int

const (
	// LineCapButt ends lines at their endpoints with no extension.
	LineCapButt LineCap = iota
	// LineCapRound ends lines with a semicircular extension.
	LineCapRound
	// LineCapSquare ends lines with a rectangular extension.
	LineCapSquare
)

// LineJoin represents line join styles.
type LineJoin int

const (
	// LineJoinMiter joins lines with a sharp corner.
	LineJoinMiter LineJoin = iota
	// LineJoinRound joins lines with a rounded corner.
	LineJoinRound
	// LineJoinBevel joins lines with a beveled corner.
	LineJoinBevel
)

// Paint represents a fill or stroke paint.
type Paint interface {
	isPaint()
}

// Color represents an RGBA color.
type Color struct {
	R, G, B, A uint8
}

func (Color) isPaint() {}

// NewColor creates a new color from RGBA values.
func NewColor(r, g, b, a uint8) Color {
	return Color{R: r, G: g, B: b, A: a}
}

// NewRGB creates an opaque color from RGB values.
func NewRGB(r, g, b uint8) Color {
	return Color{R: r, G: g, B: b, A: 255}
}

// Predefined colors for convenience.
var (
	Black = NewRGB(0, 0, 0)
	White = NewRGB(255, 255, 255)
	Gray  = NewRGB(128, 128, 128)
)

// Axis represents the direction of a grid line.
type Axis int

const (
	// AxisX is horizontal (rows, hlines).
	AxisX Axis = iota
	// AxisY is vertical (columns, vlines).
	AxisY
)

// TrackSizing represents how a track (row or column) is sized.
type TrackSizing interface {
	isTrackSizing()
}

// AutoTrack indicates automatic sizing based on content.
type AutoTrack struct{}

func (AutoTrack) isTrackSizing() {}

// FixedTrack indicates a fixed size in absolute units.
type FixedTrack struct {
	Size layout.Abs
}

func (FixedTrack) isTrackSizing() {}

// FrTrack indicates a fractional size.
type FrTrack struct {
	Fr layout.Fr
}

func (FrTrack) isTrackSizing() {}

// RelativeTrack indicates a size relative to the container.
type RelativeTrack struct {
	Ratio float64 // 0.0 to 1.0
}

func (RelativeTrack) isTrackSizing() {}

// ResolvedLine represents a resolved grid line with position and stroke info.
type ResolvedLine struct {
	// Index is the line index (0 = before first track, 1 = after first track, etc.).
	Index int
	// Position is the position of the line in absolute units.
	Position layout.Abs
	// Stroke is the stroke to use for this line (may be nil for no line).
	Stroke *Stroke
	// Priority is the priority of this stroke.
	Priority StrokePriority
}

// CellSpan represents a cell that may span multiple rows/columns.
type CellSpan struct {
	// X is the starting column index.
	X int
	// Y is the starting row index.
	Y int
	// Colspan is the number of columns spanned (default 1).
	Colspan int
	// Rowspan is the number of rows spanned (default 1).
	Rowspan int
	// Stroke is the cell's stroke override (nil to use grid default).
	Stroke *Stroke
}

// Contains returns true if this cell contains the given position.
func (c *CellSpan) Contains(x, y int) bool {
	return x >= c.X && x < c.X+c.Colspan && y >= c.Y && y < c.Y+c.Rowspan
}

// BlocksHLine returns true if this cell blocks a horizontal line at the given row.
// A line is blocked if it would cross through the middle of a rowspan.
func (c *CellSpan) BlocksHLine(row int) bool {
	return row > c.Y && row < c.Y+c.Rowspan
}

// BlocksVLine returns true if this cell blocks a vertical line at the given column.
// A line is blocked if it would cross through the middle of a colspan.
func (c *CellSpan) BlocksVLine(col int) bool {
	return col > c.X && col < c.X+c.Colspan
}

// GridLines holds the resolved grid lines for rendering.
type GridLines struct {
	// HLines are the horizontal lines (between rows).
	HLines []ResolvedLine
	// VLines are the vertical lines (between columns).
	VLines []ResolvedLine
	// Cells contains cell spans for blocking line segments.
	Cells []CellSpan
}

// LineGenerator generates line segments for grid rendering.
type LineGenerator struct {
	// Cols are the resolved column widths.
	Cols []layout.Abs
	// Rows are the resolved row heights.
	Rows []layout.Abs
	// DefaultStroke is the default stroke for grid lines.
	DefaultStroke *Stroke
	// HLineStrokes are overrides for horizontal lines (indexed by row).
	HLineStrokes []*Stroke
	// VLineStrokes are overrides for vertical lines (indexed by column).
	VLineStrokes []*Stroke
	// Cells are the cell spans for blocking detection.
	Cells []CellSpan
	// IsRTL indicates right-to-left layout.
	IsRTL bool
}

// GenerateHLineSegments generates horizontal line segments for a given row.
// The row parameter is the line index (0 = top edge, len(Rows) = bottom edge).
func (g *LineGenerator) GenerateHLineSegments(row int) []LineSegment {
	if len(g.Cols) == 0 {
		return nil
	}

	// Determine the stroke for this row
	stroke := g.getHLineStroke(row)
	if stroke == nil {
		return nil
	}

	// Calculate the y position
	y := g.getHLinePosition(row)

	// Generate segments, breaking at cell spans
	var segments []LineSegment
	var x layout.Abs
	segmentLen := layout.Abs(0)

	for col := 0; col <= len(g.Cols); col++ {
		// Check if this position is blocked by a cell span
		blocked := false
		for _, cell := range g.Cells {
			if cell.BlocksHLine(row) && col > cell.X && col <= cell.X+cell.Colspan {
				blocked = true
				break
			}
		}

		if blocked {
			// Emit current segment if any
			if segmentLen > 0 {
				segments = append(segments, LineSegment{
					Stroke:   stroke,
					Offset:   y,
					Length:   segmentLen,
					Priority: g.getHLinePriority(row),
				})
			}
			// Skip past the blocked cell
			if col < len(g.Cols) {
				x += g.Cols[col]
			}
			segmentLen = 0
		} else {
			// Extend current segment
			if col < len(g.Cols) {
				segmentLen += g.Cols[col]
				x += g.Cols[col]
			}
		}
	}

	// Emit final segment
	if segmentLen > 0 {
		segments = append(segments, LineSegment{
			Stroke:   stroke,
			Offset:   y,
			Length:   segmentLen,
			Priority: g.getHLinePriority(row),
		})
	}

	return segments
}

// GenerateVLineSegments generates vertical line segments for a given column.
// The col parameter is the line index (0 = left edge, len(Cols) = right edge).
func (g *LineGenerator) GenerateVLineSegments(col int) []LineSegment {
	if len(g.Rows) == 0 {
		return nil
	}

	// Determine the stroke for this column
	stroke := g.getVLineStroke(col)
	if stroke == nil {
		return nil
	}

	// Calculate the x position
	x := g.getVLinePosition(col)

	// Generate segments, breaking at cell spans
	var segments []LineSegment
	var y layout.Abs
	segmentLen := layout.Abs(0)

	for row := 0; row <= len(g.Rows); row++ {
		// Check if this position is blocked by a cell span
		blocked := false
		for _, cell := range g.Cells {
			if cell.BlocksVLine(col) && row > cell.Y && row <= cell.Y+cell.Rowspan {
				blocked = true
				break
			}
		}

		if blocked {
			// Emit current segment if any
			if segmentLen > 0 {
				segments = append(segments, LineSegment{
					Stroke:   stroke,
					Offset:   x,
					Length:   segmentLen,
					Priority: g.getVLinePriority(col),
				})
			}
			// Skip past the blocked cell
			if row < len(g.Rows) {
				y += g.Rows[row]
			}
			segmentLen = 0
		} else {
			// Extend current segment
			if row < len(g.Rows) {
				segmentLen += g.Rows[row]
				y += g.Rows[row]
			}
		}
	}

	// Emit final segment
	if segmentLen > 0 {
		segments = append(segments, LineSegment{
			Stroke:   stroke,
			Offset:   x,
			Length:   segmentLen,
			Priority: g.getVLinePriority(col),
		})
	}

	return segments
}

// GenerateAllSegments generates all line segments for the grid.
func (g *LineGenerator) GenerateAllSegments() (hSegments, vSegments []LineSegment) {
	// Need both dimensions for a valid grid
	if len(g.Cols) == 0 || len(g.Rows) == 0 {
		return nil, nil
	}

	// Generate horizontal lines
	for row := 0; row <= len(g.Rows); row++ {
		hSegments = append(hSegments, g.GenerateHLineSegments(row)...)
	}

	// Generate vertical lines
	for col := 0; col <= len(g.Cols); col++ {
		vSegments = append(vSegments, g.GenerateVLineSegments(col)...)
	}

	return hSegments, vSegments
}

// getHLinePosition returns the y position for a horizontal line.
func (g *LineGenerator) getHLinePosition(row int) layout.Abs {
	var y layout.Abs
	for i := 0; i < row && i < len(g.Rows); i++ {
		y += g.Rows[i]
	}
	return y
}

// getVLinePosition returns the x position for a vertical line.
func (g *LineGenerator) getVLinePosition(col int) layout.Abs {
	var x layout.Abs
	for i := 0; i < col && i < len(g.Cols); i++ {
		x += g.Cols[i]
	}
	// Handle RTL layout
	if g.IsRTL {
		totalWidth := layout.Abs(0)
		for _, w := range g.Cols {
			totalWidth += w
		}
		x = totalWidth - x
	}
	return x
}

// getHLineStroke returns the stroke for a horizontal line.
func (g *LineGenerator) getHLineStroke(row int) *Stroke {
	if row >= 0 && row < len(g.HLineStrokes) && g.HLineStrokes[row] != nil {
		return g.HLineStrokes[row]
	}
	return g.DefaultStroke
}

// getVLineStroke returns the stroke for a vertical line.
func (g *LineGenerator) getVLineStroke(col int) *Stroke {
	if col >= 0 && col < len(g.VLineStrokes) && g.VLineStrokes[col] != nil {
		return g.VLineStrokes[col]
	}
	return g.DefaultStroke
}

// getHLinePriority returns the priority for a horizontal line.
func (g *LineGenerator) getHLinePriority(row int) StrokePriority {
	if row >= 0 && row < len(g.HLineStrokes) && g.HLineStrokes[row] != nil {
		return ExplicitLine
	}
	return GridStroke
}

// getVLinePriority returns the priority for a vertical line.
func (g *LineGenerator) getVLinePriority(col int) StrokePriority {
	if col >= 0 && col < len(g.VLineStrokes) && g.VLineStrokes[col] != nil {
		return ExplicitLine
	}
	return GridStroke
}

// LineSegmentBuilder provides a state-machine approach to building line segments.
// This is useful when generating segments incrementally while iterating tracks.
type LineSegmentBuilder struct {
	segments     []LineSegment
	currentStart layout.Abs
	currentLen   layout.Abs
	currentStroke *Stroke
	currentPriority StrokePriority
	offset       layout.Abs
}

// NewLineSegmentBuilder creates a new builder at the given offset position.
func NewLineSegmentBuilder(offset layout.Abs) *LineSegmentBuilder {
	return &LineSegmentBuilder{
		offset: offset,
	}
}

// Extend extends the current segment by the given length with the given stroke.
// If the stroke changes, the current segment is finalized and a new one started.
func (b *LineSegmentBuilder) Extend(length layout.Abs, stroke *Stroke, priority StrokePriority) {
	if stroke == nil {
		// No stroke means a gap - finalize current segment
		b.Finalize()
		b.currentStart += length
		return
	}

	if b.currentStroke == nil {
		// Start new segment
		b.currentStroke = stroke
		b.currentPriority = priority
		b.currentLen = length
	} else if b.strokesMatch(stroke) && priority == b.currentPriority {
		// Extend current segment
		b.currentLen += length
	} else {
		// Stroke changed - finalize and start new
		b.Finalize()
		b.currentStroke = stroke
		b.currentPriority = priority
		b.currentLen = length
	}
}

// Gap adds a gap (no stroke) of the given length.
func (b *LineSegmentBuilder) Gap(length layout.Abs) {
	b.Finalize()
	b.currentStart += length
}

// Finalize emits the current segment if any.
func (b *LineSegmentBuilder) Finalize() {
	if b.currentLen > 0 && b.currentStroke != nil {
		b.segments = append(b.segments, LineSegment{
			Stroke:   b.currentStroke,
			Offset:   b.offset,
			Length:   b.currentLen,
			Priority: b.currentPriority,
		})
	}
	b.currentStart += b.currentLen
	b.currentLen = 0
	b.currentStroke = nil
}

// Segments returns all finalized segments.
// Call Finalize() first to ensure the last segment is included.
func (b *LineSegmentBuilder) Segments() []LineSegment {
	return b.segments
}

// strokesMatch checks if two strokes are equivalent for merging.
func (b *LineSegmentBuilder) strokesMatch(other *Stroke) bool {
	if b.currentStroke == nil || other == nil {
		return b.currentStroke == other
	}
	// For simplicity, compare by pointer (strokes are typically shared)
	// A more complete implementation would compare all fields
	return b.currentStroke == other
}
