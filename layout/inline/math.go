package inline

import (
	"github.com/boergens/gotypst/eval"
)

// MathLayoutConfig holds configuration for math layout.
type MathLayoutConfig struct {
	// FontSize is the base font size for math.
	FontSize Abs
	// ScriptScale is the scale factor for sub/superscripts (typically 0.7).
	ScriptScale float64
	// ScriptScriptScale is the scale for second-level scripts (typically 0.5).
	ScriptScriptScale float64
	// RuleThickness is the thickness of fraction lines.
	RuleThickness Abs
}

// DefaultMathLayoutConfig returns default math layout configuration.
func DefaultMathLayoutConfig(fontSize Abs) MathLayoutConfig {
	return MathLayoutConfig{
		FontSize:          fontSize,
		ScriptScale:       0.7,
		ScriptScriptScale: 0.5,
		RuleThickness:     fontSize * 0.04,
	}
}

// MathContext holds context for laying out math content.
type MathContext struct {
	Config   MathLayoutConfig
	Shaping  *ShapingContext
	FontSize Abs
}

// NewMathContext creates a new math layout context.
func NewMathContext(shaping *ShapingContext, fontSize Abs) *MathContext {
	return &MathContext{
		Config:   DefaultMathLayoutConfig(fontSize),
		Shaping:  shaping,
		FontSize: fontSize,
	}
}

// LayoutMathRoot lays out a MathRootElement and returns a MathRootItem.
func LayoutMathRoot(ctx *MathContext, elem *eval.MathRootElement) *MathRootItem {
	// Layout the radicand (content under the root sign)
	radicandFrame := layoutMathContent(ctx, &elem.Radicand)

	// Layout the optional index
	var indexFrame *MathFrame
	if len(elem.Index.Elements) > 0 {
		indexCtx := scaledContext(ctx, ctx.Config.ScriptScale)
		indexFrame = layoutMathContent(indexCtx, &elem.Index)
	}

	// Calculate dimensions
	item := &MathRootItem{
		Index:    indexFrame,
		Radicand: radicandFrame,
	}

	calculateRootDimensions(ctx, item)
	return item
}

// LayoutMathFrac lays out a MathFracElement and returns a MathFracItem.
func LayoutMathFrac(ctx *MathContext, elem *eval.MathFracElement) *MathFracItem {
	// Layout numerator and denominator
	numFrame := layoutMathContent(ctx, &elem.Num)
	denomFrame := layoutMathContent(ctx, &elem.Denom)

	item := &MathFracItem{
		Num:   numFrame,
		Denom: denomFrame,
	}

	calculateFracDimensions(ctx, item)
	return item
}

// LayoutMathAttach lays out a MathAttachElement and returns a MathAttachItem.
func LayoutMathAttach(ctx *MathContext, elem *eval.MathAttachElement) *MathAttachItem {
	// Layout base
	baseFrame := layoutMathContent(ctx, &elem.Base)

	// Layout subscript and superscript with smaller font
	scriptCtx := scaledContext(ctx, ctx.Config.ScriptScale)

	var subFrame, supFrame *MathFrame
	if len(elem.Subscript.Elements) > 0 {
		subFrame = layoutMathContent(scriptCtx, &elem.Subscript)
	}
	if len(elem.Superscript.Elements) > 0 {
		supFrame = layoutMathContent(scriptCtx, &elem.Superscript)
	}

	item := &MathAttachItem{
		Base:        baseFrame,
		Subscript:   subFrame,
		Superscript: supFrame,
		Primes:      elem.Primes,
	}

	calculateAttachDimensions(ctx, item)
	return item
}

// LayoutMathDelimited lays out a MathDelimitedElement and returns a MathDelimitedItem.
func LayoutMathDelimited(ctx *MathContext, elem *eval.MathDelimitedElement) *MathDelimitedItem {
	// Layout body content
	bodyFrame := layoutMathContent(ctx, &elem.Body)

	// Shape delimiters
	openFrame := shapeDelimiter(ctx, elem.Open, bodyFrame.Height)
	closeFrame := shapeDelimiter(ctx, elem.Close, bodyFrame.Height)

	item := &MathDelimitedItem{
		Open:       elem.Open,
		Close:      elem.Close,
		Body:       bodyFrame,
		OpenFrame:  openFrame,
		CloseFrame: closeFrame,
	}

	calculateDelimitedDimensions(ctx, item)
	return item
}

