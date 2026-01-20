// Package inline provides inline/paragraph layout including text shaping.
package inline

// MathClass represents the classification of a math atom.
// This follows TeX's math atom classification for spacing.
type MathClass int

const (
	// MathClassNone indicates no specific class (default).
	MathClassNone MathClass = iota
	// MathClassNormal is for ordinary symbols (letters, numbers).
	MathClassNormal
	// MathClassLarge is for large operators (∑, ∏, ∫).
	MathClassLarge
	// MathClassBinary is for binary operators (+, -, ×).
	MathClassBinary
	// MathClassRelation is for relation symbols (=, <, >).
	MathClassRelation
	// MathClassOpening is for opening delimiters ((, [, {).
	MathClassOpening
	// MathClassClosing is for closing delimiters (), ], }).
	MathClassClosing
	// MathClassPunctuation is for punctuation (,, ;).
	MathClassPunctuation
	// MathClassFence is for fences that can grow (|, ‖).
	MathClassFence
	// MathClassGlyphVariant is for glyph size variants.
	MathClassGlyphVariant
	// MathClassSpace is for explicit spacing.
	MathClassSpace
)

// String returns the name of the math class.
func (c MathClass) String() string {
	switch c {
	case MathClassNone:
		return "none"
	case MathClassNormal:
		return "normal"
	case MathClassLarge:
		return "large"
	case MathClassBinary:
		return "binary"
	case MathClassRelation:
		return "relation"
	case MathClassOpening:
		return "opening"
	case MathClassClosing:
		return "closing"
	case MathClassPunctuation:
		return "punctuation"
	case MathClassFence:
		return "fence"
	case MathClassGlyphVariant:
		return "glyph-variant"
	case MathClassSpace:
		return "space"
	default:
		return "unknown"
	}
}

// IsOperator returns true if this class represents an operator.
func (c MathClass) IsOperator() bool {
	return c == MathClassLarge || c == MathClassBinary
}

// IsDelimiter returns true if this class represents a delimiter.
func (c MathClass) IsDelimiter() bool {
	return c == MathClassOpening || c == MathClassClosing || c == MathClassFence
}

// MathFragment represents a fragment in math layout.
// Fragments are the intermediate representation between content elements
// and final frames during math typesetting.
type MathFragment interface {
	// isMathFragment is a marker method to seal the interface.
	isMathFragment()

	// Class returns the math class of this fragment.
	Class() MathClass

	// Width returns the natural width of the fragment.
	Width() Abs

	// Height returns the height above the baseline.
	Height() Abs

	// Depth returns the depth below the baseline.
	Depth() Abs

	// Ascent returns the ascent (height above baseline).
	Ascent() Abs

	// Descent returns the descent (depth below baseline).
	Descent() Abs

	// ItalicsCorrection returns the italics correction for this fragment.
	// Used when attaching superscripts to slanted characters.
	ItalicsCorrection() Abs
}

// MathGlyphFragment represents a glyph in math layout.
// This is the most fundamental math fragment, representing a single
// character or symbol with its metrics.
type MathGlyphFragment struct {
	// id is the glyph identifier from the font.
	id uint16
	// c is the Unicode codepoint (if applicable).
	c rune
	// fontSize is the font size for this glyph.
	fontSize Abs
	// width is the advance width.
	width Abs
	// ascent is the height above baseline.
	ascent Abs
	// descent is the depth below baseline.
	descent Abs
	// italicsCorr is the italics correction.
	italicsCorr Abs
	// class is the math classification.
	class MathClass
	// accent indicates if this is used as an accent.
	accent bool
	// stretchable indicates if this glyph can be stretched.
	stretchable bool
}

func (*MathGlyphFragment) isMathFragment() {}

// NewMathGlyphFragment creates a new glyph fragment.
func NewMathGlyphFragment(id uint16, c rune, fontSize Abs) *MathGlyphFragment {
	return &MathGlyphFragment{
		id:       id,
		c:        c,
		fontSize: fontSize,
		class:    MathClassNormal,
	}
}

// GlyphID returns the glyph identifier.
func (g *MathGlyphFragment) GlyphID() uint16 {
	return g.id
}

