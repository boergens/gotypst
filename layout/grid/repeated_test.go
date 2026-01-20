package grid

import (
	"testing"

	"github.com/boergens/gotypst/layout"
	"github.com/boergens/gotypst/layout/flow"
)

func TestHeaderRowCount(t *testing.T) {
	h := Header{StartRow: 0, EndRow: 3, Level: 0, Repeat: true}
	if h.RowCount() != 3 {
		t.Errorf("RowCount() = %d, want 3", h.RowCount())
	}
}

func TestFooterRowCount(t *testing.T) {
	f := Footer{StartRow: 10, EndRow: 12, Repeat: true}
	if f.RowCount() != 2 {
		t.Errorf("RowCount() = %d, want 2", f.RowCount())
	}
}

func TestRepeatingHeaderTotalHeight(t *testing.T) {
	rh := RepeatingHeader{
		Header:  &Header{StartRow: 0, EndRow: 2},
		Heights: []layout.Abs{20, 30},
	}
	if rh.TotalHeight() != 50 {
		t.Errorf("TotalHeight() = %v, want 50", rh.TotalHeight())
	}
}

func TestPendingHeaderTotalHeight(t *testing.T) {
	ph := PendingHeader{
		Header:  &Header{StartRow: 0, EndRow: 2},
		Heights: []layout.Abs{15, 25, 10},
	}
	if ph.TotalHeight() != 50 {
		t.Errorf("TotalHeight() = %v, want 50", ph.TotalHeight())
	}
}

func TestPendingHeaderPromoteToRepeating(t *testing.T) {
	header := &Header{StartRow: 0, EndRow: 2, Level: 0, Repeat: true}
	frames := []flow.Frame{
		flow.NewFrame(layout.Size{Width: 100, Height: 20}),
		flow.NewFrame(layout.Size{Width: 100, Height: 30}),
	}
	heights := []layout.Abs{20, 30}

	ph := PendingHeader{
		Header:  header,
		Frames:  frames,
		Heights: heights,
		StartY:  0,
	}

	rh := ph.PromoteToRepeating()

	if rh.Header != header {
		t.Error("Header not preserved after promotion")
	}
	if len(rh.Frames) != 2 {
		t.Errorf("Frames length = %d, want 2", len(rh.Frames))
	}
	if len(rh.Heights) != 2 {
		t.Errorf("Heights length = %d, want 2", len(rh.Heights))
	}
	if rh.TotalHeight() != 50 {
		t.Errorf("TotalHeight() = %v, want 50", rh.TotalHeight())
	}
}

func TestHeaderManagerNewEmpty(t *testing.T) {
	m := NewHeaderManager()

	if len(m.RepeatingHeaders) != 0 {
		t.Error("New manager should have no repeating headers")
	}
	if len(m.PendingHeaders) != 0 {
		t.Error("New manager should have no pending headers")
	}
	if m.Footer != nil {
		t.Error("New manager should have no footer")
	}
}

func TestHeaderManagerRepeatingHeaderHeight(t *testing.T) {
	m := NewHeaderManager()

	m.RepeatingHeaders = []*RepeatingHeader{
		{Heights: []layout.Abs{20, 30}},
		{Heights: []layout.Abs{15}},
	}

	if m.RepeatingHeaderHeight() != 65 {
		t.Errorf("RepeatingHeaderHeight() = %v, want 65", m.RepeatingHeaderHeight())
	}
}

func TestHeaderManagerAddPendingHeader(t *testing.T) {
	m := NewHeaderManager()
	header := &Header{StartRow: 0, EndRow: 2, Level: 0, Repeat: true}
	frames := []flow.Frame{flow.NewFrame(layout.Size{Width: 100, Height: 20})}
	heights := []layout.Abs{20}

	m.AddPendingHeader(header, frames, heights, 0)

	if len(m.PendingHeaders) != 1 {
		t.Fatalf("PendingHeaders length = %d, want 1", len(m.PendingHeaders))
	}
	if m.PendingHeaders[0].Header != header {
		t.Error("Wrong header in pending")
	}
	if m.PendingHeaderHeight() != 20 {
		t.Errorf("PendingHeaderHeight() = %v, want 20", m.PendingHeaderHeight())
	}
}

func TestHeaderManagerPromoteAllPendingHeaders(t *testing.T) {
	m := NewHeaderManager()

	// Add two pending headers, one with Repeat=true, one with Repeat=false
	h1 := &Header{StartRow: 0, EndRow: 1, Level: 0, Repeat: true}
	h2 := &Header{StartRow: 1, EndRow: 2, Level: 1, Repeat: false}

	m.AddPendingHeader(h1, []flow.Frame{}, []layout.Abs{20}, 0)
	m.AddPendingHeader(h2, []flow.Frame{}, []layout.Abs{15}, 20)

	m.PromoteAllPendingHeaders()

	// Only h1 should become repeating (Repeat=true)
	if len(m.RepeatingHeaders) != 1 {
		t.Fatalf("RepeatingHeaders length = %d, want 1", len(m.RepeatingHeaders))
	}
	if m.RepeatingHeaders[0].Header != h1 {
		t.Error("Wrong header promoted")
	}
	if len(m.PendingHeaders) != 0 {
		t.Error("PendingHeaders should be empty after promotion")
	}
}

func TestHeaderManagerClearPendingHeaders(t *testing.T) {
	m := NewHeaderManager()
	h := &Header{StartRow: 0, EndRow: 1, Level: 0, Repeat: true}
	m.AddPendingHeader(h, []flow.Frame{}, []layout.Abs{20}, 0)

	m.ClearPendingHeaders()

	if len(m.PendingHeaders) != 0 {
		t.Error("PendingHeaders should be empty after clear")
	}
}