// layoutMathContent recursively lays out math content.
func layoutMathContent(ctx *MathContext, content *eval.Content) *MathFrame {
	if content == nil || len(content.Elements) == 0 {
		return &MathFrame{
			Items:    nil,
			Width:    0,
			Height:   ctx.FontSize * 1.2,
			Baseline: ctx.FontSize,
		}
	}

	frame := &MathFrame{
		Items: make([]Item, 0, len(content.Elements)),
	}

	var totalWidth Abs
	var maxAscent, maxDescent Abs

	for _, elem := range content.Elements {
		item := layoutMathElement(ctx, elem)
		if item == nil {
			continue
		}

		frame.Items = append(frame.Items, item)
		totalWidth += item.NaturalWidth()

		// Track vertical extent
		ascent, descent := getItemVerticalExtent(item, ctx.FontSize)
		if ascent > maxAscent {
			maxAscent = ascent
		}
		if descent > maxDescent {
			maxDescent = descent
		}
	}

	frame.Width = totalWidth
	frame.Height = maxAscent + maxDescent
	frame.Baseline = maxAscent

	return frame
}

// layoutMathElement dispatches to the appropriate layout function for each element type.
func layoutMathElement(ctx *MathContext, elem eval.ContentElement) Item {
	switch e := elem.(type) {
	case *eval.MathRootElement:
		return LayoutMathRoot(ctx, e)
	case *eval.MathFracElement:
		return LayoutMathFrac(ctx, e)
	case *eval.MathAttachElement:
		return LayoutMathAttach(ctx, e)
	case *eval.MathDelimitedElement:
		return LayoutMathDelimited(ctx, e)
	case *eval.MathSymbolElement:
		return layoutMathSymbol(ctx, e)
	case *eval.TextElement:
		return layoutMathText(ctx, e)
	default:
		// Unknown element - skip
		return nil
	}
}

// layoutMathSymbol lays out a math symbol.
func layoutMathSymbol(ctx *MathContext, elem *eval.MathSymbolElement) Item {
	if ctx.Shaping == nil {
		// No shaping context - create a placeholder
		return &TextItem{shaped: &ShapedText{
			Text:   elem.Symbol,
			Glyphs: NewGlyphsFromSlice(nil),
		}}
	}

	shaped := Shape(ctx.Shaping, 0, elem.Symbol, DirLTR, "", nil)
	return &TextItem{shaped: shaped}
}

// layoutMathText lays out math text.
func layoutMathText(ctx *MathContext, elem *eval.TextElement) Item {
	if ctx.Shaping == nil {
		// No shaping context - create a placeholder
		return &TextItem{shaped: &ShapedText{
			Text:   elem.Text,
			Glyphs: NewGlyphsFromSlice(nil),
		}}
	}

	shaped := Shape(ctx.Shaping, 0, elem.Text, DirLTR, "", nil)
	return &TextItem{shaped: shaped}
}

// scaledContext creates a context with scaled font size.
func scaledContext(ctx *MathContext, scale float64) *MathContext {
	return &MathContext{
		Config:   ctx.Config,
		Shaping:  ctx.Shaping,
		FontSize: ctx.FontSize * Abs(scale),
	}
}

// calculateRootDimensions calculates dimensions for a root item.
func calculateRootDimensions(ctx *MathContext, item *MathRootItem) {
	// Radical symbol dimensions
	radicalWidth := ctx.FontSize * 0.7  // Width of the radical symbol
	radicalGap := ctx.FontSize * 0.1    // Gap between radical and radicand
	indexGap := ctx.FontSize * 0.05     // Gap for index positioning

	// Get radicand dimensions
	radicandWidth := Abs(0)
	radicandHeight := ctx.FontSize * 1.2
	radicandBaseline := ctx.FontSize
	if item.Radicand != nil {
		radicandWidth = item.Radicand.Width
		radicandHeight = item.Radicand.Height
		radicandBaseline = item.Radicand.Baseline
	}

	// Add space for the overline
	overlineHeight := ctx.Config.RuleThickness
	totalRadicandHeight := radicandHeight + radicalGap + overlineHeight

	// Calculate width considering index
	totalWidth := radicalWidth + radicandWidth
	if item.Index != nil {
		// Index overlaps with the radical symbol's surd
		indexWidth := item.Index.Width
		indexOverlap := radicalWidth * 0.4 // How much index overlaps radical
		if indexWidth > indexOverlap {
			totalWidth += (indexWidth - indexOverlap + indexGap)
		}
	}

	// Height is the maximum of radical height and radicand + overline
	item.width = totalWidth
	item.height = totalRadicandHeight
	item.baseline = radicandBaseline + radicalGap + overlineHeight
}

