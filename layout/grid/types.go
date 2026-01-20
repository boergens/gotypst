package grid

import (
	"github.com/boergens/gotypst/layout"
	"github.com/boergens/gotypst/layout/flow"
)

// Sizing represents a track (row or column) sizing specification.
type Sizing interface {
	isSizing()
}

// SizingAuto indicates the track should be sized to fit its content.
type SizingAuto struct{}

func (SizingAuto) isSizing() {}

// SizingRel indicates a relative/absolute size.
type SizingRel struct {
	// Abs is the absolute component.
	Abs layout.Abs
	// Ratio is the relative component (0.0-1.0).
	Ratio float64
}

func (SizingRel) isSizing() {}

// RelativeTo resolves the relative sizing to an absolute value.
func (s SizingRel) RelativeTo(base layout.Abs) layout.Abs {
	return s.Abs + layout.Abs(s.Ratio*float64(base))
}

// SizingFr indicates a fractional size.
type SizingFr struct {
	Fr layout.Fr
}

func (SizingFr) isSizing() {}

// Cell represents an individual cell in the grid.
type Cell struct {
	// Body is the cell's content, to be laid out.
	Body interface{}
	// X is the column index (0-based).
	X int
	// Y is the row index (0-based).
	Y int
	// Colspan is the number of columns this cell spans.
	Colspan int
	// Rowspan is the number of rows this cell spans.
	Rowspan int
	// Fill is the cell's background fill (nil for no fill).
	Fill interface{}
	// Stroke holds per-side stroke settings (nil entries use default).
	Stroke Sides[*Stroke]
	// Breakable indicates if this cell can break across regions.
	Breakable bool
	// Align holds the horizontal and vertical alignment for cell content.
	// X is horizontal (start=left, center, end=right).
	// Y is vertical (start=top, center=middle, end=bottom).
	Align flow.Axes[flow.FixedAlignment]
}

// Entry represents a single slot in the resolved grid.
// A slot can be empty, contain a cell, or be merged into another cell.
type Entry interface {
	isEntry()
}

// EntryCell indicates a slot contains a cell.
type EntryCell struct {
	Cell *Cell
}

func (EntryCell) isEntry() {}

// EntryMerged indicates a slot is merged into another cell (via colspan/rowspan).
type EntryMerged struct {
	// Parent points to the original cell this is merged into.
	Parent *Cell
}

func (EntryMerged) isEntry() {}

// Grid holds the fully resolved grid structure.
type Grid struct {
	// Cols contains the column sizing specifications.
	Cols []Sizing
	// Rows contains the row sizing specifications.
	Rows []Sizing
	// Entries is a 2D grid of entries, indexed as [y*cols + x].
	Entries []Entry
	// ColCount is the number of columns.
	ColCount int
	// RowCount is the number of rows.
	RowCount int
	// HasGutter indicates if gutter rows/cols are present.
	HasGutter bool
	// ColGutter is the gutter between columns.
	ColGutter layout.Abs
	// RowGutter is the gutter between rows.
	RowGutter layout.Abs
	// Fill is the default fill for cells.
	Fill interface{}
	// Stroke is the default stroke for grid lines.
	Stroke *Stroke
}

// EntryAt returns the entry at (x, y), or nil if out of bounds.
func (g *Grid) EntryAt(x, y int) Entry {
	if x < 0 || x >= g.ColCount || y < 0 || y >= g.RowCount {
		return nil
	}
	return g.Entries[y*g.ColCount+x]
}

// CellAt returns the cell at (x, y), or nil if the slot is empty or merged.
func (g *Grid) CellAt(x, y int) *Cell {
	entry := g.EntryAt(x, y)
	if ec, ok := entry.(EntryCell); ok {
		return ec.Cell
	}
	return nil
}

// EffectiveRowCount returns the number of rows excluding trailing gutter.
func (g *Grid) EffectiveRowCount() int {
	if g.HasGutter && g.RowCount > 0 {
		return g.RowCount - 1
	}
	return g.RowCount
}

// RowState tracks the state of rows in a region.
type RowState struct {
	// Heights maps row index to resolved height.
	Heights map[int]layout.Abs
	// IsGutter maps row index to whether it's a gutter row.
	IsGutter map[int]bool
}

