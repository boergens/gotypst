package math

import (
	"math"
	"testing"

	"github.com/boergens/gotypst/eval"
)

// approxEqual checks if two Abs values are approximately equal
// within floating point tolerance.
func approxEqual(a, b Abs) bool {
	const epsilon = 1e-9
	return math.Abs(float64(a-b)) < epsilon
}

func TestLayoutFrac(t *testing.T) {
	// Create a simple fraction: a/b
	frac := &eval.MathFracElement{
		Num: eval.Content{
			Elements: []eval.ContentElement{
				&eval.TextElement{Text: "a"},
			},
		},
		Denom: eval.Content{
			Elements: []eval.ContentElement{
				&eval.TextElement{Text: "b"},
			},
		},
	}

	ctx := &MathContext{
		FontSize: Abs(12),
		Style:    StyleDisplay,
		Cramped:  false,
	}
	constants := DefaultMathConstants()

	frame := LayoutFrac(frac, ctx, constants)

	// Verify frame has content
	if frame == nil {
		t.Fatal("LayoutFrac returned nil")
	}

	// Verify frame has items (numerator, fraction bar, denominator)
	if len(frame.Items) < 3 {
		t.Errorf("expected at least 3 items (num, bar, denom), got %d", len(frame.Items))
	}

	// Verify frame dimensions are positive
	if frame.Width() <= 0 {
		t.Errorf("expected positive width, got %v", frame.Width())
	}
	if frame.Height() <= 0 {
		t.Errorf("expected positive height, got %v", frame.Height())
	}

	// Verify baseline is within the frame
	if frame.Baseline < 0 || frame.Baseline > frame.Height() {
		t.Errorf("baseline %v should be between 0 and height %v", frame.Baseline, frame.Height())
	}
}

func TestLayoutFracNested(t *testing.T) {
	// Create a nested fraction: (a/b) / c
	innerFrac := &eval.MathFracElement{
		Num: eval.Content{
			Elements: []eval.ContentElement{
				&eval.TextElement{Text: "a"},
			},
		},
		Denom: eval.Content{
			Elements: []eval.ContentElement{
				&eval.TextElement{Text: "b"},
			},
		},
	}

	outerFrac := &eval.MathFracElement{
		Num: eval.Content{
			Elements: []eval.ContentElement{
				innerFrac,
			},
		},
		Denom: eval.Content{
			Elements: []eval.ContentElement{
				&eval.TextElement{Text: "c"},
			},
		},
	}

	ctx := &MathContext{
		FontSize: Abs(12),
		Style:    StyleDisplay,
		Cramped:  false,
	}
	constants := DefaultMathConstants()

	frame := LayoutFrac(outerFrac, ctx, constants)

	if frame == nil {
		t.Fatal("LayoutFrac returned nil for nested fraction")
	}

	// Nested fraction should be taller than a simple fraction
	simpleFrac := &eval.MathFracElement{
		Num: eval.Content{
			Elements: []eval.ContentElement{
				&eval.TextElement{Text: "x"},
			},
		},
		Denom: eval.Content{
			Elements: []eval.ContentElement{
				&eval.TextElement{Text: "y"},
			},
		},
	}
	simpleFrame := LayoutFrac(simpleFrac, ctx, constants)

	if frame.Height() <= simpleFrame.Height() {
		t.Errorf("nested fraction height %v should be greater than simple fraction height %v",
			frame.Height(), simpleFrame.Height())
	}
}

func TestLayoutFracDisplayVsInline(t *testing.T) {
	frac := &eval.MathFracElement{
		Num: eval.Content{
			Elements: []eval.ContentElement{
				&eval.TextElement{Text: "a"},
			},
		},
		Denom: eval.Content{
			Elements: []eval.ContentElement{
				&eval.TextElement{Text: "b"},
			},
		},
	}

	constants := DefaultMathConstants()

	// Display style
	displayCtx := &MathContext{
		FontSize: Abs(12),
		Style:    StyleDisplay,
		Cramped:  false,
	}
	displayFrame := LayoutFrac(frac, displayCtx, constants)

	// Text (inline) style
	textCtx := &MathContext{
		FontSize: Abs(12),
		Style:    StyleText,
		Cramped:  false,
	}
	textFrame := LayoutFrac(frac, textCtx, constants)

	// Both should produce valid frames
	if displayFrame == nil || textFrame == nil {
		t.Fatal("one of the frames is nil")
	}

	// Display style typically has larger gaps, so should be taller
	// (This depends on the constants configuration)
	if displayFrame.Height() < textFrame.Height() {
		t.Logf("Note: display frame height %v, text frame height %v",
			displayFrame.Height(), textFrame.Height())
	}
}

