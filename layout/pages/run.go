package pages

import (
	"github.com/boergens/gotypst/eval"
	"github.com/boergens/gotypst/layout"
)

// LayoutBlankPage lays out a single blank page suitable for parity adjustment.
func LayoutBlankPage(engine *Engine, locator Locator, initial StyleChain) (*LayoutedPage, error) {
	pages, err := LayoutPageRun(engine, nil, locator, initial)
	if err != nil {
		return nil, err
	}
	if len(pages) == 0 {
		return nil, nil
	}
	return &pages[0], nil
}

// LayoutPageRun lays out a page run with uniform properties.
func LayoutPageRun(engine *Engine, children []Pair, locator Locator, initial StyleChain) ([]LayoutedPage, error) {
	return layoutPageRunImpl(engine, children, locator, initial)
}

// layoutPageRunImpl is the internal implementation of LayoutPageRun.
func layoutPageRunImpl(engine *Engine, children []Pair, locator Locator, initial StyleChain) ([]LayoutedPage, error) {
	splitLocator := (&locator).Split()

	// Determine page-wide styles
	styles := resolveStyles(children, initial)

	// Resolve page dimensions
	width := resolvePageWidth(styles)
	height := resolvePageHeight(styles)
	size := layout.Size{Width: width, Height: height}

	if resolveFlipped(styles) {
		size.Width, size.Height = size.Height, size.Width
	}

	// Calculate minimum dimension for default margins
	minDim := width
	if height < minDim {
		minDim = height
	}
	if !isFinite(minDim) {
		minDim = paperA4Width
	}

	// Determine margins
	defaultMargin := layout.Abs((2.5 / 21.0) * float64(minDim))
	margin := resolveMargins(styles, defaultMargin, size)
	twoSided := resolveTwoSided(styles)

	// Resolve other page properties
	fill := resolveFill(styles)
	foreground := resolveForeground(styles)
	background := resolveBackground(styles)
	headerAscent := resolveHeaderAscent(styles, margin.Top)
	footerDescent := resolveFooterDescent(styles, margin.Bottom)
	numbering := resolveNumbering(styles)
	supplement := resolveSupplement(styles)
	binding := resolveBinding(styles)

	// Resolve header and footer
	header, footer := resolveHeaderFooter(styles, numbering)

	// Calculate content area
	area := layout.Size{
		Width:  size.Width - margin.Left - margin.Right,
		Height: size.Height - margin.Top - margin.Bottom,
	}

	// Layout the flow content
	// TODO: This needs the flow layout module (dependency hq-429)
	// For now, create a placeholder implementation
	fragment, err := layoutFlow(engine, children, splitLocator, styles, area)
	if err != nil {
		return nil, err
	}

	// Layout marginals and assemble pages
	var layouted []LayoutedPage
	for _, inner := range fragment {
		headerSize := layout.Size{Width: inner.Size.Width, Height: margin.Top - headerAscent}
		footerSize := layout.Size{Width: inner.Size.Width, Height: margin.Bottom - footerDescent}
		fullSize := layout.Size{
			Width:  inner.Size.Width + margin.Left + margin.Right,
			Height: inner.Size.Height + margin.Top + margin.Bottom,
		}

		// Layout static marginals (background/foreground) now
		backgroundFrame := layoutMarginal(engine, background, fullSize, splitLocator, 0)
		foregroundFrame := layoutMarginal(engine, foreground, fullSize, splitLocator, 0)

		// Store header/footer content for deferred layout in Finalize
		// (they need the page number for running content)
		layouted = append(layouted, LayoutedPage{
			Inner:         inner,
			Fill:          fill,
			Numbering:     numbering,
			Supplement:    supplement,
			HeaderContent: header,
			FooterContent: footer,
			HeaderSize:    headerSize,
			FooterSize:    footerSize,
			Background:    backgroundFrame,
			Foreground:    foregroundFrame,
			Margin:        margin,
			Binding:       binding,
			TwoSided:      twoSided,
		})
	}

	return layouted, nil
}

// Engine provides the layout engine context.
type Engine struct {
	// World provides access to fonts and files.
	World interface{}
	// TODO: Add more engine fields as needed
}

