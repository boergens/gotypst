// Package html provides HTML export functionality for Typst documents.
package html

import (
	"fmt"
	"io"
	"strings"

	"github.com/boergens/gotypst/layout"
	"github.com/boergens/gotypst/layout/inline"
)

// Color represents an RGBA color for CSS.
type Color struct {
	R, G, B, A uint8
}

// CSS returns the CSS color string.
func (c *Color) CSS() string {
	if c == nil {
		return ""
	}
	if c.A == 255 {
		return fmt.Sprintf("rgb(%d, %d, %d)", c.R, c.G, c.B)
	}
	return fmt.Sprintf("rgba(%d, %d, %d, %.3f)", c.R, c.G, c.B, float64(c.A)/255.0)
}

// Hex returns the hex color string (without alpha).
func (c *Color) Hex() string {
	if c == nil {
		return ""
	}
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

// StyleBuilder builds CSS style declarations.
type StyleBuilder struct {
	buf strings.Builder
}

// NewStyleBuilder creates a new style builder.
func NewStyleBuilder() *StyleBuilder {
	return &StyleBuilder{}
}

// String returns the built CSS style string.
func (s *StyleBuilder) String() string {
	return s.buf.String()
}

// WriteTo writes the style to a writer.
func (s *StyleBuilder) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write([]byte(s.buf.String()))
	return int64(n), err
}

// Reset clears the builder.
func (s *StyleBuilder) Reset() {
	s.buf.Reset()
}

// write writes a property-value pair.
func (s *StyleBuilder) write(property, value string) {
	if value == "" {
		return
	}
	if s.buf.Len() > 0 {
		s.buf.WriteString("; ")
	}
	s.buf.WriteString(property)
	s.buf.WriteString(": ")
	s.buf.WriteString(value)
}

// Position sets the position property.
func (s *StyleBuilder) Position(pos string) *StyleBuilder {
	s.write("position", pos)
	return s
}

// Absolute sets position to absolute.
func (s *StyleBuilder) Absolute() *StyleBuilder {
	return s.Position("absolute")
}

// Relative sets position to relative.
func (s *StyleBuilder) Relative() *StyleBuilder {
	return s.Position("relative")
}

// Left sets the left position.
func (s *StyleBuilder) Left(v layout.Abs) *StyleBuilder {
	s.write("left", formatPt(v))
	return s
}

// Top sets the top position.
func (s *StyleBuilder) Top(v layout.Abs) *StyleBuilder {
	s.write("top", formatPt(v))
	return s
}

// Right sets the right position.
func (s *StyleBuilder) Right(v layout.Abs) *StyleBuilder {
	s.write("right", formatPt(v))
	return s
}

// Bottom sets the bottom position.
func (s *StyleBuilder) Bottom(v layout.Abs) *StyleBuilder {
	s.write("bottom", formatPt(v))
	return s
}

// Width sets the width.
func (s *StyleBuilder) Width(v layout.Abs) *StyleBuilder {
	s.write("width", formatPt(v))
	return s
}

// Height sets the height.
func (s *StyleBuilder) Height(v layout.Abs) *StyleBuilder {
	s.write("height", formatPt(v))
	return s
}

// Size sets both width and height.
func (s *StyleBuilder) Size(size layout.Size) *StyleBuilder {
	s.Width(size.Width)
	s.Height(size.Height)
	return s
}

// Margin sets the margin.
func (s *StyleBuilder) Margin(v layout.Abs) *StyleBuilder {
	s.write("margin", formatPt(v))
	return s
}

// MarginSides sets individual margins.
func (s *StyleBuilder) MarginSides(top, right, bottom, left layout.Abs) *StyleBuilder {
	s.write("margin", fmt.Sprintf("%s %s %s %s",
		formatPt(top), formatPt(right), formatPt(bottom), formatPt(left)))
	return s
}

// Padding sets the padding.
func (s *StyleBuilder) Padding(v layout.Abs) *StyleBuilder {
	s.write("padding", formatPt(v))
	return s
}

// PaddingSides sets individual paddings.
func (s *StyleBuilder) PaddingSides(top, right, bottom, left layout.Abs) *StyleBuilder {
	s.write("padding", fmt.Sprintf("%s %s %s %s",
		formatPt(top), formatPt(right), formatPt(bottom), formatPt(left)))
	return s
}

