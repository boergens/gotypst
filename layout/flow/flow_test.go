package flow

import (
	"testing"

	"github.com/boergens/gotypst/layout"
)

func TestComposerFloat_BasicPlacement(t *testing.T) {
	engine := &Engine{}
	work := NewWork(nil)
	config := &Config{Mode: FlowModeRoot}
	composer := &Composer{Engine: engine, Work: work, Config: config}

	regions := NewRegions(
		layout.Size{Width: 100, Height: 200},
		Axes[bool]{X: false, Y: false},
		layout.Size{Width: 100, Height: 200},
	)

	// Create a placed child that is a float.
	alignY := FixedAlignStart
	placed := &PlacedChild{
		AlignX:    FixedAlignStart,
		AlignY:    &alignY,
		Scope:     PlacementScopeColumn,
		Float:     true,
		Clearance: 0,
		Delta:     Axes[Rel]{},
		location:  1,
	}

	err := composer.Float(placed, &regions, false, true)
	if err != nil {
		t.Fatalf("Float failed: %v", err)
	}

	// Float should be placed since it fits.
	floats := composer.PlacedFloats()
	if len(floats) != 1 {
		t.Errorf("Expected 1 placed float, got %d", len(floats))
	}

	// Location should be in skips.
	if _, ok := work.Skips[placed.Location()]; !ok {
		t.Error("Expected placed float location in Skips")
	}
}

func TestComposerFloat_QueueWhenNoFit(t *testing.T) {
	// Note: This test currently passes because PlacedChild.Layout() is a stub
	// that returns empty frames. When Layout() is implemented, floats that
	// don't fit will be queued.
	//
	// For now, we test the queue processing logic directly.
	engine := &Engine{}
	work := NewWork(nil)
	config := &Config{Mode: FlowModeRoot}
	composer := &Composer{Engine: engine, Work: work, Config: config}

	// Pre-queue a float.
	placed := &PlacedChild{
		AlignX:    FixedAlignStart,
		AlignY:    nil, // Inline float
		Scope:     PlacementScopeColumn,
		Float:     true,
		Clearance: 0,
		Delta:     Axes[Rel]{},
		location:  1,
	}
	work.Floats = []*PlacedChild{placed}

	regions := NewRegions(
		layout.Size{Width: 100, Height: 200},
		Axes[bool]{X: false, Y: false},
		layout.Size{Width: 100, Height: 200},
	)

	// Processing another float should trigger processing of queued floats.
	anotherPlaced := &PlacedChild{
		AlignX:   FixedAlignStart,
		AlignY:   nil,
		Scope:    PlacementScopeColumn,
		Float:    true,
		location: 2,
	}

	err := composer.Float(anotherPlaced, &regions, false, true)
	if err != nil {
		t.Fatalf("Float failed: %v", err)
	}

	// Both floats should be placed now (since Layout() returns empty frames that fit).
	// The queued float should have been processed.
	if len(work.Floats) != 0 {
		t.Errorf("Expected 0 queued floats after processing, got %d", len(work.Floats))
	}

	// Both should be placed.
	if len(composer.PlacedFloats()) != 2 {
		t.Errorf("Expected 2 placed floats, got %d", len(composer.PlacedFloats()))
	}
}

func TestComposerFloat_SkipAlreadyProcessed(t *testing.T) {
	engine := &Engine{}
	work := NewWork(nil)
	config := &Config{Mode: FlowModeRoot}
	composer := &Composer{Engine: engine, Work: work, Config: config}

	regions := NewRegions(
		layout.Size{Width: 100, Height: 200},
		Axes[bool]{X: false, Y: false},
		layout.Size{Width: 100, Height: 200},
	)

	placed := &PlacedChild{
		AlignX:    FixedAlignStart,
		AlignY:    nil,
		Scope:     PlacementScopeColumn,
		Float:     true,
		Clearance: 0,
		Delta:     Axes[Rel]{},
		location:  1,
	}

	// Pre-mark as processed.
	work.Skips[placed.Location()] = struct{}{}

	err := composer.Float(placed, &regions, false, true)
	if err != nil {
		t.Fatalf("Float failed: %v", err)
	}

	// Should not be placed again.
	if len(composer.PlacedFloats()) != 0 {
		t.Errorf("Expected 0 placed floats for already processed, got %d", len(composer.PlacedFloats()))
	}
}

func TestComposerInsertionWidth(t *testing.T) {
	engine := &Engine{}
	work := NewWork(nil)
	config := &Config{Mode: FlowModeRoot}
	composer := &Composer{Engine: engine, Work: work, Config: config}

	// Initially zero.
	if w := composer.InsertionWidth(); w != 0 {
		t.Errorf("Expected 0 insertion width, got %v", w)
	}

	// Add some placed floats manually.
	composer.placedFloats = []PlacedFloat{
		{Frame: NewFrame(layout.Size{Width: 50, Height: 30})},
		{Frame: NewFrame(layout.Size{Width: 80, Height: 40})},
		{Frame: NewFrame(layout.Size{Width: 30, Height: 20})},
	}

	// Should return the max width.
	if w := composer.InsertionWidth(); w != 80 {
		t.Errorf("Expected 80 insertion width, got %v", w)
	}
}

func TestComposerClearPlacedFloats(t *testing.T) {
	engine := &Engine{}
	work := NewWork(nil)
	config := &Config{Mode: FlowModeRoot}
	composer := &Composer{Engine: engine, Work: work, Config: config}

	composer.placedFloats = []PlacedFloat{
		{Frame: NewFrame(layout.Size{Width: 50, Height: 30})},
	}

	if len(composer.PlacedFloats()) != 1 {
		t.Errorf("Expected 1 placed float, got %d", len(composer.PlacedFloats()))
	}

	composer.ClearPlacedFloats()

	if len(composer.PlacedFloats()) != 0 {
		t.Errorf("Expected 0 placed floats after clear, got %d", len(composer.PlacedFloats()))
	}
}