// calculateFracDimensions calculates dimensions for a fraction item.
func calculateFracDimensions(ctx *MathContext, item *MathFracItem) {
	// Spacing around the fraction line
	lineGap := ctx.FontSize * 0.1
	lineThickness := ctx.Config.RuleThickness

	// Get numerator and denominator dimensions
	numWidth := Abs(0)
	numHeight := ctx.FontSize * 1.2
	if item.Num != nil {
		numWidth = item.Num.Width
		numHeight = item.Num.Height
	}

	denomWidth := Abs(0)
	denomHeight := ctx.FontSize * 1.2
	if item.Denom != nil {
		denomWidth = item.Denom.Width
		denomHeight = item.Denom.Height
	}

	// Width is the maximum of num and denom
	width := numWidth
	if denomWidth > width {
		width = denomWidth
	}

	// Add padding
	padding := ctx.FontSize * 0.1
	width += padding * 2

	// Height is num + gap + line + gap + denom
	height := numHeight + lineGap + lineThickness + lineGap + denomHeight

	// Baseline is at the fraction line (math axis)
	// The math axis is typically at half x-height
	mathAxis := ctx.FontSize * 0.5

	item.width = width
	item.height = height
	item.baseline = numHeight + lineGap + lineThickness/2 - mathAxis + ctx.FontSize
}

// calculateAttachDimensions calculates dimensions for attach item.
func calculateAttachDimensions(ctx *MathContext, item *MathAttachItem) {
	// Get base dimensions
	baseWidth := Abs(0)
	baseHeight := ctx.FontSize * 1.2
	baseBaseline := ctx.FontSize
	if item.Base != nil {
		baseWidth = item.Base.Width
		baseHeight = item.Base.Height
		baseBaseline = item.Base.Baseline
	}

	// Script positioning
	supShift := baseBaseline * 0.5                    // Superscript rises
	subShift := (baseHeight - baseBaseline) * 0.3    // Subscript drops

	scriptWidth := Abs(0)
	extraHeight := Abs(0)
	extraTop := Abs(0)

	if item.Superscript != nil {
		scriptWidth = item.Superscript.Width
		supHeight := item.Superscript.Height
		// Check if superscript extends above base
		supTop := supShift + supHeight - item.Superscript.Baseline
		if supTop > baseBaseline {
			extraTop = supTop - baseBaseline
		}
	}

	if item.Subscript != nil {
		if item.Subscript.Width > scriptWidth {
			scriptWidth = item.Subscript.Width
		}
		subBottom := subShift + item.Subscript.Baseline
		if subBottom > (baseHeight - baseBaseline) {
			extraHeight = subBottom - (baseHeight - baseBaseline)
		}
	}

	// Add prime width
	primeWidth := Abs(0)
	if item.Primes > 0 {
		primeWidth = ctx.FontSize * 0.3 * Abs(item.Primes)
	}

	item.width = baseWidth + scriptWidth + primeWidth
	item.height = baseHeight + extraTop + extraHeight
	item.baseline = baseBaseline + extraTop
}

// calculateDelimitedDimensions calculates dimensions for delimited item.
func calculateDelimitedDimensions(ctx *MathContext, item *MathDelimitedItem) {
	bodyWidth := Abs(0)
	bodyHeight := ctx.FontSize * 1.2
	bodyBaseline := ctx.FontSize
	if item.Body != nil {
		bodyWidth = item.Body.Width
		bodyHeight = item.Body.Height
		bodyBaseline = item.Body.Baseline
	}

	openWidth := Abs(0)
	closeWidth := Abs(0)
	if item.OpenFrame != nil {
		openWidth = item.OpenFrame.Width
	}
	if item.CloseFrame != nil {
		closeWidth = item.CloseFrame.Width
	}

	item.width = openWidth + bodyWidth + closeWidth
	item.height = bodyHeight
	item.baseline = bodyBaseline
}

