package inline

import (
	"testing"
)

// TestDefaultMathConstants verifies the default math constants are reasonable.
func TestDefaultMathConstants(t *testing.T) {
	c := DefaultMathConstants()

	// Script scale should be between 0 and 1
	if c.ScriptPercentScaleDown <= 0 || c.ScriptPercentScaleDown >= 1 {
		t.Errorf("ScriptPercentScaleDown = %v, want between 0 and 1", c.ScriptPercentScaleDown)
	}

	// ScriptScript should be smaller than Script
	if c.ScriptScriptPercentScaleDown >= c.ScriptPercentScaleDown {
		t.Errorf("ScriptScriptPercentScaleDown = %v >= ScriptPercentScaleDown = %v",
			c.ScriptScriptPercentScaleDown, c.ScriptPercentScaleDown)
	}

	// Superscript should shift up
	if c.SuperscriptShiftUp <= 0 {
		t.Errorf("SuperscriptShiftUp = %v, want > 0", c.SuperscriptShiftUp)
	}

	// Subscript should shift down
	if c.SubscriptShiftDown <= 0 {
		t.Errorf("SubscriptShiftDown = %v, want > 0", c.SubscriptShiftDown)
	}
}

// TestMathContext verifies the math context calculations.
func TestMathContext(t *testing.T) {
	fontSize := Abs(12.0) // 12pt font
	ctx := NewMathContext(fontSize)

	// Script font size should be smaller
	scriptSize := ctx.ScriptFontSize()
	if scriptSize >= fontSize {
		t.Errorf("ScriptFontSize = %v, want < %v", scriptSize, fontSize)
	}

	// ScriptScript font size should be smaller than script
	scriptScriptSize := ctx.ScriptScriptFontSize()
	if scriptScriptSize >= scriptSize {
		t.Errorf("ScriptScriptFontSize = %v, want < %v", scriptScriptSize, scriptSize)
	}

	// Expected values
	expectedScript := fontSize * Abs(ctx.Constants.ScriptPercentScaleDown)
	if scriptSize != expectedScript {
		t.Errorf("ScriptFontSize = %v, want %v", scriptSize, expectedScript)
	}
}

// TestLayoutMathScriptNilBase verifies nil base handling.
func TestLayoutMathScriptNilBase(t *testing.T) {
	ctx := NewMathContext(12.0)
	result := LayoutMathScript(ctx, nil, nil, nil, 0)
	if result != nil {
		t.Error("LayoutMathScript(nil base) should return nil")
	}
}

// TestLayoutMathScriptBaseOnly verifies base-only layout.
func TestLayoutMathScriptBaseOnly(t *testing.T) {
	ctx := NewMathContext(12.0)

	base := &FinalFrame{
		Size:     FinalSize{Width: 10, Height: 12},
		Baseline: 9,
	}

	result := LayoutMathScript(ctx, base, nil, nil, 0)
	if result == nil {
		t.Fatal("LayoutMathScript returned nil for valid base")
	}

	// Width should equal base width
	if result.Size.Width != base.Size.Width {
		t.Errorf("Width = %v, want %v", result.Size.Width, base.Size.Width)
	}

	// Height should at least equal base height
	if result.Size.Height < base.Size.Height {
		t.Errorf("Height = %v, want >= %v", result.Size.Height, base.Size.Height)
	}
}

// TestLayoutMathScriptWithSuperscript verifies superscript positioning.
func TestLayoutMathScriptWithSuperscript(t *testing.T) {
	ctx := NewMathContext(12.0)

	base := &FinalFrame{
		Size:     FinalSize{Width: 10, Height: 12},
		Baseline: 9,
	}

	super := &FinalFrame{
		Size:     FinalSize{Width: 6, Height: 8},
		Baseline: 6,
	}

	result := LayoutMathScript(ctx, base, super, nil, 0)
	if result == nil {
		t.Fatal("LayoutMathScript returned nil")
	}

	// Width should be base + script width
	expectedWidth := base.Size.Width + super.Size.Width
	if result.Size.Width != expectedWidth {
		t.Errorf("Width = %v, want %v", result.Size.Width, expectedWidth)
	}

	// Height should accommodate raised superscript
	if result.Size.Height <= base.Size.Height {
		t.Errorf("Height = %v, want > %v (for raised superscript)", result.Size.Height, base.Size.Height)
	}
}

