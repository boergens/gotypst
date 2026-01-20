package html

import (
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/boergens/gotypst/layout"
	"github.com/boergens/gotypst/layout/pages"
)

// Renderer generates HTML content from laid out frames.
type Renderer struct {
	// buf accumulates HTML output.
	buf strings.Builder
	// indent tracks current indentation level.
	indent int
}

// NewRenderer creates a new HTML renderer.
func NewRenderer() *Renderer {
	return &Renderer{}
}

// RenderDocument renders a full document to HTML.
func (r *Renderer) RenderDocument(doc *pages.PagedDocument, w io.Writer) error {
	r.buf.Reset()

	// Write HTML preamble
	r.writeln("<!DOCTYPE html>")
	r.writeln("<html>")
	r.writeln("<head>")
	r.indent++
	r.writeln(`<meta charset="UTF-8">`)
	r.writeln(`<meta name="viewport" content="width=device-width, initial-scale=1.0">`)

	// Document metadata
	if doc.Info.Title != nil {
		r.writef("<title>%s</title>\n", escapeHTML(*doc.Info.Title))
	} else {
		r.writeln("<title>Document</title>")
	}

	// Add base styles
	r.writeln("<style>")
	r.indent++
	r.writeln("* { margin: 0; padding: 0; box-sizing: border-box; }")
	r.writeln(".page { position: relative; margin: 20px auto; background: white; box-shadow: 0 2px 8px rgba(0,0,0,0.1); overflow: hidden; }")
	r.writeln(".frame { position: absolute; }")
	r.writeln(".text { position: absolute; white-space: pre; font-family: serif; }")
	r.writeln(".image { position: absolute; }")
	r.indent--
	r.writeln("</style>")
	r.indent--
	r.writeln("</head>")
	r.writeln("<body>")
	r.indent++

	// Render each page
	for i, page := range doc.Pages {
		if err := r.renderPage(&page, i); err != nil {
			return fmt.Errorf("rendering page %d: %w", i, err)
		}
	}

	r.indent--
	r.writeln("</body>")
	r.writeln("</html>")

	_, err := w.Write([]byte(r.buf.String()))
	return err
}

// renderPage renders a single page to HTML.
func (r *Renderer) renderPage(page *pages.Page, pageNum int) error {
	width := float64(page.Frame.Size.Width)
	height := float64(page.Frame.Size.Height)

	// Page container with background fill
	style := fmt.Sprintf("width: %.2fpt; height: %.2fpt;", width, height)
	if page.Fill != nil && page.Fill.Color != nil {
		c := page.Fill.Color
		style += fmt.Sprintf(" background-color: rgba(%d, %d, %d, %.2f);", c.R, c.G, c.B, float64(c.A)/255)
	} else {
		style += " background-color: white;"
	}

	r.writef(`<div class="page" data-page="%d" style="%s">`+"\n", pageNum+1, style)
	r.indent++

	// Render frame content
	r.renderFrame(&page.Frame, layout.Point{X: 0, Y: 0})

	r.indent--
	r.writeln("</div>")

	return nil
}

// renderFrame renders a pages.Frame recursively.
func (r *Renderer) renderFrame(frame *pages.Frame, origin layout.Point) {
	for _, item := range frame.Items {
		pos := layout.Point{
			X: origin.X + item.Pos.X,
			Y: origin.Y + item.Pos.Y,
		}
		r.renderFrameItem(item.Item, pos)
	}
}

// renderFrameItem renders a single frame item.
func (r *Renderer) renderFrameItem(item pages.FrameItem, pos layout.Point) {
	switch it := item.(type) {
	case pages.GroupItem:
		// Nested frame - render as a positioned div
		width := float64(it.Frame.Size.Width)
		height := float64(it.Frame.Size.Height)

		r.writef(`<div class="frame" style="left: %.2fpt; top: %.2fpt; width: %.2fpt; height: %.2fpt;">`+"\n",
			float64(pos.X), float64(pos.Y), width, height)
		r.indent++
		r.renderFrame(&it.Frame, layout.Point{X: 0, Y: 0})
		r.indent--
		r.writeln("</div>")

	case pages.TagItem:
		// Tags are metadata, rendered as HTML comments for debugging
		r.writef("<!-- tag: kind=%d -->\n", it.Tag.Kind)

	case pages.TextItem:
		r.renderText(it, pos)

	case pages.ImageItem:
		r.renderImage(it, pos)
	}
}

// renderText renders a text item.
func (r *Renderer) renderText(item pages.TextItem, pos layout.Point) {
	fontSize := float64(item.FontSize)

	r.writef(`<span class="text" style="left: %.2fpt; top: %.2fpt; font-size: %.2fpt;">%s</span>`+"\n",
		float64(pos.X), float64(pos.Y), fontSize, escapeHTML(item.Text))
}

// renderImage renders an image item.
func (r *Renderer) renderImage(item pages.ImageItem, pos layout.Point) {
	width := float64(item.Size.Width)
	height := float64(item.Size.Height)

	// Determine MIME type
	var mimeType string
	switch item.Image.Format {
	case pages.ImageFormatJPEG:
		mimeType = "image/jpeg"
	case pages.ImageFormatPNG:
		mimeType = "image/png"
	default:
		mimeType = "image/png" // Default to PNG for raw data
	}

	// Encode image data as base64 data URL
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(item.Image.Data))

	r.writef(`<img class="image" src="%s" style="left: %.2fpt; top: %.2fpt; width: %.2fpt; height: %.2fpt;" alt="">`+"\n",
		dataURL, float64(pos.X), float64(pos.Y), width, height)
}

// writeln writes an indented line.
func (r *Renderer) writeln(s string) {
	r.writeIndent()
	r.buf.WriteString(s)
	r.buf.WriteByte('\n')
}

// writef writes a formatted indented line.
func (r *Renderer) writef(format string, args ...interface{}) {
	r.writeIndent()
	fmt.Fprintf(&r.buf, format, args...)
}

// writeIndent writes the current indentation.
func (r *Renderer) writeIndent() {
	for i := 0; i < r.indent; i++ {
		r.buf.WriteString("  ")
	}
}

// escapeHTML escapes special HTML characters.
func escapeHTML(s string) string {
	var b strings.Builder
	for _, c := range s {
		switch c {
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '&':
			b.WriteString("&amp;")
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&#39;")
		default:
			b.WriteRune(c)
		}
	}
	return b.String()
}

// Export is a convenience function that exports a PagedDocument to HTML.
func Export(doc *pages.PagedDocument, out io.Writer) error {
	r := NewRenderer()
	return r.RenderDocument(doc, out)
}
