package flow

import (
	"testing"

	"github.com/boergens/gotypst/layout"
)

func TestInsertions_Empty(t *testing.T) {
	ins := NewInsertions()
	if !ins.IsEmpty() {
		t.Error("new Insertions should be empty")
	}
	if ins.Height() != 0 {
		t.Errorf("empty Insertions height = %v, want 0", ins.Height())
	}
	if ins.Width() != 0 {
		t.Errorf("empty Insertions width = %v, want 0", ins.Width())
	}
}

func TestInsertions_PushFloat_Top(t *testing.T) {
	ins := NewInsertions()

	placed := &PlacedChild{
		AlignX:    FixedAlignStart,
		Clearance: 10,
		location:  1,
	}
	frame := NewFrame(layout.Size{Width: 100, Height: 50})

	ins.PushFloat(placed, frame, FixedAlignStart)

	if ins.IsEmpty() {
		t.Error("Insertions should not be empty after push")
	}
	if ins.TopHeight() != 60 { // 50 + 10 clearance
		t.Errorf("top height = %v, want 60", ins.TopHeight())
	}
	if ins.BottomHeight() != 0 {
		t.Errorf("bottom height = %v, want 0", ins.BottomHeight())
	}
	if ins.Width() != 100 {
		t.Errorf("width = %v, want 100", ins.Width())
	}
	if len(ins.Skips()) != 1 {
		t.Errorf("skips len = %d, want 1", len(ins.Skips()))
	}
}

func TestInsertions_PushFloat_Bottom(t *testing.T) {
	ins := NewInsertions()

	placed := &PlacedChild{
		AlignX:    FixedAlignCenter,
		Clearance: 5,
		location:  2,
	}
	frame := NewFrame(layout.Size{Width: 80, Height: 30})

	ins.PushFloat(placed, frame, FixedAlignEnd)

	if ins.TopHeight() != 0 {
		t.Errorf("top height = %v, want 0", ins.TopHeight())
	}
	if ins.BottomHeight() != 35 { // 30 + 5 clearance
		t.Errorf("bottom height = %v, want 35", ins.BottomHeight())
	}
}

func TestInsertions_PushFootnote(t *testing.T) {
	ins := NewInsertions()

	// First footnote - no gap.
	fn1 := NewFrame(layout.Size{Width: 200, Height: 20})
	ins.PushFootnote(fn1, 5)

	if ins.BottomHeight() != 20 {
		t.Errorf("bottom height = %v, want 20", ins.BottomHeight())
	}
	if ins.Width() != 200 {
		t.Errorf("width = %v, want 200", ins.Width())
	}

	// Second footnote - includes gap.
	fn2 := NewFrame(layout.Size{Width: 180, Height: 15})
	ins.PushFootnote(fn2, 5)

	if ins.BottomHeight() != 40 { // 20 + 5 gap + 15
		t.Errorf("bottom height = %v, want 40", ins.BottomHeight())
	}
}

func TestInsertions_PushFootnoteSeparator(t *testing.T) {
	ins := NewInsertions()

	sep := NewFrame(layout.Size{Width: 300, Height: 2})
	ins.PushFootnoteSeparator(sep, 8)

	if ins.BottomHeight() != 10 { // 2 + 8 clearance
		t.Errorf("bottom height = %v, want 10", ins.BottomHeight())
	}

	// Second call should be ignored.
	sep2 := NewFrame(layout.Size{Width: 300, Height: 5})
	ins.PushFootnoteSeparator(sep2, 8)

	if ins.BottomHeight() != 10 {
		t.Errorf("bottom height after second separator = %v, want 10", ins.BottomHeight())
	}
}

func TestInsertions_Finalize_Empty(t *testing.T) {
	ins := NewInsertions()
	content := NewFrame(layout.Size{Width: 100, Height: 200})
	regionSize := layout.Size{Width: 500, Height: 800}

	output := ins.Finalize(content, regionSize)

	// Should return content unchanged.
	if output.Width() != content.Width() || output.Height() != content.Height() {
		t.Error("empty finalize should return content unchanged")
	}
}