// Color sets the text color.
func (s *StyleBuilder) Color(c *Color) *StyleBuilder {
	if c != nil {
		s.write("color", c.CSS())
	}
	return s
}

// Background sets the background color.
func (s *StyleBuilder) Background(c *Color) *StyleBuilder {
	if c != nil {
		s.write("background", c.CSS())
	}
	return s
}

// BackgroundColor sets the background-color property.
func (s *StyleBuilder) BackgroundColor(c *Color) *StyleBuilder {
	if c != nil {
		s.write("background-color", c.CSS())
	}
	return s
}

// FontSize sets the font size.
func (s *StyleBuilder) FontSize(v layout.Abs) *StyleBuilder {
	s.write("font-size", formatPt(v))
	return s
}

// FontFamily sets the font family.
func (s *StyleBuilder) FontFamily(family string) *StyleBuilder {
	if family != "" {
		s.write("font-family", family)
	}
	return s
}

// FontWeight sets the font weight.
func (s *StyleBuilder) FontWeight(weight inline.FontWeight) *StyleBuilder {
	s.write("font-weight", fmt.Sprintf("%d", weight))
	return s
}

// FontStyle sets the font style.
func (s *StyleBuilder) FontStyle(style inline.FontStyle) *StyleBuilder {
	switch style {
	case inline.FontStyleItalic:
		s.write("font-style", "italic")
	case inline.FontStyleOblique:
		s.write("font-style", "oblique")
	default:
		s.write("font-style", "normal")
	}
	return s
}

// FontStretch sets the font stretch.
func (s *StyleBuilder) FontStretch(stretch inline.FontStretch) *StyleBuilder {
	switch stretch {
	case inline.FontStretchCondensed:
		s.write("font-stretch", "condensed")
	case inline.FontStretchExpanded:
		s.write("font-stretch", "expanded")
	default:
		s.write("font-stretch", "normal")
	}
	return s
}

// FontVariant sets font style, weight, and stretch from a FontVariant.
func (s *StyleBuilder) FontVariant(v inline.FontVariant) *StyleBuilder {
	s.FontStyle(v.Style)
	s.FontWeight(v.Weight)
	s.FontStretch(v.Stretch)
	return s
}

// LineHeight sets the line height.
func (s *StyleBuilder) LineHeight(v float64) *StyleBuilder {
	s.write("line-height", fmt.Sprintf("%.3f", v))
	return s
}

// TextAlign sets text alignment.
func (s *StyleBuilder) TextAlign(align layout.Alignment) *StyleBuilder {
	switch align {
	case layout.AlignStart:
		s.write("text-align", "start")
	case layout.AlignCenter:
		s.write("text-align", "center")
	case layout.AlignEnd:
		s.write("text-align", "end")
	}
	return s
}

// TextDecoration sets text decoration.
func (s *StyleBuilder) TextDecoration(decoration string) *StyleBuilder {
	s.write("text-decoration", decoration)
	return s
}

// TextDecorationLine sets the text-decoration-line property.
func (s *StyleBuilder) TextDecorationLine(line string) *StyleBuilder {
	s.write("text-decoration-line", line)
	return s
}

// TextDecorationColor sets the text-decoration-color property.
func (s *StyleBuilder) TextDecorationColor(c *Color) *StyleBuilder {
	if c != nil {
		s.write("text-decoration-color", c.CSS())
	}
	return s
}

// TextDecorationStyle sets the text-decoration-style property.
func (s *StyleBuilder) TextDecorationStyle(style string) *StyleBuilder {
	s.write("text-decoration-style", style)
	return s
}

// TextDecorationThickness sets the text-decoration-thickness property.
func (s *StyleBuilder) TextDecorationThickness(v layout.Abs) *StyleBuilder {
	s.write("text-decoration-thickness", formatPt(v))
	return s
}

// TextUnderlineOffset sets the text-underline-offset property.
func (s *StyleBuilder) TextUnderlineOffset(v layout.Abs) *StyleBuilder {
	s.write("text-underline-offset", formatPt(v))
	return s
}

