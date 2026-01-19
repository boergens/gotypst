package math

// MaxRepeats is the maximum number of times extenders can be repeated
// when assembling glyphs from parts.
const MaxRepeats = 1024

// MathFragment represents a piece of laid out math content.
// This is a sum type interface - use type switches to handle specific types.
type MathFragment interface {
	// Size returns the dimensions of this fragment.
	Size() Size

	// Width returns the width of this fragment.
	Width() Abs

	// Height returns the height of this fragment.
	Height() Abs

	// Ascent returns the distance from baseline to top.
	Ascent() Abs

	// Descent returns the distance from baseline to bottom.
	Descent() Abs

	// Class returns the math class of this fragment.
	Class() Class

	// MathSize returns the math size if applicable.
	MathSize() *MathSize

	// IsTextLike returns whether this fragment is text-like.
	IsTextLike() bool

	// ItalicsCorrection returns the italics correction value.
	ItalicsCorrection() Abs

	// AccentAttach returns the top and bottom accent attachment positions.
	AccentAttach() (Abs, Abs)

	// IntoFrame converts this fragment into a Frame.
	IntoFrame() *Frame

	// isMathFragment is a marker method to seal the interface.
	isMathFragment()
}

// GlyphFragment represents a single glyph or glyph assembly in math.
type GlyphFragment struct {
	// Item is the underlying text item.
	Item *TextItem

	// FragSize is the bounding box size of the glyph.
	FragSize Size

	// Baseline is the distance from top to baseline (nil means use height).
	Baseline *Abs

	// ItalicsCorr is the italics correction.
	ItalicsCorr Abs

	// AccentAttachTop is the top accent attachment position.
	AccentAttachTop Abs

	// AccentAttachBottom is the bottom accent attachment position.
	AccentAttachBottom Abs

	// FragMathSize is the math size context.
	FragMathSize MathSize

	// FragClass is the math class.
	FragClass Class

	// ExtendedShape indicates if this is an extended shape.
	ExtendedShape bool

	// Shift is the baseline shift.
	Shift Abs

	// Align is the alignment offset.
	Align Abs
}

func (g *GlyphFragment) Size() Size      { return g.FragSize }
func (g *GlyphFragment) Width() Abs      { return g.FragSize.X }
func (g *GlyphFragment) Height() Abs     { return g.FragSize.Y }
func (g *GlyphFragment) Class() Class { return g.FragClass }
func (g *GlyphFragment) MathSize() *MathSize    { return &g.FragMathSize }
func (g *GlyphFragment) IsTextLike() bool       { return !g.ExtendedShape }
func (g *GlyphFragment) ItalicsCorrection() Abs { return g.ItalicsCorr }
func (g *GlyphFragment) AccentAttach() (Abs, Abs) {
	return g.AccentAttachTop, g.AccentAttachBottom
}
func (*GlyphFragment) isMathFragment() {}

// Ascent returns the distance from the baseline to the top of the glyph.
func (g *GlyphFragment) Ascent() Abs {
	if g.Baseline != nil {
		return *g.Baseline
	}
	return g.FragSize.Y
}

// Descent returns the distance from the baseline to the bottom of the glyph.
func (g *GlyphFragment) Descent() Abs {
	return g.FragSize.Y - g.Ascent()
}

// IntoFrame converts this glyph fragment into a frame.
func (g *GlyphFragment) IntoFrame() *Frame {
	frame := NewSoftFrame(g.FragSize)
	frame.SetBaseline(g.Ascent())
	pos := Point{Y: g.Ascent() + g.Shift + g.Align}
	frame.PushText(pos, g.Item)
	return frame
}

// CenterOnAxis vertically adjusts the fragment's frame so that it is
// centered on the math axis.
func (g *GlyphFragment) CenterOnAxis() {
	g.AlignOnAxis(VAlignHorizon)
}

// AlignOnAxis vertically adjusts the fragment's frame so that it is
// aligned to the given alignment on the axis.
func (g *GlyphFragment) AlignOnAxis(align VAlignment) {
	h := g.FragSize.Y
	axis := g.Item.Font.Math().AxisHeight.At(g.Item.Size)
	g.Align += g.Ascent()
	newBaseline := align.Inv().Position(h + axis*2.0)
	g.Baseline = &newBaseline
	g.Align -= *g.Baseline
}

