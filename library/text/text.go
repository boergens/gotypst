// Package text provides the text element and styling for Typst documents.
//
// The text element is the fundamental building block for text content in Typst.
// It controls how text is rendered, including font selection, size, weight,
// style, fill, stroke, and decorations.
package text

import (
	"github.com/boergens/gotypst/layout/inline"
)

// TextElem represents a text element with styling properties.
// This is the Go translation of Typst's text element.
type TextElem struct {
	// Body is the text content to render.
	Body string

	// Font specifies the font family or families to use.
	// Multiple families can be provided for fallback.
	Font []string

	// Size is the font size. Default is 11pt.
	Size Size

	// Weight is the font weight (100-900 or predefined values).
	Weight FontWeight

	// Style is the font style (normal, italic, oblique).
	Style FontStyle

	// Stretch is the font stretch (normal, condensed, expanded).
	Stretch FontStretch

	// Fill is the paint used to fill the text glyphs.
	// Can be a color, gradient, or pattern.
	Fill Paint

	// Stroke is the paint used to stroke the text glyphs.
	// If nil, text is not stroked.
	Stroke *Stroke

	// Tracking adjusts spacing between characters (in em units).
	Tracking Em

	// Spacing adjusts word spacing as a ratio (1.0 = 100%).
	Spacing float64

	// Baseline shifts the text baseline (in em units).
	Baseline Em

	// Underline controls the underline decoration.
	Underline *Underline

	// Strikethrough controls the strikethrough decoration.
	Strikethrough *Strikethrough

	// Overline controls the overline decoration.
	Overline *Overline

	// Hyphenate controls whether to hyphenate text at line breaks.
	Hyphenate *bool

	// Lang is the text language for hyphenation and shaping.
	Lang string

	// Region is the text region (e.g., "US", "GB") for locale-specific behavior.
	Region string

	// Dir is the text direction (ltr or rtl).
	Dir Dir

	// Fallback enables font fallback for missing glyphs.
	Fallback bool

	// Features are OpenType font features to enable.
	Features []string

	// Discretionary controls discretionary ligatures.
	Discretionary bool

	// Historical controls historical ligatures.
	Historical bool

	// NumberType controls number styles (lining, oldstyle).
	NumberType NumberType

	// NumberWidth controls number widths (proportional, tabular).
	NumberWidth NumberWidth

	// Slashed controls slashed zero.
	Slashed bool

	// Fractions controls fraction formatting.
	Fractions bool

	// SmallCaps enables small capitals.
	SmallCaps bool
}

// New creates a new text element with default values.
func New(body string) *TextElem {
	return &TextElem{
		Body:     body,
		Size:     SizeFromPt(11), // Default 11pt
		Weight:   FontWeightNormal,
		Style:    FontStyleNormal,
		Stretch:  FontStretchNormal,
		Spacing:  1.0, // 100%
		Fallback: true,
	}
}

// WithFont sets the font families.
func (t *TextElem) WithFont(families ...string) *TextElem {
	t.Font = families
	return t
}

// WithSize sets the font size.
func (t *TextElem) WithSize(size Size) *TextElem {
	t.Size = size
	return t
}

// WithWeight sets the font weight.
func (t *TextElem) WithWeight(weight FontWeight) *TextElem {
	t.Weight = weight
	return t
}

// WithStyle sets the font style.
func (t *TextElem) WithStyle(style FontStyle) *TextElem {
	t.Style = style
	return t
}

// WithStretch sets the font stretch.
func (t *TextElem) WithStretch(stretch FontStretch) *TextElem {
	t.Stretch = stretch
	return t
}

// WithFill sets the fill paint.
func (t *TextElem) WithFill(fill Paint) *TextElem {
	t.Fill = fill
	return t
}

// WithStroke sets the stroke.
func (t *TextElem) WithStroke(stroke *Stroke) *TextElem {
	t.Stroke = stroke
	return t
}

// WithUnderline sets the underline decoration.
func (t *TextElem) WithUnderline(u *Underline) *TextElem {
	t.Underline = u
	return t
}

// WithStrikethrough sets the strikethrough decoration.
func (t *TextElem) WithStrikethrough(s *Strikethrough) *TextElem {
	t.Strikethrough = s
	return t
}

// WithOverline sets the overline decoration.
func (t *TextElem) WithOverline(o *Overline) *TextElem {
	t.Overline = o
	return t
}

// ToFontVariant converts the text element's font properties to a FontVariant.
func (t *TextElem) ToFontVariant() inline.FontVariant {
	return inline.FontVariant{
		Style:   t.Style.ToInline(),
		Weight:  t.Weight.ToInline(),
		Stretch: t.Stretch.ToInline(),
	}
}

// HasDecoration returns true if the text has any decoration.
func (t *TextElem) HasDecoration() bool {
	return t.Underline != nil || t.Strikethrough != nil || t.Overline != nil
}

// Decorations returns all decorations on this text element.
func (t *TextElem) Decorations() []Decoration {
	var decos []Decoration
	if t.Underline != nil {
		decos = append(decos, t.Underline)
	}
	if t.Strikethrough != nil {
		decos = append(decos, t.Strikethrough)
	}
	if t.Overline != nil {
		decos = append(decos, t.Overline)
	}
	return decos
}

// IsContentElement marks TextElem as a content element.
func (*TextElem) IsContentElement() {}
