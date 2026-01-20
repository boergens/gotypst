package grid

import (
	"github.com/boergens/gotypst/layout"
	"github.com/boergens/gotypst/layout/flow"
)

// Header represents a table header that may repeat across page breaks.
type Header struct {
	// StartRow is the first row of the header (0-based).
	StartRow int
	// EndRow is one past the last row of the header.
	EndRow int
	// Level indicates header nesting level (0 is outermost).
	Level int
	// Repeat indicates if this header should repeat on subsequent pages.
	Repeat bool
}

// RowCount returns the number of rows in this header.
func (h *Header) RowCount() int {
	return h.EndRow - h.StartRow
}

// Footer represents a table footer that may repeat across page breaks.
type Footer struct {
	// StartRow is the first row of the footer (0-based).
	StartRow int
	// EndRow is one past the last row of the footer.
	EndRow int
	// Repeat indicates if this footer should repeat on pages before the final one.
	Repeat bool
}

// RowCount returns the number of rows in this footer.
func (f *Footer) RowCount() int {
	return f.EndRow - f.StartRow
}

// RepeatingHeader represents a header that is being repeated across regions.
type RepeatingHeader struct {
	// Header is the original header definition.
	Header *Header
	// Frames contains the laid out frames for each row.
	Frames []flow.Frame
	// Heights contains the height of each row.
	Heights []layout.Abs
}

// TotalHeight returns the total height of all header rows.
func (r *RepeatingHeader) TotalHeight() layout.Abs {
	var total layout.Abs
	for _, h := range r.Heights {
		total += h
	}
	return total
}

// PendingHeader represents a header that hasn't been confirmed as repeating yet.
// A header becomes repeating only after at least one content row is placed after it.
type PendingHeader struct {
	// Header is the header definition.
	Header *Header
	// Frames contains the laid out frames for each row.
	Frames []flow.Frame
	// Heights contains the height of each row.
	Heights []layout.Abs
	// StartY is the Y position where this header started in the region.
	StartY layout.Abs
}

// TotalHeight returns the total height of all pending header rows.
func (p *PendingHeader) TotalHeight() layout.Abs {
	var total layout.Abs
	for _, h := range p.Heights {
		total += h
	}
	return total
}

// PromoteToRepeating converts this pending header to a repeating header.
func (p *PendingHeader) PromoteToRepeating() *RepeatingHeader {
	return &RepeatingHeader{
		Header:  p.Header,
		Frames:  p.Frames,
		Heights: p.Heights,
	}
}

// FinishedHeaderRowInfo tracks information about finished header rows
// for use in orphan detection.
type FinishedHeaderRowInfo struct {
	// HeaderLevel is the nesting level of the header.
	HeaderLevel int
	// EndY is the Y position after this header row in the region.
	EndY layout.Abs
}

// HeaderManager manages headers and footers during grid layout.
type HeaderManager struct {
	// RepeatingHeaders are headers confirmed to repeat across pages.
	RepeatingHeaders []*RepeatingHeader
	// PendingHeaders are headers awaiting confirmation (need content row after).
	PendingHeaders []*PendingHeader
	// Footer is the grid's footer if any.
	Footer *Footer
	// FooterFrames contains laid out footer frames.
	FooterFrames []flow.Frame
	// FooterHeights contains the height of each footer row.
	FooterHeights []layout.Abs
	// FinishedHeaderRows tracks laid out header rows for orphan detection.
	FinishedHeaderRows []FinishedHeaderRowInfo
}

// NewHeaderManager creates a new header manager.
func NewHeaderManager() *HeaderManager {
	return &HeaderManager{}
}

// RepeatingHeaderHeight returns the total height of all repeating headers.
func (m *HeaderManager) RepeatingHeaderHeight() layout.Abs {
	var total layout.Abs
	for _, h := range m.RepeatingHeaders {
		total += h.TotalHeight()
	}
	return total
}

// PendingHeaderHeight returns the total height of all pending headers.
func (m *HeaderManager) PendingHeaderHeight() layout.Abs {
	var total layout.Abs
	for _, h := range m.PendingHeaders {
		total += h.TotalHeight()
	}
	return total
}

// FooterHeight returns the total height of the footer.
func (m *HeaderManager) FooterHeight() layout.Abs {
	var total layout.Abs
	for _, h := range m.FooterHeights {
		total += h
	}
	return total
}

// AddPendingHeader adds a new pending header.
func (m *HeaderManager) AddPendingHeader(header *Header, frames []flow.Frame, heights []layout.Abs, startY layout.Abs) {
	m.PendingHeaders = append(m.PendingHeaders, &PendingHeader{
		Header:  header,
		Frames:  frames,
		Heights: heights,
		StartY:  startY,
	})
}

