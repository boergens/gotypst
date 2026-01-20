package svg

import (
	"encoding/base64"
	"fmt"
	"math"
	"strings"

	"github.com/boergens/gotypst/layout"
	"github.com/boergens/gotypst/layout/flow"
	"github.com/boergens/gotypst/layout/inline"
	"github.com/boergens/gotypst/layout/pages"
	"github.com/boergens/gotypst/library/visualize"
)

// Renderer generates SVG content from laid out frames.
type Renderer struct {
	// FontMap maps font faces to font family names for SVG.
	FontMap FontMapper
}

// renderContext holds state during a single render pass.
type renderContext struct {
	gradients   []*visualize.Gradient
	gradientIDs map[*visualize.Gradient]string
	nextGradID  int
}

// newRenderContext creates a new render context.
func newRenderContext() *renderContext {
	return &renderContext{
		gradientIDs: make(map[*visualize.Gradient]string),
	}
}

// registerGradient registers a gradient and returns its ID.
func (ctx *renderContext) registerGradient(g *visualize.Gradient) string {
	if id, ok := ctx.gradientIDs[g]; ok {
		return id
	}
	id := fmt.Sprintf("grad%d", ctx.nextGradID)
	ctx.nextGradID++
	ctx.gradients = append(ctx.gradients, g)
	ctx.gradientIDs[g] = id
	return id
}

// FontMapper maps fonts to SVG font family names.
type FontMapper interface {
	// FontFamily returns the CSS font-family value for a font.
	// Returns empty string if font is not registered.
	FontFamily(face interface{}) string
}

// DefaultFontMapper is a simple font mapper that returns a default font.
type DefaultFontMapper struct{}

// FontFamily returns a default font family.
func (DefaultFontMapper) FontFamily(face interface{}) string {
	return "serif"
}

// NewRenderer creates a new SVG renderer.
func NewRenderer() *Renderer {
	return &Renderer{
		FontMap: DefaultFontMapper{},
	}
}

// RenderDocument renders a full document to SVG strings.
// Returns a slice of SVG documents, one per page.
func (r *Renderer) RenderDocument(doc *pages.PagedDocument) []string {
	svgs := make([]string, len(doc.Pages))

	for i, page := range doc.Pages {
		svgs[i] = r.RenderPage(&page)
	}

	return svgs
}

// RenderPage renders a single page to an SVG string.
func (r *Renderer) RenderPage(page *pages.Page) string {
	ctx := newRenderContext()
	var content strings.Builder

	width := page.Frame.Size.Width
	height := page.Frame.Size.Height

	// Fill page background if specified
	if page.Fill != nil && page.Fill.Color != nil {
		content.WriteString(fmt.Sprintf(`<rect width="%g" height="%g" fill="%s"/>`,
			float64(width), float64(height), colorToSVG(page.Fill.Color)))
		content.WriteString("\n")
	}

	// Render frame content (this may register gradients)
	r.renderPagesFrameWithContext(ctx, &content, &page.Frame, layout.Point{X: 0, Y: 0})

	// Build final SVG
	var b strings.Builder

	// SVG header
	b.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%g" height="%g" viewBox="0 0 %g %g">`,
		float64(width), float64(height), float64(width), float64(height)))
	b.WriteString("\n")

	// Add defs section if we have gradients
	if len(ctx.gradients) > 0 {
		b.WriteString("<defs>\n")
		for _, g := range ctx.gradients {
			id := ctx.gradientIDs[g]
			writeGradientDef(&b, g, id)
		}
		b.WriteString("</defs>\n")
	}

	// Add rendered content
	b.WriteString(content.String())

	b.WriteString("</svg>")

	return b.String()
}

// renderPagesFrame renders a pages.Frame (from layout/pages package).
func (r *Renderer) renderPagesFrame(b *strings.Builder, frame *pages.Frame, origin layout.Point) {
	r.renderPagesFrameWithContext(nil, b, frame, origin)
}

// renderPagesFrameWithContext renders a pages.Frame with gradient context.
func (r *Renderer) renderPagesFrameWithContext(ctx *renderContext, b *strings.Builder, frame *pages.Frame, origin layout.Point) {
	for _, item := range frame.Items {
		pos := layout.Point{
			X: origin.X + item.Pos.X,
			Y: origin.Y + item.Pos.Y,
		}
		r.renderPagesFrameItemWithContext(ctx, b, item.Item, pos)
	}
}

// renderPagesFrameItem renders a single item from a pages.Frame.
func (r *Renderer) renderPagesFrameItem(b *strings.Builder, item pages.FrameItem, pos layout.Point) {
	r.renderPagesFrameItemWithContext(nil, b, item, pos)
}