// WhiteSpace sets the white-space property.
func (s *StyleBuilder) WhiteSpace(value string) *StyleBuilder {
	s.write("white-space", value)
	return s
}

// Direction sets the text direction.
func (s *StyleBuilder) Direction(dir layout.Dir) *StyleBuilder {
	if dir == layout.DirRTL {
		s.write("direction", "rtl")
	} else {
		s.write("direction", "ltr")
	}
	return s
}

// Border sets a simple border.
func (s *StyleBuilder) Border(width layout.Abs, style string, c *Color) *StyleBuilder {
	if c != nil {
		s.write("border", fmt.Sprintf("%s %s %s", formatPt(width), style, c.CSS()))
	}
	return s
}

// BorderWidth sets the border width.
func (s *StyleBuilder) BorderWidth(v layout.Abs) *StyleBuilder {
	s.write("border-width", formatPt(v))
	return s
}

// BorderStyle sets the border style.
func (s *StyleBuilder) BorderStyle(style string) *StyleBuilder {
	s.write("border-style", style)
	return s
}

// BorderColor sets the border color.
func (s *StyleBuilder) BorderColor(c *Color) *StyleBuilder {
	if c != nil {
		s.write("border-color", c.CSS())
	}
	return s
}

// BorderRadius sets the border radius.
func (s *StyleBuilder) BorderRadius(v layout.Abs) *StyleBuilder {
	s.write("border-radius", formatPt(v))
	return s
}

// BorderRadiusCorners sets individual corner radii.
func (s *StyleBuilder) BorderRadiusCorners(topLeft, topRight, bottomRight, bottomLeft layout.Abs) *StyleBuilder {
	s.write("border-radius", fmt.Sprintf("%s %s %s %s",
		formatPt(topLeft), formatPt(topRight), formatPt(bottomRight), formatPt(bottomLeft)))
	return s
}

// BoxShadow sets a box shadow.
func (s *StyleBuilder) BoxShadow(x, y, blur, spread layout.Abs, c *Color) *StyleBuilder {
	if c != nil {
		s.write("box-shadow", fmt.Sprintf("%s %s %s %s %s",
			formatPt(x), formatPt(y), formatPt(blur), formatPt(spread), c.CSS()))
	}
	return s
}

// Opacity sets the opacity.
func (s *StyleBuilder) Opacity(v float64) *StyleBuilder {
	s.write("opacity", fmt.Sprintf("%.3f", v))
	return s
}

// Display sets the display property.
func (s *StyleBuilder) Display(value string) *StyleBuilder {
	s.write("display", value)
	return s
}

// Overflow sets the overflow property.
func (s *StyleBuilder) Overflow(value string) *StyleBuilder {
	s.write("overflow", value)
	return s
}

// ZIndex sets the z-index.
func (s *StyleBuilder) ZIndex(v int) *StyleBuilder {
	s.write("z-index", fmt.Sprintf("%d", v))
	return s
}

// Transform sets a CSS transform.
func (s *StyleBuilder) Transform(value string) *StyleBuilder {
	s.write("transform", value)
	return s
}

// TransformOrigin sets the transform origin.
func (s *StyleBuilder) TransformOrigin(value string) *StyleBuilder {
	s.write("transform-origin", value)
	return s
}

// Raw writes a raw property-value pair.
func (s *StyleBuilder) Raw(property, value string) *StyleBuilder {
	s.write(property, value)
	return s
}

// StrokeStyle converts an inline.FixedStroke to CSS border properties.
func (s *StyleBuilder) StrokeStyle(stroke *inline.FixedStroke) *StyleBuilder {
	if stroke == nil {
		return s
	}

	s.BorderWidth(stroke.Thickness)

	// Convert dash array to border style
	if len(stroke.DashArray) > 0 {
		s.BorderStyle("dashed")
	} else {
		s.BorderStyle("solid")
	}

	// Convert paint to color
	if c, ok := stroke.Paint.(*Color); ok {
		s.BorderColor(c)
	}

	return s
}

// LineCap converts inline.LineCap to CSS stroke-linecap value.
func LineCap(cap inline.LineCap) string {
	switch cap {
	case inline.LineCapRound:
		return "round"
	case inline.LineCapSquare:
		return "square"
	default:
		return "butt"
	}
}