// Parallelize runs layout tasks in parallel.
func (e *Engine) Parallelize(items []RunItem, fn func(*Engine, RunItem) ([]LayoutedPage, error)) []layoutResult {
	results := make([]layoutResult, len(items))
	// TODO: Implement parallel execution
	// For now, run sequentially
	for i, item := range items {
		pages, err := fn(e, item)
		results[i] = layoutResult{pages: pages, err: err}
	}
	return results
}

type layoutResult struct {
	pages []LayoutedPage
	err   error
}

// Constants
const (
	paperA4Width layout.Abs = 595.276 // A4 width in points
)

// Helper functions for style resolution

func resolveStyles(children []Pair, initial StyleChain) StyleChain {
	// TODO: Merge styles from children with initial
	return initial
}

func resolvePageWidth(styles StyleChain) layout.Abs {
	if w := styles.Get("page.width"); w != nil {
		if abs, ok := w.(layout.Abs); ok {
			return abs
		}
	}
	return paperA4Width
}

func resolvePageHeight(styles StyleChain) layout.Abs {
	if h := styles.Get("page.height"); h != nil {
		if abs, ok := h.(layout.Abs); ok {
			return abs
		}
	}
	return 841.89 // A4 height in points
}

func resolveFlipped(styles StyleChain) bool {
	if f := styles.Get("page.flipped"); f != nil {
		if b, ok := f.(bool); ok {
			return b
		}
	}
	return false
}

func resolveMargins(styles StyleChain, defaultMargin layout.Abs, size layout.Size) Sides[layout.Abs] {
	// TODO: Resolve individual margins from styles
	return Sides[layout.Abs]{
		Left:   defaultMargin,
		Top:    defaultMargin,
		Right:  defaultMargin,
		Bottom: defaultMargin,
	}
}

func resolveTwoSided(styles StyleChain) bool {
	if t := styles.Get("page.margin.two-sided"); t != nil {
		if b, ok := t.(bool); ok {
			return b
		}
	}
	return false
}

func resolveFill(styles StyleChain) *Paint {
	if f := styles.Get("page.fill"); f != nil {
		if p, ok := f.(*Paint); ok {
			return p
		}
	}
	return nil
}

func resolveForeground(styles StyleChain) *Content {
	if f := styles.Get("page.foreground"); f != nil {
		if c, ok := f.(*Content); ok {
			return c
		}
	}
	return nil
}

func resolveBackground(styles StyleChain) *Content {
	if b := styles.Get("page.background"); b != nil {
		if c, ok := b.(*Content); ok {
			return c
		}
	}
	return nil
}

func resolveHeaderAscent(styles StyleChain, topMargin layout.Abs) layout.Abs {
	if h := styles.Get("page.header-ascent"); h != nil {
		if abs, ok := h.(layout.Abs); ok {
			return abs
		}
	}
	return topMargin * 0.3 // Default to 30% of top margin
}

func resolveFooterDescent(styles StyleChain, bottomMargin layout.Abs) layout.Abs {
	if f := styles.Get("page.footer-descent"); f != nil {
		if abs, ok := f.(layout.Abs); ok {
			return abs
		}
	}
	return bottomMargin * 0.3 // Default to 30% of bottom margin
}

func resolveNumbering(styles StyleChain) *Numbering {
	if n := styles.Get("page.numbering"); n != nil {
		if num, ok := n.(*Numbering); ok {
			return num
		}
	}
	return nil
}

func resolveSupplement(styles StyleChain) Content {
	if s := styles.Get("page.supplement"); s != nil {
		if c, ok := s.(Content); ok {
			return c
		}
	}
	return Content{}
}

func resolveBinding(styles StyleChain) Binding {
	if b := styles.Get("page.binding"); b != nil {
		if binding, ok := b.(Binding); ok {
			return binding
		}
	}
	// Default based on text direction
	if dir := styles.Get("text.dir"); dir != nil {
		if d, ok := dir.(layout.Dir); ok && d == layout.DirRTL {
			return BindingRight
		}
	}
	return BindingLeft
}