// shapeDelimiter shapes a delimiter to match a target height.
func shapeDelimiter(ctx *MathContext, delim string, targetHeight Abs) *MathFrame {
	if ctx.Shaping == nil || delim == "" {
		return &MathFrame{
			Width:    0,
			Height:   targetHeight,
			Baseline: targetHeight * 0.8,
		}
	}

	shaped := Shape(ctx.Shaping, 0, delim, DirLTR, "", nil)
	return &MathFrame{
		Items:    []Item{&TextItem{shaped: shaped}},
		Width:    shaped.Width(),
		Height:   targetHeight,
		Baseline: targetHeight * 0.8,
	}
}

// getItemVerticalExtent returns the ascent and descent for an item.
func getItemVerticalExtent(item Item, fontSize Abs) (ascent, descent Abs) {
	switch it := item.(type) {
	case *MathRootItem:
		return it.baseline, it.height - it.baseline
	case *MathFracItem:
		return it.baseline, it.height - it.baseline
	case *MathAttachItem:
		return it.baseline, it.height - it.baseline
	case *MathDelimitedItem:
		return it.baseline, it.height - it.baseline
	case *TextItem:
		// Text uses approximate metrics
		ascent = fontSize * 0.8
		descent = fontSize * 0.2
		return ascent, descent
	default:
		return fontSize * 0.8, fontSize * 0.2
	}
}

// BuildMathRootFrame builds a FinalFrame from a MathRootItem.
func BuildMathRootFrame(item *MathRootItem, ctx *MathContext) *FinalFrame {
	frame := &FinalFrame{
		Size:     FinalSize{Width: item.width, Height: item.height},
		Baseline: item.baseline,
	}

	// Position radical symbol
	radicalWidth := ctx.FontSize * 0.7
	radicalGap := ctx.FontSize * 0.1
	overlineHeight := ctx.Config.RuleThickness

	// Calculate positions
	radicandX := radicalWidth
	radicandY := overlineHeight + radicalGap

	// Add index if present (positioned above and to the left of the surd)
	if item.Index != nil {
		indexX := Abs(0)
		indexY := Abs(0) // At top
		indexFrame := buildMathFrameFrame(item.Index)
		frame.PushFrame(FinalPoint{X: indexX, Y: indexY}, indexFrame)
	}

	// Add radical symbol (as text)
	radicalText := &ShapedText{
		Text:   "\u221A",
		Glyphs: NewGlyphsFromSlice(nil),
	}
	radicalFrame := &FinalFrame{
		Size:     FinalSize{Width: radicalWidth, Height: item.height},
		Baseline: item.baseline,
	}
	radicalFrame.Push(FinalPoint{X: 0, Y: 0}, FinalTextItem{Text: radicalText})

	radicalX := Abs(0)
	if item.Index != nil {
		indexWidth := item.Index.Width
		indexOverlap := radicalWidth * 0.4
		if indexWidth > indexOverlap {
			radicalX = indexWidth - indexOverlap + ctx.FontSize*0.05
		}
	}
	frame.PushFrame(FinalPoint{X: radicalX, Y: 0}, radicalFrame)

	// Add radicand
	if item.Radicand != nil {
		radicandFrame := buildMathFrameFrame(item.Radicand)
		frame.PushFrame(FinalPoint{X: radicandX + radicalX, Y: radicandY}, radicandFrame)
	}

	// Add overline (as a shape item - simplified here)
	// In a full implementation, this would be a rule/line item

	return frame
}

// BuildMathFracFrame builds a FinalFrame from a MathFracItem.
func BuildMathFracFrame(item *MathFracItem, ctx *MathContext) *FinalFrame {
	frame := &FinalFrame{
		Size:     FinalSize{Width: item.width, Height: item.height},
		Baseline: item.baseline,
	}

	lineGap := ctx.FontSize * 0.1
	lineThickness := ctx.Config.RuleThickness
	padding := ctx.FontSize * 0.1

	// Center numerator
	numX := padding
	numY := Abs(0)
	if item.Num != nil {
		numX = (item.width - item.Num.Width) / 2
		numFrame := buildMathFrameFrame(item.Num)
		frame.PushFrame(FinalPoint{X: numX, Y: numY}, numFrame)
	}

	// Center denominator
	denomX := padding
	numHeight := ctx.FontSize * 1.2
	if item.Num != nil {
		numHeight = item.Num.Height
	}
	denomY := numHeight + lineGap + lineThickness + lineGap
	if item.Denom != nil {
		denomX = (item.width - item.Denom.Width) / 2
		denomFrame := buildMathFrameFrame(item.Denom)
		frame.PushFrame(FinalPoint{X: denomX, Y: denomY}, denomFrame)
	}

	// Fraction line would be added here as a rule item

	return frame
}

