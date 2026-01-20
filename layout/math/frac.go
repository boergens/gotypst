package math

import (
	"github.com/boergens/gotypst/eval"
)

// LayoutFrac lays out a fraction element.
//
// The fraction layout consists of:
//   - Numerator positioned above the fraction bar
//   - Fraction bar (horizontal line) at the math axis
//   - Denominator positioned below the fraction bar
//
// The baseline of the resulting frame is at the math axis (center of fraction bar).
func LayoutFrac(elem *eval.MathFracElement, ctx *MathContext, constants MathConstants) *MathFrame {
	fontSize := ctx.FontSizeForStyle(ctx.Style)

	// Layout numerator and denominator
	numFrame := LayoutContent(&elem.Num, ctx, constants)
	denomFrame := LayoutContent(&elem.Denom, ctx, constants)

	// Calculate the width (max of numerator/denominator, with some padding)
	const horizontalPadding = 0.1 // 10% padding on each side
	numWidth := numFrame.Width()
	denomWidth := denomFrame.Width()
	contentWidth := numWidth
	if denomWidth > contentWidth {
		contentWidth = denomWidth
	}
	padding := Em(horizontalPadding).At(fontSize)
	barWidth := contentWidth + 2*padding

	// Get constants at current font size
	axisHeight := constants.AxisHeight.At(fontSize)
	ruleThickness := constants.FractionRuleThickness.At(fontSize)

	// Calculate vertical positioning based on math style
	var numShift, denomShift Abs
	if ctx.Style == StyleDisplay {
		// Display style: larger gaps
		numShift = constants.FractionNumeratorShiftUp.At(fontSize)
		denomShift = constants.FractionDenominatorShiftDown.At(fontSize)
	} else {
		// Text/script styles: smaller gaps
		numShift = constants.StackTopShiftUp.At(fontSize)
		denomShift = constants.StackBottomShiftDown.At(fontSize)
	}

	// Ensure minimum gaps
	numGapMin := constants.FractionNumeratorGapMin.At(fontSize)
	denomGapMin := constants.FractionDenominatorGapMin.At(fontSize)

	// The fraction bar is centered on the axis
	// Numerator bottom should be above the bar top
	// Denominator top should be below the bar bottom
	barTop := axisHeight + ruleThickness/2
	barBottom := axisHeight - ruleThickness/2

	// Calculate numerator position
	// The baseline of the numerator should be at numShift above the axis
	// But we need to ensure minimum gap from numerator bottom to bar top
	numBaseline := numFrame.Baseline
	numHeight := numFrame.Height()
	numBottom := numHeight - numBaseline // Distance from baseline to bottom

	// Position numerator so its bottom is at least numGapMin above bar top
	numY := barTop + numGapMin + numBottom
	if numY < numShift {
		numY = numShift + numBottom
	}

	// Calculate denominator position
	// Position denominator so its top is at least denomGapMin below bar bottom
	denomBaseline := denomFrame.Baseline

	// The denominator's top (y=0) should be at barBottom - denomGapMin
	// But we're measuring from the frame's baseline (math axis)
	// So denominator's y position is -(barBottom - denomGapMin)
	denomY := barBottom - denomGapMin
	if denomY > -denomShift {
		denomY = -denomShift
	}

	// Total height: from numerator top to denominator bottom
	totalTop := numY + numBaseline           // Distance from axis to numerator top
	totalBottom := -denomY + denomFrame.Height() - denomBaseline // Distance from axis to denominator bottom

	totalHeight := totalTop + totalBottom

	// Create the result frame
	// Baseline is at the axis, which is totalTop from the top
	frame := &MathFrame{
		Size: Size{
			Width:  barWidth,
			Height: totalHeight,
		},
		Baseline: totalTop, // Distance from top to axis
	}

	// Position numerator (centered horizontally)
	numX := (barWidth - numWidth) / 2
	// numY is the distance from axis to numerator bottom
	// Frame y=0 is at top, so numerator top is at totalTop - numY - numHeight + numBottom
	// Simplify: numerator y = totalTop - numY - numBaseline
	numFrameY := totalTop - numY - numBaseline
	frame.PushFrame(Point{X: numX, Y: numFrameY}, numFrame)

	// Position fraction bar (horizontal line at the axis)
	// The bar is drawn from its left edge
	barX := Abs(0)
	barY := totalTop - axisHeight // Y position of the bar center
	frame.Push(Point{X: barX, Y: barY}, LineItem{
		Length:    barWidth,
		Thickness: ruleThickness,
	})

	// Position denominator (centered horizontally)
	denomX := (barWidth - denomWidth) / 2
	// denomY is the distance from axis to denominator top (negative value)
	// Denominator top y = totalTop - denomY
	denomFrameY := totalTop - denomY
	frame.PushFrame(Point{X: denomX, Y: denomFrameY}, denomFrame)

	return frame
}

