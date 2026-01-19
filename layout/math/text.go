package math

import (
	"strings"
	"unicode"
)

// TextItemData represents text content for math layout.
type TextItemData struct {
	Text string
}

// layoutTextImpl lays out text in math mode.
func layoutTextImpl(
	item *MathItem,
	ctx *MathContext,
	styles *StyleChain,
	props *MathProperties,
) error {
	text := item.Text
	span := props.Span

	// Check for newlines
	if strings.ContainsFunc(text, isNewline) {
		return layoutTextLines(strings.FieldsFunc(text, isNewline), span, ctx, styles, props)
	}

	fragment, err := layoutInlineText(text, span, ctx, styles, props)
	if err != nil {
		return err
	}
	ctx.Push(fragment)
	return nil
}

// isNewline checks if a rune is a newline character.
func isNewline(r rune) bool {
	return r == '\n' || r == '\r'
}

// layoutTextLines lays out multiple lines of text.
func layoutTextLines(
	lines []string,
	span Span,
	ctx *MathContext,
	styles *StyleChain,
	props *MathProperties,
) error {
	var fragments []MathFragment

	for i, line := range lines {
		if i != 0 {
			fragments = append(fragments, &LinebreakFragment{})
		}
		if line != "" {
			frag, err := layoutInlineText(line, span, ctx, styles, props)
			if err != nil {
				return err
			}
			fragments = append(fragments, frag)
		}
	}

	frame := FragmentsIntoFrame(fragments, styles)
	axis := ctx.Font().Math().AxisHeight.Resolve(styles.ResolveTextSize())
	frame.SetBaseline(frame.Height()/2.0 + axis)

	ctx.Push(NewFrameFragment(props, styles.ResolveTextSize(), frame))
	return nil
}

// layoutInlineText lays out inline text with math font styling.
func layoutInlineText(
	text string,
	span Span,
	ctx *MathContext,
	styles *StyleChain,
	props *MathProperties,
) (*FrameFragment, error) {
	// Check if text is only digits and decimal points
	if isNumeric(text) {
		// Fast path for numbers
		var fragments []MathFragment
		for _, c := range text {
			glyph := newGlyphFragmentChar(ctx, styles, c, span)
			if glyph != nil {
				fragments = append(fragments, glyph)
			}
		}
		frame := FragmentsIntoFrame(fragments, styles)
		ff := NewFrameFragment(props, styles.ResolveTextSize(), frame)
		ff.TextLike = true
		return ff, nil
	}

	// Regular text layout
	// TODO: Implement full inline text layout with paragraph layout
	frame := NewSoftFrame(Size{})
	ff := NewFrameFragment(props, styles.ResolveTextSize(), frame)
	ff.TextLike = true
	return ff, nil
}

// isNumeric checks if text contains only digits and decimal points.
func isNumeric(text string) bool {
	for _, c := range text {
		if !unicode.IsDigit(c) && c != '.' {
			return false
		}
	}
	return true
}

// GlyphItemData represents a glyph item for layout.
type GlyphItemData struct {
	Text    string
	Flac    bool // Flattened accent feature
	Stretch *StretchInfo
}

// StretchInfo holds information about glyph stretching.
type StretchInfo struct {
	Target     Rel
	RelativeTo *Abs
	ShortFall  Em
	FontSize   *Abs
}

// layoutGlyphImpl lays out a single glyph in math font.
func layoutGlyphImpl(
	item *MathItem,
	ctx *MathContext,
	styles *StyleChain,
	props *MathProperties,
) error {
	// TODO: Handle flac feature
	// TODO: Handle dtls feature for dotless variants

	text := item.Text
	if text == "" {
		return nil
	}

	glyph := newGlyphFragment(ctx, styles, text, props.Span)
	if glyph == nil {
		return nil
	}

	glyph.FragClass = props.Class

	// TODO: Handle stretching

	// Center large operators on axis
	if glyph.FragClass == Large {
		glyph.CenterOnAxis()
	}

	ctx.Push(glyph)
	return nil
}

// newGlyphFragmentChar creates a GlyphFragment from a single character.
func newGlyphFragmentChar(ctx *MathContext, styles *StyleChain, c rune, span Span) *GlyphFragment {
	return newGlyphFragment(ctx, styles, string(c), span)
}

// newGlyphFragment creates a GlyphFragment from text.
func newGlyphFragment(ctx *MathContext, styles *StyleChain, text string, span Span) *GlyphFragment {
	// Shape the text
	font, glyphs := shape(ctx, styles, text)
	if font == nil || len(glyphs) == 0 {
		return nil
	}

	// Create text item
	fontSize := styles.ResolveTextSize()
	textItem := &TextItem{
		Text:   text,
		Font:   font,
		Size:   fontSize,
		Glyphs: glyphs,
	}

	// Get first character for class lookup
	c := []rune(text)[0]
	class := defaultMathClass(c)

	// Calculate metrics
	gf := &GlyphFragment{
		Item:         textItem,
		FragMathSize: props_size_from_styles(styles),
		FragClass:    class,
	}
	gf.updateGlyph()

	return gf
}

// updateGlyph updates glyph metrics after shaping.
func (g *GlyphFragment) updateGlyph() {
	if len(g.Item.Glyphs) == 0 {
		return
	}

	// TODO: Get actual glyph metrics from font
	// For now use placeholder values
	width := g.Item.Width()
	height := g.Item.Size // Approximate height as font size

	g.FragSize = Size{X: width, Y: height}
	baseline := height * 0.8 // Approximate baseline
	g.Baseline = &baseline
	g.AccentAttachTop = width / 2.0
	g.AccentAttachBottom = width / 2.0
}

// shape performs text shaping for math.
func shape(ctx *MathContext, styles *StyleChain, text string) (*Font, []Glyph) {
	// TODO: Implement full shaping with harfbuzz
	// For now return placeholder
	font := ctx.Font()
	if font == nil {
		return nil, nil
	}

	// Create basic glyph data
	var glyphs []Glyph
	for i, c := range text {
		glyphs = append(glyphs, Glyph{
			ID:       uint16(c), // Placeholder - should be actual glyph ID
			XAdvance: 0.5,       // Placeholder
			Range:    Range{Start: i, End: i + len(string(c))},
		})
	}

	return font, glyphs
}

// defaultMathClass returns the default math class for a character.
func defaultMathClass(c rune) Class {
	// TODO: Use proper unicode math class lookup
	switch {
	case c >= '0' && c <= '9':
		return Normal
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z':
		return Normal
	case c == '+', c == '-':
		return Binary
	case c == '=', c == '<', c == '>':
		return Relation
	case c == '(', c == '[', c == '{':
		return Opening
	case c == ')', c == ']', c == '}':
		return Closing
	case c == ',', c == ';':
		return Punctuation
	default:
		return Normal
	}
}

// props_size_from_styles extracts MathSize from styles.
func props_size_from_styles(styles *StyleChain) MathSize {
	// TODO: Get actual math size from styles
	return MathSizeText
}

