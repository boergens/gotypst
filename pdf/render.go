package pdf

import (
	"github.com/boergens/gotypst/layout"
	"github.com/boergens/gotypst/layout/flow"
	"github.com/boergens/gotypst/layout/inline"
	"github.com/boergens/gotypst/layout/pages"
)

// Renderer generates PDF content streams from laid out frames.
type Renderer struct {
	// FontMap maps font faces to PDF font resource names (e.g., "/F1").
	FontMap FontMapper

	// pageHeight is used to convert coordinates (PDF has origin at bottom-left).
	pageHeight layout.Abs
}

// FontMapper maps fonts to PDF resource names.
type FontMapper interface {
	// FontName returns the PDF resource name for a font (e.g., "/F1").
	// Returns empty string if font is not registered.
	FontName(face interface{}) string
}

// DefaultFontMapper is a simple font mapper that returns a default font.
type DefaultFontMapper struct{}

// FontName returns a default font name.
func (DefaultFontMapper) FontName(face interface{}) string {
	return "/F1"
}

// NewRenderer creates a new PDF renderer.
func NewRenderer() *Renderer {
	return &Renderer{
		FontMap: DefaultFontMapper{},
	}
}

// RenderDocument renders a full document to content streams.
// Returns a slice of content streams, one per page.
func (r *Renderer) RenderDocument(doc *pages.PagedDocument) []*ContentStream {
	streams := make([]*ContentStream, len(doc.Pages))

	for i, page := range doc.Pages {
		streams[i] = r.RenderPage(&page)
	}

	return streams
}

// RenderPage renders a single page to a content stream.
func (r *Renderer) RenderPage(page *pages.Page) *ContentStream {
	cs := NewContentStream()
	r.pageHeight = page.Frame.Size.Height

	// Fill page background if specified
	if page.Fill != nil && page.Fill.Color != nil {
		cs.SaveState()
		cs.SetFillColor(&Color{
			R: page.Fill.Color.R,
			G: page.Fill.Color.G,
			B: page.Fill.Color.B,
			A: page.Fill.Color.A,
		})
		cs.Rectangle(0, 0, page.Frame.Size.Width, page.Frame.Size.Height)
		cs.Fill()
		cs.RestoreState()
	}

	// Render frame content
	r.renderPagesFrame(cs, &page.Frame, layout.Point{X: 0, Y: 0})

	return cs
}

// renderPagesFrame renders a pages.Frame (from layout/pages package).
func (r *Renderer) renderPagesFrame(cs *ContentStream, frame *pages.Frame, origin layout.Point) {
	for _, item := range frame.Items {
		pos := layout.Point{
			X: origin.X + item.Pos.X,
			Y: origin.Y + item.Pos.Y,
		}
		r.renderPagesFrameItem(cs, item.Item, pos)
	}
}

// renderPagesFrameItem renders a single item from a pages.Frame.
func (r *Renderer) renderPagesFrameItem(cs *ContentStream, item pages.FrameItem, pos layout.Point) {
	switch it := item.(type) {
	case pages.GroupItem:
		// Nested frame - recurse
		r.renderPagesFrame(cs, &it.Frame, pos)
	case pages.TagItem:
		// Tags are metadata, not rendered
	case pages.TextItem:
		// Render text directly
		r.renderSimpleText(cs, it.Text, it.FontSize, pos)
	}
}

// RenderFlowFrame renders a flow.Frame to a content stream.
func (r *Renderer) RenderFlowFrame(cs *ContentStream, frame *flow.Frame, origin layout.Point) {
	for _, entry := range frame.Items() {
		pos := layout.Point{
			X: origin.X + entry.Pos.X,
			Y: origin.Y + entry.Pos.Y,
		}
		r.renderFlowFrameItem(cs, entry.Item, pos)
	}
}

