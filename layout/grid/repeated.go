package grid

import (
	"github.com/boergens/gotypst/layout"
	"github.com/boergens/gotypst/layout/flow"
)

// HeaderManager manages repeating headers across regions.
type HeaderManager struct {
	// RepeatingHeaders holds headers that repeat on each page.
	RepeatingHeaders []Header
	// PendingHeaders holds headers waiting to become repeating.
	PendingHeaders []Header
	// FinishedRows tracks info about completed header rows.
	FinishedRows []FinishedHeaderRowInfo
	// TotalHeaderHeight is the combined height of all repeating headers.
	TotalHeaderHeight layout.Abs
}

// NewHeaderManager creates a new header manager.
func NewHeaderManager() *HeaderManager {
	return &HeaderManager{}
}

// AddPendingHeader registers a new pending header.
// Headers become "repeating" once content is placed after them.
func (hm *HeaderManager) AddPendingHeader(header Header) {
	hm.PendingHeaders = append(hm.PendingHeaders, header)
}

// PromotePendingHeaders promotes pending headers to repeating status.
// This is called when non-header content is placed after the headers.
func (hm *HeaderManager) PromotePendingHeaders() {
	hm.RepeatingHeaders = append(hm.RepeatingHeaders, hm.PendingHeaders...)
	hm.PendingHeaders = nil

	// Recalculate total header height.
	hm.recalculateHeaderHeight()
}

// RecordFinishedRow records a finished header row.
func (hm *HeaderManager) RecordFinishedRow(y int, height layout.Abs) {
	hm.FinishedRows = append(hm.FinishedRows, FinishedHeaderRowInfo{
		Y:      y,
		Height: height,
	})
}

// recalculateHeaderHeight recalculates the total header height.
func (hm *HeaderManager) recalculateHeaderHeight() {
	hm.TotalHeaderHeight = 0
	for _, h := range hm.RepeatingHeaders {
		if h.Frame != nil {
			hm.TotalHeaderHeight += h.Frame.Height()
		}
	}
}

// ClearPendingOnOrphan clears pending headers when they would be orphaned.
// This happens when headers are the only content in a region.
func (hm *HeaderManager) ClearPendingOnOrphan() {
	hm.PendingHeaders = nil
}

// HasRepeatingHeaders returns true if there are repeating headers.
func (hm *HeaderManager) HasRepeatingHeaders() bool {
	return len(hm.RepeatingHeaders) > 0
}

// HasPendingHeaders returns true if there are pending headers.
func (hm *HeaderManager) HasPendingHeaders() bool {
	return len(hm.PendingHeaders) > 0
}

// GetHeadersForRegion returns headers to layout at the start of a region.
func (hm *HeaderManager) GetHeadersForRegion(regionIdx int) []Header {
	if regionIdx == 0 {
		// First region: no repeating headers yet.
		return nil
	}
	return hm.RepeatingHeaders
}

// ResolveHeaderConflicts resolves conflicts between headers of different levels.
// Higher-level headers can override lower-level ones at the same position.
func (hm *HeaderManager) ResolveHeaderConflicts() {
	if len(hm.RepeatingHeaders) <= 1 {
		return
	}

	// Build a map of row coverage.
	coverage := make(map[int]int) // row -> max level

	for _, h := range hm.RepeatingHeaders {
		for y := h.StartY; y < h.EndY; y++ {
			if existing, ok := coverage[y]; !ok || h.Level > existing {
				coverage[y] = h.Level
			}
		}
	}

	// Filter headers to keep only the highest-level ones.
	var filtered []Header
	for _, h := range hm.RepeatingHeaders {
		keep := true
		for y := h.StartY; y < h.EndY; y++ {
			if coverage[y] > h.Level {
				keep = false
				break
			}
		}
		if keep {
			filtered = append(filtered, h)
		}
	}

	hm.RepeatingHeaders = filtered
}

// FooterManager manages footer sections.
type FooterManager struct {
	// Footer is the current footer (if any).
	Footer *Footer
	// FooterHeight is the height reserved for the footer.
	FooterHeight layout.Abs
}

// NewFooterManager creates a new footer manager.
func NewFooterManager() *FooterManager {
	return &FooterManager{}
}

// SetFooter sets the footer section.
func (fm *FooterManager) SetFooter(footer *Footer) {
	fm.Footer = footer
	if footer != nil && footer.Frame != nil {
		fm.FooterHeight = footer.Frame.Height()
	} else {
		fm.FooterHeight = 0
	}
}

// HasFooter returns true if there is a footer.
func (fm *FooterManager) HasFooter() bool {
	return fm.Footer != nil
}

// GetFooterHeight returns the height reserved for the footer.
func (fm *FooterManager) GetFooterHeight() layout.Abs {
	return fm.FooterHeight
}

// ShouldShowFooter determines if the footer should be shown in the given region.
// Footers typically only appear on the last region.
func (fm *FooterManager) ShouldShowFooter(isFinalRegion bool) bool {
	return fm.Footer != nil && isFinalRegion
}

// RepeatableSection represents a section that can repeat (header or footer).
type RepeatableSection struct {
	// Type indicates if this is a header or footer.
	Type RepeatableSectionType
	// StartY is the starting row.
	StartY int
	// EndY is the ending row (exclusive).
	EndY int
	// Level is the nesting level (for headers).
	Level int
	// Frame is the laid out content.
	Frame *flow.Frame
}

// RepeatableSectionType indicates the type of repeatable section.
type RepeatableSectionType int

const (
	RepeatableSectionHeader RepeatableSectionType = iota
	RepeatableSectionFooter
)

// HeaderRowRange returns the row range for a header.
func (h *Header) HeaderRowRange() (start, end int) {
	return h.StartY, h.EndY
}

// RowCount returns the number of rows in the header.
func (h *Header) RowCount() int {
	return h.EndY - h.StartY
}

// ContainsRow returns true if the header contains the given row.
func (h *Header) ContainsRow(y int) bool {
	return y >= h.StartY && y < h.EndY
}

// IsNestedWithin returns true if this header is nested within another.
func (h *Header) IsNestedWithin(other *Header) bool {
	return h.StartY >= other.StartY && h.EndY <= other.EndY
}

// OverlapsWith returns true if this header overlaps with another.
func (h *Header) OverlapsWith(other *Header) bool {
	return h.StartY < other.EndY && h.EndY > other.StartY
}

// OrphanDetector detects orphaned headers in a region.
type OrphanDetector struct {
	// HeaderHeight is the total height of headers.
	HeaderHeight layout.Abs
	// ContentHeight is the height of non-header content.
	ContentHeight layout.Abs
}

// NewOrphanDetector creates a new orphan detector.
func NewOrphanDetector() *OrphanDetector {
	return &OrphanDetector{}
}

// AddHeaderHeight adds header height.
func (od *OrphanDetector) AddHeaderHeight(height layout.Abs) {
	od.HeaderHeight += height
}

// AddContentHeight adds non-header content height.
func (od *OrphanDetector) AddContentHeight(height layout.Abs) {
	od.ContentHeight += height
}

// IsOrphaned returns true if the region contains only headers.
func (od *OrphanDetector) IsOrphaned() bool {
	return od.HeaderHeight > 0 && od.ContentHeight == 0
}

// Reset resets the detector for a new region.
func (od *OrphanDetector) Reset() {
	od.HeaderHeight = 0
	od.ContentHeight = 0
}