// FrameFragment represents a laid-out frame in math context.
type FrameFragment struct {
	// Frame is the underlying frame.
	Frame *Frame

	// FontSize is the font size used.
	FontSize Abs

	// FragClass is the math class.
	FragClass Class

	// FragMathSize is the math size context.
	FragMathSize MathSize

	// BaseAscent is the base element's ascent.
	BaseAscent Abs

	// BaseDescent is the base element's descent.
	BaseDescent Abs

	// ItalicsCorr is the italics correction.
	ItalicsCorr Abs

	// AccentAttachTop is the top accent attachment position.
	AccentAttachTop Abs

	// AccentAttachBottom is the bottom accent attachment position.
	AccentAttachBottom Abs

	// TextLike indicates if this is text-like.
	TextLike bool

	// Ignorant indicates if this fragment should be ignored for some purposes.
	Ignorant bool
}

// NewFrameFragment creates a new FrameFragment with the given properties.
func NewFrameFragment(props *MathProperties, fontSize Abs, frame *Frame) *FrameFragment {
	baseAscent := frame.Ascent()
	baseDescent := frame.Descent()
	accentAttach := frame.Width() / 2.0
	return &FrameFragment{
		Frame:              frame,
		FontSize:           fontSize,
		FragClass:          props.Class,
		FragMathSize:       props.Size,
		BaseAscent:         baseAscent,
		BaseDescent:        baseDescent,
		ItalicsCorr:        0,
		AccentAttachTop:    accentAttach,
		AccentAttachBottom: accentAttach,
		TextLike:           false,
		Ignorant:           props.Ignorant,
	}
}

func (f *FrameFragment) Size() Size      { return f.Frame.Size() }
func (f *FrameFragment) Width() Abs      { return f.Frame.Width() }
func (f *FrameFragment) Height() Abs     { return f.Frame.Height() }
func (f *FrameFragment) Ascent() Abs     { return f.Frame.Ascent() }
func (f *FrameFragment) Descent() Abs    { return f.Frame.Descent() }
func (f *FrameFragment) Class() Class { return f.FragClass }
func (f *FrameFragment) MathSize() *MathSize    { return &f.FragMathSize }
func (f *FrameFragment) IsTextLike() bool       { return f.TextLike }
func (f *FrameFragment) ItalicsCorrection() Abs { return f.ItalicsCorr }
func (f *FrameFragment) AccentAttach() (Abs, Abs) {
	return f.AccentAttachTop, f.AccentAttachBottom
}
func (f *FrameFragment) IntoFrame() *Frame { return f.Frame }
func (*FrameFragment) isMathFragment()     {}

// WithBaseAscent returns a copy with the base ascent set.
func (f *FrameFragment) WithBaseAscent(ascent Abs) *FrameFragment {
	f.BaseAscent = ascent
	return f
}

// WithBaseDescent returns a copy with the base descent set.
func (f *FrameFragment) WithBaseDescent(descent Abs) *FrameFragment {
	f.BaseDescent = descent
	return f
}

// WithItalicsCorrection returns a copy with the italics correction set.
func (f *FrameFragment) WithItalicsCorrection(italics Abs) *FrameFragment {
	f.ItalicsCorr = italics
	return f
}

// WithAccentAttach returns a copy with the accent attachment positions set.
func (f *FrameFragment) WithAccentAttach(top, bottom Abs) *FrameFragment {
	f.AccentAttachTop = top
	f.AccentAttachBottom = bottom
	return f
}

// WithTextLike returns a copy with the text-like flag set.
func (f *FrameFragment) WithTextLike(textLike bool) *FrameFragment {
	f.TextLike = textLike
	return f
}

// SpaceFragment represents horizontal space in math.
type SpaceFragment struct {
	Amount Abs
}

func (s *SpaceFragment) Size() Size               { return Size{X: s.Amount, Y: 0} }
func (s *SpaceFragment) Width() Abs               { return s.Amount }
func (s *SpaceFragment) Height() Abs              { return 0 }
func (s *SpaceFragment) Ascent() Abs              { return 0 }
func (s *SpaceFragment) Descent() Abs             { return 0 }
func (s *SpaceFragment) Class() Class   { return Space }
func (s *SpaceFragment) MathSize() *MathSize      { return nil }
func (s *SpaceFragment) IsTextLike() bool         { return false }
func (s *SpaceFragment) ItalicsCorrection() Abs   { return 0 }
func (s *SpaceFragment) AccentAttach() (Abs, Abs) { return s.Amount / 2.0, s.Amount / 2.0 }
func (s *SpaceFragment) IntoFrame() *Frame        { return NewSoftFrame(s.Size()) }
func (*SpaceFragment) isMathFragment()            {}

// LinebreakFragment represents a line break in math.
type LinebreakFragment struct{}

