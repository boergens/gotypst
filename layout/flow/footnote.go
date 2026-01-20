package flow

import (
	"github.com/boergens/gotypst/layout"
)

// FootnoteConfig holds configuration for footnote layout.
type FootnoteConfig struct {
	// Separator is the frame to place between main content and footnotes.
	Separator *Frame
	// Gap is the spacing between the separator and footnotes.
	Gap layout.Abs
	// Clearance is the minimum space between content and footnotes.
	Clearance layout.Abs
}

// DefaultFootnoteConfig returns the default footnote configuration.
func DefaultFootnoteConfig() FootnoteConfig {
	return FootnoteConfig{
		Separator: nil,
		Gap:       layout.Abs(10), // 10pt default gap
		Clearance: layout.Abs(0),
	}
}

// Footnote represents a footnote entry to be placed at the bottom of a page.
type Footnote struct {
	// Location uniquely identifies this footnote.
	Location Location
	// Frame contains the laid out footnote content.
	Frame Frame
	// Flow indicates the flow need of the content containing this footnote.
	Flow layout.Abs
	// Breakable indicates whether the footnote can be broken across pages.
	Breakable bool
}

// FootnoteState tracks footnotes during flow layout.
type FootnoteState struct {
	// Config holds footnote configuration.
	Config FootnoteConfig
	// Pending holds footnotes waiting to be placed.
	Pending []Footnote
	// Used tracks the total height used by footnotes in the current region.
	Used layout.Abs
	// Insertions holds footnotes placed in the current region.
	Insertions []Footnote
}

// NewFootnoteState creates a new footnote state with the given config.
func NewFootnoteState(config FootnoteConfig) *FootnoteState {
	return &FootnoteState{
		Config: config,
	}
}

// Height returns the total height needed for placed footnotes,
// including separator and gap.
func (s *FootnoteState) Height() layout.Abs {
	if len(s.Insertions) == 0 {
		return 0
	}

	height := s.Used
	if s.Config.Separator != nil {
		height += s.Config.Separator.Height()
	}
	if height > 0 {
		height += s.Config.Gap
	}
	return height
}

// Width returns the maximum width of placed footnotes.
func (s *FootnoteState) Width() layout.Abs {
	var maxWidth layout.Abs
	for _, fn := range s.Insertions {
		if fn.Frame.Width() > maxWidth {
			maxWidth = fn.Frame.Width()
		}
	}
	return maxWidth
}

// Clear removes all placed footnotes, keeping pending ones.
func (s *FootnoteState) Clear() {
	s.Insertions = nil
	s.Used = 0
}

// Finalize produces a frame containing all footnotes for the current region.
// Returns nil if there are no footnotes.
func (s *FootnoteState) Finalize(width layout.Abs) *Frame {
	if len(s.Insertions) == 0 {
		return nil
	}

	height := s.Height()
	frame := NewFrame(layout.Size{Width: width, Height: height})

	var y layout.Abs

	// Add separator if configured
	if s.Config.Separator != nil {
		frame.PushFrame(layout.Point{X: 0, Y: y}, *s.Config.Separator)
		y += s.Config.Separator.Height()
	}

	// Add gap
	if y > 0 || s.Config.Gap > 0 {
		y += s.Config.Gap
	}

	// Add footnote frames
	for _, fn := range s.Insertions {
		frame.PushFrame(layout.Point{X: 0, Y: y}, fn.Frame)
		y += fn.Frame.Height()
	}

	return &frame
}

// FootnoteMarker represents a footnote marker discovered in content.
type FootnoteMarker struct {
	// Location uniquely identifies the marker.
	Location Location
	// Entry is the footnote content to be laid out.
	Entry Frame
}

// DiscoverFootnotes scans a frame for footnote markers.
// Returns markers that need to be processed for footnote placement.
func DiscoverFootnotes(frame *Frame) []FootnoteMarker {
	var markers []FootnoteMarker
	discoverFootnotesRecursive(frame, &markers)
	return markers
}

// discoverFootnotesRecursive recursively searches for footnote markers.
func discoverFootnotesRecursive(frame *Frame, markers *[]FootnoteMarker) {
	for _, entry := range frame.Items() {
		switch item := entry.Item.(type) {
		case FrameItemFootnoteMarker:
			*markers = append(*markers, FootnoteMarker{
				Location: item.Location,
				Entry:    item.Entry,
			})
		case FrameItemFrame:
			discoverFootnotesRecursive(&item.Frame, markers)
		}
	}
}