// TestLayoutMathScriptWithSubscript verifies subscript positioning.
func TestLayoutMathScriptWithSubscript(t *testing.T) {
	ctx := NewMathContext(12.0)

	base := &FinalFrame{
		Size:     FinalSize{Width: 10, Height: 12},
		Baseline: 9,
	}

	sub := &FinalFrame{
		Size:     FinalSize{Width: 6, Height: 8},
		Baseline: 6,
	}

	result := LayoutMathScript(ctx, base, nil, sub, 0)
	if result == nil {
		t.Fatal("LayoutMathScript returned nil")
	}

	// Width should be base + script width
	expectedWidth := base.Size.Width + sub.Size.Width
	if result.Size.Width != expectedWidth {
		t.Errorf("Width = %v, want %v", result.Size.Width, expectedWidth)
	}
}

// TestLayoutMathScriptWithBoth verifies both superscript and subscript positioning.
func TestLayoutMathScriptWithBoth(t *testing.T) {
	ctx := NewMathContext(12.0)

	base := &FinalFrame{
		Size:     FinalSize{Width: 10, Height: 12},
		Baseline: 9,
	}

	super := &FinalFrame{
		Size:     FinalSize{Width: 6, Height: 8},
		Baseline: 6,
	}

	sub := &FinalFrame{
		Size:     FinalSize{Width: 5, Height: 8},
		Baseline: 6,
	}

	result := LayoutMathScript(ctx, base, super, sub, 0)
	if result == nil {
		t.Fatal("LayoutMathScript returned nil")
	}

	// Width should be base + max(super, sub) width
	maxScriptWidth := super.Size.Width
	if sub.Size.Width > maxScriptWidth {
		maxScriptWidth = sub.Size.Width
	}
	expectedWidth := base.Size.Width + maxScriptWidth
	if result.Size.Width != expectedWidth {
		t.Errorf("Width = %v, want %v", result.Size.Width, expectedWidth)
	}
}

// TestLayoutMathScriptWithPrimes verifies prime marks add width.
func TestLayoutMathScriptWithPrimes(t *testing.T) {
	ctx := NewMathContext(12.0)

	base := &FinalFrame{
		Size:     FinalSize{Width: 10, Height: 12},
		Baseline: 9,
	}

	// Without primes
	resultNoPrimes := LayoutMathScript(ctx, base, nil, nil, 0)
	// With 2 primes
	resultWithPrimes := LayoutMathScript(ctx, base, nil, nil, 2)

	if resultWithPrimes.Size.Width <= resultNoPrimes.Size.Width {
		t.Error("Width should increase with prime marks")
	}
}

// TestLayoutMathLimitsNilNucleus verifies nil nucleus handling.
func TestLayoutMathLimitsNilNucleus(t *testing.T) {
	ctx := NewMathContext(12.0)
	result := LayoutMathLimits(ctx, nil, nil, nil)
	if result != nil {
		t.Error("LayoutMathLimits(nil nucleus) should return nil")
	}
}

// TestLayoutMathLimitsNucleusOnly verifies nucleus-only layout.
func TestLayoutMathLimitsNucleusOnly(t *testing.T) {
	ctx := NewMathContext(12.0)

	nucleus := &FinalFrame{
		Size:     FinalSize{Width: 15, Height: 20},
		Baseline: 15,
	}

	result := LayoutMathLimits(ctx, nucleus, nil, nil)
	if result == nil {
		t.Fatal("LayoutMathLimits returned nil for valid nucleus")
	}

	// Width should equal nucleus width
	if result.Size.Width != nucleus.Size.Width {
		t.Errorf("Width = %v, want %v", result.Size.Width, nucleus.Size.Width)
	}

	// Height should equal nucleus height
	if result.Size.Height != nucleus.Size.Height {
		t.Errorf("Height = %v, want %v", result.Size.Height, nucleus.Size.Height)
	}
}

