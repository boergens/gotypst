package flow

import (
	"testing"

	"github.com/boergens/gotypst/layout"
)

func TestDefaultFootnoteConfig(t *testing.T) {
	config := DefaultFootnoteConfig()

	if config.Separator != nil {
		t.Error("Expected nil separator by default")
	}
	if config.Gap != layout.Abs(10) {
		t.Errorf("Expected gap of 10pt, got %v", config.Gap)
	}
	if config.Clearance != 0 {
		t.Error("Expected zero clearance by default")
	}
}

func TestNewFootnoteState(t *testing.T) {
	config := DefaultFootnoteConfig()
	state := NewFootnoteState(config)

	if state == nil {
		t.Fatal("NewFootnoteState returned nil")
	}
	if len(state.Pending) != 0 {
		t.Error("Expected empty pending list")
	}
	if len(state.Insertions) != 0 {
		t.Error("Expected empty insertions list")
	}
	if state.Used != 0 {
		t.Error("Expected zero used height")
	}
}

func TestFootnoteStateHeight(t *testing.T) {
	config := DefaultFootnoteConfig()
	state := NewFootnoteState(config)

	// Empty state should have zero height
	if state.Height() != 0 {
		t.Errorf("Empty state height = %v, want 0", state.Height())
	}

	// Add a footnote
	fn := Footnote{
		Location: 1,
		Frame:    NewFrame(layout.Size{Width: 100, Height: 50}),
	}
	state.Insertions = append(state.Insertions, fn)
	state.Used = 50

	// Height should include used space plus gap
	expectedHeight := layout.Abs(50 + 10) // 50pt content + 10pt gap
	if state.Height() != expectedHeight {
		t.Errorf("State height = %v, want %v", state.Height(), expectedHeight)
	}
}

func TestFootnoteStateHeightWithSeparator(t *testing.T) {
	sepFrame := NewFrame(layout.Size{Width: 100, Height: 5})
	config := FootnoteConfig{
		Separator: &sepFrame,
		Gap:       layout.Abs(8),
		Clearance: 0,
	}
	state := NewFootnoteState(config)

	// Add a footnote
	fn := Footnote{
		Location: 1,
		Frame:    NewFrame(layout.Size{Width: 100, Height: 50}),
	}
	state.Insertions = append(state.Insertions, fn)
	state.Used = 50

	// Height should include separator + gap + used
	expectedHeight := layout.Abs(5 + 8 + 50) // 5pt separator + 8pt gap + 50pt content
	if state.Height() != expectedHeight {
		t.Errorf("State height = %v, want %v", state.Height(), expectedHeight)
	}
}

func TestFootnoteStateWidth(t *testing.T) {
	config := DefaultFootnoteConfig()
	state := NewFootnoteState(config)

	// Empty state should have zero width
	if state.Width() != 0 {
		t.Errorf("Empty state width = %v, want 0", state.Width())
	}

	// Add footnotes of varying widths
	fn1 := Footnote{
		Location: 1,
		Frame:    NewFrame(layout.Size{Width: 100, Height: 50}),
	}
	fn2 := Footnote{
		Location: 2,
		Frame:    NewFrame(layout.Size{Width: 150, Height: 30}),
	}
	state.Insertions = append(state.Insertions, fn1, fn2)

	// Width should be the max of all footnote widths
	if state.Width() != 150 {
		t.Errorf("State width = %v, want 150", state.Width())
	}
}

func TestFootnoteStateClear(t *testing.T) {
	config := DefaultFootnoteConfig()
	state := NewFootnoteState(config)

	// Add footnotes
	fn := Footnote{
		Location: 1,
		Frame:    NewFrame(layout.Size{Width: 100, Height: 50}),
	}
	state.Insertions = append(state.Insertions, fn)
	state.Used = 50
	state.Pending = append(state.Pending, fn)

	state.Clear()

	if len(state.Insertions) != 0 {
		t.Error("Insertions not cleared")
	}
	if state.Used != 0 {
		t.Error("Used not cleared")
	}
	// Pending should remain
	if len(state.Pending) != 1 {
		t.Error("Pending should not be cleared")
	}
}

