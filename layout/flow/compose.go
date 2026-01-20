package flow

import (
	"github.com/boergens/gotypst/layout"
)

// Insertions accumulates out-of-flow elements (floats and footnotes) during
// frame composition. It tracks their positions and calculates the space they
// occupy in the output frame.
type Insertions struct {
	// topFloats holds floats aligned to the top of the region.
	topFloats []placedFrame
	// bottomFloats holds floats aligned to the bottom of the region.
	bottomFloats []placedFrame
	// footnotes holds footnote frames in document order.
	footnotes []Frame
	// footnoteSeparator is the optional separator line above footnotes.
	footnoteSeparator *Frame
	// topSize is the total height of top floats including clearances.
	topSize layout.Abs
	// bottomSize is the total height of bottom floats and footnotes.
	bottomSize layout.Abs
	// width is the maximum width of any insertion.
	width layout.Abs
	// skips tracks locations that should be skipped (already processed).
	skips []Location
}

// placedFrame pairs a placed child with its laid out frame.
type placedFrame struct {
	placed *PlacedChild
	frame  Frame
}

// NewInsertions creates an empty Insertions tracker.
func NewInsertions() *Insertions {
	return &Insertions{}
}

// PushFloat adds a float to the appropriate position (top or bottom).
func (ins *Insertions) PushFloat(placed *PlacedChild, frame Frame, alignY FixedAlignment) {
	pf := placedFrame{placed: placed, frame: frame}

	// Track maximum width.
	if frame.Width() > ins.width {
		ins.width = frame.Width()
	}

	// Add to top or bottom based on vertical alignment.
	switch alignY {
	case FixedAlignStart:
		ins.topFloats = append(ins.topFloats, pf)
		ins.topSize += frame.Height() + placed.Clearance
	case FixedAlignEnd:
		ins.bottomFloats = append(ins.bottomFloats, pf)
		ins.bottomSize += frame.Height() + placed.Clearance
	default:
		// Center-aligned floats go to top by default.
		ins.topFloats = append(ins.topFloats, pf)
		ins.topSize += frame.Height() + placed.Clearance
	}

	// Record location to skip on relayout.
	ins.skips = append(ins.skips, placed.Location())
}

// PushFootnote adds a footnote frame to the bottom area.
func (ins *Insertions) PushFootnote(frame Frame, gap layout.Abs) {
	if frame.Width() > ins.width {
		ins.width = frame.Width()
	}

	// Add gap between footnotes.
	if len(ins.footnotes) > 0 {
		ins.bottomSize += gap
	}

	ins.footnotes = append(ins.footnotes, frame)
	ins.bottomSize += frame.Height()
}

// PushFootnoteSeparator sets the separator frame above footnotes.
func (ins *Insertions) PushFootnoteSeparator(frame Frame, clearance layout.Abs) {
	if ins.footnoteSeparator != nil {
		return // Already have a separator.
	}

	if frame.Width() > ins.width {
		ins.width = frame.Width()
	}

	ins.footnoteSeparator = &frame
	ins.bottomSize += frame.Height() + clearance
}

// Height returns the total height occupied by insertions.
func (ins *Insertions) Height() layout.Abs {
	return ins.topSize + ins.bottomSize
}

// TopHeight returns the height of top floats.
func (ins *Insertions) TopHeight() layout.Abs {
	return ins.topSize
}

// BottomHeight returns the height of bottom floats and footnotes.
func (ins *Insertions) BottomHeight() layout.Abs {
	return ins.bottomSize
}

// Width returns the maximum width of any insertion.
func (ins *Insertions) Width() layout.Abs {
	return ins.width
}

// IsEmpty returns true if there are no insertions.
func (ins *Insertions) IsEmpty() bool {
	return len(ins.topFloats) == 0 &&
		len(ins.bottomFloats) == 0 &&
		len(ins.footnotes) == 0
}

// Skips returns the locations that should be skipped.
func (ins *Insertions) Skips() []Location {
	return ins.skips
}