// LayoutContent lays out math content recursively.
// This is the main entry point for laying out arbitrary math content.
func LayoutContent(content *eval.Content, ctx *MathContext, constants MathConstants) *MathFrame {
	if content == nil || len(content.Elements) == 0 {
		// Return empty frame
		return &MathFrame{
			Size:     Size{Width: 0, Height: 0},
			Baseline: 0,
		}
	}

	// For a single element, layout just that element
	if len(content.Elements) == 1 {
		return LayoutElement(content.Elements[0], ctx, constants)
	}

	// For multiple elements, arrange them horizontally
	return LayoutHorizontal(content.Elements, ctx, constants)
}

// LayoutHorizontal arranges multiple elements horizontally.
func LayoutHorizontal(elements []eval.ContentElement, ctx *MathContext, constants MathConstants) *MathFrame {
	if len(elements) == 0 {
		return &MathFrame{}
	}

	// Layout each element
	frames := make([]*MathFrame, len(elements))
	for i, elem := range elements {
		frames[i] = LayoutElement(elem, ctx, constants)
	}

	// Calculate total width and find max baseline/height
	var totalWidth Abs
	var maxAboveBaseline Abs
	var maxBelowBaseline Abs

	for _, f := range frames {
		totalWidth += f.Width()
		aboveBaseline := f.Baseline
		belowBaseline := f.Height() - f.Baseline
		if aboveBaseline > maxAboveBaseline {
			maxAboveBaseline = aboveBaseline
		}
		if belowBaseline > maxBelowBaseline {
			maxBelowBaseline = belowBaseline
		}
	}

	totalHeight := maxAboveBaseline + maxBelowBaseline

	// Create result frame
	result := &MathFrame{
		Size: Size{
			Width:  totalWidth,
			Height: totalHeight,
		},
		Baseline: maxAboveBaseline,
	}

	// Position each frame, aligned on baseline
	x := Abs(0)
	for _, f := range frames {
		y := maxAboveBaseline - f.Baseline
		result.PushFrame(Point{X: x, Y: y}, f)
		x += f.Width()
	}

	return result
}

// LayoutElement lays out a single content element.
func LayoutElement(elem eval.ContentElement, ctx *MathContext, constants MathConstants) *MathFrame {
	if elem == nil {
		return &MathFrame{}
	}

	switch e := elem.(type) {
	case *eval.MathFracElement:
		return LayoutFrac(e, ctx, constants)
	case *eval.TextElement:
		return LayoutText(e, ctx)
	case *eval.MathSymbolElement:
		return LayoutSymbol(e, ctx)
	case *eval.MathAttachElement:
		return LayoutAttach(e, ctx, constants)
	case *eval.MathRootElement:
		return LayoutRoot(e, ctx, constants)
	case *eval.MathDelimitedElement:
		return LayoutDelimited(e, ctx, constants)
	default:
		// For unknown elements, return empty frame
		return &MathFrame{}
	}
}