func TestFootnoteStateReset(t *testing.T) {
	config := DefaultFootnoteConfig()
	state := NewFootnoteState(config)

	// Add footnotes
	fn := Footnote{
		Location: 1,
		Frame:    NewFrame(layout.Size{Width: 100, Height: 50}),
	}
	state.Insertions = append(state.Insertions, fn)
	state.Used = 50
	state.Pending = append(state.Pending, fn)

	state.Reset()

	if len(state.Insertions) != 0 {
		t.Error("Insertions not reset")
	}
	if state.Used != 0 {
		t.Error("Used not reset")
	}
	// Pending should remain (they carry over)
	if len(state.Pending) != 1 {
		t.Error("Pending should be preserved across reset")
	}
}

func TestFootnoteStateFinalize(t *testing.T) {
	config := DefaultFootnoteConfig()
	state := NewFootnoteState(config)

	// Empty state should return nil
	frame := state.Finalize(200)
	if frame != nil {
		t.Error("Expected nil frame for empty state")
	}

	// Add a footnote
	fnFrame := NewFrame(layout.Size{Width: 100, Height: 50})
	fn := Footnote{
		Location: 1,
		Frame:    fnFrame,
	}
	state.Insertions = append(state.Insertions, fn)
	state.Used = 50

	frame = state.Finalize(200)
	if frame == nil {
		t.Fatal("Expected non-nil frame")
	}

	// Check frame dimensions
	expectedHeight := state.Height()
	if frame.Height() != expectedHeight {
		t.Errorf("Frame height = %v, want %v", frame.Height(), expectedHeight)
	}
	if frame.Width() != 200 {
		t.Errorf("Frame width = %v, want 200", frame.Width())
	}
}

func TestDiscoverFootnotesEmpty(t *testing.T) {
	frame := NewFrame(layout.Size{Width: 100, Height: 100})
	markers := DiscoverFootnotes(&frame)

	if len(markers) != 0 {
		t.Errorf("Expected no markers, got %d", len(markers))
	}
}

func TestDiscoverFootnotesWithMarker(t *testing.T) {
	frame := NewFrame(layout.Size{Width: 100, Height: 100})
	entryFrame := NewFrame(layout.Size{Width: 80, Height: 30})
	frame.Push(layout.Point{X: 0, Y: 0}, FrameItemFootnoteMarker{
		Location: 42,
		Entry:    entryFrame,
	})

	markers := DiscoverFootnotes(&frame)

	if len(markers) != 1 {
		t.Fatalf("Expected 1 marker, got %d", len(markers))
	}
	if markers[0].Location != 42 {
		t.Errorf("Marker location = %v, want 42", markers[0].Location)
	}
}

func TestDiscoverFootnotesNested(t *testing.T) {
	// Create a nested frame structure with a marker inside
	innerFrame := NewFrame(layout.Size{Width: 50, Height: 50})
	entryFrame := NewFrame(layout.Size{Width: 40, Height: 20})
	innerFrame.Push(layout.Point{X: 0, Y: 0}, FrameItemFootnoteMarker{
		Location: 123,
		Entry:    entryFrame,
	})

	outerFrame := NewFrame(layout.Size{Width: 100, Height: 100})
	outerFrame.PushFrame(layout.Point{X: 10, Y: 10}, innerFrame)

	markers := DiscoverFootnotes(&outerFrame)

	if len(markers) != 1 {
		t.Fatalf("Expected 1 marker, got %d", len(markers))
	}
	if markers[0].Location != 123 {
		t.Errorf("Marker location = %v, want 123", markers[0].Location)
	}
}