func resolveHeaderFooter(styles StyleChain, numbering *Numbering) (*Content, *Content) {
	header := styles.Get("page.header")
	footer := styles.Get("page.footer")

	var headerContent, footerContent *Content

	if h, ok := header.(*Content); ok {
		headerContent = h
	}
	if f, ok := footer.(*Content); ok {
		footerContent = f
	}

	// If no explicit header/footer and numbering exists,
	// place numbering based on number-align style
	if numbering != nil {
		numberAlign := styles.Get("page.number-align")
		if numberAlign == "top" && headerContent == nil {
			// Create numbering content for header (centered by default)
			headerContent = createNumberingContent(numbering.Pattern, layout.AlignCenter)
		} else if footerContent == nil {
			// Create numbering content for footer (default, centered)
			footerContent = createNumberingContent(numbering.Pattern, layout.AlignCenter)
		}
	}

	return headerContent, footerContent
}

// createNumberingContent creates content with a page number element.
func createNumberingContent(pattern string, align layout.Alignment) *Content {
	numberElem := &NumberingElem{Pattern: pattern}
	alignElem := &AlignElem{
		Align: align,
		Body:  Content{Elements: []ContentElement{numberElem}},
	}
	return &Content{Elements: []ContentElement{alignElem}}
}

func isFinite(v layout.Abs) bool {
	return v > 0 && v < 1e10
}

// layoutFlow lays out content in a flow layout.
// TODO: This is a placeholder - needs flow layout module (hq-429)
func layoutFlow(engine *Engine, children []Pair, locator *SplitLocator, styles StyleChain, area layout.Size) ([]Frame, error) {
	if len(children) == 0 {
		// Return a single empty frame for blank pages
		return []Frame{{Size: area}}, nil
	}

	// Minimal implementation: extract text and create simple positioned items
	frame := Frame{Size: area}
	var y layout.Abs = 12 // Start with some top margin
	fontSize := layout.Abs(12) // Default font size

	for _, pair := range children {
		// Check if this is a text element
		if textElem, ok := pair.Element.(*eval.TextElement); ok {
			// Create a placeholder TextItem with estimated size
			// TODO: Proper text shaping with glyphs when inline module is ready
			textWidth := estimateTextWidth(textElem.Text)
			frame.Push(
				layout.Point{X: 0, Y: y},
				TextItem{
					Glyphs: nil, // Would be filled by text shaping
					Size:   layout.Size{Width: textWidth, Height: fontSize},
				},
			)
			y += fontSize * 1.2 // Simple line spacing
		}
		// Skip other element types for now (spaces, etc.)
	}

	return []Frame{frame}, nil
}

// layoutMarginal lays out a marginal (header, footer, background, foreground).
// For headers/footers with running content, pass the page number for formatting.
// For static content like background/foreground, pass 0 for pageNum.
func layoutMarginal(engine *Engine, content *Content, area layout.Size, locator *SplitLocator, pageNum int) *Frame {
	if content == nil {
		return nil
	}

	frame := Frame{Size: area}

	// Layout each content element
	for _, elem := range content.Elements {
		layoutContentElement(&frame, elem, area, pageNum)
	}

	return &frame
}