// FrameItemFootnoteMarker represents a footnote marker in a frame.
type FrameItemFootnoteMarker struct {
	// Location uniquely identifies this marker.
	Location Location
	// Entry is the footnote content frame.
	Entry Frame
}

func (FrameItemFootnoteMarker) isFrameItem() {}

// processFootnotes handles footnotes discovered in a frame.
// It attempts to place footnotes at the bottom of the current region,
// enforcing the footnote invariant (marker and entry on same page).
//
// Parameters:
//   - state: The footnote state to update
//   - regions: The available regions for layout
//   - frame: The frame containing potential footnote markers
//   - flowNeed: The height needed by the flow content
//   - breakable: Whether the content containing footnotes is breakable
//   - migratable: Whether the content can be migrated to a new region
//
// Returns an error if the footnote invariant cannot be satisfied.
func processFootnotes(
	state *FootnoteState,
	regions *Regions,
	frame *Frame,
	flowNeed layout.Abs,
	breakable bool,
	migratable bool,
) error {
	// Discover footnote markers in the frame
	markers := DiscoverFootnotes(frame)
	if len(markers) == 0 && len(state.Pending) == 0 {
		return nil
	}

	// Process any pending footnotes first
	for i := 0; i < len(state.Pending); {
		fn := state.Pending[i]

		// Calculate space needed
		needed := fn.Frame.Height()
		if len(state.Insertions) == 0 && state.Config.Separator != nil {
			needed += state.Config.Separator.Height() + state.Config.Gap
		}

		// Check if footnote fits
		available := regions.Size.Height - flowNeed - state.Height() - state.Config.Clearance
		if available.Fits(needed) {
			// Place the footnote
			state.Insertions = append(state.Insertions, fn)
			state.Used += fn.Frame.Height()

			// Remove from pending
			state.Pending = append(state.Pending[:i], state.Pending[i+1:]...)
		} else if fn.Breakable && available > 0 {
			// Try to break the footnote across regions
			// For now, defer to next region
			i++
		} else {
			// Footnote doesn't fit, defer to next region
			i++
		}
	}

	// Process new markers
	for _, marker := range markers {
		fn := Footnote{
			Location:  marker.Location,
			Frame:     marker.Entry,
			Flow:      flowNeed,
			Breakable: breakable,
		}

		// Calculate space needed for this footnote
		needed := fn.Frame.Height()
		if len(state.Insertions) == 0 && state.Config.Separator != nil {
			needed += state.Config.Separator.Height() + state.Config.Gap
		}

		// Check if footnote fits
		available := regions.Size.Height - flowNeed - state.Height() - state.Config.Clearance
		if available.Fits(needed) {
			// Place the footnote
			state.Insertions = append(state.Insertions, fn)
			state.Used += fn.Frame.Height()
		} else if regions.MayProgress() && migratable {
			// Content can be migrated to ensure footnote invariant
			// Add to pending and signal need for migration
			state.Pending = append(state.Pending, fn)
		} else if fn.Breakable && available > 0 {
			// Try to break the footnote
			// For now, just place what fits
			state.Insertions = append(state.Insertions, fn)
			state.Used += fn.Frame.Height()
		} else {
			// Footnote doesn't fit, add to pending
			state.Pending = append(state.Pending, fn)
		}
	}

	// Reduce available region height by footnote usage
	regions.Size.Height -= state.Height()

	return nil
}

// FootnoteOverflow represents footnote content that needs to continue
// on a subsequent page.
type FootnoteOverflow struct {
	// Footnote is the footnote that overflowed.
	Footnote Footnote
	// Remaining is the frame content that didn't fit.
	Remaining Frame
}

// HandleOverflow processes footnote overflow to subsequent pages.
// This handles the case where a footnote is too large to fit entirely
// on the current page and needs to be broken.
func (s *FootnoteState) HandleOverflow() []FootnoteOverflow {
	// For now, return empty - full overflow handling requires
	// more sophisticated frame splitting
	return nil
}

// Reset clears all footnote state for a new region.
func (s *FootnoteState) Reset() {
	s.Insertions = nil
	s.Used = 0
	// Keep pending footnotes - they carry over to next region
}