// BuildMathAttachFrame builds a FinalFrame from a MathAttachItem.
func BuildMathAttachFrame(item *MathAttachItem, ctx *MathContext) *FinalFrame {
	frame := &FinalFrame{
		Size:     FinalSize{Width: item.width, Height: item.height},
		Baseline: item.baseline,
	}

	baseX := Abs(0)
	baseY := item.baseline
	if item.Base != nil {
		baseY = item.baseline - item.Base.Baseline
	}

	// Position base
	if item.Base != nil {
		baseFrame := buildMathFrameFrame(item.Base)
		frame.PushFrame(FinalPoint{X: baseX, Y: baseY}, baseFrame)
	}

	// Position scripts
	scriptX := Abs(0)
	if item.Base != nil {
		scriptX = item.Base.Width
	}

	if item.Superscript != nil {
		supY := Abs(0) // At top for maximum rise
		supFrame := buildMathFrameFrame(item.Superscript)
		frame.PushFrame(FinalPoint{X: scriptX, Y: supY}, supFrame)
	}

	if item.Subscript != nil {
		subY := item.height - item.Subscript.Height
		subFrame := buildMathFrameFrame(item.Subscript)
		frame.PushFrame(FinalPoint{X: scriptX, Y: subY}, subFrame)
	}

	return frame
}

// BuildMathDelimitedFrame builds a FinalFrame from a MathDelimitedItem.
func BuildMathDelimitedFrame(item *MathDelimitedItem, ctx *MathContext) *FinalFrame {
	frame := &FinalFrame{
		Size:     FinalSize{Width: item.width, Height: item.height},
		Baseline: item.baseline,
	}

	x := Abs(0)

	// Opening delimiter
	if item.OpenFrame != nil {
		openFinalFrame := buildMathFrameFrame(item.OpenFrame)
		frame.PushFrame(FinalPoint{X: x, Y: 0}, openFinalFrame)
		x += item.OpenFrame.Width
	}

	// Body
	if item.Body != nil {
		bodyFrame := buildMathFrameFrame(item.Body)
		frame.PushFrame(FinalPoint{X: x, Y: 0}, bodyFrame)
		x += item.Body.Width
	}

	// Closing delimiter
	if item.CloseFrame != nil {
		closeFinalFrame := buildMathFrameFrame(item.CloseFrame)
		frame.PushFrame(FinalPoint{X: x, Y: 0}, closeFinalFrame)
	}

	return frame
}

// buildMathFrameFrame converts a MathFrame to a FinalFrame.
func buildMathFrameFrame(mf *MathFrame) *FinalFrame {
	frame := &FinalFrame{
		Size:     FinalSize{Width: mf.Width, Height: mf.Height},
		Baseline: mf.Baseline,
	}

	x := Abs(0)
	for _, item := range mf.Items {
		switch it := item.(type) {
		case *TextItem:
			if it.shaped != nil {
				childFrame := &FinalFrame{
					Size:     FinalSize{Width: it.NaturalWidth(), Height: mf.Height},
					Baseline: mf.Baseline,
				}
				childFrame.Push(FinalPoint{X: 0, Y: 0}, FinalTextItem{Text: it.shaped})
				frame.PushFrame(FinalPoint{X: x, Y: 0}, childFrame)
			}
		case *MathRootItem:
			// Create a context for recursive building
			ctx := &MathContext{FontSize: mf.Height * 0.8}
			childFrame := BuildMathRootFrame(it, ctx)
			frame.PushFrame(FinalPoint{X: x, Y: 0}, childFrame)
		case *MathFracItem:
			ctx := &MathContext{FontSize: mf.Height * 0.8}
			childFrame := BuildMathFracFrame(it, ctx)
			frame.PushFrame(FinalPoint{X: x, Y: 0}, childFrame)
		case *MathAttachItem:
			ctx := &MathContext{FontSize: mf.Height * 0.8}
			childFrame := BuildMathAttachFrame(it, ctx)
			frame.PushFrame(FinalPoint{X: x, Y: 0}, childFrame)
		case *MathDelimitedItem:
			ctx := &MathContext{FontSize: mf.Height * 0.8}
			childFrame := BuildMathDelimitedFrame(it, ctx)
			frame.PushFrame(FinalPoint{X: x, Y: 0}, childFrame)
		}
		x += item.NaturalWidth()
	}

	return frame
}
