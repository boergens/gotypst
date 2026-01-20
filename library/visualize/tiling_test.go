package visualize

import (
	"testing"
)

func TestNewTiling(t *testing.T) {
	body := "test content"
	tiling := NewTiling(body)

	if tiling == nil {
		t.Fatal("expected tiling to be created")
	}
	if tiling.Body != body {
		t.Errorf("expected body %q, got %v", body, tiling.Body)
	}
	if tiling.Size != nil {
		t.Error("expected size to be nil (auto)")
	}
	if tiling.Spacing[0] != 0 || tiling.Spacing[1] != 0 {
		t.Errorf("expected zero spacing, got (%f, %f)", tiling.Spacing[0], tiling.Spacing[1])
	}
	if tiling.Relative != RelativeAuto {
		t.Errorf("expected auto relative, got %v", tiling.Relative)
	}
}

func TestTilingWithOptions(t *testing.T) {
	body := "pattern"
	tiling := NewTiling(body,
		WithSize(20, 30),
		WithSpacing(5, 10),
		WithTilingRelative(RelativeParent),
	)

	if tiling.Size == nil {
		t.Fatal("expected size to be set")
	}
	if tiling.Size[0] != 20 || tiling.Size[1] != 30 {
		t.Errorf("expected size (20, 30), got (%f, %f)", tiling.Size[0], tiling.Size[1])
	}
	if tiling.Spacing[0] != 5 || tiling.Spacing[1] != 10 {
		t.Errorf("expected spacing (5, 10), got (%f, %f)", tiling.Spacing[0], tiling.Spacing[1])
	}
	if tiling.Relative != RelativeParent {
		t.Errorf("expected parent relative, got %v", tiling.Relative)
	}
}

func TestTilingGetters(t *testing.T) {
	body := "content"
	tiling := NewTiling(body,
		WithSize(15, 25),
		WithSpacing(3, 4),
	)

	size := tiling.GetSize()
	if size == nil {
		t.Fatal("expected size to be set")
	}
	if (*size)[0] != 15 || (*size)[1] != 25 {
		t.Errorf("expected size (15, 25), got (%f, %f)", (*size)[0], (*size)[1])
	}

	spacing := tiling.GetSpacing()
	if spacing[0] != 3 || spacing[1] != 4 {
		t.Errorf("expected spacing (3, 4), got (%f, %f)", spacing[0], spacing[1])
	}

	if tiling.GetRelative() != RelativeAuto {
		t.Errorf("expected auto relative, got %v", tiling.GetRelative())
	}

	if tiling.GetBody() != body {
		t.Errorf("expected body %q, got %v", body, tiling.GetBody())
	}
}

func TestTilingIsAuto(t *testing.T) {
	// Auto size
	autoTiling := NewTiling("content")
	if !autoTiling.IsAuto() {
		t.Error("expected IsAuto to be true for auto-sized tiling")
	}

	// Explicit size
	sizedTiling := NewTiling("content", WithSize(10, 10))
	if sizedTiling.IsAuto() {
		t.Error("expected IsAuto to be false for sized tiling")
	}
}

func TestTilingCellDimensions(t *testing.T) {
	tiling := NewTiling("content", WithSize(20, 30))

	if tiling.CellWidth() != 20 {
		t.Errorf("expected cell width 20, got %f", tiling.CellWidth())
	}
	if tiling.CellHeight() != 30 {
		t.Errorf("expected cell height 30, got %f", tiling.CellHeight())
	}

	// Auto tiling should return 0
	autoTiling := NewTiling("content")
	if autoTiling.CellWidth() != 0 {
		t.Errorf("expected cell width 0 for auto, got %f", autoTiling.CellWidth())
	}
}

func TestTilingSpacing(t *testing.T) {
	tiling := NewTiling("content", WithSpacing(5, 8))

	if tiling.HorizontalSpacing() != 5 {
		t.Errorf("expected horizontal spacing 5, got %f", tiling.HorizontalSpacing())
	}
	if tiling.VerticalSpacing() != 8 {
		t.Errorf("expected vertical spacing 8, got %f", tiling.VerticalSpacing())
	}
}

func TestTilingEffectiveCellSize(t *testing.T) {
	tiling := NewTiling("content",
		WithSize(20, 30),
		WithSpacing(5, 10),
	)

	width, height := tiling.EffectiveCellSize()
	if width != 25 {
		t.Errorf("expected effective width 25, got %f", width)
	}
	if height != 40 {
		t.Errorf("expected effective height 40, got %f", height)
	}

	// Auto tiling should return 0, 0
	autoTiling := NewTiling("content")
	w, h := autoTiling.EffectiveCellSize()
	if w != 0 || h != 0 {
		t.Errorf("expected (0, 0) for auto tiling, got (%f, %f)", w, h)
	}
}

func TestTilingClone(t *testing.T) {
	original := NewTiling("content",
		WithSize(20, 30),
		WithSpacing(5, 10),
		WithTilingRelative(RelativeSelf),
	)

	clone := original.Clone()

	if clone == original {
		t.Error("clone should be a different instance")
	}
	if clone.Size == original.Size {
		t.Error("clone size should be a different array")
	}
	if clone.Size[0] != original.Size[0] || clone.Size[1] != original.Size[1] {
		t.Error("clone size values should match original")
	}
	if clone.Spacing != original.Spacing {
		t.Error("clone spacing should match original")
	}
	if clone.Relative != original.Relative {
		t.Error("clone relative should match original")
	}
}

func TestTilingString(t *testing.T) {
	autoTiling := NewTiling("content")
	if autoTiling.String() != "tiling(size: auto)" {
		t.Errorf("unexpected string: %s", autoTiling.String())
	}

	sizedTiling := NewTiling("content", WithSize(20, 30))
	expected := "tiling(size: (20.0pt, 30.0pt))"
	if sizedTiling.String() != expected {
		t.Errorf("expected %q, got %q", expected, sizedTiling.String())
	}
}

func TestTilingType(t *testing.T) {
	tiling := NewTiling("content")
	if tiling.Type() != "tiling" {
		t.Errorf("expected type 'tiling', got %q", tiling.Type())
	}
}

func TestNilTilingMethods(t *testing.T) {
	var tiling *Tiling = nil

	if tiling.GetSize() != nil {
		t.Error("expected nil size for nil tiling")
	}
	if tiling.GetSpacing() != ([2]Length{0, 0}) {
		t.Error("expected zero spacing for nil tiling")
	}
	if tiling.GetRelative() != RelativeAuto {
		t.Error("expected auto relative for nil tiling")
	}
	if tiling.GetBody() != nil {
		t.Error("expected nil body for nil tiling")
	}
	if !tiling.IsAuto() {
		t.Error("expected IsAuto true for nil tiling")
	}
	if tiling.CellWidth() != 0 {
		t.Error("expected zero cell width for nil tiling")
	}
	if tiling.CellHeight() != 0 {
		t.Error("expected zero cell height for nil tiling")
	}
	if tiling.HorizontalSpacing() != 0 {
		t.Error("expected zero horizontal spacing for nil tiling")
	}
	if tiling.VerticalSpacing() != 0 {
		t.Error("expected zero vertical spacing for nil tiling")
	}
	w, h := tiling.EffectiveCellSize()
	if w != 0 || h != 0 {
		t.Error("expected zero effective size for nil tiling")
	}
	if tiling.Clone() != nil {
		t.Error("expected nil clone for nil tiling")
	}
	if tiling.String() != "tiling()" {
		t.Errorf("unexpected string for nil tiling: %s", tiling.String())
	}
}
