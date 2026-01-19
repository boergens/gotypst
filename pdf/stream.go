// Package pdf provides PDF export functionality for Typst documents.
package pdf

import (
	"fmt"
	"io"
	"strings"

	"github.com/boergens/gotypst/layout"
	"github.com/boergens/gotypst/layout/inline"
)

// ContentStream writes PDF content stream operators.
type ContentStream struct {
	buf strings.Builder
}

// NewContentStream creates a new content stream writer.
func NewContentStream() *ContentStream {
	return &ContentStream{}
}

// Bytes returns the content stream as bytes.
func (cs *ContentStream) Bytes() []byte {
	return []byte(cs.buf.String())
}

// String returns the content stream as a string.
func (cs *ContentStream) String() string {
	return cs.buf.String()
}

// WriteTo writes the content stream to a writer.
func (cs *ContentStream) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write([]byte(cs.buf.String()))
	return int64(n), err
}

// writeOp writes an operator with arguments.
func (cs *ContentStream) writeOp(op string, args ...interface{}) {
	for _, arg := range args {
		cs.writeArg(arg)
		cs.buf.WriteByte(' ')
	}
	cs.buf.WriteString(op)
	cs.buf.WriteByte('\n')
}

// writeArg writes a single argument value.
func (cs *ContentStream) writeArg(arg interface{}) {
	switch v := arg.(type) {
	case float64:
		cs.buf.WriteString(formatFloat(v))
	case layout.Abs:
		cs.buf.WriteString(formatFloat(float64(v)))
	case int:
		fmt.Fprintf(&cs.buf, "%d", v)
	case string:
		cs.buf.WriteString(v)
	default:
		fmt.Fprintf(&cs.buf, "%v", v)
	}
}

// formatFloat formats a float with appropriate precision.
func formatFloat(f float64) string {
	// Use minimal precision that preserves accuracy
	s := fmt.Sprintf("%.4f", f)
	// Trim trailing zeros after decimal point
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s
}

// Graphics State Operators

// SaveState saves the current graphics state (q operator).
func (cs *ContentStream) SaveState() {
	cs.writeOp("q")
}

// RestoreState restores the previous graphics state (Q operator).
func (cs *ContentStream) RestoreState() {
	cs.writeOp("Q")
}

// Transform applies a transformation matrix (cm operator).
// The matrix is [a b c d e f] representing:
// [ a b 0 ]
// [ c d 0 ]
// [ e f 1 ]
func (cs *ContentStream) Transform(a, b, c, d, e, f float64) {
	cs.writeOp("cm", a, b, c, d, e, f)
}

// TranslateTransform applies a translation transformation.
func (cs *ContentStream) TranslateTransform(tx, ty layout.Abs) {
	cs.Transform(1, 0, 0, 1, float64(tx), float64(ty))
}

// ScaleTransform applies a scaling transformation.
func (cs *ContentStream) ScaleTransform(sx, sy float64) {
	cs.Transform(sx, 0, 0, sy, 0, 0)
}

// Color Operators

// SetFillColorRGB sets the fill color in RGB color space (rg operator).
func (cs *ContentStream) SetFillColorRGB(r, g, b float64) {
	cs.writeOp("rg", r, g, b)
}

// SetStrokeColorRGB sets the stroke color in RGB color space (RG operator).
func (cs *ContentStream) SetStrokeColorRGB(r, g, b float64) {
	cs.writeOp("RG", r, g, b)
}

// SetFillColorGray sets the fill color in grayscale (g operator).
func (cs *ContentStream) SetFillColorGray(gray float64) {
	cs.writeOp("g", gray)
}

// SetStrokeColorGray sets the stroke color in grayscale (G operator).
func (cs *ContentStream) SetStrokeColorGray(gray float64) {
	cs.writeOp("G", gray)
}

// SetFillColorCMYK sets the fill color in CMYK color space (k operator).
func (cs *ContentStream) SetFillColorCMYK(c, m, y, k float64) {
	cs.writeOp("k", c, m, y, k)
}