// LayoutText lays out a text element.
func LayoutText(elem *eval.TextElement, ctx *MathContext) *MathFrame {
	fontSize := ctx.FontSizeForStyle(ctx.Style)

	// Approximate text width (would use actual shaping in production)
	// Use approximately 0.5em per character for math text
	charWidth := Em(0.5).At(fontSize)
	width := charWidth * Abs(len(elem.Text))

	// Height is approximately the font size
	height := fontSize

	// Baseline at ~80% from top
	baseline := height * 0.8

	frame := &MathFrame{
		Size: Size{
			Width:  width,
			Height: height,
		},
		Baseline: baseline,
	}

	frame.Push(Point{X: 0, Y: 0}, TextItem{
		Text:     elem.Text,
		FontSize: fontSize,
	})

	return frame
}

// LayoutSymbol lays out a math symbol element.
func LayoutSymbol(elem *eval.MathSymbolElement, ctx *MathContext) *MathFrame {
	fontSize := ctx.FontSizeForStyle(ctx.Style)

	// Approximate symbol width
	charWidth := Em(0.5).At(fontSize)
	width := charWidth * Abs(len(elem.Symbol))

	height := fontSize
	baseline := height * 0.8

	frame := &MathFrame{
		Size: Size{
			Width:  width,
			Height: height,
		},
		Baseline: baseline,
	}

	frame.Push(Point{X: 0, Y: 0}, TextItem{
		Text:     elem.Symbol,
		FontSize: fontSize,
	})

	return frame
}

// LayoutAttach lays out subscripts and superscripts.
func LayoutAttach(elem *eval.MathAttachElement, ctx *MathContext, constants MathConstants) *MathFrame {
	fontSize := ctx.FontSizeForStyle(ctx.Style)

	// Layout base
	baseFrame := LayoutContent(&elem.Base, ctx, constants)

	// Create context for scripts (smaller size)
	scriptCtx := &MathContext{
		FontSize: ctx.FontSize,
		Style:    ctx.Style.ScriptStyle(),
		Cramped:  ctx.Cramped,
	}

	// Layout subscript and superscript
	var subFrame, supFrame *MathFrame
	hasSubscript := len(elem.Subscript.Elements) > 0
	hasSuperscript := len(elem.Superscript.Elements) > 0

	if hasSubscript {
		subFrame = LayoutContent(&elem.Subscript, scriptCtx, constants)
	}
	if hasSuperscript {
		supFrame = LayoutContent(&elem.Superscript, scriptCtx, constants)
	}

	// Calculate positioning
	// Scripts are positioned to the right of the base
	scriptX := baseFrame.Width()

	// Superscript shift up and subscript shift down
	supShift := Em(0.4).At(fontSize) // Shift up from baseline
	subShift := Em(0.2).At(fontSize) // Shift down from baseline

	// Calculate total size
	width := scriptX
	if hasSuperscript && supFrame.Width() > 0 {
		width = scriptX + supFrame.Width()
	}
	if hasSubscript && subFrame.Width() > width-scriptX {
		width = scriptX + subFrame.Width()
	}

	// Height calculation
	topExtent := baseFrame.Baseline
	bottomExtent := baseFrame.Height() - baseFrame.Baseline

	if hasSuperscript {
		supTop := baseFrame.Baseline - supShift + supFrame.Baseline
		if supTop > topExtent {
			topExtent = supTop
		}
	}
	if hasSubscript {
		subBottom := (baseFrame.Height() - baseFrame.Baseline) + subShift + (subFrame.Height() - subFrame.Baseline)
		if subBottom > bottomExtent {
			bottomExtent = subBottom
		}
	}

	height := topExtent + bottomExtent
	baseline := topExtent

	// Create result frame
	frame := &MathFrame{
		Size: Size{
			Width:  width,
			Height: height,
		},
		Baseline: baseline,
	}

	// Position base
	baseY := baseline - baseFrame.Baseline
	frame.PushFrame(Point{X: 0, Y: baseY}, baseFrame)

	// Position superscript
	if hasSuperscript {
		supY := baseline - supShift - supFrame.Baseline
		frame.PushFrame(Point{X: scriptX, Y: supY}, supFrame)
	}

	// Position subscript
	if hasSubscript {
		subY := baseline + subShift - subFrame.Baseline + (baseFrame.Height() - baseFrame.Baseline - baseline + baseFrame.Baseline)
		// Simplify: position subscript below baseline
		subY = baseline + subShift
		frame.PushFrame(Point{X: scriptX, Y: subY}, subFrame)
	}

	return frame
}

