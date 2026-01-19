// Rowspans provides tracking for cells that span multiple rows.
//
// Rowspan handling is complex because:
// - Spanning cells may cross page boundaries
// - Row heights must account for spanning cell requirements
// - Spanning cells need their content distributed across regions

package grid

import (
	"github.com/boergens/gotypst/layout"
)

// RowspanTracker tracks active rowspans during layout.
type RowspanTracker struct {
	// Active contains currently active rowspans.
	Active []*RowspanState
}

// RowspanState tracks the state of a single rowspan.
type RowspanState struct {
	// Cell is the spanning cell.
	Cell *Cell

	// StartY is the starting row (same as Cell.Y).
	StartY int

	// CurrentY is the current row being processed.
	CurrentY int

	// AllocatedHeight is the height already allocated to this span.
	AllocatedHeight layout.Abs

	// Frames contains laid out frames for each region.
	Frames []*layout.Frame

	// BacklogHeights tracks height consumed in each region.
	BacklogHeights []layout.Abs
}

// NewRowspanTracker creates a new rowspan tracker.
func NewRowspanTracker() *RowspanTracker {
	return &RowspanTracker{
		Active: nil,
	}
}

// Start begins tracking a new rowspan.
func (t *RowspanTracker) Start(cell *Cell) *RowspanState {
	state := &RowspanState{
		Cell:            cell,
		StartY:          cell.Y,
		CurrentY:        cell.Y,
		AllocatedHeight: 0,
		Frames:          nil,
		BacklogHeights:  nil,
	}
	t.Active = append(t.Active, state)
	return state
}

// ActiveAt returns all active rowspans at the given row.
func (t *RowspanTracker) ActiveAt(y int) []*RowspanState {
	var active []*RowspanState
	for _, rs := range t.Active {
		if rs.Cell.Y <= y && y < rs.Cell.EndY() {
			active = append(active, rs)
		}
	}
	return active
}

// CompletedAt returns rowspans that complete at the given row.
func (t *RowspanTracker) CompletedAt(y int) []*RowspanState {
	var completed []*RowspanState
	for _, rs := range t.Active {
		if rs.Cell.EndY()-1 == y {
			completed = append(completed, rs)
		}
	}
	return completed
}

// Remove removes a completed rowspan from tracking.
func (t *RowspanTracker) Remove(state *RowspanState) {
	for i, rs := range t.Active {
		if rs == state {
			t.Active = append(t.Active[:i], t.Active[i+1:]...)
			return
		}
	}
}

// Clear removes all tracked rowspans.
func (t *RowspanTracker) Clear() {
	t.Active = nil
}

// AllocateHeight allocates additional height to a rowspan.
func (state *RowspanState) AllocateHeight(height layout.Abs) {
	state.AllocatedHeight += height
}

// RemainingHeight calculates how much more height this span needs.
func (state *RowspanState) RemainingHeight(totalRequired layout.Abs) layout.Abs {
	remaining := totalRequired - state.AllocatedHeight
	if remaining < 0 {
		return 0
	}
	return remaining
}

// AddFrame adds a frame for a region.
func (state *RowspanState) AddFrame(frame *layout.Frame) {
	state.Frames = append(state.Frames, frame)
}

// AddBacklogHeight records height consumed in a region.
func (state *RowspanState) AddBacklogHeight(height layout.Abs) {
	state.BacklogHeights = append(state.BacklogHeights, height)
}

// IsComplete returns true if the rowspan has been fully processed.
func (state *RowspanState) IsComplete() bool {
	return state.CurrentY >= state.Cell.EndY()-1
}

// SpansRegion returns true if this rowspan crosses the given row boundary.
func (state *RowspanState) SpansRegion(breakAtY int) bool {
	return state.Cell.Y < breakAtY && breakAtY < state.Cell.EndY()
}

// Unbreakable contains cells that must stay together with a spanning cell.
// This is used for orphan prevention - we don't want a single row
// of a spanning cell to appear alone at the top of a page.
type Unbreakable struct {
	// Cells that must be kept together.
	Cells []*Cell

	// MinRows is the minimum number of rows that must appear together.
	MinRows int
}

// NewUnbreakable creates an unbreakable group from a spanning cell.
func NewUnbreakable(cell *Cell, minRows int) *Unbreakable {
	return &Unbreakable{
		Cells:   []*Cell{cell},
		MinRows: minRows,
	}
}

// Add adds a cell to the unbreakable group.
func (u *Unbreakable) Add(cell *Cell) {
	u.Cells = append(u.Cells, cell)
}

// StartY returns the first row of the unbreakable group.
func (u *Unbreakable) StartY() int {
	if len(u.Cells) == 0 {
		return 0
	}
	minY := u.Cells[0].Y
	for _, cell := range u.Cells[1:] {
		if cell.Y < minY {
			minY = cell.Y
		}
	}
	return minY
}

// EndY returns the row after the last row of the unbreakable group.
func (u *Unbreakable) EndY() int {
	if len(u.Cells) == 0 {
		return 0
	}
	maxY := u.Cells[0].EndY()
	for _, cell := range u.Cells[1:] {
		if cell.EndY() > maxY {
			maxY = cell.EndY()
		}
	}
	return maxY
}

// Contains returns true if the group contains the given row.
func (u *Unbreakable) Contains(y int) bool {
	return y >= u.StartY() && y < u.EndY()
}

// RowspanLayout handles laying out content for a spanning cell.
type RowspanLayout struct {
	// Cell is the spanning cell.
	Cell *Cell

	// Width is the resolved width for the cell.
	Width layout.Abs

	// Heights is the height allocated in each row.
	Heights []layout.Abs

	// Frames contains the laid out content for each region.
	Frames []*layout.Frame
}

// NewRowspanLayout creates a new rowspan layout helper.
func NewRowspanLayout(cell *Cell, width layout.Abs) *RowspanLayout {
	return &RowspanLayout{
		Cell:    cell,
		Width:   width,
		Heights: make([]layout.Abs, cell.Rowspan),
		Frames:  nil,
	}
}

// TotalHeight returns the total height across all rows.
func (r *RowspanLayout) TotalHeight() layout.Abs {
	var total layout.Abs
	for _, h := range r.Heights {
		total += h
	}
	return total
}

// SetRowHeight sets the height for a specific row index.
func (r *RowspanLayout) SetRowHeight(index int, height layout.Abs) {
	if index < len(r.Heights) {
		r.Heights[index] = height
	}
}

// LayoutContent lays out the cell content with the given total height.
func (r *RowspanLayout) LayoutContent(totalHeight layout.Abs) *layout.Frame {
	frame := layout.NewFrame(layout.Size{
		Width:  r.Width,
		Height: totalHeight,
	})

	// Placeholder for actual content layout
	// This would lay out r.Cell.Content into the frame

	return frame
}

// SplitAtHeight splits the content at the given height for region breaks.
func (r *RowspanLayout) SplitAtHeight(height layout.Abs) (*layout.Frame, *layout.Frame) {
	// Split the content into two frames at the given height
	// First frame: content from 0 to height
	// Second frame: content from height to end

	firstFrame := layout.NewFrame(layout.Size{
		Width:  r.Width,
		Height: height,
	})

	remainingHeight := r.TotalHeight() - height
	if remainingHeight < 0 {
		remainingHeight = 0
	}

	secondFrame := layout.NewFrame(layout.Size{
		Width:  r.Width,
		Height: remainingHeight,
	})

	// Placeholder for actual content splitting
	// This would clip/split the content appropriately

	return firstFrame, secondFrame
}