// Finalize composes all insertions with a content frame into the output.
// The layout order from top to bottom is:
// 1. Top floats
// 2. Main content
// 3. Bottom floats
// 4. Footnote separator (if any)
// 5. Footnotes
func (ins *Insertions) Finalize(content Frame, regionSize layout.Size) Frame {
	if ins.IsEmpty() {
		return content
	}

	output := Soft(regionSize)
	var offset layout.Abs

	// Position top floats.
	for _, pf := range ins.topFloats {
		x := pf.placed.AlignX.Position(regionSize.Width - pf.frame.Width())
		delta := RelAxesToPoint(pf.placed.Delta, regionSize)
		pos := layout.Point{X: x + delta.X, Y: offset + delta.Y}
		output.PushFrame(pos, pf.frame)
		offset += pf.frame.Height() + pf.placed.Clearance
	}

	// Position main content after top floats.
	contentY := offset
	output.PushFrame(layout.Point{X: 0, Y: contentY}, content)
	offset += content.Height()

	// Calculate bottom section start position.
	// Bottom elements are positioned from the bottom of the region upward.
	bottomStart := regionSize.Height

	// Position footnotes (from bottom up).
	for i := len(ins.footnotes) - 1; i >= 0; i-- {
		fn := ins.footnotes[i]
		bottomStart -= fn.Height()
		output.PushFrame(layout.Point{X: 0, Y: bottomStart}, fn)
	}

	// Position footnote separator.
	if ins.footnoteSeparator != nil {
		bottomStart -= ins.footnoteSeparator.Height()
		output.PushFrame(layout.Point{X: 0, Y: bottomStart}, *ins.footnoteSeparator)
	}

	// Position bottom floats (from bottom up, above footnotes).
	for i := len(ins.bottomFloats) - 1; i >= 0; i-- {
		pf := ins.bottomFloats[i]
		bottomStart -= pf.frame.Height()
		x := pf.placed.AlignX.Position(regionSize.Width - pf.frame.Width())
		delta := RelAxesToPoint(pf.placed.Delta, regionSize)
		pos := layout.Point{X: x + delta.X, Y: bottomStart + delta.Y}
		output.PushFrame(pos, pf.frame)
		bottomStart -= pf.placed.Clearance
	}

	return output
}

// ComposerState holds composition state beyond what's in Composer.
type ComposerState struct {
	// Insertions tracks accumulated floats and footnotes.
	Insertions *Insertions
	// FootnoteGap is the spacing between footnotes.
	FootnoteGap layout.Abs
	// FootnoteClearance is the clearance above the footnote separator.
	FootnoteClearance layout.Abs
}

// NewComposerState creates a new composer state.
func NewComposerState() *ComposerState {
	return &ComposerState{
		Insertions:        NewInsertions(),
		FootnoteGap:       layout.Abs(5.0),  // Default gap between footnotes.
		FootnoteClearance: layout.Abs(10.0), // Default clearance above separator.
	}
}

// ComposerExt extends Composer with composition state.
type ComposerExt struct {
	*Composer
	State *ComposerState
}

// NewComposerExt creates an extended composer wrapping the base composer.
func NewComposerExt(c *Composer) *ComposerExt {
	return &ComposerExt{
		Composer: c,
		State:    NewComposerState(),
	}
}

// floatFits checks if a float with the given frame and clearance fits.
func (ext *ComposerExt) floatFits(frame Frame, clearance layout.Abs, regions *Regions) bool {
	needed := frame.Height() + clearance
	available := regions.Size.Height - ext.State.Insertions.Height()
	return available.Fits(needed)
}

// handleFloat processes a single float, either placing it or queuing it.
func (ext *ComposerExt) handleFloat(
	placed *PlacedChild,
	regions *Regions,
	clearance bool,
) (Stop, error) {
	// Check if we should skip this float (already processed).
	if _, skip := ext.Work.Skips[placed.Location()]; skip {
		return nil, nil
	}

	// Layout the float.
	frame, err := placed.Layout(ext.Engine, regions.Base())
	if err != nil {
		return StopError{Err: err}, nil
	}

	// Determine vertical alignment.
	alignY := FixedAlignStart
	if placed.AlignY != nil {
		alignY = *placed.AlignY
	}

	// Check if the float fits.
	if !ext.floatFits(frame, placed.Clearance, regions) {
		// Queue the float for later processing.
		ext.Work.Floats = append(ext.Work.Floats, placed)

		// If clearance is set and we're not empty, trigger relayout.
		if clearance && placed.Clearance > 0 {
			return StopRelayout{Scope: placed.Scope}, nil
		}
		return nil, nil
	}

	// Place the float.
	ext.State.Insertions.PushFloat(placed, frame, alignY)

	// Mark this location to skip on relayout.
	ext.Work.Skips[placed.Location()] = struct{}{}

	// Shrink available region.
	reduction := frame.Height() + placed.Clearance
	regions.Size.Height -= reduction

	// Trigger relayout to redistribute content around the float.
	return StopRelayout{Scope: placed.Scope}, nil
}