// renderPagesFrameItemWithContext renders a single item from a pages.Frame with gradient context.
func (r *Renderer) renderPagesFrameItemWithContext(ctx *renderContext, b *strings.Builder, item pages.FrameItem, pos layout.Point) {
	switch it := item.(type) {
	case pages.GroupItem:
		// Nested frame - recurse
		r.renderPagesFrameWithContext(ctx, b, &it.Frame, pos)
	case pages.TagItem:
		// Tags are metadata, not rendered
	case pages.TextItem:
		// Render text directly
		r.renderSimpleText(b, it.Text, it.FontSize, pos)
	case pages.ImageItem:
		// Render image
		r.renderImage(b, &it.Image, it.Size, pos)
	}
}

// RenderFlowFrame renders a flow.Frame to SVG.
func (r *Renderer) RenderFlowFrame(b *strings.Builder, frame *flow.Frame, origin layout.Point) {
	for _, entry := range frame.Items() {
		pos := layout.Point{
			X: origin.X + entry.Pos.X,
			Y: origin.Y + entry.Pos.Y,
		}
		r.renderFlowFrameItem(b, entry.Item, pos)
	}
}

// renderFlowFrameItem renders a single item from a flow.Frame.
func (r *Renderer) renderFlowFrameItem(b *strings.Builder, item flow.FrameItem, pos layout.Point) {
	switch it := item.(type) {
	case flow.FrameItemFrame:
		// Nested frame - recurse
		r.RenderFlowFrame(b, &it.Frame, pos)
	case flow.FrameItemTag:
		// Tags are metadata, not rendered
	case flow.FrameItemLink:
		// Links could be rendered as <a> elements
		// For now, skip as main content rendering handles the visual part
	}
}

// RenderInlineFrame renders an inline.FinalFrame to SVG.
func (r *Renderer) RenderInlineFrame(b *strings.Builder, frame *inline.FinalFrame, origin layout.Point) {
	for _, entry := range frame.Items {
		pos := layout.Point{
			X: origin.X + layout.Abs(entry.Pos.X),
			Y: origin.Y + layout.Abs(entry.Pos.Y),
		}
		r.renderInlineFrameItem(b, entry.Item, pos, frame.Baseline)
	}
}

// renderInlineFrameItem renders a single item from an inline.FinalFrame.
func (r *Renderer) renderInlineFrameItem(b *strings.Builder, item inline.FinalFrameItem, pos layout.Point, baseline layout.Abs) {
	switch it := item.(type) {
	case inline.FinalTextItem:
		r.renderShapedText(b, it.Text, pos, baseline)
	case inline.FinalMathScriptItem:
		r.renderMathScript(b, it, pos, baseline)
	case inline.FinalMathLimitsItem:
		r.renderMathLimits(b, it, pos, baseline)
	}
}

// renderMathScript renders a math script item (superscript/subscript).
func (r *Renderer) renderMathScript(b *strings.Builder, item inline.FinalMathScriptItem, pos layout.Point, baseline layout.Abs) {
	// Render base content
	if item.BaseFrame != nil {
		r.RenderInlineFrame(b, item.BaseFrame, pos)
	}

	// Calculate script X position
	scriptX := pos.X + item.ScriptXOffset

	// Render superscript (positioned above baseline)
	if item.SuperFrame != nil {
		superPos := layout.Point{
			X: scriptX,
			Y: pos.Y + item.SuperOffset,
		}
		r.RenderInlineFrame(b, item.SuperFrame, superPos)
	}

	// Render subscript (positioned below baseline)
	if item.SubFrame != nil {
		subPos := layout.Point{
			X: scriptX,
			Y: pos.Y + item.SubOffset,
		}
		r.RenderInlineFrame(b, item.SubFrame, subPos)
	}
}

// renderMathLimits renders a math limits item (operator with limits above/below).
func (r *Renderer) renderMathLimits(b *strings.Builder, item inline.FinalMathLimitsItem, pos layout.Point, baseline layout.Abs) {
	// Render upper limit (centered above nucleus)
	if item.UpperFrame != nil {
		upperPos := layout.Point{
			X: pos.X + item.CenterX - item.UpperFrame.Size.Width/2,
			Y: pos.Y + item.UpperOffset,
		}
		r.RenderInlineFrame(b, item.UpperFrame, upperPos)
	}

	// Render nucleus (main operator)
	if item.NucleusFrame != nil {
		nucleusPos := layout.Point{
			X: pos.X + item.CenterX - item.NucleusFrame.Size.Width/2,
			Y: pos.Y,
		}
		r.RenderInlineFrame(b, item.NucleusFrame, nucleusPos)
	}

	// Render lower limit (centered below nucleus)
	if item.LowerFrame != nil {
		lowerPos := layout.Point{
			X: pos.X + item.CenterX - item.LowerFrame.Size.Width/2,
			Y: pos.Y + item.LowerOffset,
		}
		r.RenderInlineFrame(b, item.LowerFrame, lowerPos)
	}
}

