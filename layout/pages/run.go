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

		headerFrame := layoutMarginal(engine, header, headerSize, splitLocator)
		footerFrame := layoutMarginal(engine, footer, footerSize, splitLocator)
		backgroundFrame := layoutMarginal(engine, background, fullSize, splitLocator)
		foregroundFrame := layoutMarginal(engine, foreground, fullSize, splitLocator)

		layouted = append(layouted, LayoutedPage{
			Inner:      inner,
			Fill:       fill,
			Numbering:  numbering,
			Supplement: supplement,
			Header:     headerFrame,
			Footer:     footerFrame,
			Background: backgroundFrame,
			Foreground: foregroundFrame,
			Margin:     margin,
			Binding:    binding,
			TwoSided:   twoSided,
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
			// Create numbering content for header
			// TODO: Implement numbering content creation
		} else if footerContent == nil {
			// Create numbering content for footer (default)
			// TODO: Implement numbering content creation
		}
	}

	return headerContent, footerContent
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

	// Minimal implementation: collect text into lines separated by paragraph breaks
	frame := Frame{Size: area}
	var y layout.Abs = 0
	fontSize := layout.Abs(12) // Default font size
	lineHeight := fontSize * 1.4

	var currentLine string
	flushLine := func() {
		if currentLine != "" {
			frame.Push(
				layout.Point{X: 0, Y: y},
				TextItem{Text: currentLine, FontSize: fontSize},
			)
			y += lineHeight
			currentLine = ""
		}
	}

	for _, pair := range children {
		elem, ok := pair.Element.(eval.ContentElement)
		if !ok {
			continue
		}

		// Paragraph breaks flush the current line and add spacing
		if _, ok := elem.(*eval.ParbreakElement); ok {
			flushLine()
			y += lineHeight * 0.3 // Extra paragraph spacing
			continue
		}

		// List items get their own line
		if _, ok := elem.(*eval.ListItemElement); ok {
			flushLine()
		}

		text := extractText(elem)
		currentLine += text
	}

	// Flush any remaining text
	flushLine()

	return []Frame{frame}, nil
}

// extractText recursively extracts text content from an element.
func extractText(elem eval.ContentElement) string {
	switch e := elem.(type) {
	case *eval.TextElement:
		return e.Text
	case *eval.SpaceElement:
		return " "
	case *eval.HeadingElement:
		return extractTextFromContent(&e.Content)
	case *eval.StrongElement:
		return extractTextFromContent(&e.Content)
	case *eval.EmphElement:
		return extractTextFromContent(&e.Content)
	case *eval.ParagraphElement:
		return extractTextFromContent(&e.Body)
	case *eval.ListItemElement:
		return "• " + extractTextFromContent(&e.Content)
	case *eval.RawElement:
		return e.Text
	default:
		return ""
	}
}

// extractTextFromContent extracts text from a Content struct.
func extractTextFromContent(c *eval.Content) string {
	var result string
	for _, elem := range c.Elements {
		result += extractText(elem)
	}
	return result
}

// layoutMarginal lays out a marginal (header, footer, etc.)
func layoutMarginal(engine *Engine, content *Content, area layout.Size, locator *SplitLocator) *Frame {
	if content == nil {
		return nil
	}
	// TODO: Layout the marginal content
	frame := Frame{Size: area}
	return &frame
}