func TestLayoutFracEmptyContent(t *testing.T) {
	// Fraction with empty numerator
	frac := &eval.MathFracElement{
		Num: eval.Content{
			Elements: []eval.ContentElement{},
		},
		Denom: eval.Content{
			Elements: []eval.ContentElement{
				&eval.TextElement{Text: "b"},
			},
		},
	}

	ctx := &MathContext{
		FontSize: Abs(12),
		Style:    StyleDisplay,
		Cramped:  false,
	}
	constants := DefaultMathConstants()

	frame := LayoutFrac(frac, ctx, constants)

	// Should not panic and should return valid frame
	if frame == nil {
		t.Fatal("LayoutFrac returned nil for fraction with empty numerator")
	}

	// Should still have a fraction bar
	hasLine := false
	for _, item := range frame.Items {
		if _, ok := item.Item.(LineItem); ok {
			hasLine = true
			break
		}
	}
	if !hasLine {
		t.Error("fraction should have a fraction bar even with empty numerator")
	}
}

func TestLayoutText(t *testing.T) {
	text := &eval.TextElement{Text: "xyz"}
	ctx := &MathContext{
		FontSize: Abs(12),
		Style:    StyleText,
		Cramped:  false,
	}

	frame := LayoutText(text, ctx)

	if frame == nil {
		t.Fatal("LayoutText returned nil")
	}

	// Verify dimensions
	if frame.Width() <= 0 {
		t.Errorf("expected positive width, got %v", frame.Width())
	}
	if frame.Height() <= 0 {
		t.Errorf("expected positive height, got %v", frame.Height())
	}

	// Should have one text item
	if len(frame.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(frame.Items))
	}

	// Verify it's a text item with correct content
	if textItem, ok := frame.Items[0].Item.(TextItem); ok {
		if textItem.Text != "xyz" {
			t.Errorf("expected text 'xyz', got '%s'", textItem.Text)
		}
	} else {
		t.Errorf("expected TextItem, got %T", frame.Items[0].Item)
	}
}

func TestLayoutSymbol(t *testing.T) {
	symbol := &eval.MathSymbolElement{Symbol: "alpha"}
	ctx := &MathContext{
		FontSize: Abs(12),
		Style:    StyleText,
		Cramped:  false,
	}

	frame := LayoutSymbol(symbol, ctx)

	if frame == nil {
		t.Fatal("LayoutSymbol returned nil")
	}

	// Should have one text item
	if len(frame.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(frame.Items))
	}
}

func TestLayoutHorizontal(t *testing.T) {
	elements := []eval.ContentElement{
		&eval.TextElement{Text: "a"},
		&eval.MathSymbolElement{Symbol: "+"},
		&eval.TextElement{Text: "b"},
	}

	ctx := &MathContext{
		FontSize: Abs(12),
		Style:    StyleText,
		Cramped:  false,
	}
	constants := DefaultMathConstants()

	frame := LayoutHorizontal(elements, ctx, constants)

	if frame == nil {
		t.Fatal("LayoutHorizontal returned nil")
	}

	// Should have 3 child frames
	if len(frame.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(frame.Items))
	}

	// Width should be sum of individual widths
	var expectedWidth Abs
	for _, elem := range elements {
		f := LayoutElement(elem, ctx, constants)
		expectedWidth += f.Width()
	}

	if frame.Width() != expectedWidth {
		t.Errorf("expected width %v, got %v", expectedWidth, frame.Width())
	}
}

func TestLayoutAttach(t *testing.T) {
	// x^2
	attach := &eval.MathAttachElement{
		Base: eval.Content{
			Elements: []eval.ContentElement{
				&eval.TextElement{Text: "x"},
			},
		},
		Superscript: eval.Content{
			Elements: []eval.ContentElement{
				&eval.TextElement{Text: "2"},
			},
		},
	}

	ctx := &MathContext{
		FontSize: Abs(12),
		Style:    StyleText,
		Cramped:  false,
	}
	constants := DefaultMathConstants()

	frame := LayoutAttach(attach, ctx, constants)

	if frame == nil {
		t.Fatal("LayoutAttach returned nil")
	}

	// Should have at least 2 items (base and superscript)
	if len(frame.Items) < 2 {
		t.Errorf("expected at least 2 items, got %d", len(frame.Items))
	}

	// Width should be greater than base alone
	baseFrame := LayoutText(&eval.TextElement{Text: "x"}, ctx)
	if frame.Width() <= baseFrame.Width() {
		t.Errorf("attach frame width %v should be greater than base width %v",
			frame.Width(), baseFrame.Width())
	}
}

func TestLayoutRoot(t *testing.T) {
	// sqrt(x)
	root := &eval.MathRootElement{
		Radicand: eval.Content{
			Elements: []eval.ContentElement{
				&eval.TextElement{Text: "x"},
			},
		},
	}

	ctx := &MathContext{
		FontSize: Abs(12),
		Style:    StyleText,
		Cramped:  false,
	}
	constants := DefaultMathConstants()

	frame := LayoutRoot(root, ctx, constants)

	if frame == nil {
		t.Fatal("LayoutRoot returned nil")
	}

	// Should have items for root symbol, radicand, and overline
	if len(frame.Items) < 3 {
		t.Errorf("expected at least 3 items, got %d", len(frame.Items))
	}
}