// renderFlowFrameItem renders a single item from a flow.Frame.
func (r *Renderer) renderFlowFrameItem(cs *ContentStream, item flow.FrameItem, pos layout.Point) {
	switch it := item.(type) {
	case flow.FrameItemFrame:
		// Nested frame - recurse
		r.RenderFlowFrame(cs, &it.Frame, pos)
	case flow.FrameItemTag:
		// Tags are metadata, not rendered
	case flow.FrameItemLink:
		// Links are handled separately (PDF annotations)
	}
}

// RenderInlineFrame renders an inline.FinalFrame to a content stream.
func (r *Renderer) RenderInlineFrame(cs *ContentStream, frame *inline.FinalFrame, origin layout.Point) {
	for _, entry := range frame.Items {
		pos := layout.Point{
			X: origin.X + layout.Abs(entry.Pos.X),
			Y: origin.Y + layout.Abs(entry.Pos.Y),
		}
		r.renderInlineFrameItem(cs, entry.Item, pos, frame.Baseline)
	}
}

// renderInlineFrameItem renders a single item from an inline.FinalFrame.
func (r *Renderer) renderInlineFrameItem(cs *ContentStream, item inline.FinalFrameItem, pos layout.Point, baseline layout.Abs) {
	switch it := item.(type) {
	case inline.FinalTextItem:
		r.renderShapedText(cs, it.Text, pos, baseline)
	}
}

// renderShapedText renders shaped text to the content stream.
func (r *Renderer) renderShapedText(cs *ContentStream, text *inline.ShapedText, pos layout.Point, baseline layout.Abs) {
	if text == nil || text.Glyphs.Len() == 0 {
		return
	}

	glyphs := text.Glyphs.Kept()
	if len(glyphs) == 0 {
		return
	}

	// Get font info from first glyph
	firstGlyph := &glyphs[0]
	fontName := r.FontMap.FontName(firstGlyph.Font)
	fontSize := firstGlyph.Size

	cs.BeginText()

	// Set font
	cs.SetFont(fontName, fontSize)

	// Convert to PDF coordinates (origin at bottom-left)
	// pos.Y is top-down, so we need to flip it
	pdfX := pos.X
	pdfY := r.pageHeight - pos.Y - baseline

	cs.SetTextMatrixPos(pdfX, pdfY)

	// Build TJ array with glyph positioning
	var items []TextPositionItem
	var currentX layout.Abs

	for i := range glyphs {
		g := &glyphs[i]

		// Handle x-offset if present
		if g.XOffset != 0 {
			// TJ offsets are in thousandths of em, negative moves right
			offsetUnits := -float64(g.XOffset) * 1000
			items = append(items, TextPositionOffset(offsetUnits))
		}

		// Add the glyph character
		// Note: In production, this would use glyph IDs and proper encoding
		items = append(items, TextPositionString(string(g.Char)))

		// Track position for next glyph
		advance := g.XAdvance.At(g.Size)
		currentX += advance

		// If there's extra spacing between glyphs (beyond normal advance),
		// add a positioning adjustment
		if i+1 < len(glyphs) {
			// This would handle justification adjustments
			// For now, we rely on standard advance widths
		}
	}

	cs.ShowTextWithPositioning(items)
	cs.EndText()
}

// renderSimpleText renders simple text directly at a position.
// This is a minimal implementation for basic text rendering without shaping.
func (r *Renderer) renderSimpleText(cs *ContentStream, text string, fontSize layout.Abs, pos layout.Point) {
	if text == "" {
		return
	}

	// Convert to PDF coordinates (origin at bottom-left)
	pdfX := pos.X
	pdfY := r.pageHeight - pos.Y - fontSize

	cs.BeginText()
	cs.SetFont("/F1", fontSize)
	cs.SetTextMatrixPos(pdfX, pdfY)
	cs.ShowText(text)
	cs.EndText()
}