// processFootnotes searches a frame for footnote elements and processes them.
// This is a simplified implementation that looks for footnote markers.
func (ext *ComposerExt) processFootnotes(
	regions *Regions,
	frame *Frame,
	flowNeed layout.Abs,
	breakable bool,
	migratable bool,
) error {
	// In a full implementation, this would:
	// 1. Recursively search the frame for footnote references
	// 2. Resolve each reference to its footnote content
	// 3. Layout the footnote and add it to insertions
	// 4. Handle cases where footnotes don't fit

	// For now, this is a pass-through that doesn't process footnotes.
	// The actual footnote discovery requires introspection support.
	return nil
}

// processQueuedFloats attempts to place queued floats.
func (ext *ComposerExt) processQueuedFloats(regions *Regions) Stop {
	remaining := make([]*PlacedChild, 0, len(ext.Work.Floats))

	for _, placed := range ext.Work.Floats {
		if _, skip := ext.Work.Skips[placed.Location()]; skip {
			continue
		}

		frame, err := placed.Layout(ext.Engine, regions.Base())
		if err != nil {
			return StopError{Err: err}
		}

		alignY := FixedAlignStart
		if placed.AlignY != nil {
			alignY = *placed.AlignY
		}

		if ext.floatFits(frame, placed.Clearance, regions) {
			ext.State.Insertions.PushFloat(placed, frame, alignY)
			ext.Work.Skips[placed.Location()] = struct{}{}
			regions.Size.Height -= frame.Height() + placed.Clearance
		} else {
			remaining = append(remaining, placed)
		}
	}

	ext.Work.Floats = remaining
	return nil
}

// Compose distributes content into a region with proper handling of floats
// and footnotes. This is the main entry point for frame composition.
//
// It performs the following steps:
// 1. Process any queued floats that might now fit
// 2. Distribute content using the Distribute function
// 3. Finalize insertions into the output frame
func Compose(composer *Composer, regions Regions) (Frame, Stop) {
	// Ensure state is initialized.
	if composer.State == nil {
		composer.State = NewComposerState()
	}

	// Process any queued floats first.
	if err := composer.ProcessQueuedFloats(&regions); err != nil {
		return Frame{}, StopError{Err: err}
	}

	// Distribute content into the region.
	content, stop := Distribute(composer, regions)
	if stop != nil {
		// If we got a relayout signal, return it without finalizing.
		if _, ok := stop.(StopRelayout); ok {
			return Frame{}, stop
		}
		// For finish or error, finalize what we have.
		if _, ok := stop.(StopError); ok {
			return Frame{}, stop
		}
	}

	// Finalize insertions into the output frame.
	output := composer.FinalizeInsertions(content, regions.Size)

	return output, stop
}

// ComposeLoop repeatedly composes a region until no relayout is needed.
// This handles the case where floats cause content to be redistributed.
func ComposeLoop(composer *Composer, regions Regions, maxIterations int) (Frame, Stop) {
	for i := 0; i < maxIterations; i++ {
		frame, stop := Compose(composer, regions)

		// If we get a relayout signal, reset and try again.
		if relayout, ok := stop.(StopRelayout); ok {
			// Reset the insertions but keep the skips to avoid reprocessing.
			composer.State.Insertions = NewInsertions()

			// For column-scoped relayouts, we stay in the same column.
			// For page-scoped relayouts, the caller handles it.
			if relayout.Scope == PlacementScopePage {
				return Frame{}, stop
			}
			continue
		}

		return frame, stop
	}

	// Exceeded max iterations - return what we have.
	content, stop := Distribute(composer, regions)
	return composer.FinalizeInsertions(content, regions.Size), stop
}

// StickyState tracks sticky block migration state.
type StickyState struct {
	// Checkpoint is a saved state for potential rollback.
	Checkpoint *StickyCheckpoint
	// Active indicates if sticky handling is enabled.
	Active bool
}

// StickyCheckpoint captures state for sticky block migration.
type StickyCheckpoint struct {
	WorkSnapshot Work
	ItemCount    int
}

// NewStickyState creates a new sticky state tracker.
func NewStickyState() *StickyState {
	return &StickyState{Active: true}
}

// Save creates a checkpoint of the current state.
func (s *StickyState) Save(work *Work, itemCount int) {
	if !s.Active {
		return
	}
	s.Checkpoint = &StickyCheckpoint{
		WorkSnapshot: work.Clone(),
		ItemCount:    itemCount,
	}
}

// Clear removes the checkpoint and disables sticky handling.
func (s *StickyState) Clear() {
	s.Checkpoint = nil
	s.Active = false
}

// Restore returns to the saved checkpoint if one exists.
func (s *StickyState) Restore(work *Work) int {
	if s.Checkpoint == nil {
		return -1
	}
	*work = s.Checkpoint.WorkSnapshot
	return s.Checkpoint.ItemCount
}