// LineJoin converts inline.LineJoin to CSS stroke-linejoin value.
func LineJoin(join inline.LineJoin) string {
	switch join {
	case inline.LineJoinRound:
		return "round"
	case inline.LineJoinBevel:
		return "bevel"
	default:
		return "miter"
	}
}

// formatPt formats an Abs value as CSS points.
func formatPt(v layout.Abs) string {
	// Use minimal precision
	f := float64(v)
	if f == float64(int(f)) {
		return fmt.Sprintf("%dpt", int(f))
	}
	return fmt.Sprintf("%.2fpt", f)
}

// formatPx formats an Abs value as CSS pixels (assuming 72dpi).
func formatPx(v layout.Abs) string {
	// 1pt = 1.333px at 96dpi (CSS standard)
	px := float64(v) * 96.0 / 72.0
	if px == float64(int(px)) {
		return fmt.Sprintf("%dpx", int(px))
	}
	return fmt.Sprintf("%.2fpx", px)
}

// CSSClass builds a CSS class definition.
type CSSClass struct {
	Name       string
	Styles     *StyleBuilder
	PseudoElem string // e.g., "::before", "::after"
}

// String returns the CSS class definition.
func (c *CSSClass) String() string {
	var buf strings.Builder
	buf.WriteString(".")
	buf.WriteString(c.Name)
	if c.PseudoElem != "" {
		buf.WriteString(c.PseudoElem)
	}
	buf.WriteString(" { ")
	buf.WriteString(c.Styles.String())
	buf.WriteString(" }")
	return buf.String()
}

// Stylesheet manages a collection of CSS styles.
type Stylesheet struct {
	classes []CSSClass
	buf     strings.Builder
}

// NewStylesheet creates a new stylesheet.
func NewStylesheet() *Stylesheet {
	return &Stylesheet{}
}

// AddClass adds a CSS class to the stylesheet.
func (ss *Stylesheet) AddClass(name string, styles *StyleBuilder) {
	ss.classes = append(ss.classes, CSSClass{Name: name, Styles: styles})
}

// AddClassWithPseudo adds a CSS class with a pseudo-element.
func (ss *Stylesheet) AddClassWithPseudo(name string, pseudoElem string, styles *StyleBuilder) {
	ss.classes = append(ss.classes, CSSClass{Name: name, Styles: styles, PseudoElem: pseudoElem})
}

// AddRaw adds raw CSS to the stylesheet.
func (ss *Stylesheet) AddRaw(css string) {
	ss.buf.WriteString(css)
	ss.buf.WriteString("\n")
}

// String returns the complete stylesheet.
func (ss *Stylesheet) String() string {
	var buf strings.Builder

	// Write raw CSS first
	if ss.buf.Len() > 0 {
		buf.WriteString(ss.buf.String())
	}

	// Write class definitions
	for _, class := range ss.classes {
		buf.WriteString(class.String())
		buf.WriteString("\n")
	}

	return buf.String()
}

// WriteTo writes the stylesheet to a writer.
func (ss *Stylesheet) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write([]byte(ss.String()))
	return int64(n), err
}

// BaseStyles returns common base styles for HTML output.
func BaseStyles() *Stylesheet {
	ss := NewStylesheet()

	// Reset styles
	ss.AddRaw(`* { margin: 0; padding: 0; box-sizing: border-box; }`)

	// Page container
	pageStyle := NewStyleBuilder()
	pageStyle.Relative().Overflow("hidden")
	ss.AddClass("page", pageStyle)

	// Absolute positioned content
	contentStyle := NewStyleBuilder()
	contentStyle.Absolute()
	ss.AddClass("content", contentStyle)

	// Text span
	textStyle := NewStyleBuilder()
	textStyle.WhiteSpace("pre")
	ss.AddClass("text", textStyle)

	return ss
}

// ColorFromPagesColor converts a pages.Color to an html.Color.
func ColorFromPagesColor(c interface{}) *Color {
	// Type switch to handle different color representations
	switch v := c.(type) {
	case *Color:
		return v
	case Color:
		return &v
	default:
		// Try to extract from interface using reflection-like approach
		// This handles pages.Color and similar types
		return nil
	}
}