func TestLayoutDelimited(t *testing.T) {
	// (a + b)
	delim := &eval.MathDelimitedElement{
		Open:  "(",
		Close: ")",
		Body: eval.Content{
			Elements: []eval.ContentElement{
				&eval.TextElement{Text: "a"},
				&eval.MathSymbolElement{Symbol: "+"},
				&eval.TextElement{Text: "b"},
			},
		},
	}

	ctx := &MathContext{
		FontSize: Abs(12),
		Style:    StyleText,
		Cramped:  false,
	}
	constants := DefaultMathConstants()

	frame := LayoutDelimited(delim, ctx, constants)

	if frame == nil {
		t.Fatal("LayoutDelimited returned nil")
	}

	// Should have items for open paren, body, and close paren
	if len(frame.Items) < 3 {
		t.Errorf("expected at least 3 items, got %d", len(frame.Items))
	}
}

func TestLayoutEquation(t *testing.T) {
	// Block equation: $x/y$
	equation := &eval.EquationElement{
		Body: eval.Content{
			Elements: []eval.ContentElement{
				&eval.MathFracElement{
					Num: eval.Content{
						Elements: []eval.ContentElement{
							&eval.TextElement{Text: "x"},
						},
					},
					Denom: eval.Content{
						Elements: []eval.ContentElement{
							&eval.TextElement{Text: "y"},
						},
					},
				},
			},
		},
		Block: true,
	}

	frame := LayoutEquation(equation, Abs(12))

	if frame == nil {
		t.Fatal("LayoutEquation returned nil")
	}

	// Should have produced a valid fraction layout
	if frame.Width() <= 0 {
		t.Errorf("expected positive width, got %v", frame.Width())
	}
	if frame.Height() <= 0 {
		t.Errorf("expected positive height, got %v", frame.Height())
	}
}

func TestLayoutEquationWithResult(t *testing.T) {
	equation := &eval.EquationElement{
		Body: eval.Content{
			Elements: []eval.ContentElement{
				&eval.TextElement{Text: "x"},
			},
		},
		Block: false,
	}

	result := LayoutEquationWithResult(equation, Abs(11))

	if result == nil {
		t.Fatal("LayoutEquationWithResult returned nil")
	}

	if result.Frame == nil {
		t.Error("result.Frame is nil")
	}

	if result.FontSize != Abs(11) {
		t.Errorf("expected FontSize 11, got %v", result.FontSize)
	}

	if result.IsBlock != false {
		t.Error("expected IsBlock to be false for inline equation")
	}
}

func TestMathConstants(t *testing.T) {
	constants := DefaultMathConstants()

	// Verify constants are positive
	if constants.AxisHeight <= 0 {
		t.Errorf("AxisHeight should be positive, got %v", constants.AxisHeight)
	}
	if constants.FractionRuleThickness <= 0 {
		t.Errorf("FractionRuleThickness should be positive, got %v", constants.FractionRuleThickness)
	}
	if constants.FractionNumeratorGapMin <= 0 {
		t.Errorf("FractionNumeratorGapMin should be positive, got %v", constants.FractionNumeratorGapMin)
	}
	if constants.FractionDenominatorGapMin <= 0 {
		t.Errorf("FractionDenominatorGapMin should be positive, got %v", constants.FractionDenominatorGapMin)
	}
}

func TestMathStyleScriptStyle(t *testing.T) {
	tests := []struct {
		input    MathStyle
		expected MathStyle
	}{
		{StyleDisplay, StyleScript},
		{StyleText, StyleScript},
		{StyleScript, StyleScriptScript},
		{StyleScriptScript, StyleScriptScript},
	}

	for _, tt := range tests {
		result := tt.input.ScriptStyle()
		if result != tt.expected {
			t.Errorf("ScriptStyle(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestMathContextFontSizeForStyle(t *testing.T) {
	ctx := &MathContext{
		FontSize: Abs(12),
		Style:    StyleDisplay,
	}

	// Display and text styles should return base font size
	if size := ctx.FontSizeForStyle(StyleDisplay); size != 12 {
		t.Errorf("StyleDisplay font size = %v, want 12", size)
	}
	if size := ctx.FontSizeForStyle(StyleText); size != 12 {
		t.Errorf("StyleText font size = %v, want 12", size)
	}

	// Script style should be 70% of base (use approximate comparison for floats)
	scriptSize := ctx.FontSizeForStyle(StyleScript)
	expectedScript := Abs(12 * 0.7)
	if !approxEqual(scriptSize, expectedScript) {
		t.Errorf("StyleScript font size = %v, want %v", scriptSize, expectedScript)
	}

	// ScriptScript style should be 50% of base
	scriptScriptSize := ctx.FontSizeForStyle(StyleScriptScript)
	expectedScriptScript := Abs(12 * 0.5)
	if !approxEqual(scriptScriptSize, expectedScriptScript) {
		t.Errorf("StyleScriptScript font size = %v, want %v", scriptScriptSize, expectedScriptScript)
	}
}
