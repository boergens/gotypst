package svg

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/boergens/gotypst/layout"
	"github.com/boergens/gotypst/layout/flow"
	"github.com/boergens/gotypst/layout/inline"
	"github.com/boergens/gotypst/layout/pages"
)

// Renderer generates SVG content from laid out frames.
type Renderer struct {
	// FontMap maps font faces to font family names for SVG.
	FontMap FontMapper
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
	var b strings.Builder

	width := page.Frame.Size.Width
	height := page.Frame.Size.Height

	// SVG header
	b.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%g" height="%g" viewBox="0 0 %g %g">`,
		float64(width), float64(height), float64(width), float64(height)))
	b.WriteString("\n")

	// Fill page background if specified
	if page.Fill != nil && page.Fill.Color != nil {
		b.WriteString(fmt.Sprintf(`<rect width="%g" height="%g" fill="%s"/>`,
			float64(width), float64(height), colorToSVG(page.Fill.Color)))
		b.WriteString("\n")
	}

	// Render frame content
	r.renderPagesFrame(&b, &page.Frame, layout.Point{X: 0, Y: 0})

	b.WriteString("</svg>")

	return b.String()
}

// renderPagesFrame renders a pages.Frame (from layout/pages package).
func (r *Renderer) renderPagesFrame(b *strings.Builder, frame *pages.Frame, origin layout.Point) {
	for _, item := range frame.Items {
		pos := layout.Point{
			X: origin.X + item.Pos.X,
			Y: origin.Y + item.Pos.Y,
		}
		r.renderPagesFrameItem(b, item.Item, pos)
	}
}

// renderPagesFrameItem renders a single item from a pages.Frame.
func (r *Renderer) renderPagesFrameItem(b *strings.Builder, item pages.FrameItem, pos layout.Point) {
	switch it := item.(type) {
	case pages.GroupItem:
		// Nested frame - recurse
		r.renderPagesFrame(b, &it.Frame, pos)
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
	b.WriteString(fmt.Sprintf(`<line x1="%g" y1="%g" x2="%g" y2="%g"`,
		float64(x1), float64(y1), float64(x2), float64(y2)))

	if stroke != nil {
		b.WriteString(strokeToSVG(stroke))
	}

	b.WriteString("/>\n")
}

// renderRect renders a rectangle to SVG.
func (r *Renderer) renderRect(b *strings.Builder, x, y, w, h, radius layout.Abs, fill interface{}, stroke *inline.FixedStroke) {
	if radius > 0 {
		b.WriteString(fmt.Sprintf(`<rect x="%g" y="%g" width="%g" height="%g" rx="%g" ry="%g"`,
			float64(x), float64(y), float64(w), float64(h), float64(radius), float64(radius)))
	} else {
		b.WriteString(fmt.Sprintf(`<rect x="%g" y="%g" width="%g" height="%g"`,
			float64(x), float64(y), float64(w), float64(h)))
	}

	if fill != nil {
		b.WriteString(fillToSVG(fill))
	} else {
		b.WriteString(` fill="none"`)
	}

	if stroke != nil {
		b.WriteString(strokeToSVG(stroke))
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
	if len(segments) == 0 {
		return
	}

	pathData := segmentsToSVGPath(segments, origin)

	b.WriteString(fmt.Sprintf(`<path d="%s"`, pathData))

	if fill != nil {
		b.WriteString(fillToSVG(fill))
	} else {
		b.WriteString(` fill="none"`)
	}

	if stroke != nil {
		b.WriteString(strokeToSVG(stroke))
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

// fillToSVG converts a fill value to SVG fill attribute.
func fillToSVG(fill interface{}) string {
	switch f := fill.(type) {
	case *pages.Color:
		return fmt.Sprintf(` fill="%s"`, colorToSVG(f))
	case pages.Color:
		return fmt.Sprintf(` fill="%s"`, colorToSVG(&f))
	default:
		// For unknown fill types, use black as fallback
		return ` fill="#000000"`
	}
}

// strokeToSVG converts a FixedStroke to SVG stroke attributes.
func strokeToSVG(stroke *inline.FixedStroke) string {
	var b strings.Builder

	// Stroke color
	switch p := stroke.Paint.(type) {
	case *pages.Color:
		b.WriteString(fmt.Sprintf(` stroke="%s"`, colorToSVG(p)))
	case pages.Color:
		b.WriteString(fmt.Sprintf(` stroke="%s"`, colorToSVG(&p)))
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