// Char returns the Unicode codepoint.
func (g *MathGlyphFragment) Char() rune {
	return g.c
}

// FontSize returns the font size.
func (g *MathGlyphFragment) FontSize() Abs {
	return g.fontSize
}

// Class returns the math class.
func (g *MathGlyphFragment) Class() MathClass {
	return g.class
}

// SetClass sets the math class.
func (g *MathGlyphFragment) SetClass(class MathClass) {
	g.class = class
}

// Width returns the advance width.
func (g *MathGlyphFragment) Width() Abs {
	return g.width
}

// SetWidth sets the advance width.
func (g *MathGlyphFragment) SetWidth(w Abs) {
	g.width = w
}

// Height returns the height above baseline (same as Ascent).
func (g *MathGlyphFragment) Height() Abs {
	return g.ascent
}

// Depth returns the depth below baseline (same as Descent).
func (g *MathGlyphFragment) Depth() Abs {
	return g.descent
}

// Ascent returns the ascent.
func (g *MathGlyphFragment) Ascent() Abs {
	return g.ascent
}

// SetAscent sets the ascent.
func (g *MathGlyphFragment) SetAscent(a Abs) {
	g.ascent = a
}

// Descent returns the descent.
func (g *MathGlyphFragment) Descent() Abs {
	return g.descent
}

// SetDescent sets the descent.
func (g *MathGlyphFragment) SetDescent(d Abs) {
	g.descent = d
}

// ItalicsCorrection returns the italics correction.
func (g *MathGlyphFragment) ItalicsCorrection() Abs {
	return g.italicsCorr
}

// SetItalicsCorrection sets the italics correction.
func (g *MathGlyphFragment) SetItalicsCorrection(ic Abs) {
	g.italicsCorr = ic
}

// IsAccent returns true if this glyph is used as an accent.
func (g *MathGlyphFragment) IsAccent() bool {
	return g.accent
}

// SetAccent marks this glyph as an accent.
func (g *MathGlyphFragment) SetAccent(accent bool) {
	g.accent = accent
}

// IsStretchable returns true if this glyph can be stretched.
func (g *MathGlyphFragment) IsStretchable() bool {
	return g.stretchable
}

// SetStretchable marks this glyph as stretchable.
func (g *MathGlyphFragment) SetStretchable(stretchable bool) {
	g.stretchable = stretchable
}

// TotalHeight returns the total height (ascent + descent).
func (g *MathGlyphFragment) TotalHeight() Abs {
	return g.ascent + g.descent
}

// WithMetrics returns a copy with the given metrics.
func (g *MathGlyphFragment) WithMetrics(width, ascent, descent, italicsCorr Abs) *MathGlyphFragment {
	clone := *g
	clone.width = width
	clone.ascent = ascent
	clone.descent = descent
	clone.italicsCorr = italicsCorr
	return &clone
}

// MathFrameFragment represents a composed frame in math layout.
// This wraps an existing FinalFrame with math-specific properties.
type MathFrameFragment struct {
	// frame is the underlying composed frame.
	frame *FinalFrame
	// class is the math classification.
	class MathClass
	// italicsCorr is the italics correction for the frame.
	italicsCorr Abs
	// baseAscent is the baseline-relative ascent.
	baseAscent Abs
	// baseDescent is the baseline-relative descent.
	baseDescent Abs
}

func (*MathFrameFragment) isMathFragment() {}

// NewMathFrameFragment creates a frame fragment from a FinalFrame.
func NewMathFrameFragment(frame *FinalFrame) *MathFrameFragment {
	if frame == nil {
		return &MathFrameFragment{
			frame: &FinalFrame{},
			class: MathClassNone,
		}
	}
	return &MathFrameFragment{
		frame:       frame,
		class:       MathClassNone,
		baseAscent:  frame.Baseline,
		baseDescent: frame.Size.Height - frame.Baseline,
	}
}

// Frame returns the underlying FinalFrame.
func (f *MathFrameFragment) Frame() *FinalFrame {
	return f.frame
}

// Class returns the math class.
func (f *MathFrameFragment) Class() MathClass {
	return f.class
}

