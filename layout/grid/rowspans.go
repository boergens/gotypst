package grid

import (
	"github.com/boergens/gotypst/layout"
)

// RowspanTracker manages rowspan cells during grid layout.
type RowspanTracker struct {
	// Rowspans holds all active rowspans.
	Rowspans []Rowspan
	// ByPosition allows lookup by starting cell position.
	ByPosition map[Axes[int]]*Rowspan
}

// NewRowspanTracker creates a new rowspan tracker.
func NewRowspanTracker() *RowspanTracker {
	return &RowspanTracker{
		ByPosition: make(map[Axes[int]]*Rowspan),
	}
}

// Register adds a new rowspan to track.
func (rt *RowspanTracker) Register(rs Rowspan) {
	rt.Rowspans = append(rt.Rowspans, rs)
	rt.ByPosition[Axes[int]{X: rs.X, Y: rs.Y}] = &rt.Rowspans[len(rt.Rowspans)-1]
}

// Get returns the rowspan at the given position, or nil if none.
func (rt *RowspanTracker) Get(x, y int) *Rowspan {
	return rt.ByPosition[Axes[int]{X: x, Y: y}]
}

// ActiveAt returns all rowspans that are active at the given row.
func (rt *RowspanTracker) ActiveAt(y int) []*Rowspan {
	var active []*Rowspan
	for i := range rt.Rowspans {
		rs := &rt.Rowspans[i]
		if rs.Y <= y && rs.Y+rs.RowspanCount > y {
			active = append(active, rs)
		}
	}
	return active
}

// UpdateResolvedRow updates the max resolved row for rowspans ending after the given row.
func (rt *RowspanTracker) UpdateResolvedRow(y int) {
	for i := range rt.Rowspans {
		rs := &rt.Rowspans[i]
		endY := rs.Y + rs.RowspanCount - 1
		if endY > y {
			if rs.MaxResolvedRow == nil || *rs.MaxResolvedRow < y {
				row := y
				rs.MaxResolvedRow = &row
			}
		}
	}
}

// AccumulateHeight adds height to rowspans that span the given row.
func (rt *RowspanTracker) AccumulateHeight(y int, height layout.Abs, regionIdx int) {
	for i := range rt.Rowspans {
		rs := &rt.Rowspans[i]
		if rs.Y <= y && rs.Y+rs.RowspanCount > y {
			// Ensure we have enough height entries.
			for len(rs.Heights) <= regionIdx-rs.FirstRegion {
				rs.Heights = append(rs.Heights, 0)
			}
			idx := regionIdx - rs.FirstRegion
			if idx >= 0 && idx < len(rs.Heights) {
				rs.Heights[idx] += height
			}
		}
	}
}

// RemoveGutterFromEnd removes gutter height from rowspans ending at the given row.
func (rt *RowspanTracker) RemoveGutterFromEnd(y int, gutterHeight layout.Abs, regionIdx int) {
	for i := range rt.Rowspans {
		rs := &rt.Rowspans[i]
		endY := rs.Y + rs.RowspanCount - 1
		if endY == y {
			idx := regionIdx - rs.FirstRegion
			if idx >= 0 && idx < len(rs.Heights) {
				rs.Heights[idx] -= gutterHeight
				if rs.Heights[idx] < 0 {
					rs.Heights[idx] = 0
				}
			}
		}
	}
}

// CompletedAt returns rowspans that complete at or before the given row.
func (rt *RowspanTracker) CompletedAt(y int) []*Rowspan {
	var completed []*Rowspan
	for i := range rt.Rowspans {
		rs := &rt.Rowspans[i]
		if rs.Y+rs.RowspanCount <= y+1 {
			completed = append(completed, rs)
		}
	}
	return completed
}

// RemoveCompleted removes rowspans that have been completed.
func (rt *RowspanTracker) RemoveCompleted(y int) {
	var remaining []Rowspan
	for _, rs := range rt.Rowspans {
		if rs.Y+rs.RowspanCount > y+1 {
			remaining = append(remaining, rs)
		} else {
			delete(rt.ByPosition, Axes[int]{X: rs.X, Y: rs.Y})
		}
	}
	rt.Rowspans = remaining
}

// TotalHeight returns the total height accumulated for a rowspan across all regions.
func (rs *Rowspan) TotalHeight() layout.Abs {
	var total layout.Abs
	for _, h := range rs.Heights {
		total += h
	}
	return total
}

// AvailableHeight returns the available height in the given region for this rowspan.
func (rs *Rowspan) AvailableHeight(regionIdx int) layout.Abs {
	idx := regionIdx - rs.FirstRegion
	if idx < 0 || idx >= len(rs.Heights) {
		return 0
	}
	return rs.Heights[idx]
}

// SpansIntoRegion checks if this rowspan extends into the given region.
func (rs *Rowspan) SpansIntoRegion(regionIdx int) bool {
	return regionIdx >= rs.FirstRegion
}

// IsComplete checks if the rowspan has been fully laid out to the given row.
func (rs *Rowspan) IsComplete(currentRow int) bool {
	return currentRow >= rs.Y+rs.RowspanCount-1
}

// RemainingRows returns the number of rows remaining for this rowspan.
func (rs *Rowspan) RemainingRows(currentRow int) int {
	endRow := rs.Y + rs.RowspanCount - 1
	if currentRow >= endRow {
		return 0
	}
	return endRow - currentRow
}

// CellPosition returns the cell position for this rowspan.
func (rs *Rowspan) CellPosition() Axes[int] {
	return Axes[int]{X: rs.X, Y: rs.Y}
}

// CalculateBacklogHeights calculates heights for rowspan continuation across regions.
func (rs *Rowspan) CalculateBacklogHeights(regionHeights []layout.Abs) {
	// Clear existing heights beyond the first region.
	if len(rs.Heights) > 1 {
		rs.Heights = rs.Heights[:1]
	}

	// Add heights from each region in the backlog.
	for _, h := range regionHeights {
		rs.Heights = append(rs.Heights, h)
	}
}