// renderShapedText renders shaped text to SVG.
func (r *Renderer) renderShapedText(b *strings.Builder, text *inline.ShapedText, pos layout.Point, baseline layout.Abs) {
	if text == nil || text.Glyphs.Len() == 0 {
		return
	}

	glyphs := text.Glyphs.Kept()
	if len(glyphs) == 0 {
		return
	}

	// Get font info from first glyph
	firstGlyph := &glyphs[0]
	fontFamily := r.FontMap.FontFamily(firstGlyph.Font)
	fontSize := firstGlyph.Size

	// SVG uses top-left origin, text is positioned at baseline
	svgX := pos.X
	svgY := pos.Y + baseline

	// Build text content
	var textContent strings.Builder
	for i := range glyphs {
		g := &glyphs[i]
		textContent.WriteRune(g.Char)
	}

	// Output SVG text element
	b.WriteString(fmt.Sprintf(`<text x="%g" y="%g" font-family="%s" font-size="%g">`,
		float64(svgX), float64(svgY), escapeXML(fontFamily), float64(fontSize)))
	b.WriteString(escapeXML(textContent.String()))
	b.WriteString("</text>\n")
}

// renderSimpleText renders simple text directly at a position.
func (r *Renderer) renderSimpleText(b *strings.Builder, text string, fontSize layout.Abs, pos layout.Point) {
	if text == "" {
		return
	}

	// SVG text is positioned at baseline; add font size as approximation
	svgX := pos.X
	svgY := pos.Y + fontSize

	b.WriteString(fmt.Sprintf(`<text x="%g" y="%g" font-family="serif" font-size="%g">`,
		float64(svgX), float64(svgY), float64(fontSize)))
	b.WriteString(escapeXML(text))
	b.WriteString("</text>\n")
}

// GlyphOutliner provides glyph outline data for text-to-paths conversion.
// Implementations should extract glyph outlines from font data.
type GlyphOutliner interface {
	// OutlineGlyph returns the outline of a glyph as path segments.
	// The segments are in font units and should be scaled by fontSize/unitsPerEm.
	// Returns nil, false if the glyph outline is not available.
	OutlineGlyph(glyphID uint16) ([]inline.PathSegment, bool)

	// UnitsPerEm returns the font's units per em value.
	UnitsPerEm() float64
}

