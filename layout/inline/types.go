package inline

import (
	"github.com/boergens/gotypst/layout"
)

// Range represents a byte range in the text.
type Range struct {
	Start, End int
}

// Len returns the length of the range.
func (r Range) Len() int {
	return r.End - r.Start
}

// Item represents a prepared item in inline layout.
type Item interface {
	// isItem is a marker method to seal the interface.
	isItem()
	// NaturalWidth returns the natural layouted width of the item.
	NaturalWidth() layout.Abs
	// Textual returns the textual representation of this item.
	Textual() string
}

// TextItem represents a shaped text run.
type TextItem struct {
	shaped *ShapedText
}

func (*TextItem) isItem() {}

// NaturalWidth returns the width of the shaped text.
func (t *TextItem) NaturalWidth() layout.Abs {
	if t.shaped == nil {
		return 0
	}
	return t.shaped.Width()
}

// Textual returns the text content.
func (t *TextItem) Textual() string {
	if t.shaped == nil {
		return ""
	}
	return t.shaped.Text
}

// Text returns the shaped text, if this is a text item.
func (t *TextItem) Text() *ShapedText {
	return t.shaped
}

// AbsoluteItem represents absolute spacing between items.
type AbsoluteItem struct {
	Amount layout.Abs
	Weak   bool
}

func (*AbsoluteItem) isItem() {}

// NaturalWidth returns the spacing amount.
func (a *AbsoluteItem) NaturalWidth() layout.Abs {
	return a.Amount
}

// Textual returns a space character as representation.
func (*AbsoluteItem) Textual() string {
	return " "
}

// FractionalItem represents fractional spacing between items.
type FractionalItem struct {
	Amount layout.Fr
}

func (*FractionalItem) isItem() {}

// NaturalWidth returns zero for fractional items.
func (*FractionalItem) NaturalWidth() layout.Abs {
	return 0
}

// Textual returns a space character as representation.
func (*FractionalItem) Textual() string {
	return " "
}

// FrameItem represents layouted inline-level content.
type FrameItem struct {
	width layout.Abs
}

func (*FrameItem) isItem() {}

// NaturalWidth returns the frame width.
func (f *FrameItem) NaturalWidth() layout.Abs {
	return f.width
}

// Textual returns an object replacement character.
func (*FrameItem) Textual() string {
	return "\uFFFC"
}

// TagItem represents a tag in the content.
type TagItem struct{}

func (*TagItem) isItem() {}

// NaturalWidth returns zero for tags.
func (*TagItem) NaturalWidth() layout.Abs {
	return 0
}

// Textual returns an empty string for tags.
func (*TagItem) Textual() string {
	return ""
}

// SkipItem represents an invisible item that should be skipped.
type SkipItem struct {
	Content string
}

func (*SkipItem) isItem() {}

// NaturalWidth returns zero for skip items.
func (*SkipItem) NaturalWidth() layout.Abs {
	return 0
}

// Textual returns the skip content.
func (s *SkipItem) Textual() string {
	return s.Content
}

// Dash represents a dash at the end of a line.
type Dash int

const (
	// DashSoft is a soft hyphen added to break a word.
	DashSoft Dash = iota + 1
	// DashHard is a regular hyphen in a compound word.
	DashHard
	// DashOther is another kind of dash.
	DashOther
)

// Line represents a layouted line of inline items.
type Line struct {
	// Items contains the items the line is made of.
	Items []Item
	// Width is the exact natural width of the line.
	Width layout.Abs
	// Justify indicates whether the line should be justified.
	Justify bool
	// Dash indicates if the line ends with a hyphen or dash.
	Dash Dash
}

// Empty creates an empty line.
func EmptyLine() Line {
	return Line{
		Items:   nil,
		Width:   0,
		Justify: false,
		Dash:    0,
	}
}

// Justifiables returns the number of glyphs where additional space can be inserted.
func (l *Line) Justifiables() int {
	count := 0
	var lastText *ShapedText
	for _, item := range l.Items {
		if ti, ok := item.(*TextItem); ok && ti.shaped != nil {
			count += ti.shaped.Justifiables()
			lastText = ti.shaped
		}
	}
	// CJK character at line end should not be adjusted.
	if lastText != nil && lastText.CJKJustifiableAtLast() {
		count--
	}
	return count
}

// Stretchability returns how much the line can stretch.
func (l *Line) Stretchability() layout.Abs {
	var total layout.Abs
	for _, item := range l.Items {
		if ti, ok := item.(*TextItem); ok && ti.shaped != nil {
			total += ti.shaped.Stretchability()
		}
	}
	return total
}

// Shrinkability returns how much the line can shrink.
func (l *Line) Shrinkability() layout.Abs {
	var total layout.Abs
	for _, item := range l.Items {
		if ti, ok := item.(*TextItem); ok && ti.shaped != nil {
			total += ti.shaped.Shrinkability()
		}
	}
	return total
}