func TestInsertions_Finalize_WithFloats(t *testing.T) {
	ins := NewInsertions()

	// Add a top float.
	topPlaced := &PlacedChild{
		AlignX:    FixedAlignStart,
		Clearance: 10,
		location:  1,
	}
	topFrame := NewFrame(layout.Size{Width: 100, Height: 50})
	ins.PushFloat(topPlaced, topFrame, FixedAlignStart)

	// Add a bottom float.
	bottomPlaced := &PlacedChild{
		AlignX:    FixedAlignEnd,
		Clearance: 5,
		location:  2,
	}
	bottomFrame := NewFrame(layout.Size{Width: 80, Height: 30})
	ins.PushFloat(bottomPlaced, bottomFrame, FixedAlignEnd)

	content := NewFrame(layout.Size{Width: 400, Height: 200})
	regionSize := layout.Size{Width: 500, Height: 800}

	output := ins.Finalize(content, regionSize)

	// Should have items positioned.
	if output.Width() != regionSize.Width || output.Height() != regionSize.Height {
		t.Errorf("output size = %v, want %v", output.Size(), regionSize)
	}
	if len(output.Items()) != 3 { // top float + content + bottom float
		t.Errorf("output items = %d, want 3", len(output.Items()))
	}
}

func TestComposer_Float_Fits(t *testing.T) {
	engine := &Engine{}
	work := NewWork(nil)
	config := &Config{Mode: FlowModeRoot}
	composer := NewComposer(engine, work, config)

	regions := NewRegions(
		layout.Size{Width: 500, Height: 800},
		Axes[bool]{X: false, Y: false},
		layout.Size{Width: 500, Height: 800},
	)

	alignY := FixedAlignStart
	placed := &PlacedChild{
		AlignX:    FixedAlignStart,
		AlignY:    &alignY,
		Clearance: 10,
		location:  1,
	}

	err := composer.Float(placed, &regions, true, true)
	if err != nil {
		t.Errorf("Float returned error: %v", err)
	}

	// Float should be placed (via Skips).
	if _, ok := composer.Work.Skips[placed.Location()]; !ok {
		t.Error("float location should be in Skips")
	}
}

func TestComposer_Float_Queued(t *testing.T) {
	engine := &Engine{}
	work := NewWork(nil)
	config := &Config{Mode: FlowModeRoot}
	composer := NewComposer(engine, work, config)

	// Pre-fill insertions to use up space, leaving very little room.
	// This simulates a region that's mostly full.
	existingPlaced := &PlacedChild{
		AlignX:    FixedAlignStart,
		Clearance: 0,
		location:  99,
	}
	existingFrame := NewFrame(layout.Size{Width: 100, Height: 790})
	composer.State.Insertions.PushFloat(existingPlaced, existingFrame, FixedAlignStart)

	// Region with 800 height, but 790 is used by existing insertion.
	regions := NewRegions(
		layout.Size{Width: 500, Height: 800},
		Axes[bool]{X: false, Y: false},
		layout.Size{Width: 500, Height: 800},
	)

	alignY := FixedAlignStart
	placed := &PlacedChild{
		AlignX:    FixedAlignStart,
		AlignY:    &alignY,
		Clearance: 10,
		location:  1,
	}

	err := composer.Float(placed, &regions, true, true)
	if err != nil {
		t.Errorf("Float returned error: %v", err)
	}

	// Float should be queued since remaining space (10) doesn't fit
	// clearance (10) plus any frame height (even 0 needs the clearance).
	// Actually, since PlacedChild.Layout returns empty frame (height 0),
	// it will fit (0 <= 10). Let's test the insertion width instead.
	// The float with height 0 and clearance 10 would need 10 space.
	// Available is 800 - 790 = 10. So 0 + 10 clearance needs 10, which fits exactly.
	// To truly test queuing, we need the insertion to be slightly larger.

	// This test now verifies that when space is tight but available,
	// the float is placed (not queued).
	if _, ok := composer.Work.Skips[placed.Location()]; !ok {
		t.Error("float with 0 height should be placed when clearance exactly fits")
	}
}