// RenderTextAsPath renders shaped text as SVG path elements instead of <text> elements.
// This ensures exact rendering without requiring fonts to be installed.
// The outliner parameter provides glyph outline data; if nil, falls back to <text> element.
func (r *Renderer) RenderTextAsPath(b *strings.Builder, text *inline.ShapedText, pos layout.Point, baseline layout.Abs, outliner GlyphOutliner, fill interface{}) {
	if text == nil || text.Glyphs.Len() == 0 {
		return
	}

	glyphs := text.Glyphs.Kept()
	if len(glyphs) == 0 {
		return
	}

	// If no outliner provided, fall back to text element
	if outliner == nil {
		r.renderShapedText(b, text, pos, baseline)
		return
	}

	// SVG uses top-left origin, text is positioned at baseline
	svgX := float64(pos.X)
	svgY := float64(pos.Y + baseline)

	// Scale factor: font size / units per em
	unitsPerEm := outliner.UnitsPerEm()
	if unitsPerEm == 0 {
		unitsPerEm = 1000 // Default
	}

	// Build combined path for all glyphs
	var pathData strings.Builder
	x := svgX

	for i := range glyphs {
		g := &glyphs[i]
		scale := float64(g.Size) / unitsPerEm

		// Get glyph outline
		segments, ok := outliner.OutlineGlyph(g.GlyphID)
		if !ok || len(segments) == 0 {
			// Skip glyphs without outlines (e.g., spaces)
			x += float64(g.XAdvance.At(g.Size))
			continue
		}

		// Apply glyph offsets
		glyphX := x + float64(g.XOffset.At(g.Size))
		glyphY := svgY + float64(g.YOffset.At(g.Size))

		// Convert glyph segments to path data
		for _, seg := range segments {
			switch s := seg.(type) {
			case *inline.LineSegment:
				// Line segment - scale and translate
				x0 := glyphX + s.X0*scale
				y0 := glyphY - s.Y0*scale // Y is inverted in SVG
				x1 := glyphX + s.X1*scale
				y1 := glyphY - s.Y1*scale
				pathData.WriteString(fmt.Sprintf("M%.2f %.2f L%.2f %.2f ", x0, y0, x1, y1))

			case *inline.QuadSegment:
				// Quadratic bezier - scale and translate
				x0 := glyphX + s.X0*scale
				y0 := glyphY - s.Y0*scale
				x1 := glyphX + s.X1*scale
				y1 := glyphY - s.Y1*scale
				x2 := glyphX + s.X2*scale
				y2 := glyphY - s.Y2*scale
				pathData.WriteString(fmt.Sprintf("M%.2f %.2f Q%.2f %.2f %.2f %.2f ", x0, y0, x1, y1, x2, y2))

			case *inline.CubicSegment:
				// Cubic bezier - scale and translate
				x0 := glyphX + s.X0*scale
				y0 := glyphY - s.Y0*scale
				x1 := glyphX + s.X1*scale
				y1 := glyphY - s.Y1*scale
				x2 := glyphX + s.X2*scale
				y2 := glyphY - s.Y2*scale
				x3 := glyphX + s.X3*scale
				y3 := glyphY - s.Y3*scale
				pathData.WriteString(fmt.Sprintf("M%.2f %.2f C%.2f %.2f %.2f %.2f %.2f %.2f ", x0, y0, x1, y1, x2, y2, x3, y3))
			}
		}

		// Advance x position
		x += float64(g.XAdvance.At(g.Size))
	}

	// Output path element
	if pathData.Len() > 0 {
		b.WriteString(fmt.Sprintf(`<path d="%s"`, strings.TrimSpace(pathData.String())))
		if fill != nil {
			b.WriteString(fillToSVG(fill))
		} else {
			b.WriteString(` fill="#000000"`)
		}
		b.WriteString("/>\n")
	}
}

// RenderTextAsPathWithContext renders text as paths with gradient context support.
func (r *Renderer) RenderTextAsPathWithContext(ctx *renderContext, b *strings.Builder, text *inline.ShapedText, pos layout.Point, baseline layout.Abs, outliner GlyphOutliner, fill interface{}) {
	if text == nil || text.Glyphs.Len() == 0 {
		return
	}

	glyphs := text.Glyphs.Kept()
	if len(glyphs) == 0 {
		return
	}

	// If no outliner provided, fall back to text element
	if outliner == nil {
		r.renderShapedText(b, text, pos, baseline)
		return
	}

	// SVG uses top-left origin, text is positioned at baseline
	svgX := float64(pos.X)
	svgY := float64(pos.Y + baseline)

	// Scale factor: font size / units per em
	unitsPerEm := outliner.UnitsPerEm()
	if unitsPerEm == 0 {
		unitsPerEm = 1000 // Default
	}

	// Build combined path for all glyphs
	var pathData strings.Builder
	x := svgX

	for i := range glyphs {
		g := &glyphs[i]
		scale := float64(g.Size) / unitsPerEm

		// Get glyph outline
		segments, ok := outliner.OutlineGlyph(g.GlyphID)
		if !ok || len(segments) == 0 {
			x += float64(g.XAdvance.At(g.Size))
			continue
		}

		// Apply glyph offsets
		glyphX := x + float64(g.XOffset.At(g.Size))
		glyphY := svgY + float64(g.YOffset.At(g.Size))

		// Convert glyph segments to path data
		for _, seg := range segments {
			switch s := seg.(type) {
			case *inline.LineSegment:
				x0 := glyphX + s.X0*scale
				y0 := glyphY - s.Y0*scale
				x1 := glyphX + s.X1*scale
				y1 := glyphY - s.Y1*scale
				pathData.WriteString(fmt.Sprintf("M%.2f %.2f L%.2f %.2f ", x0, y0, x1, y1))

			case *inline.QuadSegment:
				x0 := glyphX + s.X0*scale
				y0 := glyphY - s.Y0*scale
				x1 := glyphX + s.X1*scale
				y1 := glyphY - s.Y1*scale
				x2 := glyphX + s.X2*scale
				y2 := glyphY - s.Y2*scale
				pathData.WriteString(fmt.Sprintf("M%.2f %.2f Q%.2f %.2f %.2f %.2f ", x0, y0, x1, y1, x2, y2))

			case *inline.CubicSegment:
				x0 := glyphX + s.X0*scale
				y0 := glyphY - s.Y0*scale
				x1 := glyphX + s.X1*scale
				y1 := glyphY - s.Y1*scale
				x2 := glyphX + s.X2*scale
				y2 := glyphY - s.Y2*scale
				x3 := glyphX + s.X3*scale
				y3 := glyphY - s.Y3*scale
				pathData.WriteString(fmt.Sprintf("M%.2f %.2f C%.2f %.2f %.2f %.2f %.2f %.2f ", x0, y0, x1, y1, x2, y2, x3, y3))
			}
		}

		x += float64(g.XAdvance.At(g.Size))
	}

	// Output path element
	if pathData.Len() > 0 {
		b.WriteString(fmt.Sprintf(`<path d="%s"`, strings.TrimSpace(pathData.String())))
		if fill != nil {
			b.WriteString(fillToSVGWithContext(ctx, fill))
		} else {
			b.WriteString(` fill="#000000"`)
		}
		b.WriteString("/>\n")
	}
}