// SetStrokeColorCMYK sets the stroke color in CMYK color space (K operator).
func (cs *ContentStream) SetStrokeColorCMYK(c, m, y, k float64) {
	cs.writeOp("K", c, m, y, k)
}

// Text Operators

// BeginText begins a text object (BT operator).
func (cs *ContentStream) BeginText() {
	cs.writeOp("BT")
}

// EndText ends a text object (ET operator).
func (cs *ContentStream) EndText() {
	cs.writeOp("ET")
}

// SetTextMatrix sets the text matrix (Tm operator).
func (cs *ContentStream) SetTextMatrix(a, b, c, d, e, f float64) {
	cs.writeOp("Tm", a, b, c, d, e, f)
}

// SetTextMatrixPos sets the text matrix for a simple position.
func (cs *ContentStream) SetTextMatrixPos(x, y layout.Abs) {
	cs.SetTextMatrix(1, 0, 0, 1, float64(x), float64(y))
}

// MoveText moves to start of next line, offset from start of current line (Td operator).
func (cs *ContentStream) MoveText(tx, ty layout.Abs) {
	cs.writeOp("Td", tx, ty)
}

// SetFont sets the font and size (Tf operator).
// fontName should be a PDF name (e.g., "/F1").
func (cs *ContentStream) SetFont(fontName string, size layout.Abs) {
	cs.writeOp("Tf", fontName, size)
}

// ShowText shows a text string (Tj operator).
func (cs *ContentStream) ShowText(text string) {
	cs.buf.WriteString(encodeString(text))
	cs.buf.WriteString(" Tj\n")
}

// ShowTextWithPositioning shows text with individual glyph positioning (TJ operator).
func (cs *ContentStream) ShowTextWithPositioning(items []TextPositionItem) {
	cs.buf.WriteByte('[')
	for _, item := range items {
		switch v := item.(type) {
		case TextPositionString:
			cs.buf.WriteString(encodeString(string(v)))
		case TextPositionOffset:
			// Negative value moves right in PDF coordinates
			cs.buf.WriteString(formatFloat(float64(v)))
		}
	}
	cs.buf.WriteString("] TJ\n")
}

// TextPositionItem is an item in a TJ array.
type TextPositionItem interface {
	isTextPositionItem()
}

// TextPositionString is a text string in a TJ array.
type TextPositionString string

func (TextPositionString) isTextPositionItem() {}

// TextPositionOffset is a positioning offset in a TJ array (in thousandths of em).
type TextPositionOffset float64

func (TextPositionOffset) isTextPositionItem() {}

// SetCharacterSpacing sets the character spacing (Tc operator).
func (cs *ContentStream) SetCharacterSpacing(spacing layout.Abs) {
	cs.writeOp("Tc", spacing)
}

// SetWordSpacing sets the word spacing (Tw operator).
func (cs *ContentStream) SetWordSpacing(spacing layout.Abs) {
	cs.writeOp("Tw", spacing)
}

// SetTextRenderingMode sets the text rendering mode (Tr operator).
func (cs *ContentStream) SetTextRenderingMode(mode int) {
	cs.writeOp("Tr", mode)
}

// SetTextRise sets the text rise (Ts operator).
func (cs *ContentStream) SetTextRise(rise layout.Abs) {
	cs.writeOp("Ts", rise)
}

// encodeString encodes a string for PDF.
func encodeString(s string) string {
	var buf strings.Builder
	buf.WriteByte('(')
	for _, b := range []byte(s) {
		switch b {
		case '\\', '(', ')':
			buf.WriteByte('\\')
			buf.WriteByte(b)
		case '\n':
			buf.WriteString("\\n")
		case '\r':
			buf.WriteString("\\r")
		case '\t':
			buf.WriteString("\\t")
		default:
			if b < 32 || b > 126 {
				fmt.Fprintf(&buf, "\\%03o", b)
			} else {
				buf.WriteByte(b)
			}
		}
	}
	buf.WriteByte(')')
	return buf.String()
}