func TestProcessFootnotesEmpty(t *testing.T) {
	config := DefaultFootnoteConfig()
	state := NewFootnoteState(config)
	regions := NewRegions(layout.Size{Width: 200, Height: 500}, Axes[bool]{X: true, Y: true}, layout.Size{Width: 200, Height: 500})
	frame := NewFrame(layout.Size{Width: 100, Height: 50})

	err := processFootnotes(state, &regions, &frame, 50, false, true)
	if err != nil {
		t.Errorf("processFootnotes failed: %v", err)
	}

	// No footnotes should be placed
	if len(state.Insertions) != 0 {
		t.Errorf("Expected no insertions, got %d", len(state.Insertions))
	}
}

func TestProcessFootnotesWithMarker(t *testing.T) {
	config := DefaultFootnoteConfig()
	state := NewFootnoteState(config)
	regions := NewRegions(layout.Size{Width: 200, Height: 500}, Axes[bool]{X: true, Y: true}, layout.Size{Width: 200, Height: 500})

	// Create a frame with a footnote marker
	frame := NewFrame(layout.Size{Width: 100, Height: 50})
	entryFrame := NewFrame(layout.Size{Width: 80, Height: 30})
	frame.Push(layout.Point{X: 0, Y: 0}, FrameItemFootnoteMarker{
		Location: 1,
		Entry:    entryFrame,
	})

	err := processFootnotes(state, &regions, &frame, 50, false, true)
	if err != nil {
		t.Errorf("processFootnotes failed: %v", err)
	}

	// Footnote should be placed
	if len(state.Insertions) != 1 {
		t.Errorf("Expected 1 insertion, got %d", len(state.Insertions))
	}
}

func TestProcessFootnotesOverflow(t *testing.T) {
	config := DefaultFootnoteConfig()
	state := NewFootnoteState(config)
	// Small region that can't fit the footnote
	regions := NewRegions(layout.Size{Width: 200, Height: 80}, Axes[bool]{X: true, Y: true}, layout.Size{Width: 200, Height: 80})

	// Create a frame with a large footnote marker
	frame := NewFrame(layout.Size{Width: 100, Height: 50})
	entryFrame := NewFrame(layout.Size{Width: 80, Height: 100}) // Larger than remaining space
	frame.Push(layout.Point{X: 0, Y: 0}, FrameItemFootnoteMarker{
		Location: 1,
		Entry:    entryFrame,
	})

	err := processFootnotes(state, &regions, &frame, 50, false, true)
	if err != nil {
		t.Errorf("processFootnotes failed: %v", err)
	}

	// Footnote should be pending (doesn't fit)
	if len(state.Pending) != 1 {
		t.Errorf("Expected 1 pending footnote, got %d", len(state.Pending))
	}
}

func TestComposerFootnotesNonRoot(t *testing.T) {
	config := &Config{Mode: FlowModeBlock}
	composer := &Composer{Config: config}
	regions := NewRegions(layout.Size{Width: 200, Height: 500}, Axes[bool]{X: true, Y: true}, layout.Size{Width: 200, Height: 500})
	frame := NewFrame(layout.Size{Width: 100, Height: 50})

	// Footnotes should be ignored in non-root mode
	err := composer.Footnotes(&regions, &frame, 50, false, true)
	if err != nil {
		t.Errorf("Footnotes failed: %v", err)
	}

	if composer.FootnoteState != nil {
		t.Error("FootnoteState should be nil in non-root mode")
	}
}

func TestComposerFootnotesRoot(t *testing.T) {
	config := &Config{Mode: FlowModeRoot, Footnotes: DefaultFootnoteConfig()}
	composer := &Composer{Config: config}
	regions := NewRegions(layout.Size{Width: 200, Height: 500}, Axes[bool]{X: true, Y: true}, layout.Size{Width: 200, Height: 500})
	frame := NewFrame(layout.Size{Width: 100, Height: 50})

	err := composer.Footnotes(&regions, &frame, 50, false, true)
	if err != nil {
		t.Errorf("Footnotes failed: %v", err)
	}

	// FootnoteState should be initialized in root mode
	if composer.FootnoteState == nil {
		t.Error("FootnoteState should be initialized in root mode")
	}
}