// LayoutRoot lays out a root (square root, nth root) element.
func LayoutRoot(elem *eval.MathRootElement, ctx *MathContext, constants MathConstants) *MathFrame {
	fontSize := ctx.FontSizeForStyle(ctx.Style)

	// Layout radicand (content under the root)
	radicandFrame := LayoutContent(&elem.Radicand, ctx, constants)

	// Root symbol dimensions
	rootSymbolWidth := Em(0.5).At(fontSize)
	rootExtension := Em(0.1).At(fontSize) // Extra height above radicand

	// Calculate size
	width := rootSymbolWidth + radicandFrame.Width()
	height := radicandFrame.Height() + rootExtension
	baseline := radicandFrame.Baseline + rootExtension

	frame := &MathFrame{
		Size: Size{
			Width:  width,
			Height: height,
		},
		Baseline: baseline,
	}

	// Add root symbol (represented as text for now)
	frame.Push(Point{X: 0, Y: 0}, TextItem{
		Text:     "√",
		FontSize: fontSize,
	})

	// Position radicand
	frame.PushFrame(Point{X: rootSymbolWidth, Y: rootExtension}, radicandFrame)

	// Add overline above radicand
	overlineY := Abs(0)
	overlineThickness := constants.FractionRuleThickness.At(fontSize)
	frame.Push(Point{X: rootSymbolWidth, Y: overlineY}, LineItem{
		Length:    radicandFrame.Width(),
		Thickness: overlineThickness,
	})

	// Handle index (for nth roots) if present
	if len(elem.Index.Elements) > 0 {
		indexCtx := &MathContext{
			FontSize: ctx.FontSize,
			Style:    StyleScriptScript,
			Cramped:  true,
		}
		indexFrame := LayoutContent(&elem.Index, indexCtx, constants)
		// Position index in the "v" of the root symbol
		indexX := Abs(0)
		indexY := height * 0.3
		frame.PushFrame(Point{X: indexX, Y: indexY}, indexFrame)
	}

	return frame
}

// LayoutDelimited lays out delimited content (parentheses, brackets, etc.).
func LayoutDelimited(elem *eval.MathDelimitedElement, ctx *MathContext, constants MathConstants) *MathFrame {
	fontSize := ctx.FontSizeForStyle(ctx.Style)

	// Layout body content
	bodyFrame := LayoutContent(&elem.Body, ctx, constants)

	// Delimiter dimensions
	delimWidth := Em(0.3).At(fontSize)

	// Calculate total size
	width := delimWidth + bodyFrame.Width() + delimWidth
	height := bodyFrame.Height()
	baseline := bodyFrame.Baseline

	frame := &MathFrame{
		Size: Size{
			Width:  width,
			Height: height,
		},
		Baseline: baseline,
	}

	// Add opening delimiter
	frame.Push(Point{X: 0, Y: 0}, TextItem{
		Text:     elem.Open,
		FontSize: fontSize,
	})

	// Position body
	frame.PushFrame(Point{X: delimWidth, Y: 0}, bodyFrame)

	// Add closing delimiter
	frame.Push(Point{X: delimWidth + bodyFrame.Width(), Y: 0}, TextItem{
		Text:     elem.Close,
		FontSize: fontSize,
	})

	return frame
}