func TestHeaderManagerSetFooter(t *testing.T) {
	m := NewHeaderManager()
	footer := &Footer{StartRow: 10, EndRow: 12, Repeat: true}
	frames := []flow.Frame{
		flow.NewFrame(layout.Size{Width: 100, Height: 15}),
		flow.NewFrame(layout.Size{Width: 100, Height: 15}),
	}
	heights := []layout.Abs{15, 15}

	m.SetFooter(footer, frames, heights)

	if m.Footer != footer {
		t.Error("Footer not set correctly")
	}
	if m.FooterHeight() != 30 {
		t.Errorf("FooterHeight() = %v, want 30", m.FooterHeight())
	}
}

func TestHeaderManagerShouldRepeatFooter(t *testing.T) {
	m := NewHeaderManager()

	// No footer
	if m.ShouldRepeatFooter(false) {
		t.Error("Should not repeat when no footer")
	}
	if m.ShouldRepeatFooter(true) {
		t.Error("Should not repeat when no footer (final)")
	}

	// Non-repeating footer
	m.Footer = &Footer{Repeat: false}
	if m.ShouldRepeatFooter(false) {
		t.Error("Non-repeating footer should not appear on non-final page")
	}
	if !m.ShouldRepeatFooter(true) {
		t.Error("Non-repeating footer should appear on final page")
	}

	// Repeating footer
	m.Footer = &Footer{Repeat: true}
	if !m.ShouldRepeatFooter(false) {
		t.Error("Repeating footer should appear on non-final page")
	}
	if !m.ShouldRepeatFooter(true) {
		t.Error("Repeating footer should appear on final page")
	}
}

func TestHeaderManagerCheckOrphanHeaders(t *testing.T) {
	m := NewHeaderManager()

	// No pending headers - no orphans
	if m.CheckOrphanHeaders(false) {
		t.Error("No orphans when no pending headers")
	}

	// Pending headers with content - no orphans
	h := &Header{StartRow: 0, EndRow: 1}
	m.AddPendingHeader(h, []flow.Frame{}, []layout.Abs{20}, 0)
	if m.CheckOrphanHeaders(true) {
		t.Error("No orphans when content rows exist")
	}

	// Pending headers without content - orphans!
	if !m.CheckOrphanHeaders(false) {
		t.Error("Should detect orphan headers")
	}
}

func TestHeaderManagerHandleConflictingHeaders(t *testing.T) {
	m := NewHeaderManager()

	// Add headers at different levels
	m.RepeatingHeaders = []*RepeatingHeader{
		{Header: &Header{Level: 0}},
		{Header: &Header{Level: 1}},
		{Header: &Header{Level: 2}},
	}

	// New header at level 1 should remove levels 1 and 2
	m.HandleConflictingHeaders(1)

	if len(m.RepeatingHeaders) != 1 {
		t.Fatalf("RepeatingHeaders length = %d, want 1", len(m.RepeatingHeaders))
	}
	if m.RepeatingHeaders[0].Header.Level != 0 {
		t.Error("Wrong header kept after conflict resolution")
	}
}

func TestHeaderManagerCloneForRegion(t *testing.T) {
	m := NewHeaderManager()
	m.RepeatingHeaders = []*RepeatingHeader{
		{Header: &Header{Level: 0}, Heights: []layout.Abs{20}},
	}
	m.AddPendingHeader(&Header{Level: 1}, []flow.Frame{}, []layout.Abs{15}, 0)
	m.Footer = &Footer{Repeat: true}
	m.FooterHeights = []layout.Abs{10}

	clone := m.CloneForRegion()

	// Should have repeating headers
	if len(clone.RepeatingHeaders) != 1 {
		t.Errorf("Clone RepeatingHeaders length = %d, want 1", len(clone.RepeatingHeaders))
	}

	// Should NOT have pending headers (they're region-specific)
	if len(clone.PendingHeaders) != 0 {
		t.Errorf("Clone PendingHeaders length = %d, want 0", len(clone.PendingHeaders))
	}

	// Should have footer
	if clone.Footer == nil {
		t.Error("Clone should have footer")
	}
}

func TestHeaderManagerPrepareHeadersForNewRegion(t *testing.T) {
	m := NewHeaderManager()

	// No headers
	frames, heights := m.PrepareHeadersForNewRegion()
	if frames != nil || heights != nil {
		t.Error("Should return nil when no repeating headers")
	}

	// With repeating headers
	m.RepeatingHeaders = []*RepeatingHeader{
		{
			Frames:  []flow.Frame{flow.NewFrame(layout.Size{Width: 100, Height: 20})},
			Heights: []layout.Abs{20},
		},
		{
			Frames:  []flow.Frame{flow.NewFrame(layout.Size{Width: 100, Height: 15})},
			Heights: []layout.Abs{15},
		},
	}

	frames, heights = m.PrepareHeadersForNewRegion()
	if len(frames) != 2 {
		t.Errorf("Frames length = %d, want 2", len(frames))
	}
	if len(heights) != 2 {
		t.Errorf("Heights length = %d, want 2", len(heights))
	}
}

func TestCurrentAvailableHeight(t *testing.T) {
	c := Current{
		RepeatingHeaderHeight: 30,
		FooterHeight:          20,
		UsedHeight:            50,
	}

	regionHeight := layout.Abs(200)
	available := c.AvailableHeight(regionHeight)

	// 200 - 30 - 20 - 50 = 100
	if available != 100 {
		t.Errorf("AvailableHeight() = %v, want 100", available)
	}
}