// renderImage renders an image to SVG.
func (r *Renderer) renderImage(b *strings.Builder, img *pages.Image, size layout.Size, pos layout.Point) {
	// Determine MIME type
	var mimeType string
	switch img.Format {
	case pages.ImageFormatJPEG:
		mimeType = "image/jpeg"
	case pages.ImageFormatPNG:
		mimeType = "image/png"
	default:
		// For raw format, skip or encode as PNG
		return
	}

	// Encode image data as base64
	encoded := base64.StdEncoding.EncodeToString(img.Data)

	b.WriteString(fmt.Sprintf(`<image x="%g" y="%g" width="%g" height="%g" href="data:%s;base64,%s"/>`,
		float64(pos.X), float64(pos.Y), float64(size.Width), float64(size.Height),
		mimeType, encoded))
	b.WriteString("\n")
}

// RenderDecoFrame renders a decoration frame to SVG.
func (r *Renderer) RenderDecoFrame(b *strings.Builder, frame *inline.DecoFrame, origin layout.Point) {
	for _, entry := range frame.Items {
		pos := layout.Point{
			X: origin.X + layout.Abs(entry.Pos.X),
			Y: origin.Y + layout.Abs(entry.Pos.Y),
		}
		r.renderDecoFrameItem(b, entry.Item, pos)
	}
}

// renderDecoFrameItem renders a single decoration frame item.
func (r *Renderer) renderDecoFrameItem(b *strings.Builder, item inline.DecoFrameItem, pos layout.Point) {
	switch it := item.(type) {
	case inline.DecoShapeItem:
		r.renderDecoShape(b, it.Shape, pos)
	case inline.DecoTextFrameItem:
		// Text in deco frames - would need baseline info
		// For now, skip as main text rendering handles this
	}
}

// renderDecoShape renders a decoration shape to SVG.
func (r *Renderer) renderDecoShape(b *strings.Builder, shape interface{}, pos layout.Point) {
	switch s := shape.(type) {
	case inline.DecoLineShape:
		r.renderLine(b, pos.X, pos.Y, pos.X+layout.Abs(s.Target.X), pos.Y+layout.Abs(s.Target.Y), &s.Stroke)

	case inline.DecoRectShape:
		r.renderRect(b, pos.X, pos.Y, layout.Abs(s.Size.Width), layout.Abs(s.Size.Height), s.Radius, s.Fill, s.Stroke)
	}
}

// renderLine renders a line to SVG.
func (r *Renderer) renderLine(b *strings.Builder, x1, y1, x2, y2 layout.Abs, stroke *inline.FixedStroke) {
	r.renderLineWithContext(nil, b, x1, y1, x2, y2, stroke)
}

// renderLineWithContext renders a line to SVG with gradient context.
func (r *Renderer) renderLineWithContext(ctx *renderContext, b *strings.Builder, x1, y1, x2, y2 layout.Abs, stroke *inline.FixedStroke) {
	b.WriteString(fmt.Sprintf(`<line x1="%g" y1="%g" x2="%g" y2="%g"`,
		float64(x1), float64(y1), float64(x2), float64(y2)))

	if stroke != nil {
		b.WriteString(strokeToSVGWithContext(ctx, stroke))
	}

	b.WriteString("/>\n")
}

// renderRect renders a rectangle to SVG.
func (r *Renderer) renderRect(b *strings.Builder, x, y, w, h, radius layout.Abs, fill interface{}, stroke *inline.FixedStroke) {
	r.renderRectWithContext(nil, b, x, y, w, h, radius, fill, stroke)
}