// HasNegativeWidthItems returns true if the line has items with negative width.
func (l *Line) HasNegativeWidthItems() bool {
	for _, item := range l.Items {
		switch it := item.(type) {
		case *AbsoluteItem:
			if it.Amount < 0 {
				return true
			}
		case *FrameItem:
			if it.width < 0 {
				return true
			}
		}
	}
	return false
}

// Fr returns the sum of fractions in the line.
func (l *Line) Fr() layout.Fr {
	var total layout.Fr
	for _, item := range l.Items {
		if fi, ok := item.(*FractionalItem); ok {
			total += fi.Amount
		}
	}
	return total
}

// TrailingText returns the last text item in the line.
func (l *Line) TrailingText() *ShapedText {
	for i := len(l.Items) - 1; i >= 0; i-- {
		if ti, ok := l.Items[i].(*TextItem); ok && ti.shaped != nil {
			return ti.shaped
		}
	}
	return nil
}

// ShapedText represents shaped text from the shaping engine.
// This is a placeholder that will be filled in by the shaping module.
type ShapedText struct {
	Text   string
	Glyphs []ShapedGlyph
	Lang   layout.Lang
	Size   layout.Abs
}

// Width returns the total width of the shaped text.
func (s *ShapedText) Width() layout.Abs {
	var total layout.Abs
	for _, g := range s.Glyphs {
		total += g.XAdvance.At(g.Size)
	}
	return total
}

// Justifiables returns the number of justifiable glyphs.
func (s *ShapedText) Justifiables() int {
	count := 0
	for _, g := range s.Glyphs {
		if g.IsJustifiable {
			count++
		}
	}
	return count
}

// CJKJustifiableAtLast returns true if the last glyph is CJK justifiable.
func (s *ShapedText) CJKJustifiableAtLast() bool {
	if len(s.Glyphs) == 0 {
		return false
	}
	return s.Glyphs[len(s.Glyphs)-1].IsCJKJustifiable
}

// Stretchability returns how much the text can stretch.
func (s *ShapedText) Stretchability() layout.Abs {
	var total layout.Abs
	for _, g := range s.Glyphs {
		stretch := g.Adjustability.StretchLeft + g.Adjustability.StretchRight
		total += stretch.At(g.Size)
	}
	return total
}

// Shrinkability returns how much the text can shrink.
func (s *ShapedText) Shrinkability() layout.Abs {
	var total layout.Abs
	for _, g := range s.Glyphs {
		shrink := g.Adjustability.ShrinkLeft + g.Adjustability.ShrinkRight
		total += shrink.At(g.Size)
	}
	return total
}

// ShapedGlyph represents a single shaped glyph.
type ShapedGlyph struct {
	GlyphID         uint16
	XAdvance        layout.Em
	XOffset         layout.Em
	Size            layout.Abs
	IsJustifiable   bool
	IsCJKJustifiable bool
	Adjustability   Adjustability
	Range           Range
}

// Adjustability represents how much a glyph can be adjusted.
type Adjustability struct {
	StretchLeft  layout.Em
	StretchRight layout.Em
	ShrinkLeft   layout.Em
	ShrinkRight  layout.Em
}

// Costs represents costs for various layout decisions.
type Costs struct {
	Hyphenation float64
	Runt        float64
}

// DefaultCosts returns the default cost values.
func DefaultCosts() Costs {
	return Costs{
		Hyphenation: 1.0,
		Runt:        1.0,
	}
}

// Config represents shared configuration for inline layout.
type Config struct {
	// Justify indicates whether to justify text.
	Justify bool
	// Linebreaks is the line breaking algorithm to use.
	Linebreaks layout.Linebreaks
	// FirstLineIndent is the indent for the first line.
	FirstLineIndent layout.Abs
	// HangingIndent is the indent for subsequent lines.
	HangingIndent layout.Abs
	// Align is the horizontal alignment.
	Align layout.Alignment
	// FontSize is the text size.
	FontSize layout.Abs
	// Dir is the dominant text direction.
	Dir layout.Dir
	// Hyphenate is the hyphenation setting (nil means auto).
	Hyphenate *bool
	// Lang is the text language (nil means auto per-item).
	Lang *layout.Lang
	// Fallback indicates whether font fallback is enabled.
	Fallback bool
	// CJKLatinSpacing indicates whether to add CJK-Latin spacing.
	CJKLatinSpacing bool
	// Costs for layout decisions.
	Costs Costs
}

// Preparation holds prepared data for line breaking.
type Preparation struct {
	// Text is the full text content.
	Text string
	// Items are the prepared items with their byte ranges.
	Items []PreparedItem
	// Config is the shared configuration.
	Config *Config
}

// PreparedItem associates a byte range with an item.
type PreparedItem struct {
	Range Range
	Item  Item
}

// Get returns the item at the given byte offset.
func (p *Preparation) Get(offset int) (Range, Item) {
	for _, pi := range p.Items {
		if offset >= pi.Range.Start && offset < pi.Range.End {
			return pi.Range, pi.Item
		}
	}
	// Return the last item if offset is at the end.
	if len(p.Items) > 0 && offset == len(p.Text) {
		last := p.Items[len(p.Items)-1]
		return last.Range, last.Item
	}
	return Range{}, nil
}