func (*LinebreakFragment) Size() Size               { return Size{} }
func (*LinebreakFragment) Width() Abs               { return 0 }
func (*LinebreakFragment) Height() Abs              { return 0 }
func (*LinebreakFragment) Ascent() Abs              { return 0 }
func (*LinebreakFragment) Descent() Abs             { return 0 }
func (*LinebreakFragment) Class() Class   { return Space }
func (*LinebreakFragment) MathSize() *MathSize      { return nil }
func (*LinebreakFragment) IsTextLike() bool         { return false }
func (*LinebreakFragment) ItalicsCorrection() Abs   { return 0 }
func (*LinebreakFragment) AccentAttach() (Abs, Abs) { return 0, 0 }
func (*LinebreakFragment) IntoFrame() *Frame        { return NewSoftFrame(Size{}) }
func (*LinebreakFragment) isMathFragment()          {}

// AlignFragment represents an alignment point in math.
type AlignFragment struct{}

func (*AlignFragment) Size() Size               { return Size{} }
func (*AlignFragment) Width() Abs               { return 0 }
func (*AlignFragment) Height() Abs              { return 0 }
func (*AlignFragment) Ascent() Abs              { return 0 }
func (*AlignFragment) Descent() Abs             { return 0 }
func (*AlignFragment) Class() Class   { return Special }
func (*AlignFragment) MathSize() *MathSize      { return nil }
func (*AlignFragment) IsTextLike() bool         { return false }
func (*AlignFragment) ItalicsCorrection() Abs   { return 0 }
func (*AlignFragment) AccentAttach() (Abs, Abs) { return 0, 0 }
func (*AlignFragment) IntoFrame() *Frame        { return NewSoftFrame(Size{}) }
func (*AlignFragment) isMathFragment()          {}

// TagFragment represents a tag in math.
type TagFragment struct {
	Tag Tag
}

func (*TagFragment) Size() Size               { return Size{} }
func (*TagFragment) Width() Abs               { return 0 }
func (*TagFragment) Height() Abs              { return 0 }
func (*TagFragment) Ascent() Abs              { return 0 }
func (*TagFragment) Descent() Abs             { return 0 }
func (*TagFragment) Class() Class   { return Special }
func (*TagFragment) MathSize() *MathSize      { return nil }
func (*TagFragment) IsTextLike() bool         { return false }
func (*TagFragment) ItalicsCorrection() Abs   { return 0 }
func (*TagFragment) AccentAttach() (Abs, Abs) { return 0, 0 }
func (t *TagFragment) IntoFrame() *Frame {
	frame := NewSoftFrame(Size{})
	frame.PushTag(Point{}, t.Tag)
	return frame
}
func (*TagFragment) isMathFragment() {}

// BaseAscent returns the base ascent for a fragment.
func BaseAscent(f MathFragment) Abs {
	if ff, ok := f.(*FrameFragment); ok {
		return ff.BaseAscent
	}
	return f.Ascent()
}

// BaseDescent returns the base descent for a fragment.
func BaseDescent(f MathFragment) Abs {
	if ff, ok := f.(*FrameFragment); ok {
		return ff.BaseDescent
	}
	return f.Descent()
}

// IsIgnorant returns whether a fragment should be ignored.
func IsIgnorant(f MathFragment) bool {
	switch frag := f.(type) {
	case *FrameFragment:
		return frag.Ignorant
	case *TagFragment:
		return true
	default:
		return false
	}
}

// FragmentFill returns the fill color of a fragment if it has one.
func FragmentFill(f MathFragment) Paint {
	if gf, ok := f.(*GlyphFragment); ok {
		return gf.Item.Fill
	}
	return nil
}

// FragmentStroke returns the stroke of a fragment if it has one.
func FragmentStroke(f MathFragment) *FixedStroke {
	if gf, ok := f.(*GlyphFragment); ok {
		return gf.Item.Stroke
	}
	return nil
}

// KernAtHeight calculates the kerning value at a specific corner and height.
func KernAtHeight(f MathFragment, corner Corner, height Abs) Abs {
	gf, ok := f.(*GlyphFragment)
	if !ok {
		return 0
	}

	// For glyph assemblies we pick either the start or end glyph
	// depending on the corner.
	glyphs := gf.Item.Glyphs
	if len(glyphs) == 0 {
		return 0
	}

	isVertical := true
	for _, g := range glyphs {
		if g.YAdvance == 0 {
			isVertical = false
			break
		}
	}

	var glyphIndex int
	switch {
	case isVertical && (corner == CornerTopLeft || corner == CornerTopRight):
		glyphIndex = len(glyphs) - 1
	case !isVertical && (corner == CornerTopRight || corner == CornerBottomRight):
		glyphIndex = len(glyphs) - 1
	default:
		glyphIndex = 0
	}

	heightEm := Em(height / gf.Item.Size)
	kern := kernAtHeightFromFont(gf.Item.Font, GlyphID(glyphs[glyphIndex].ID), corner, heightEm)
	return kern.At(gf.Item.Size)
}
