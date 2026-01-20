package text

import (
	"github.com/boergens/gotypst/layout/inline"
)

// Decoration is an interface for text decorations.
type Decoration interface {
	isDecoration()
	// ToDecoLine converts to the inline package DecoLine type.
	ToDecoLine(textFill Paint, fontSize inline.Abs) inline.DecoLine
}

// Underline represents an underline decoration.
type Underline struct {
	// Stroke is the line stroke. If nil, derived from text properties.
	Stroke *Stroke
	// Offset overrides the default underline position.
	// Positive values move the line down (away from text).
	Offset *inline.Abs
	// Evade causes the line to skip over glyph descenders.
	Evade bool
	// Background renders the line behind the text.
	Background bool
}

func (*Underline) isDecoration() {}

// NewUnderline creates a new underline with defaults.
func NewUnderline() *Underline {
	return &Underline{
		Evade: true, // Default: avoid descenders
	}
}

// WithStroke sets the underline stroke.
func (u *Underline) WithStroke(s *Stroke) *Underline {
	u.Stroke = s
	return u
}

// WithOffset sets the underline offset.
func (u *Underline) WithOffset(offset inline.Abs) *Underline {
	u.Offset = &offset
	return u
}

// WithEvade sets whether the underline evades descenders.
func (u *Underline) WithEvade(evade bool) *Underline {
	u.Evade = evade
	return u
}

// WithBackground sets whether the underline renders behind text.
func (u *Underline) WithBackground(bg bool) *Underline {
	u.Background = bg
	return u
}

// ToDecoLine converts to the inline package DecoLine type.
func (u *Underline) ToDecoLine(textFill Paint, fontSize inline.Abs) inline.DecoLine {
	var stroke *inline.FixedStroke
	if u.Stroke != nil {
		fs := u.Stroke.ToFixedStroke()
		stroke = &fs
	}
	return &inline.UnderlineDeco{
		Stroke:     stroke,
		Offset:     u.Offset,
		Evade:      u.Evade,
		Background: u.Background,
	}
}

// Strikethrough represents a strikethrough decoration.
type Strikethrough struct {
	// Stroke is the line stroke. If nil, derived from text properties.
	Stroke *Stroke
	// Offset overrides the default strikethrough position.
	Offset *inline.Abs
	// Background renders the line behind the text.
	Background bool
}

func (*Strikethrough) isDecoration() {}

// NewStrikethrough creates a new strikethrough with defaults.
func NewStrikethrough() *Strikethrough {
	return &Strikethrough{}
}

// WithStroke sets the strikethrough stroke.
func (s *Strikethrough) WithStroke(stroke *Stroke) *Strikethrough {
	s.Stroke = stroke
	return s
}

// WithOffset sets the strikethrough offset.
func (s *Strikethrough) WithOffset(offset inline.Abs) *Strikethrough {
	s.Offset = &offset
	return s
}

// WithBackground sets whether the strikethrough renders behind text.
func (s *Strikethrough) WithBackground(bg bool) *Strikethrough {
	s.Background = bg
	return s
}

// ToDecoLine converts to the inline package DecoLine type.
func (s *Strikethrough) ToDecoLine(textFill Paint, fontSize inline.Abs) inline.DecoLine {
	var stroke *inline.FixedStroke
	if s.Stroke != nil {
		fs := s.Stroke.ToFixedStroke()
		stroke = &fs
	}
	return &inline.StrikethroughDeco{
		Stroke:     stroke,
		Offset:     s.Offset,
		Background: s.Background,
	}
}

// Overline represents an overline decoration.
type Overline struct {
	// Stroke is the line stroke. If nil, derived from text properties.
	Stroke *Stroke
	// Offset overrides the default overline position.
	// Positive values move the line up (away from text).
	Offset *inline.Abs
	// Evade causes the line to skip over glyph ascenders.
	Evade bool
	// Background renders the line behind the text.
	Background bool
}

func (*Overline) isDecoration() {}

// NewOverline creates a new overline with defaults.
func NewOverline() *Overline {
	return &Overline{
		Evade: true, // Default: avoid ascenders
	}
}

// WithStroke sets the overline stroke.
func (o *Overline) WithStroke(s *Stroke) *Overline {
	o.Stroke = s
	return o
}

// WithOffset sets the overline offset.
func (o *Overline) WithOffset(offset inline.Abs) *Overline {
	o.Offset = &offset
	return o
}

// WithEvade sets whether the overline evades ascenders.
func (o *Overline) WithEvade(evade bool) *Overline {
	o.Evade = evade
	return o
}

// WithBackground sets whether the overline renders behind text.
func (o *Overline) WithBackground(bg bool) *Overline {
	o.Background = bg
	return o
}

// ToDecoLine converts to the inline package DecoLine type.
func (o *Overline) ToDecoLine(textFill Paint, fontSize inline.Abs) inline.DecoLine {
	var stroke *inline.FixedStroke
	if o.Stroke != nil {
		fs := o.Stroke.ToFixedStroke()
		stroke = &fs
	}
	return &inline.OverlineDeco{
		Stroke:     stroke,
		Offset:     o.Offset,
		Evade:      o.Evade,
		Background: o.Background,
	}
}

// DecorationExtent returns the extra extent to add to decorations.
// This allows decorations to extend slightly beyond text bounds.
func DecorationExtent(fontSize inline.Abs) inline.Abs {
	return fontSize * 0.02 // 2% of font size
}