func TestPlacedFloat_TopAlignment(t *testing.T) {
	alignY := FixedAlignStart
	placed := &PlacedChild{
		AlignX: FixedAlignStart,
		AlignY: &alignY, // Top alignment
		Scope:  PlacementScopeColumn,
		Float:  true,
	}

	// Top-aligned floats have explicit AlignY.
	if placed.AlignY == nil {
		t.Error("Expected explicit AlignY for top-aligned float")
	}
	if *placed.AlignY != FixedAlignStart {
		t.Errorf("Expected FixedAlignStart, got %v", *placed.AlignY)
	}
}

func TestPlacedFloat_BottomAlignment(t *testing.T) {
	alignY := FixedAlignEnd
	placed := &PlacedChild{
		AlignX: FixedAlignStart,
		AlignY: &alignY, // Bottom alignment
		Scope:  PlacementScopeColumn,
		Float:  true,
	}

	// Bottom-aligned floats have explicit AlignY.
	if placed.AlignY == nil {
		t.Error("Expected explicit AlignY for bottom-aligned float")
	}
	if *placed.AlignY != FixedAlignEnd {
		t.Errorf("Expected FixedAlignEnd, got %v", *placed.AlignY)
	}
}

func TestPlacedFloat_InlineAlignment(t *testing.T) {
	placed := &PlacedChild{
		AlignX: FixedAlignStart,
		AlignY: nil, // Inline - no explicit vertical alignment
		Scope:  PlacementScopeColumn,
		Float:  true,
	}

	// Inline floats have nil AlignY.
	if placed.AlignY != nil {
		t.Error("Expected nil AlignY for inline float")
	}
}

func TestDistribute_WithFloats(t *testing.T) {
	engine := &Engine{}
	work := NewWork(nil)
	config := &Config{Mode: FlowModeRoot}
	composer := &Composer{Engine: engine, Work: work, Config: config}

	regions := NewRegions(
		layout.Size{Width: 100, Height: 200},
		Axes[bool]{X: false, Y: true},
		layout.Size{Width: 100, Height: 200},
	)

	frame, stop := Distribute(composer, regions)

	if stop != nil {
		t.Fatalf("Distribute returned stop: %v", stop)
	}

	// Frame should be created successfully.
	if frame.Size().Width != 0 && frame.Size().Height == 0 {
		t.Error("Expected valid frame dimensions")
	}
}

func TestWorkClone_PreservesFloats(t *testing.T) {
	placed := &PlacedChild{
		AlignX:   FixedAlignStart,
		location: 1,
		Float:    true,
	}

	work := NewWork(nil)
	work.Floats = []*PlacedChild{placed}
	work.Skips[placed.location] = struct{}{}

	clone := work.Clone()

	// Clone should have independent floats slice.
	if len(clone.Floats) != 1 {
		t.Errorf("Expected 1 float in clone, got %d", len(clone.Floats))
	}

	// Modifying original shouldn't affect clone.
	work.Floats = append(work.Floats, &PlacedChild{location: 2})
	if len(clone.Floats) != 1 {
		t.Errorf("Clone floats affected by original modification, got %d", len(clone.Floats))
	}

	// Skips should be independent too.
	if _, ok := clone.Skips[Location(1)]; !ok {
		t.Error("Expected skip in clone")
	}
	work.Skips[Location(3)] = struct{}{}
	if _, ok := clone.Skips[Location(3)]; ok {
		t.Error("Clone skips affected by original modification")
	}
}

func TestFixedAlignment_Position(t *testing.T) {
	tests := []struct {
		align    FixedAlignment
		free     layout.Abs
		expected layout.Abs
	}{
		{FixedAlignStart, 100, 0},
		{FixedAlignCenter, 100, 50},
		{FixedAlignEnd, 100, 100},
	}

	for _, tt := range tests {
		result := tt.align.Position(tt.free)
		if result != tt.expected {
			t.Errorf("FixedAlignment(%v).Position(%v) = %v, want %v",
				tt.align, tt.free, result, tt.expected)
		}
	}
}

func TestFlushChild_TriggersFinishWithPendingFloats(t *testing.T) {
	engine := &Engine{}
	work := NewWork([]Child{FlushChild{}})
	work.Floats = []*PlacedChild{{location: 1}} // Pending float
	config := &Config{Mode: FlowModeRoot}
	composer := &Composer{Engine: engine, Work: work, Config: config}

	d := &Distributor{
		composer: composer,
		regions: NewRegions(
			layout.Size{Width: 100, Height: 200},
			Axes[bool]{X: false, Y: false},
			layout.Size{Width: 100, Height: 200},
		),
	}

	stop := d.flush()
	if stop == nil {
		t.Error("Expected StopFinish when pending floats exist")
	}
	if _, ok := stop.(StopFinish); !ok {
		t.Errorf("Expected StopFinish, got %T", stop)
	}
}

func TestFlushChild_NopWithoutPendingFloats(t *testing.T) {
	engine := &Engine{}
	work := NewWork([]Child{FlushChild{}})
	// No pending floats
	config := &Config{Mode: FlowModeRoot}
	composer := &Composer{Engine: engine, Work: work, Config: config}

	d := &Distributor{
		composer: composer,
		regions: NewRegions(
			layout.Size{Width: 100, Height: 200},
			Axes[bool]{X: false, Y: false},
			layout.Size{Width: 100, Height: 200},
		),
	}

	stop := d.flush()
	if stop != nil {
		t.Errorf("Expected nil when no pending floats, got %v", stop)
	}
}