// RenderDecoFrame renders a decoration frame to the content stream.
func (r *Renderer) RenderDecoFrame(cs *ContentStream, frame *inline.DecoFrame, origin layout.Point) {
	for _, entry := range frame.Items {
		pos := layout.Point{
			X: origin.X + layout.Abs(entry.Pos.X),
			Y: origin.Y + layout.Abs(entry.Pos.Y),
		}
		r.renderDecoFrameItem(cs, entry.Item, pos)
	}
}

// renderDecoFrameItem renders a single decoration frame item.
func (r *Renderer) renderDecoFrameItem(cs *ContentStream, item inline.DecoFrameItem, pos layout.Point) {
	switch it := item.(type) {
	case inline.DecoShapeItem:
		r.renderDecoShape(cs, it.Shape, pos)
	case inline.DecoTextFrameItem:
		// Text in deco frames - would need baseline info
		// For now, skip as main text rendering handles this
	}
}

// renderDecoShape renders a decoration shape.
func (r *Renderer) renderDecoShape(cs *ContentStream, shape interface{}, pos layout.Point) {
	// Convert to PDF coordinates
	pdfX := pos.X
	pdfY := r.pageHeight - pos.Y

	switch s := shape.(type) {
	case inline.DecoLineShape:
		cs.SaveState()
		cs.ApplyStrokeStyle(&s.Stroke)

		cs.MoveTo(pdfX, pdfY)
		cs.LineTo(pdfX+layout.Abs(s.Target.X), pdfY-layout.Abs(s.Target.Y))
		cs.Stroke()

		cs.RestoreState()

	case inline.DecoRectShape:
		cs.SaveState()

		if s.Fill != nil {
			if c, ok := s.Fill.(*Color); ok {
				cs.SetFillColor(c)
			}
		}

		if s.Stroke != nil {
			cs.ApplyStrokeStyle(s.Stroke)
		}

		// Draw rectangle (y-flipped)
		rectY := pdfY - layout.Abs(s.Size.Height)
		if s.Radius > 0 {
			cs.RoundedRectangle(pdfX, rectY, layout.Abs(s.Size.Width), layout.Abs(s.Size.Height), s.Radius)
		} else {
			cs.Rectangle(pdfX, rectY, layout.Abs(s.Size.Width), layout.Abs(s.Size.Height))
		}

		if s.Fill != nil && s.Stroke != nil {
			cs.FillAndStroke()
		} else if s.Fill != nil {
			cs.Fill()
		} else if s.Stroke != nil {
			cs.Stroke()
		}

		cs.RestoreState()
	}
}

// DrawLine draws a line from (x1, y1) to (x2, y2).
func (r *Renderer) DrawLine(cs *ContentStream, x1, y1, x2, y2 layout.Abs, stroke *inline.FixedStroke) {
	cs.SaveState()
	cs.ApplyStrokeStyle(stroke)

	// Convert to PDF coordinates
	pdfY1 := r.pageHeight - y1
	pdfY2 := r.pageHeight - y2

	cs.MoveTo(x1, pdfY1)
	cs.LineTo(x2, pdfY2)
	cs.Stroke()

	cs.RestoreState()
}

// DrawRect draws a rectangle.
func (r *Renderer) DrawRect(cs *ContentStream, x, y, w, h layout.Abs, fill *Color, stroke *inline.FixedStroke) {
	cs.SaveState()

	if fill != nil {
		cs.SetFillColor(fill)
	}
	if stroke != nil {
		cs.ApplyStrokeStyle(stroke)
	}

	// Convert to PDF coordinates
	pdfY := r.pageHeight - y - h

	cs.Rectangle(x, pdfY, w, h)

	if fill != nil && stroke != nil {
		cs.FillAndStroke()
	} else if fill != nil {
		cs.Fill()
	} else if stroke != nil {
		cs.Stroke()
	}

	cs.RestoreState()
}