// NewRowState creates an empty RowState.
func NewRowState() RowState {
	return RowState{
		Heights:  make(map[int]layout.Abs),
		IsGutter: make(map[int]bool),
	}
}

// Clone creates a copy of the RowState.
func (r RowState) Clone() RowState {
	heights := make(map[int]layout.Abs, len(r.Heights))
	for k, v := range r.Heights {
		heights[k] = v
	}
	isGutter := make(map[int]bool, len(r.IsGutter))
	for k, v := range r.IsGutter {
		isGutter[k] = v
	}
	return RowState{Heights: heights, IsGutter: isGutter}
}

// Current tracks the current region's state during layout.
type Current struct {
	// RegionIdx is the index of the current region.
	RegionIdx int
	// Height is the current height consumed in this region.
	Height layout.Abs
	// Row is the current row index.
	Row int
	// Initial is the initial position in this region.
	Initial layout.Abs
}

// Rowspan tracks cells spanning multiple rows.
type Rowspan struct {
	// X is the column index.
	X int
	// Y is the starting row index.
	Y int
	// Disambiguator is used to distinguish cells at the same position.
	Disambiguator int
	// RowspanCount is the number of rows this cell spans.
	RowspanCount int
	// IsUnbreakable indicates if the rowspan cannot be broken.
	IsUnbreakable bool
	// DX is the horizontal offset in the grid.
	DX layout.Abs
	// DY is the vertical offset in the first region.
	DY layout.Abs
	// FirstRegion is the region index where this rowspan starts.
	FirstRegion int
	// RegionFull is the full height available in the first region.
	RegionFull layout.Abs
	// Heights holds per-region available heights.
	Heights []layout.Abs
	// MaxResolvedRow is the maximum row resolved so far (nil = none).
	MaxResolvedRow *int
	// IsBeingRepeated indicates if this rowspan is part of a repeating header.
	IsBeingRepeated bool
}

// Header represents a repeating header or pending header.
type Header struct {
	// StartY is the starting row index of the header.
	StartY int
	// EndY is the ending row index (exclusive).
	EndY int
	// Level is the header nesting level.
	Level int
	// Frame is the laid out header content (nil until layout).
	Frame *flow.Frame
}

// Footer represents a footer section.
type Footer struct {
	// StartY is the starting row index.
	StartY int
	// EndY is the ending row index (exclusive).
	EndY int
	// Frame is the laid out footer content.
	Frame *flow.Frame
}

// FinishedHeaderRowInfo tracks info about finished header rows.
type FinishedHeaderRowInfo struct {
	// Y is the row index.
	Y int
	// Height is the resolved height.
	Height layout.Abs
}

// Stroke represents stroke styling for grid lines.
type Stroke struct {
	// Paint is the stroke color/pattern.
	Paint interface{}
	// Thickness is the stroke thickness.
	Thickness layout.Abs
	// Dash is the dash pattern (nil for solid).
	Dash []layout.Abs
	// Cap is the line cap style.
	Cap StrokeCap
	// Join is the line join style.
	Join StrokeJoin
}

// StrokeCap represents line cap styles.
type StrokeCap int

const (
	StrokeCapButt StrokeCap = iota
	StrokeCapRound
	StrokeCapSquare
)

// StrokeJoin represents line join styles.
type StrokeJoin int

const (
	StrokeJoinMiter StrokeJoin = iota
	StrokeJoinRound
	StrokeJoinBevel
)

// Sides holds values for all four sides.
type Sides[T any] struct {
	Left, Top, Right, Bottom T
}

// LineSegment represents a drawable grid line segment.
type LineSegment struct {
	// Stroke is the stroke style for this segment.
	Stroke *Stroke
	// Offset is the position along the perpendicular axis.
	Offset layout.Abs
	// Length is the length of the segment.
	Length layout.Abs
	// Priority determines which stroke wins on overlaps.
	Priority StrokePriority
}

// StrokePriority determines which stroke takes precedence on overlaps.
type StrokePriority int

const (
	// GridStrokePriority is for global grid styling (lowest).
	GridStrokePriority StrokePriority = iota
	// CellStrokePriority is for per-cell stroke overrides.
	CellStrokePriority
	// ExplicitLinePriority is for user-placed hline/vline (highest).
	ExplicitLinePriority
)

// Axes holds a pair of values for horizontal (X) and vertical (Y) axes.
type Axes[T any] struct {
	X, Y T
}
