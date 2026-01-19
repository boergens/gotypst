package flow

import (
	"testing"

	"github.com/boergens/gotypst/layout"
)

func TestNewFrame(t *testing.T) {
	size := layout.Size{Width: 100, Height: 200}

	soft := NewSoftFrame(size)
	if soft.Kind != FrameSoft {
		t.Errorf("expected FrameSoft, got %v", soft.Kind)
	}
	if soft.Width() != 100 {
		t.Errorf("expected width 100, got %v", soft.Width())
	}
	if soft.Height() != 200 {
		t.Errorf("expected height 200, got %v", soft.Height())
	}

	hard := NewHardFrame(size)
	if hard.Kind != FrameHard {
		t.Errorf("expected FrameHard, got %v", hard.Kind)
	}
}

func TestFramePushFrame(t *testing.T) {
	parent := NewSoftFrame(layout.Size{Width: 100, Height: 100})
	child := NewSoftFrame(layout.Size{Width: 50, Height: 50})

	parent.PushFrame(layout.Point{X: 10, Y: 20}, child)

	if len(parent.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(parent.Items))
	}

	item := parent.Items[0]
	if item.Pos.X != 10 || item.Pos.Y != 20 {
		t.Errorf("expected position (10, 20), got (%v, %v)", item.Pos.X, item.Pos.Y)
	}
}

func TestFrameTranslate(t *testing.T) {
	frame := NewSoftFrame(layout.Size{Width: 100, Height: 100})
	child := NewSoftFrame(layout.Size{Width: 50, Height: 50})
	frame.PushFrame(layout.Point{X: 10, Y: 20}, child)

	frame.Translate(layout.Point{X: 5, Y: 10})

	if len(frame.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(frame.Items))
	}
	if frame.Items[0].Pos.X != 15 || frame.Items[0].Pos.Y != 30 {
		t.Errorf("expected position (15, 30), got (%v, %v)",
			frame.Items[0].Pos.X, frame.Items[0].Pos.Y)
	}
}

func TestInsertions(t *testing.T) {
	ins := NewInsertions()

	if !ins.IsEmpty() {
		t.Error("expected empty insertions")
	}
	if ins.Height() != 0 {
		t.Errorf("expected height 0, got %v", ins.Height())
	}

	// Add a top float
	placed := &PlacedChild{
		Scope:     PlacementScopeColumn,
		AlignX:    FixedAlignmentStart,
		Clearance: 5,
	}
	frame := NewSoftFrame(layout.Size{Width: 50, Height: 30})
	ins.PushFloat(placed, frame, FixedAlignmentStart)

	if ins.IsEmpty() {
		t.Error("expected non-empty insertions after adding float")
	}
	if ins.Height() != 35 { // 30 + 5 clearance
		t.Errorf("expected height 35, got %v", ins.Height())
	}

	// Add a bottom float
	placed2 := &PlacedChild{
		Scope:     PlacementScopeColumn,
		AlignX:    FixedAlignmentEnd,
		Clearance: 10,
	}
	frame2 := NewSoftFrame(layout.Size{Width: 60, Height: 40})
	ins.PushFloat(placed2, frame2, FixedAlignmentEnd)

	if ins.Height() != 85 { // 35 + 40 + 10
		t.Errorf("expected height 85, got %v", ins.Height())
	}
}

func TestInsertionsFinalize(t *testing.T) {
	config := &Config{
		Footnote: FootnoteConfig{
			Clearance: 5,
			Gap:       3,
		},
	}
	work := &Work{
		Skips: make(map[Location]struct{}),
	}
	ins := NewInsertions()

	// Add top float
	placed := &PlacedChild{
		Scope:     PlacementScopeColumn,
		AlignX:    FixedAlignmentStart,
		Clearance: 5,
	}
	topFrame := NewSoftFrame(layout.Size{Width: 50, Height: 20})
	ins.PushFloat(placed, topFrame, FixedAlignmentStart)

	// Inner content
	inner := NewSoftFrame(layout.Size{Width: 100, Height: 50})

	// Finalize
	output := ins.Finalize(work, config, inner)

	// Total height should be: top float (20) + clearance (5) + inner (50) = 75
	expectedHeight := layout.Abs(75)
	if output.Height() != expectedHeight {
		t.Errorf("expected height %v, got %v", expectedHeight, output.Height())
	}

	// Should have 2 items: top float and inner
	if len(output.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(output.Items))
	}
}

func TestRegions(t *testing.T) {
	regions := Regions{
		Size:    layout.Size{Width: 100, Height: 200},
		Backlog: []layout.Abs{180, 180},
		Expand:  Axes[bool]{X: true, Y: false},
		Full:    layout.Size{Width: 100, Height: 200},
		Last:    false,
	}

	if !regions.MayProgress() {
		t.Error("expected MayProgress to be true with backlog")
	}

	base := regions.Base()
	if base.Width != 100 || base.Height != 200 {
		t.Errorf("expected base (100, 200), got (%v, %v)", base.Width, base.Height)
	}

	sizes := regions.Iter()
	if len(sizes) != 3 {
		t.Errorf("expected 3 sizes, got %d", len(sizes))
	}

	regions.Next()
	if regions.Size.Height != 180 {
		t.Errorf("expected height 180 after Next(), got %v", regions.Size.Height)
	}
	if len(regions.Backlog) != 1 {
		t.Errorf("expected 1 backlog item after Next(), got %d", len(regions.Backlog))
	}
}