// Path Operators

// MoveTo begins a new subpath (m operator).
func (cs *ContentStream) MoveTo(x, y layout.Abs) {
	cs.writeOp("m", x, y)
}

// LineTo appends a straight line segment (l operator).
func (cs *ContentStream) LineTo(x, y layout.Abs) {
	cs.writeOp("l", x, y)
}

// CurveTo appends a cubic Bezier curve (c operator).
func (cs *ContentStream) CurveTo(x1, y1, x2, y2, x3, y3 layout.Abs) {
	cs.writeOp("c", x1, y1, x2, y2, x3, y3)
}

// CurveToV appends a cubic Bezier with first control point at current point (v operator).
func (cs *ContentStream) CurveToV(x2, y2, x3, y3 layout.Abs) {
	cs.writeOp("v", x2, y2, x3, y3)
}

// CurveToY appends a cubic Bezier with second control point at endpoint (y operator).
func (cs *ContentStream) CurveToY(x1, y1, x3, y3 layout.Abs) {
	cs.writeOp("y", x1, y1, x3, y3)
}

// ClosePath closes the current subpath (h operator).
func (cs *ContentStream) ClosePath() {
	cs.writeOp("h")
}

// Rectangle appends a rectangle to the path (re operator).
func (cs *ContentStream) Rectangle(x, y, w, h layout.Abs) {
	cs.writeOp("re", x, y, w, h)
}

// RoundedRectangle draws a rounded rectangle using curves.
func (cs *ContentStream) RoundedRectangle(x, y, w, h, radius layout.Abs) {
	if radius <= 0 {
		cs.Rectangle(x, y, w, h)
		return
	}

	// Clamp radius to half the smallest dimension
	maxRadius := w / 2
	if h/2 < maxRadius {
		maxRadius = h / 2
	}
	if radius > maxRadius {
		radius = maxRadius
	}

	// Kappa for approximating circular arcs with cubic beziers
	const kappa = 0.5522847498

	k := radius * layout.Abs(kappa)

	// Start at top-left, after the corner radius
	cs.MoveTo(x+radius, y)

	// Top edge and top-right corner
	cs.LineTo(x+w-radius, y)
	cs.CurveTo(x+w-radius+k, y, x+w, y+radius-k, x+w, y+radius)

	// Right edge and bottom-right corner
	cs.LineTo(x+w, y+h-radius)
	cs.CurveTo(x+w, y+h-radius+k, x+w-radius+k, y+h, x+w-radius, y+h)

	// Bottom edge and bottom-left corner
	cs.LineTo(x+radius, y+h)
	cs.CurveTo(x+radius-k, y+h, x, y+h-radius+k, x, y+h-radius)

	// Left edge and top-left corner
	cs.LineTo(x, y+radius)
	cs.CurveTo(x, y+radius-k, x+radius-k, y, x+radius, y)

	cs.ClosePath()
}

// Path Painting Operators

// Stroke strokes the current path (S operator).
func (cs *ContentStream) Stroke() {
	cs.writeOp("S")
}

// CloseAndStroke closes and strokes the current path (s operator).
func (cs *ContentStream) CloseAndStroke() {
	cs.writeOp("s")
}

// Fill fills the current path using non-zero winding rule (f operator).
func (cs *ContentStream) Fill() {
	cs.writeOp("f")
}

// FillEvenOdd fills the current path using even-odd rule (f* operator).
func (cs *ContentStream) FillEvenOdd() {
	cs.writeOp("f*")
}

// FillAndStroke fills and strokes the current path (B operator).
func (cs *ContentStream) FillAndStroke() {
	cs.writeOp("B")
}

// FillEvenOddAndStroke fills (even-odd) and strokes the current path (B* operator).
func (cs *ContentStream) FillEvenOddAndStroke() {
	cs.writeOp("B*")
}

// CloseFillAndStroke closes, fills, and strokes the current path (b operator).
func (cs *ContentStream) CloseFillAndStroke() {
	cs.writeOp("b")
}