// renderRectWithContext renders a rectangle to SVG with gradient context.
func (r *Renderer) renderRectWithContext(ctx *renderContext, b *strings.Builder, x, y, w, h, radius layout.Abs, fill interface{}, stroke *inline.FixedStroke) {
	if radius > 0 {
		b.WriteString(fmt.Sprintf(`<rect x="%g" y="%g" width="%g" height="%g" rx="%g" ry="%g"`,
			float64(x), float64(y), float64(w), float64(h), float64(radius), float64(radius)))
	} else {
		b.WriteString(fmt.Sprintf(`<rect x="%g" y="%g" width="%g" height="%g"`,
			float64(x), float64(y), float64(w), float64(h)))
	}

	if fill != nil {
		b.WriteString(fillToSVGWithContext(ctx, fill))
	} else {
		b.WriteString(` fill="none"`)
	}

	if stroke != nil {
		b.WriteString(strokeToSVGWithContext(ctx, stroke))
	}

	b.WriteString("/>\n")
}

// DrawLine draws a line from (x1, y1) to (x2, y2).
func (r *Renderer) DrawLine(b *strings.Builder, x1, y1, x2, y2 layout.Abs, stroke *inline.FixedStroke) {
	r.renderLine(b, x1, y1, x2, y2, stroke)
}

// DrawRect draws a rectangle.
func (r *Renderer) DrawRect(b *strings.Builder, x, y, w, h layout.Abs, fill interface{}, stroke *inline.FixedStroke) {
	r.renderRect(b, x, y, w, h, 0, fill, stroke)
}

// DrawRoundedRect draws a rounded rectangle.
func (r *Renderer) DrawRoundedRect(b *strings.Builder, x, y, w, h, radius layout.Abs, fill interface{}, stroke *inline.FixedStroke) {
	r.renderRect(b, x, y, w, h, radius, fill, stroke)
}

// DrawPath draws a path from segments.
func (r *Renderer) DrawPath(b *strings.Builder, segments []inline.PathSegment, origin layout.Point, stroke *inline.FixedStroke, fill interface{}) {
	r.DrawPathWithContext(nil, b, segments, origin, stroke, fill)
}

// DrawPathWithContext draws a path from segments with gradient context.
func (r *Renderer) DrawPathWithContext(ctx *renderContext, b *strings.Builder, segments []inline.PathSegment, origin layout.Point, stroke *inline.FixedStroke, fill interface{}) {
	if len(segments) == 0 {
		return
	}

	pathData := segmentsToSVGPath(segments, origin)

	b.WriteString(fmt.Sprintf(`<path d="%s"`, pathData))

	if fill != nil {
		b.WriteString(fillToSVGWithContext(ctx, fill))
	} else {
		b.WriteString(` fill="none"`)
	}

	if stroke != nil {
		b.WriteString(strokeToSVGWithContext(ctx, stroke))
	}

	b.WriteString("/>\n")
}

// segmentsToSVGPath converts path segments to SVG path data.
func segmentsToSVGPath(segments []inline.PathSegment, origin layout.Point) string {
	var b strings.Builder

	for _, seg := range segments {
		switch s := seg.(type) {
		case *inline.LineSegment:
			// Move to start, line to end
			b.WriteString(fmt.Sprintf("M%g %g L%g %g ",
				s.X0+float64(origin.X), s.Y0+float64(origin.Y),
				s.X1+float64(origin.X), s.Y1+float64(origin.Y)))

		case *inline.QuadSegment:
			// Move to start, quadratic curve
			b.WriteString(fmt.Sprintf("M%g %g Q%g %g %g %g ",
				s.X0+float64(origin.X), s.Y0+float64(origin.Y),
				s.X1+float64(origin.X), s.Y1+float64(origin.Y),
				s.X2+float64(origin.X), s.Y2+float64(origin.Y)))

		case *inline.CubicSegment:
			// Move to start, cubic curve
			b.WriteString(fmt.Sprintf("M%g %g C%g %g %g %g %g %g ",
				s.X0+float64(origin.X), s.Y0+float64(origin.Y),
				s.X1+float64(origin.X), s.Y1+float64(origin.Y),
				s.X2+float64(origin.X), s.Y2+float64(origin.Y),
				s.X3+float64(origin.X), s.Y3+float64(origin.Y)))
		}
	}

	return strings.TrimSpace(b.String())
}

// colorToSVG converts a pages.Color to SVG color string.
func colorToSVG(c *pages.Color) string {
	if c.A == 255 {
		return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
	}
	return fmt.Sprintf("rgba(%d,%d,%d,%.3f)", c.R, c.G, c.B, float64(c.A)/255.0)
}