// PromoteAllPendingHeaders promotes all pending headers to repeating.
// Called when a content row is successfully placed after pending headers.
func (m *HeaderManager) PromoteAllPendingHeaders() {
	for _, pending := range m.PendingHeaders {
		if pending.Header.Repeat {
			m.RepeatingHeaders = append(m.RepeatingHeaders, pending.PromoteToRepeating())
		}
	}
	m.PendingHeaders = nil
}

// ClearPendingHeaders removes all pending headers without promoting them.
// Used for orphan prevention when headers are alone in a region.
func (m *HeaderManager) ClearPendingHeaders() {
	m.PendingHeaders = nil
}

// SetFooter sets the footer definition and its laid out frames.
func (m *HeaderManager) SetFooter(footer *Footer, frames []flow.Frame, heights []layout.Abs) {
	m.Footer = footer
	m.FooterFrames = frames
	m.FooterHeights = heights
}

// ShouldRepeatFooter returns true if the footer should appear on this region.
// The footer repeats on all pages if Footer.Repeat is true, or only on the
// final page if Repeat is false.
func (m *HeaderManager) ShouldRepeatFooter(isFinalRegion bool) bool {
	if m.Footer == nil {
		return false
	}
	return m.Footer.Repeat || isFinalRegion
}

// PrepareHeadersForNewRegion prepares headers for placement in a new region.
// Returns the frames and heights to place at the start of the region.
func (m *HeaderManager) PrepareHeadersForNewRegion() ([]flow.Frame, []layout.Abs) {
	if len(m.RepeatingHeaders) == 0 {
		return nil, nil
	}

	// Collect all repeating header frames and heights
	var frames []flow.Frame
	var heights []layout.Abs
	for _, h := range m.RepeatingHeaders {
		frames = append(frames, h.Frames...)
		heights = append(heights, h.Heights...)
	}
	return frames, heights
}

// CheckOrphanHeaders checks if only headers are present in the region without
// any content rows. If so, returns true and the headers should be removed.
func (m *HeaderManager) CheckOrphanHeaders(hasContentRows bool) bool {
	// Headers are orphaned if we have pending headers but no content rows
	if len(m.PendingHeaders) > 0 && !hasContentRows {
		return true
	}
	return false
}

// RecordFinishedHeaderRow records a header row that has been placed for orphan tracking.
func (m *HeaderManager) RecordFinishedHeaderRow(level int, endY layout.Abs) {
	m.FinishedHeaderRows = append(m.FinishedHeaderRows, FinishedHeaderRowInfo{
		HeaderLevel: level,
		EndY:        endY,
	})
}

// ClearFinishedHeaderRows clears the finished header row tracking.
func (m *HeaderManager) ClearFinishedHeaderRows() {
	m.FinishedHeaderRows = nil
}

// HandleConflictingHeaders handles the case where a new header at a given level
// conflicts with existing repeating headers at the same or deeper level.
// Headers at conflicting levels are removed from the repeating set.
func (m *HeaderManager) HandleConflictingHeaders(newHeaderLevel int) {
	// Remove repeating headers at the same or deeper level
	var kept []*RepeatingHeader
	for _, h := range m.RepeatingHeaders {
		if h.Header.Level < newHeaderLevel {
			kept = append(kept, h)
		}
	}
	m.RepeatingHeaders = kept
}

// CloneForRegion creates a copy of the header manager state for a new region.
// This is used when starting layout of a new region.
func (m *HeaderManager) CloneForRegion() *HeaderManager {
	clone := &HeaderManager{
		Footer:        m.Footer,
		FooterFrames:  m.FooterFrames,
		FooterHeights: m.FooterHeights,
	}

	// Copy repeating headers (they carry over to new regions)
	clone.RepeatingHeaders = make([]*RepeatingHeader, len(m.RepeatingHeaders))
	copy(clone.RepeatingHeaders, m.RepeatingHeaders)

	// Pending headers don't carry over - they're region-specific
	return clone
}

// PlaceRepeatingHeaders places all repeating headers into a frame at the start
// of a new region.
func (m *HeaderManager) PlaceRepeatingHeaders(output *flow.Frame, startY layout.Abs, width layout.Abs) layout.Abs {
	y := startY
	for _, header := range m.RepeatingHeaders {
		for i, frame := range header.Frames {
			output.PushFrame(layout.Point{X: 0, Y: y}, frame)
			y += header.Heights[i]
		}
	}
	return y
}

// PlaceFooter places the footer into a frame at the end of a region.
func (m *HeaderManager) PlaceFooter(output *flow.Frame, y layout.Abs, width layout.Abs) {
	if m.Footer == nil || len(m.FooterFrames) == 0 {
		return
	}
	for i, frame := range m.FooterFrames {
		output.PushFrame(layout.Point{X: 0, Y: y}, frame)
		y += m.FooterHeights[i]
	}
}