// SetClass sets the math class.
func (f *MathFrameFragment) SetClass(class MathClass) {
	f.class = class
}

// Width returns the frame width.
func (f *MathFrameFragment) Width() Abs {
	if f.frame == nil {
		return 0
	}
	return f.frame.Size.Width
}

// Height returns the height above baseline.
func (f *MathFrameFragment) Height() Abs {
	return f.baseAscent
}

// Depth returns the depth below baseline.
func (f *MathFrameFragment) Depth() Abs {
	return f.baseDescent
}

// Ascent returns the ascent (same as Height).
func (f *MathFrameFragment) Ascent() Abs {
	return f.baseAscent
}

// Descent returns the descent (same as Depth).
func (f *MathFrameFragment) Descent() Abs {
	return f.baseDescent
}

// ItalicsCorrection returns the italics correction.
func (f *MathFrameFragment) ItalicsCorrection() Abs {
	return f.italicsCorr
}

// SetItalicsCorrection sets the italics correction.
func (f *MathFrameFragment) SetItalicsCorrection(ic Abs) {
	f.italicsCorr = ic
}

// TotalHeight returns the total height (ascent + descent).
func (f *MathFrameFragment) TotalHeight() Abs {
	return f.baseAscent + f.baseDescent
}

// SetBaseline adjusts the baseline position.
func (f *MathFrameFragment) SetBaseline(ascent, descent Abs) {
	f.baseAscent = ascent
	f.baseDescent = descent
}

// MathSpaceFragment represents horizontal space in math layout.
// Space fragments are used for spacing between math atoms according
// to TeX's math spacing rules.
type MathSpaceFragment struct {
	// width is the space width.
	width Abs
	// class is always MathClassSpace.
	class MathClass
}

func (*MathSpaceFragment) isMathFragment() {}

// NewMathSpaceFragment creates a space fragment.
func NewMathSpaceFragment(width Abs) *MathSpaceFragment {
	return &MathSpaceFragment{
		width: width,
		class: MathClassSpace,
	}
}

// Class returns MathClassSpace.
func (s *MathSpaceFragment) Class() MathClass {
	return s.class
}

// Width returns the space width.
func (s *MathSpaceFragment) Width() Abs {
	return s.width
}

// Height returns zero (spaces have no height).
func (s *MathSpaceFragment) Height() Abs {
	return 0
}

// Depth returns zero (spaces have no depth).
func (s *MathSpaceFragment) Depth() Abs {
	return 0
}

// Ascent returns zero.
func (s *MathSpaceFragment) Ascent() Abs {
	return 0
}

// Descent returns zero.
func (s *MathSpaceFragment) Descent() Abs {
	return 0
}

// ItalicsCorrection returns zero (no italics correction for space).
func (s *MathSpaceFragment) ItalicsCorrection() Abs {
	return 0
}

// MathLinebreakFragment represents a line break point in math.
type MathLinebreakFragment struct{}

func (*MathLinebreakFragment) isMathFragment() {}

// Class returns MathClassNone.
func (l *MathLinebreakFragment) Class() MathClass {
	return MathClassNone
}

// Width returns zero.
func (l *MathLinebreakFragment) Width() Abs {
	return 0
}

// Height returns zero.
func (l *MathLinebreakFragment) Height() Abs {
	return 0
}

// Depth returns zero.
func (l *MathLinebreakFragment) Depth() Abs {
	return 0
}

// Ascent returns zero.
func (l *MathLinebreakFragment) Ascent() Abs {
	return 0
}

// Descent returns zero.
func (l *MathLinebreakFragment) Descent() Abs {
	return 0
}

// ItalicsCorrection returns zero.
func (l *MathLinebreakFragment) ItalicsCorrection() Abs {
	return 0
}

// MathAlignFragment represents an alignment point in math.
type MathAlignFragment struct{}

func (*MathAlignFragment) isMathFragment() {}

// Class returns MathClassNone.
func (a *MathAlignFragment) Class() MathClass {
	return MathClassNone
}

// Width returns zero.
func (a *MathAlignFragment) Width() Abs {
	return 0
}

// Height returns zero.
func (a *MathAlignFragment) Height() Abs {
	return 0
}