// visualizeColorToSVG converts a visualize.Color to SVG color string.
func visualizeColorToSVG(c *visualize.Color) string {
	if c == nil {
		return "#000000"
	}
	if c.A == 255 {
		return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
	}
	return fmt.Sprintf("rgba(%d,%d,%d,%.3f)", c.R, c.G, c.B, float64(c.A)/255.0)
}

// writeGradientDef writes an SVG gradient definition.
func writeGradientDef(b *strings.Builder, g *visualize.Gradient, id string) {
	if g == nil {
		return
	}

	switch g.Kind {
	case visualize.GradientKindLinear:
		writeLinearGradientDef(b, g, id)
	case visualize.GradientKindRadial:
		writeRadialGradientDef(b, g, id)
	case visualize.GradientKindConic:
		// SVG doesn't support conic gradients directly
		// Fall back to linear gradient as approximation
		writeLinearGradientDef(b, g, id)
	}
}

// writeLinearGradientDef writes an SVG linear gradient definition.
func writeLinearGradientDef(b *strings.Builder, g *visualize.Gradient, id string) {
	// Calculate gradient direction from angle
	angle := 0.0
	if a := g.GetAngle(); a != nil {
		angle = *a
	}

	// Convert angle to x1,y1,x2,y2 coordinates
	// SVG gradients use percentage coordinates (0-100%)
	// Angle 0 = left to right, pi/2 = top to bottom
	x1 := 50 - 50*math.Cos(angle)
	y1 := 50 - 50*math.Sin(angle)
	x2 := 50 + 50*math.Cos(angle)
	y2 := 50 + 50*math.Sin(angle)

	b.WriteString(fmt.Sprintf(`<linearGradient id="%s" x1="%.1f%%" y1="%.1f%%" x2="%.1f%%" y2="%.1f%%">`,
		id, x1, y1, x2, y2))
	b.WriteString("\n")

	writeGradientStops(b, g)

	b.WriteString("</linearGradient>\n")
}

// writeRadialGradientDef writes an SVG radial gradient definition.
func writeRadialGradientDef(b *strings.Builder, g *visualize.Gradient, id string) {
	// Get center point (default 50%, 50%)
	cx, cy := 50.0, 50.0
	if center := g.GetCenter(); center != nil {
		cx = center[0] * 100
		cy = center[1] * 100
	}

	// Get radius (default 50%)
	r := 50.0
	if radius := g.GetRadius(); radius != nil {
		r = *radius * 100
	}

	// Get focal point
	fx, fy := cx, cy
	if focal := g.GetFocalCenter(); focal != nil {
		fx = focal[0] * 100
		fy = focal[1] * 100
	}

	// Get focal radius
	fr := 0.0
	if focalRadius := g.GetFocalRadius(); focalRadius != nil {
		fr = *focalRadius * 100
	}

	if fr > 0 || (fx != cx || fy != cy) {
		// Use focal point attributes
		b.WriteString(fmt.Sprintf(`<radialGradient id="%s" cx="%.1f%%" cy="%.1f%%" r="%.1f%%" fx="%.1f%%" fy="%.1f%%" fr="%.1f%%">`,
			id, cx, cy, r, fx, fy, fr))
	} else {
		b.WriteString(fmt.Sprintf(`<radialGradient id="%s" cx="%.1f%%" cy="%.1f%%" r="%.1f%%">`,
			id, cx, cy, r))
	}
	b.WriteString("\n")

	writeGradientStops(b, g)

	b.WriteString("</radialGradient>\n")
}

// writeGradientStops writes the stop elements for a gradient.
func writeGradientStops(b *strings.Builder, g *visualize.Gradient) {
	stops := g.GetStops()
	if len(stops) == 0 {
		return
	}

	// Normalize stops to have offsets
	normalizedStops := normalizeGradientStops(stops)

	for _, stop := range normalizedStops {
		offset := 0.0
		if stop.Offset != nil {
			offset = *stop.Offset * 100
		}

		color := visualizeColorToSVG(stop.Color)
		b.WriteString(fmt.Sprintf(`  <stop offset="%.1f%%" stop-color="%s"/>`, offset, color))
		b.WriteString("\n")
	}
}