// layoutContentElement lays out a single content element into a frame.
func layoutContentElement(frame *Frame, elem ContentElement, area layout.Size, pageNum int) {
	switch e := elem.(type) {
	case *TextElem:
		// Create a simple text item
		// For proper text rendering, we'd use the inline shaping module
		// For now, create a placeholder with estimated size
		textItem := TextItem{
			Glyphs: nil, // Would be filled by text shaping
			Size:   layout.Size{Width: estimateTextWidth(e.Text), Height: 12}, // Placeholder
		}
		// Center the text vertically in the area
		y := (area.Height - textItem.Size.Height) / 2
		frame.Push(layout.Point{X: 0, Y: y}, textItem)

	case *NumberingElem:
		// Format the page number according to the pattern
		text := formatPageNumber(pageNum, e.Pattern)
		textItem := TextItem{
			Glyphs: nil,
			Size:   layout.Size{Width: estimateTextWidth(text), Height: 12},
		}
		y := (area.Height - textItem.Size.Height) / 2
		frame.Push(layout.Point{X: 0, Y: y}, textItem)

	case *AlignElem:
		// Create a sub-frame for aligned content
		subFrame := Frame{Size: area}
		for _, subElem := range e.Body.Elements {
			layoutContentElement(&subFrame, subElem, area, pageNum)
		}
		// Calculate content width and position based on alignment
		contentWidth := calculateContentWidth(&subFrame)
		var x layout.Abs
		switch e.Align {
		case layout.AlignCenter:
			x = (area.Width - contentWidth) / 2
		case layout.AlignEnd:
			x = area.Width - contentWidth
		default: // AlignStart
			x = 0
		}
		// Reposition items with alignment offset
		for i := range subFrame.Items {
			subFrame.Items[i].Pos.X += x
		}
		frame.PushMultiple(subFrame.Items)

	case *SpaceElem:
		// Space doesn't add visual items, just affects positioning
		// This is handled by the alignment logic
	}
}

// estimateTextWidth provides a rough estimate of text width.
// In a real implementation, this would use font metrics.
func estimateTextWidth(text string) layout.Abs {
	// Rough estimate: ~7 points per character at 12pt font
	return layout.Abs(len(text) * 7)
}

// formatPageNumber formats a page number according to a pattern.
func formatPageNumber(num int, pattern string) string {
	switch pattern {
	case "1", "": // Arabic numerals (default)
		return formatArabic(num)
	case "i": // Lowercase roman numerals
		return formatRomanLower(num)
	case "I": // Uppercase roman numerals
		return formatRomanUpper(num)
	case "a": // Lowercase letters
		return formatLetterLower(num)
	case "A": // Uppercase letters
		return formatLetterUpper(num)
	default:
		return formatArabic(num)
	}
}

// formatArabic formats a number as Arabic numerals.
func formatArabic(num int) string {
	if num <= 0 {
		return "0"
	}
	result := ""
	for num > 0 {
		result = string(rune('0'+num%10)) + result
		num /= 10
	}
	return result
}

// formatRomanLower formats a number as lowercase Roman numerals.
func formatRomanLower(num int) string {
	return formatRoman(num, false)
}

// formatRomanUpper formats a number as uppercase Roman numerals.
func formatRomanUpper(num int) string {
	return formatRoman(num, true)
}

// formatRoman formats a number as Roman numerals.
func formatRoman(num int, upper bool) string {
	if num <= 0 || num > 3999 {
		return formatArabic(num)
	}
	values := []int{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	symbols := []string{"m", "cm", "d", "cd", "c", "xc", "l", "xl", "x", "ix", "v", "iv", "i"}
	if upper {
		symbols = []string{"M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"}
	}
	result := ""
	for i, v := range values {
		for num >= v {
			result += symbols[i]
			num -= v
		}
	}
	return result
}

// formatLetterLower formats a number as lowercase letters (a, b, ..., z, aa, ab, ...).
func formatLetterLower(num int) string {
	return formatLetter(num, 'a')
}

// formatLetterUpper formats a number as uppercase letters (A, B, ..., Z, AA, AB, ...).
func formatLetterUpper(num int) string {
	return formatLetter(num, 'A')
}

// formatLetter formats a number as letters.
func formatLetter(num int, base rune) string {
	if num <= 0 {
		return string(base)
	}
	result := ""
	for num > 0 {
		num-- // Convert to 0-indexed
		result = string(base+rune(num%26)) + result
		num /= 26
	}
	return result
}

// calculateContentWidth calculates the total width of content in a frame.
func calculateContentWidth(frame *Frame) layout.Abs {
	var maxWidth layout.Abs
	for _, item := range frame.Items {
		var itemWidth layout.Abs
		switch it := item.Item.(type) {
		case TextItem:
			itemWidth = item.Pos.X + it.Size.Width
		case GroupItem:
			itemWidth = item.Pos.X + it.Frame.Size.Width
		}
		if itemWidth > maxWidth {
			maxWidth = itemWidth
		}
	}
	return maxWidth
}