// TestLayoutMathLimitsWithUpper verifies upper limit positioning.
func TestLayoutMathLimitsWithUpper(t *testing.T) {
	ctx := NewMathContext(12.0)

	nucleus := &FinalFrame{
		Size:     FinalSize{Width: 15, Height: 20},
		Baseline: 15,
	}

	upper := &FinalFrame{
		Size:     FinalSize{Width: 10, Height: 8},
		Baseline: 6,
	}

	result := LayoutMathLimits(ctx, nucleus, upper, nil)
	if result == nil {
		t.Fatal("LayoutMathLimits returned nil")
	}

	// Width should be max of nucleus and upper
	expectedWidth := nucleus.Size.Width
	if upper.Size.Width > expectedWidth {
		expectedWidth = upper.Size.Width
	}
	if result.Size.Width != expectedWidth {
		t.Errorf("Width = %v, want %v", result.Size.Width, expectedWidth)
	}

	// Height should include upper + gap + nucleus
	if result.Size.Height <= nucleus.Size.Height+upper.Size.Height {
		t.Errorf("Height = %v, want > %v", result.Size.Height, nucleus.Size.Height+upper.Size.Height)
	}
}

// TestLayoutMathLimitsWithLower verifies lower limit positioning.
func TestLayoutMathLimitsWithLower(t *testing.T) {
	ctx := NewMathContext(12.0)

	nucleus := &FinalFrame{
		Size:     FinalSize{Width: 15, Height: 20},
		Baseline: 15,
	}

	lower := &FinalFrame{
		Size:     FinalSize{Width: 10, Height: 8},
		Baseline: 6,
	}

	result := LayoutMathLimits(ctx, nucleus, nil, lower)
	if result == nil {
		t.Fatal("LayoutMathLimits returned nil")
	}

	// Height should include nucleus + gap + lower
	if result.Size.Height <= nucleus.Size.Height+lower.Size.Height {
		t.Errorf("Height = %v, want > %v", result.Size.Height, nucleus.Size.Height+lower.Size.Height)
	}
}

// TestLayoutMathLimitsWithBoth verifies both limits positioning.
func TestLayoutMathLimitsWithBoth(t *testing.T) {
	ctx := NewMathContext(12.0)

	nucleus := &FinalFrame{
		Size:     FinalSize{Width: 15, Height: 20},
		Baseline: 15,
	}

	upper := &FinalFrame{
		Size:     FinalSize{Width: 20, Height: 8}, // Wider than nucleus
		Baseline: 6,
	}

	lower := &FinalFrame{
		Size:     FinalSize{Width: 10, Height: 8},
		Baseline: 6,
	}

	result := LayoutMathLimits(ctx, nucleus, upper, lower)
	if result == nil {
		t.Fatal("LayoutMathLimits returned nil")
	}

	// Width should be max of all three
	expectedWidth := upper.Size.Width // upper is widest
	if result.Size.Width != expectedWidth {
		t.Errorf("Width = %v, want %v", result.Size.Width, expectedWidth)
	}

	// Height should include upper + gaps + nucleus + lower
	minHeight := upper.Size.Height + nucleus.Size.Height + lower.Size.Height
	if result.Size.Height <= minHeight {
		t.Errorf("Height = %v, want > %v", result.Size.Height, minHeight)
	}
}

// TestIsLargeOperator verifies large operator detection.
func TestIsLargeOperator(t *testing.T) {
	largeOps := []string{"∑", "∏", "∐", "⋀", "⋁", "⋂", "⋃", "lim", "sup", "inf", "max", "min"}
	for _, op := range largeOps {
		if !IsLargeOperator(op) {
			t.Errorf("IsLargeOperator(%q) = false, want true", op)
		}
	}

	notLargeOps := []string{"+", "-", "=", "x", "a", "sin", "cos"}
	for _, op := range notLargeOps {
		if IsLargeOperator(op) {
			t.Errorf("IsLargeOperator(%q) = true, want false", op)
		}
	}
}

// TestMathContextCrampedMode verifies cramped mode affects superscript position.
func TestMathContextCrampedMode(t *testing.T) {
	ctx := NewMathContext(12.0)
	normalShift := ctx.Constants.SuperscriptShiftUp

	ctx.Cramped = true
	crampedShift := ctx.Constants.SuperscriptShiftUpCramped

	if crampedShift >= normalShift {
		t.Errorf("Cramped superscript shift %v should be less than normal %v",
			crampedShift, normalShift)
	}
}