func TestFixedAlignment(t *testing.T) {
	tests := []struct {
		align    FixedAlignment
		avail    layout.Abs
		expected layout.Abs
	}{
		{FixedAlignmentStart, 100, 0},
		{FixedAlignmentCenter, 100, 50},
		{FixedAlignmentEnd, 100, 100},
	}

	for _, tt := range tests {
		pos := tt.align.Position(tt.avail)
		if pos != tt.expected {
			t.Errorf("align %v with avail %v: expected %v, got %v",
				tt.align, tt.avail, tt.expected, pos)
		}
	}
}

func TestFixedAlignmentInv(t *testing.T) {
	if FixedAlignmentStart.Inv() != FixedAlignmentEnd {
		t.Error("expected Start.Inv() == End")
	}
	if FixedAlignmentEnd.Inv() != FixedAlignmentStart {
		t.Error("expected End.Inv() == Start")
	}
	if FixedAlignmentCenter.Inv() != FixedAlignmentCenter {
		t.Error("expected Center.Inv() == Center")
	}
}

func TestSmartValue(t *testing.T) {
	auto := SmartAuto[int]()
	if !auto.IsAuto() {
		t.Error("expected IsAuto() to be true")
	}
	if auto.GetOr(42) != 42 {
		t.Errorf("expected GetOr to return default, got %v", auto.GetOr(42))
	}

	custom := SmartCustom(123)
	if custom.IsAuto() {
		t.Error("expected IsAuto() to be false")
	}
	if custom.Get() != 123 {
		t.Errorf("expected Get() == 123, got %v", custom.Get())
	}
	if custom.GetOr(42) != 123 {
		t.Errorf("expected GetOr to return custom value, got %v", custom.GetOr(42))
	}
}

func TestComposeSingleColumn(t *testing.T) {
	config := &Config{
		Mode: FlowModeRoot,
		Columns: ColumnConfig{
			Count:  1,
			Width:  100,
			Gutter: 10,
			Dir:    layout.DirLTR,
		},
		Footnote: FootnoteConfig{
			Clearance: 5,
			Gap:       3,
		},
	}

	work := &Work{
		Skips: make(map[Location]struct{}),
	}

	regions := Regions{
		Size:    layout.Size{Width: 100, Height: 200},
		Backlog: nil,
		Expand:  Axes[bool]{X: true, Y: false},
		Full:    layout.Size{Width: 100, Height: 200},
		Last:    true,
	}

	frame, err := Compose(nil, work, config, NewLocator(), regions)
	if err != nil {
		t.Fatalf("Compose failed: %v", err)
	}
	if frame == nil {
		t.Fatal("expected non-nil frame")
	}
}

func TestComposeMultiColumn(t *testing.T) {
	config := &Config{
		Mode: FlowModeRoot,
		Columns: ColumnConfig{
			Count:  2,
			Width:  45,
			Gutter: 10,
			Dir:    layout.DirLTR,
		},
		Footnote: FootnoteConfig{
			Clearance: 5,
			Gap:       3,
		},
	}

	work := &Work{
		Skips: make(map[Location]struct{}),
	}

	regions := Regions{
		Size:    layout.Size{Width: 100, Height: 200},
		Backlog: nil,
		Expand:  Axes[bool]{X: true, Y: false},
		Full:    layout.Size{Width: 100, Height: 200},
		Last:    true,
	}

	frame, err := Compose(nil, work, config, NewLocator(), regions)
	if err != nil {
		t.Fatalf("Compose failed: %v", err)
	}
	if frame == nil {
		t.Fatal("expected non-nil frame")
	}

	// Multi-column should produce a hard frame
	if frame.Kind != FrameHard {
		t.Error("expected hard frame for multi-column layout")
	}
}

func TestWorkClone(t *testing.T) {
	work := &Work{
		Children: []Child{LineChild{Frame: NewSoftFrame(layout.Size{Width: 10, Height: 10})}},
		Skips:    map[Location]struct{}{{hash: 1}: {}},
	}

	clone := work.Clone()

	// Modify original
	work.Children = append(work.Children, LineChild{})
	work.Skips[Location{hash: 2}] = struct{}{}

	// Clone should be unaffected
	if len(clone.Children) != 1 {
		t.Errorf("expected clone to have 1 child, got %d", len(clone.Children))
	}
	if len(clone.Skips) != 1 {
		t.Errorf("expected clone to have 1 skip, got %d", len(clone.Skips))
	}
}

func TestFindInFrame(t *testing.T) {
	// Create a frame with some tags
	frame := NewSoftFrame(layout.Size{Width: 100, Height: 100})
	marker := &ParLineMarker{
		Numbering: Numbering{Pattern: "1"},
	}
	frame.Items = append(frame.Items, FrameItem{
		Pos: layout.Point{X: 0, Y: 20},
		Item: TagItem{
			Tag: StartTag{Elem: marker},
		},
	})

	found := FindInFrame(frame, func(elem interface{}) (*ParLineMarker, bool) {
		if m, ok := elem.(*ParLineMarker); ok {
			return m, true
		}
		return nil, false
	})

	if len(found) != 1 {
		t.Fatalf("expected 1 found element, got %d", len(found))
	}
	if found[0].Y != 20 {
		t.Errorf("expected Y=20, got %v", found[0].Y)
	}
	if found[0].Element != marker {
		t.Error("expected same marker element")
	}
}