// normalizeGradientStops ensures all stops have offsets.
func normalizeGradientStops(stops []visualize.GradientStop) []visualize.GradientStop {
	if len(stops) == 0 {
		return stops
	}

	result := make([]visualize.GradientStop, len(stops))
	copy(result, stops)

	// First and last default to 0 and 1
	if result[0].Offset == nil {
		offset := 0.0
		result[0].Offset = &offset
	}
	if len(result) > 1 && result[len(result)-1].Offset == nil {
		offset := 1.0
		result[len(result)-1].Offset = &offset
	}

	// Distribute remaining stops evenly
	for i := 1; i < len(result)-1; i++ {
		if result[i].Offset == nil {
			// Find next defined stop
			j := i + 1
			for j < len(result) && result[j].Offset == nil {
				j++
			}
			// Interpolate
			startOffset := *result[i-1].Offset
			endOffset := *result[j].Offset
			count := j - i + 1
			for k := i; k < j; k++ {
				offset := startOffset + (endOffset-startOffset)*float64(k-i+1)/float64(count)
				result[k].Offset = &offset
			}
		}
	}

	return result
}

// fillToSVG converts a fill value to SVG fill attribute.
func fillToSVG(fill interface{}) string {
	return fillToSVGWithContext(nil, fill)
}

// fillToSVGWithContext converts a fill value to SVG fill attribute, registering gradients.
func fillToSVGWithContext(ctx *renderContext, fill interface{}) string {
	switch f := fill.(type) {
	case *pages.Color:
		return fmt.Sprintf(` fill="%s"`, colorToSVG(f))
	case pages.Color:
		return fmt.Sprintf(` fill="%s"`, colorToSVG(&f))
	case *visualize.Gradient:
		if ctx != nil && f != nil {
			id := ctx.registerGradient(f)
			return fmt.Sprintf(` fill="url(#%s)"`, id)
		}
		return ` fill="#000000"`
	case *visualize.Color:
		return fmt.Sprintf(` fill="%s"`, visualizeColorToSVG(f))
	default:
		// For unknown fill types, use black as fallback
		return ` fill="#000000"`
	}
}

// strokeToSVG converts a FixedStroke to SVG stroke attributes.
func strokeToSVG(stroke *inline.FixedStroke) string {
	return strokeToSVGWithContext(nil, stroke)
}

// strokeToSVGWithContext converts a FixedStroke to SVG stroke attributes, registering gradients.
func strokeToSVGWithContext(ctx *renderContext, stroke *inline.FixedStroke) string {
	var b strings.Builder

	// Stroke color/gradient
	switch p := stroke.Paint.(type) {
	case *pages.Color:
		b.WriteString(fmt.Sprintf(` stroke="%s"`, colorToSVG(p)))
	case pages.Color:
		b.WriteString(fmt.Sprintf(` stroke="%s"`, colorToSVG(&p)))
	case *visualize.Gradient:
		if ctx != nil && p != nil {
			id := ctx.registerGradient(p)
			b.WriteString(fmt.Sprintf(` stroke="url(#%s)"`, id))
		} else {
			b.WriteString(` stroke="#000000"`)
		}
	case *visualize.Color:
		b.WriteString(fmt.Sprintf(` stroke="%s"`, visualizeColorToSVG(p)))
	default:
		b.WriteString(` stroke="#000000"`)
	}

	// Stroke width
	b.WriteString(fmt.Sprintf(` stroke-width="%g"`, float64(stroke.Thickness)))

	// Line cap
	switch stroke.LineCap {
	case inline.LineCapButt:
		b.WriteString(` stroke-linecap="butt"`)
	case inline.LineCapRound:
		b.WriteString(` stroke-linecap="round"`)
	case inline.LineCapSquare:
		b.WriteString(` stroke-linecap="square"`)
	}

	// Line join
	switch stroke.LineJoin {
	case inline.LineJoinMiter:
		b.WriteString(` stroke-linejoin="miter"`)
	case inline.LineJoinRound:
		b.WriteString(` stroke-linejoin="round"`)
	case inline.LineJoinBevel:
		b.WriteString(` stroke-linejoin="bevel"`)
	}

	// Dash array
	if len(stroke.DashArray) > 0 {
		var dashes []string
		for _, d := range stroke.DashArray {
			dashes = append(dashes, fmt.Sprintf("%g", float64(d)))
		}
		b.WriteString(fmt.Sprintf(` stroke-dasharray="%s"`, strings.Join(dashes, " ")))

		// Dash offset
		if stroke.DashPhase != 0 {
			b.WriteString(fmt.Sprintf(` stroke-dashoffset="%g"`, float64(stroke.DashPhase)))
		}
	}

	return b.String()
}

// escapeXML escapes special XML characters.
func escapeXML(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '&':
			b.WriteString("&amp;")
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&apos;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