// Depth returns zero.
func (a *MathAlignFragment) Depth() Abs {
	return 0
}

// Ascent returns zero.
func (a *MathAlignFragment) Ascent() Abs {
	return 0
}

// Descent returns zero.
func (a *MathAlignFragment) Descent() Abs {
	return 0
}

// ItalicsCorrection returns zero.
func (a *MathAlignFragment) ItalicsCorrection() Abs {
	return 0
}

// MathSpacing returns the spacing amount between two math classes.
// This implements TeX's math spacing rules.
func MathSpacing(left, right MathClass, scriptLevel int) Abs {
	// No spacing in script/scriptscript styles
	if scriptLevel > 0 {
		return 0
	}

	// TeX spacing table (in mu, converted to approximate em fractions)
	// Thin space = 3mu ≈ 0.167em
	// Medium space = 4mu ≈ 0.222em
	// Thick space = 5mu ≈ 0.278em
	const (
		thin   = Em(0.167)
		medium = Em(0.222)
		thick  = Em(0.278)
	)

	// Default font size for spacing calculation
	const defaultSize = Abs(10)

	switch left {
	case MathClassNormal:
		switch right {
		case MathClassLarge:
			return thin.At(defaultSize)
		case MathClassBinary:
			return medium.At(defaultSize)
		case MathClassRelation:
			return thick.At(defaultSize)
		case MathClassOpening:
			return 0
		case MathClassClosing:
			return 0
		case MathClassPunctuation:
			return 0
		}
	case MathClassLarge:
		switch right {
		case MathClassNormal:
			return thin.At(defaultSize)
		case MathClassLarge:
			return thin.At(defaultSize)
		case MathClassBinary:
			return medium.At(defaultSize)
		case MathClassRelation:
			return thick.At(defaultSize)
		case MathClassOpening:
			return 0
		case MathClassPunctuation:
			return 0
		}
	case MathClassBinary:
		switch right {
		case MathClassNormal:
			return medium.At(defaultSize)
		case MathClassLarge:
			return medium.At(defaultSize)
		case MathClassOpening:
			return medium.At(defaultSize)
		}
	case MathClassRelation:
		switch right {
		case MathClassNormal:
			return thick.At(defaultSize)
		case MathClassLarge:
			return thick.At(defaultSize)
		case MathClassOpening:
			return thick.At(defaultSize)
		}
	case MathClassOpening:
		// No space after opening
		return 0
	case MathClassClosing:
		switch right {
		case MathClassLarge:
			return thin.At(defaultSize)
		case MathClassBinary:
			return medium.At(defaultSize)
		case MathClassRelation:
			return thick.At(defaultSize)
		case MathClassPunctuation:
			return 0
		}
	case MathClassPunctuation:
		switch right {
		case MathClassNormal:
			return thin.At(defaultSize)
		case MathClassLarge:
			return thin.At(defaultSize)
		case MathClassRelation:
			return thin.At(defaultSize)
		case MathClassOpening:
			return thin.At(defaultSize)
		}
	}

	return 0
}

// ClassifyMathChar returns the math class for a character.
func ClassifyMathChar(c rune) MathClass {
	switch c {
	// Binary operators
	case '+', '-', '±', '∓', '×', '÷', '·', '∗', '⊕', '⊖', '⊗', '⊘':
		return MathClassBinary
	// Relation symbols
	case '=', '<', '>', '≤', '≥', '≠', '≈', '∼', '≡', '⊂', '⊃', '∈', '∋':
		return MathClassRelation
	// Large operators
	case '∑', '∏', '∐', '∫', '∬', '∭', '∮', '⋀', '⋁', '⋂', '⋃':
		return MathClassLarge
	// Opening delimiters
	case '(', '[', '{', '⟨', '⌈', '⌊':
		return MathClassOpening
	// Closing delimiters
	case ')', ']', '}', '⟩', '⌉', '⌋':
		return MathClassClosing
	// Punctuation
	case ',', ';', ':':
		return MathClassPunctuation
	// Fence
	case '|', '‖':
		return MathClassFence
	default:
		return MathClassNormal
	}
}