// DrawRoundedRect draws a rounded rectangle.
func (r *Renderer) DrawRoundedRect(cs *ContentStream, x, y, w, h, radius layout.Abs, fill *Color, stroke *inline.FixedStroke) {
	cs.SaveState()

	if fill != nil {
		cs.SetFillColor(fill)
	}
	if stroke != nil {
		cs.ApplyStrokeStyle(stroke)
	}

	// Convert to PDF coordinates
	pdfY := r.pageHeight - y - h

	cs.RoundedRectangle(x, pdfY, w, h, radius)

	if fill != nil && stroke != nil {
		cs.FillAndStroke()
	} else if fill != nil {
		cs.Fill()
	} else if stroke != nil {
		cs.Stroke()
	}

	cs.RestoreState()
}

// DrawPath draws a path from segments.
func (r *Renderer) DrawPath(cs *ContentStream, segments []inline.PathSegment, origin layout.Point, stroke *inline.FixedStroke, fill *Color) {
	if len(segments) == 0 {
		return
	}

	cs.SaveState()

	if fill != nil {
		cs.SetFillColor(fill)
	}
	if stroke != nil {
		cs.ApplyStrokeStyle(stroke)
	}

	for _, seg := range segments {
		switch s := seg.(type) {
		case *inline.LineSegment:
			// First point - move to
			pdfY0 := r.pageHeight - layout.Abs(s.Y0) - origin.Y
			cs.MoveTo(layout.Abs(s.X0)+origin.X, pdfY0)

			// Line to end
			pdfY1 := r.pageHeight - layout.Abs(s.Y1) - origin.Y
			cs.LineTo(layout.Abs(s.X1)+origin.X, pdfY1)

		case *inline.QuadSegment:
			// Convert quadratic to cubic bezier for PDF
			// P0 = start, P1 = control, P2 = end
			// Cubic: P0, P0 + 2/3*(P1-P0), P2 + 2/3*(P1-P2), P2
			pdfY0 := float64(r.pageHeight) - s.Y0 - float64(origin.Y)
			pdfY1 := float64(r.pageHeight) - s.Y1 - float64(origin.Y)
			pdfY2 := float64(r.pageHeight) - s.Y2 - float64(origin.Y)

			cs.MoveTo(layout.Abs(s.X0)+origin.X, layout.Abs(pdfY0))

			// Convert quadratic to cubic control points
			cp1x := s.X0 + 2.0/3.0*(s.X1-s.X0)
			cp1y := pdfY0 + 2.0/3.0*(pdfY1-pdfY0)
			cp2x := s.X2 + 2.0/3.0*(s.X1-s.X2)
			cp2y := pdfY2 + 2.0/3.0*(pdfY1-pdfY2)

			cs.CurveTo(
				layout.Abs(cp1x)+origin.X, layout.Abs(cp1y),
				layout.Abs(cp2x)+origin.X, layout.Abs(cp2y),
				layout.Abs(s.X2)+origin.X, layout.Abs(pdfY2),
			)

		case *inline.CubicSegment:
			pdfY0 := float64(r.pageHeight) - s.Y0 - float64(origin.Y)
			pdfY1 := float64(r.pageHeight) - s.Y1 - float64(origin.Y)
			pdfY2 := float64(r.pageHeight) - s.Y2 - float64(origin.Y)
			pdfY3 := float64(r.pageHeight) - s.Y3 - float64(origin.Y)

			cs.MoveTo(layout.Abs(s.X0)+origin.X, layout.Abs(pdfY0))
			cs.CurveTo(
				layout.Abs(s.X1)+origin.X, layout.Abs(pdfY1),
				layout.Abs(s.X2)+origin.X, layout.Abs(pdfY2),
				layout.Abs(s.X3)+origin.X, layout.Abs(pdfY3),
			)
		}
	}

	if fill != nil && stroke != nil {
		cs.FillAndStroke()
	} else if fill != nil {
		cs.Fill()
	} else if stroke != nil {
		cs.Stroke()
	}

	cs.RestoreState()
}