// CloseFillEvenOddAndStroke closes, fills (even-odd), and strokes (b* operator).
func (cs *ContentStream) CloseFillEvenOddAndStroke() {
	cs.writeOp("b*")
}

// EndPath ends the path without painting (n operator).
func (cs *ContentStream) EndPath() {
	cs.writeOp("n")
}

// Graphics State Operators

// SetLineWidth sets the line width (w operator).
func (cs *ContentStream) SetLineWidth(width layout.Abs) {
	cs.writeOp("w", width)
}

// SetLineCap sets the line cap style (J operator).
// 0 = butt cap, 1 = round cap, 2 = projecting square cap
func (cs *ContentStream) SetLineCap(cap int) {
	cs.writeOp("J", cap)
}

// SetLineJoin sets the line join style (j operator).
// 0 = miter join, 1 = round join, 2 = bevel join
func (cs *ContentStream) SetLineJoin(join int) {
	cs.writeOp("j", join)
}

// SetMiterLimit sets the miter limit (M operator).
func (cs *ContentStream) SetMiterLimit(limit float64) {
	cs.writeOp("M", limit)
}

// SetDashPattern sets the line dash pattern (d operator).
func (cs *ContentStream) SetDashPattern(array []layout.Abs, phase layout.Abs) {
	cs.buf.WriteByte('[')
	for i, v := range array {
		if i > 0 {
			cs.buf.WriteByte(' ')
		}
		cs.buf.WriteString(formatFloat(float64(v)))
	}
	cs.buf.WriteString("] ")
	cs.buf.WriteString(formatFloat(float64(phase)))
	cs.buf.WriteString(" d\n")
}

// Clipping Operators

// Clip intersects the current clipping path with the current path (W operator).
func (cs *ContentStream) Clip() {
	cs.writeOp("W")
}

// ClipEvenOdd intersects using even-odd rule (W* operator).
func (cs *ContentStream) ClipEvenOdd() {
	cs.writeOp("W*")
}

// Color Helper Functions

// SetFillColor sets the fill color from an RGBA color.
func (cs *ContentStream) SetFillColor(c *Color) {
	if c == nil {
		return
	}
	r := float64(c.R) / 255.0
	g := float64(c.G) / 255.0
	b := float64(c.B) / 255.0
	cs.SetFillColorRGB(r, g, b)
}

// SetStrokeColor sets the stroke color from an RGBA color.
func (cs *ContentStream) SetStrokeColor(c *Color) {
	if c == nil {
		return
	}
	r := float64(c.R) / 255.0
	g := float64(c.G) / 255.0
	b := float64(c.B) / 255.0
	cs.SetStrokeColorRGB(r, g, b)
}

// Color represents an RGBA color.
type Color struct {
	R, G, B, A uint8
}

// ApplyStrokeStyle applies a stroke style to the content stream.
func (cs *ContentStream) ApplyStrokeStyle(stroke *inline.FixedStroke) {
	if stroke == nil {
		return
	}

	cs.SetLineWidth(stroke.Thickness)

	// Set line cap
	switch stroke.LineCap {
	case inline.LineCapButt:
		cs.SetLineCap(0)
	case inline.LineCapRound:
		cs.SetLineCap(1)
	case inline.LineCapSquare:
		cs.SetLineCap(2)
	}

	// Set line join
	switch stroke.LineJoin {
	case inline.LineJoinMiter:
		cs.SetLineJoin(0)
	case inline.LineJoinRound:
		cs.SetLineJoin(1)
	case inline.LineJoinBevel:
		cs.SetLineJoin(2)
	}

	// Set dash pattern if specified
	if len(stroke.DashArray) > 0 {
		cs.SetDashPattern(stroke.DashArray, stroke.DashPhase)
	}

	// Set stroke color from paint
	if c, ok := stroke.Paint.(*Color); ok {
		cs.SetStrokeColor(c)
	}
}