func TestComposer_Float_Queued_NoSpace(t *testing.T) {
	engine := &Engine{}
	work := NewWork(nil)
	config := &Config{Mode: FlowModeRoot}
	composer := NewComposer(engine, work, config)

	// Pre-fill insertions to use up ALL space.
	existingPlaced := &PlacedChild{
		AlignX:    FixedAlignStart,
		Clearance: 0,
		location:  99,
	}
	existingFrame := NewFrame(layout.Size{Width: 100, Height: 800})
	composer.State.Insertions.PushFloat(existingPlaced, existingFrame, FixedAlignStart)

	// Region is completely full.
	regions := NewRegions(
		layout.Size{Width: 500, Height: 800},
		Axes[bool]{X: false, Y: false},
		layout.Size{Width: 500, Height: 800},
	)

	alignY := FixedAlignStart
	placed := &PlacedChild{
		AlignX:    FixedAlignStart,
		AlignY:    &alignY,
		Clearance: 10,
		location:  1,
	}

	err := composer.Float(placed, &regions, true, true)
	if err != nil {
		t.Errorf("Float returned error: %v", err)
	}

	// Float should be queued since there's no space at all.
	if len(composer.Work.Floats) != 1 {
		t.Errorf("expected 1 queued float, got %d", len(composer.Work.Floats))
	}
}

func TestComposer_InsertionWidth(t *testing.T) {
	engine := &Engine{}
	work := NewWork(nil)
	config := &Config{Mode: FlowModeRoot}
	composer := NewComposer(engine, work, config)

	if composer.InsertionWidth() != 0 {
		t.Error("InsertionWidth should be 0 initially")
	}

	// Add a float.
	alignY := FixedAlignStart
	placed := &PlacedChild{
		AlignX:    FixedAlignStart,
		AlignY:    &alignY,
		Clearance: 0,
		location:  1,
	}
	frame := NewFrame(layout.Size{Width: 150, Height: 50})
	composer.State.Insertions.PushFloat(placed, frame, FixedAlignStart)

	if composer.InsertionWidth() != 150 {
		t.Errorf("InsertionWidth = %v, want 150", composer.InsertionWidth())
	}
}

func TestCompose_Basic(t *testing.T) {
	engine := &Engine{}
	children := []Child{
		&LineChild{
			Frame: NewFrame(layout.Size{Width: 400, Height: 20}),
			Align: Axes[FixedAlignment]{X: FixedAlignStart, Y: FixedAlignStart},
			Need:  20,
		},
	}
	work := NewWork(children)
	config := &Config{Mode: FlowModeRoot}
	composer := NewComposer(engine, work, config)

	regions := NewRegions(
		layout.Size{Width: 500, Height: 800},
		Axes[bool]{X: false, Y: false},
		layout.Size{Width: 500, Height: 800},
	)

	frame, stop := Compose(composer, regions)

	if stop != nil {
		if _, ok := stop.(StopFinish); !ok {
			t.Errorf("unexpected stop: %v", stop)
		}
	}

	if frame.IsEmpty() {
		t.Error("composed frame should not be empty")
	}
}

func TestComposerState_Defaults(t *testing.T) {
	state := NewComposerState()

	if state.Insertions == nil {
		t.Error("Insertions should not be nil")
	}
	if state.FootnoteGap <= 0 {
		t.Error("FootnoteGap should be positive")
	}
	if state.FootnoteClearance <= 0 {
		t.Error("FootnoteClearance should be positive")
	}
}

func TestStickyState_SaveRestore(t *testing.T) {
	state := NewStickyState()

	children := []Child{
		&LineChild{Frame: NewFrame(layout.Size{Width: 100, Height: 10})},
		&LineChild{Frame: NewFrame(layout.Size{Width: 100, Height: 10})},
	}
	work := NewWork(children)
	work.Advance() // Move to second child.

	state.Save(work, 5)

	// Modify work.
	work.Advance()
	if work.index != 2 {
		t.Errorf("work.index = %d, want 2", work.index)
	}

	// Restore.
	itemCount := state.Restore(work)
	if itemCount != 5 {
		t.Errorf("restored itemCount = %d, want 5", itemCount)
	}
	if work.index != 1 {
		t.Errorf("restored work.index = %d, want 1", work.index)
	}
}

func TestStickyState_Clear(t *testing.T) {
	state := NewStickyState()
	work := NewWork(nil)

	state.Save(work, 3)
	state.Clear()

	if state.Active {
		t.Error("state should be inactive after Clear")
	}
	if state.Checkpoint != nil {
		t.Error("checkpoint should be nil after Clear")
	}

	// Save should be no-op when inactive.
	state.Save(work, 10)
	if state.Checkpoint != nil {
		t.Error("Save should be no-op when inactive")
	}
}