func TestComposerInsertionWidth(t *testing.T) {
	composer := &Composer{}

	// Without footnote state
	if composer.InsertionWidth() != 0 {
		t.Error("Expected zero width without footnote state")
	}

	// With footnote state and footnotes
	config := DefaultFootnoteConfig()
	composer.FootnoteState = NewFootnoteState(config)
	fn := Footnote{
		Location: 1,
		Frame:    NewFrame(layout.Size{Width: 150, Height: 50}),
	}
	composer.FootnoteState.Insertions = append(composer.FootnoteState.Insertions, fn)

	if composer.InsertionWidth() != 150 {
		t.Errorf("InsertionWidth = %v, want 150", composer.InsertionWidth())
	}
}

func TestComposerFootnoteHeight(t *testing.T) {
	composer := &Composer{}

	// Without footnote state
	if composer.FootnoteHeight() != 0 {
		t.Error("Expected zero height without footnote state")
	}

	// With footnote state and footnotes
	config := DefaultFootnoteConfig()
	composer.FootnoteState = NewFootnoteState(config)
	fn := Footnote{
		Location: 1,
		Frame:    NewFrame(layout.Size{Width: 100, Height: 50}),
	}
	composer.FootnoteState.Insertions = append(composer.FootnoteState.Insertions, fn)
	composer.FootnoteState.Used = 50

	expectedHeight := composer.FootnoteState.Height()
	if composer.FootnoteHeight() != expectedHeight {
		t.Errorf("FootnoteHeight = %v, want %v", composer.FootnoteHeight(), expectedHeight)
	}
}

func TestComposerFinalizeFootnotes(t *testing.T) {
	composer := &Composer{}

	// Without footnote state
	if composer.FinalizeFootnotes(200) != nil {
		t.Error("Expected nil without footnote state")
	}

	// With footnote state and footnotes
	config := DefaultFootnoteConfig()
	composer.FootnoteState = NewFootnoteState(config)
	fn := Footnote{
		Location: 1,
		Frame:    NewFrame(layout.Size{Width: 100, Height: 50}),
	}
	composer.FootnoteState.Insertions = append(composer.FootnoteState.Insertions, fn)
	composer.FootnoteState.Used = 50

	frame := composer.FinalizeFootnotes(200)
	if frame == nil {
		t.Error("Expected non-nil frame")
	}
}

func TestComposerHasPendingFootnotes(t *testing.T) {
	composer := &Composer{}

	// Without footnote state
	if composer.HasPendingFootnotes() {
		t.Error("Expected false without footnote state")
	}

	// With empty pending list
	config := DefaultFootnoteConfig()
	composer.FootnoteState = NewFootnoteState(config)
	if composer.HasPendingFootnotes() {
		t.Error("Expected false with empty pending list")
	}

	// With pending footnotes
	fn := Footnote{Location: 1}
	composer.FootnoteState.Pending = append(composer.FootnoteState.Pending, fn)
	if !composer.HasPendingFootnotes() {
		t.Error("Expected true with pending footnotes")
	}
}

func TestComposerResetFootnotes(t *testing.T) {
	config := DefaultFootnoteConfig()
	composer := &Composer{
		FootnoteState: NewFootnoteState(config),
	}

	// Add insertions and pending
	fn := Footnote{
		Location: 1,
		Frame:    NewFrame(layout.Size{Width: 100, Height: 50}),
	}
	composer.FootnoteState.Insertions = append(composer.FootnoteState.Insertions, fn)
	composer.FootnoteState.Used = 50
	composer.FootnoteState.Pending = append(composer.FootnoteState.Pending, fn)

	composer.ResetFootnotes()

	if len(composer.FootnoteState.Insertions) != 0 {
		t.Error("Insertions not reset")
	}
	if composer.FootnoteState.Used != 0 {
		t.Error("Used not reset")
	}
	// Pending should remain
	if len(composer.FootnoteState.Pending) != 1 {
		t.Error("Pending should be preserved")
	}
}
