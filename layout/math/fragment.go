// Package math provides mathematical expression layout.
package math

import (
	"github.com/boergens/gotypst/layout"
	"github.com/go-text/typesetting/font"
)

// MathFragment represents a component of mathematical layout.
// This is a sealed interface - only types in this package can implement it.
type MathFragment interface {
	// isMathFragment is a marker method to seal the interface.
	isMathFragment()
	// Width returns the width of the fragment.
	Width() layout.Abs
	// Height returns the height of the fragment (ascent + descent).
	Height() layout.Abs
	// Ascent returns the distance from baseline to top.
	Ascent() layout.Abs
	// Descent returns the distance from baseline to bottom.
	Descent() layout.Abs
	// Class returns the math class for spacing decisions.
	Class() MathClass
	// ItalicsCorrection returns the italics correction for the fragment.
	ItalicsCorrection() layout.Abs
}

// MathClass represents the mathematical spacing class of an element.
// These classes determine inter-element spacing according to TeX rules.
type MathClass int

const (
	// ClassOrd is an ordinary symbol (letters, numbers).
	ClassOrd MathClass = iota
	// ClassOp is a large operator (sum, integral, product).
	ClassOp
	// ClassBin is a binary operator (+, -, ×).
	ClassBin
	// ClassRel is a relation (=, <, >).
	ClassRel
	// ClassOpen is an opening delimiter (left paren, bracket).
	ClassOpen
	// ClassClose is a closing delimiter (right paren, bracket).
	ClassClose
	// ClassPunct is punctuation (comma, semicolon).
	ClassPunct
	// ClassInner is an inner subformula (fractions in display style).
	ClassInner
	// ClassNone indicates no class (space, alignment).
	ClassNone
)

// String returns the string representation of the math class.
func (c MathClass) String() string {
	switch c {
	case ClassOrd:
		return "Ord"
	case ClassOp:
		return "Op"
	case ClassBin:
		return "Bin"
	case ClassRel:
		return "Rel"
	case ClassOpen:
		return "Open"
	case ClassClose:
		return "Close"
	case ClassPunct:
		return "Punct"
	case ClassInner:
		return "Inner"
	case ClassNone:
		return "None"
	default:
		return "Unknown"
	}
}

// GlyphFragment represents shaped glyphs with math-specific properties.
type GlyphFragment struct {
	// Font is the font face used for these glyphs.
	Font *font.Face
	// FontSize is the size at which the glyphs are rendered.
	FontSize layout.Abs
	// Glyphs contains the individual glyph IDs.
	Glyphs []MathGlyph
	// MathClass is the spacing class for this fragment.
	MathClass MathClass
	// Italics is the italics correction value.
	Italics layout.Abs
	// Limits indicates if this is a limits-style operator.
	Limits bool
	// MidBaseline is the middle baseline position for operators.
	MidBaseline *layout.Abs
}

func (*GlyphFragment) isMathFragment() {}

// Width returns the total width of the glyph fragment.
func (g *GlyphFragment) Width() layout.Abs {
	var total layout.Abs
	for _, glyph := range g.Glyphs {
		total += glyph.Advance
	}
	return total
}

// Height returns the total height (ascent + descent).
func (g *GlyphFragment) Height() layout.Abs {
	return g.Ascent() + g.Descent()
}

// Ascent returns the maximum ascent of all glyphs.
func (g *GlyphFragment) Ascent() layout.Abs {
	var maxAscent layout.Abs
	for _, glyph := range g.Glyphs {
		if glyph.Ascent > maxAscent {
			maxAscent = glyph.Ascent
		}
	}
	return maxAscent
}

// Descent returns the maximum descent of all glyphs.
func (g *GlyphFragment) Descent() layout.Abs {
	var maxDescent layout.Abs
	for _, glyph := range g.Glyphs {
		if glyph.Descent > maxDescent {
			maxDescent = glyph.Descent
		}
	}
	return maxDescent
}

// Class returns the math class of the fragment.
func (g *GlyphFragment) Class() MathClass {
	return g.MathClass
}

// ItalicsCorrection returns the italics correction.
func (g *GlyphFragment) ItalicsCorrection() layout.Abs {
	return g.Italics
}

// MathGlyph represents a single glyph in a math fragment.
type MathGlyph struct {
	// ID is the glyph ID in the font.
	ID uint16
	// Advance is the horizontal advance width.
	Advance layout.Abs
	// XOffset is the horizontal offset from the pen position.
	XOffset layout.Abs
	// YOffset is the vertical offset from the baseline.
	YOffset layout.Abs
	// Ascent is the distance from baseline to top of the glyph.
	Ascent layout.Abs
	// Descent is the distance from baseline to bottom of the glyph.
	Descent layout.Abs
}

// FrameFragment represents a nested frame for composed math content.
// This is used for fractions, radicals, scripts, matrices, and other
// compound math structures.
type FrameFragment struct {
	// Size is the dimensions of the frame.
	Size layout.Size
	// Baseline is the distance from the top to the baseline.
	Baseline layout.Abs
	// MathClass is the spacing class for this fragment.
	MathClass MathClass
	// Italics is the italics correction value.
	Italics layout.Abs
	// Items contains the positioned items within the frame.
	Items []FrameItem
}

func (*FrameFragment) isMathFragment() {}

// Width returns the frame width.
func (f *FrameFragment) Width() layout.Abs {
	return f.Size.Width
}

// Height returns the frame height.
func (f *FrameFragment) Height() layout.Abs {
	return f.Size.Height
}

// Ascent returns the distance from baseline to top.
func (f *FrameFragment) Ascent() layout.Abs {
	return f.Baseline
}

// Descent returns the distance from baseline to bottom.
func (f *FrameFragment) Descent() layout.Abs {
	return f.Size.Height - f.Baseline
}

// Class returns the math class of the fragment.
func (f *FrameFragment) Class() MathClass {
	return f.MathClass
}

// ItalicsCorrection returns the italics correction.
func (f *FrameFragment) ItalicsCorrection() layout.Abs {
	return f.Italics
}

// FrameItem represents a positioned item within a frame.
type FrameItem struct {
	// Pos is the position of the item within the frame.
	Pos layout.Point
	// Fragment is the math fragment at this position.
	Fragment MathFragment
}

// SpaceFragment represents spacing between math elements.
type SpaceFragment struct {
	// Amount is the width of the space.
	Amount layout.Abs
}

func (*SpaceFragment) isMathFragment() {}

// Width returns the space width.
func (s *SpaceFragment) Width() layout.Abs {
	return s.Amount
}

// Height returns zero for space.
func (s *SpaceFragment) Height() layout.Abs {
	return 0
}

// Ascent returns zero for space.
func (s *SpaceFragment) Ascent() layout.Abs {
	return 0
}

// Descent returns zero for space.
func (s *SpaceFragment) Descent() layout.Abs {
	return 0
}

// Class returns ClassNone for space.
func (s *SpaceFragment) Class() MathClass {
	return ClassNone
}

// ItalicsCorrection returns zero for space.
func (s *SpaceFragment) ItalicsCorrection() layout.Abs {
	return 0
}

// LinebreakFragment marks a line break point in math layout.
type LinebreakFragment struct{}

func (*LinebreakFragment) isMathFragment() {}

// Width returns zero.
func (l *LinebreakFragment) Width() layout.Abs { return 0 }

// Height returns zero.
func (l *LinebreakFragment) Height() layout.Abs { return 0 }

// Ascent returns zero.
func (l *LinebreakFragment) Ascent() layout.Abs { return 0 }

// Descent returns zero.
func (l *LinebreakFragment) Descent() layout.Abs { return 0 }

// Class returns ClassNone.
func (l *LinebreakFragment) Class() MathClass { return ClassNone }

// ItalicsCorrection returns zero.
func (l *LinebreakFragment) ItalicsCorrection() layout.Abs { return 0 }

// AlignFragment marks an alignment point in math layout.
type AlignFragment struct{}

func (*AlignFragment) isMathFragment() {}

// Width returns zero.
func (a *AlignFragment) Width() layout.Abs { return 0 }

// Height returns zero.
func (a *AlignFragment) Height() layout.Abs { return 0 }

// Ascent returns zero.
func (a *AlignFragment) Ascent() layout.Abs { return 0 }

// Descent returns zero.
func (a *AlignFragment) Descent() layout.Abs { return 0 }

// Class returns ClassNone.
func (a *AlignFragment) Class() MathClass { return ClassNone }

// ItalicsCorrection returns zero.
func (a *AlignFragment) ItalicsCorrection() layout.Abs { return 0 }

// AccentKind represents the type of math accent decoration.
type AccentKind int

const (
	// AccentHat is the circumflex/hat accent (̂).
	AccentHat AccentKind = iota
	// AccentTilde is the tilde accent (̃).
	AccentTilde
	// AccentBar is the macron/bar/overline accent (̄).
	AccentBar
	// AccentVec is the vector arrow accent (⃗).
	AccentVec
	// AccentDot is the dot accent (̇).
	AccentDot
	// AccentDDot is the double dot/umlaut accent (̈).
	AccentDDot
	// AccentBreve is the breve accent (̆).
	AccentBreve
	// AccentAcute is the acute accent (́).
	AccentAcute
	// AccentGrave is the grave accent (̀).
	AccentGrave
)

// String returns the name of the accent kind.
func (k AccentKind) String() string {
	switch k {
	case AccentHat:
		return "hat"
	case AccentTilde:
		return "tilde"
	case AccentBar:
		return "bar"
	case AccentVec:
		return "vec"
	case AccentDot:
		return "dot"
	case AccentDDot:
		return "ddot"
	case AccentBreve:
		return "breve"
	case AccentAcute:
		return "acute"
	case AccentGrave:
		return "grave"
	default:
		return "unknown"
	}
}

// AccentGlyphID returns the Unicode code point for the accent glyph.
// These are the standalone (non-combining) forms used for display.
func (k AccentKind) AccentGlyphID() rune {
	switch k {
	case AccentHat:
		return '\u0302' // COMBINING CIRCUMFLEX ACCENT
	case AccentTilde:
		return '\u0303' // COMBINING TILDE
	case AccentBar:
		return '\u0304' // COMBINING MACRON
	case AccentVec:
		return '\u20D7' // COMBINING RIGHT ARROW ABOVE
	case AccentDot:
		return '\u0307' // COMBINING DOT ABOVE
	case AccentDDot:
		return '\u0308' // COMBINING DIAERESIS
	case AccentBreve:
		return '\u0306' // COMBINING BREVE
	case AccentAcute:
		return '\u0301' // COMBINING ACUTE ACCENT
	case AccentGrave:
		return '\u0300' // COMBINING GRAVE ACCENT
	default:
		return '\u0302' // Default to hat
	}
}

// AccentFragment represents an accent positioned over base content.
// This implements math accents like hat (̂), tilde (̃), bar (̄), vec (⃗), etc.
type AccentFragment struct {
	// Base is the content fragment being accented.
	Base MathFragment
	// Accent is the accent glyph fragment positioned above the base.
	Accent MathFragment
	// Kind is the type of accent (hat, tilde, bar, vec, etc.).
	Kind AccentKind
	// AccentGap is the vertical gap between the base and accent.
	AccentGap layout.Abs
}

func (*AccentFragment) isMathFragment() {}

// Width returns the maximum of base and accent widths.
func (a *AccentFragment) Width() layout.Abs {
	baseW := a.Base.Width()
	accentW := a.Accent.Width()
	if accentW > baseW {
		return accentW
	}
	return baseW
}

// Height returns the total height (base height + gap + accent height).
func (a *AccentFragment) Height() layout.Abs {
	return a.Ascent() + a.Descent()
}

// Ascent returns the distance from baseline to top (base ascent + gap + accent).
func (a *AccentFragment) Ascent() layout.Abs {
	return a.Base.Ascent() + a.AccentGap + a.Accent.Height()
}

// Descent returns the descent from the base.
func (a *AccentFragment) Descent() layout.Abs {
	return a.Base.Descent()
}

// Class returns the math class of the base (accents inherit base class).
func (a *AccentFragment) Class() MathClass {
	return a.Base.Class()
}

// ItalicsCorrection returns the italics correction from the base.
func (a *AccentFragment) ItalicsCorrection() layout.Abs {
	return a.Base.ItalicsCorrection()
}

// BaseOffset returns the horizontal offset to center the base under the accent.
func (a *AccentFragment) BaseOffset() layout.Abs {
	accentW := a.Accent.Width()
	baseW := a.Base.Width()
	if accentW > baseW {
		return (accentW - baseW) / 2
	}
	return 0
}

// AccentOffset returns the horizontal offset to center the accent over the base.
func (a *AccentFragment) AccentOffset() layout.Abs {
	baseW := a.Base.Width()
	accentW := a.Accent.Width()
	if baseW > accentW {
		return (baseW - accentW) / 2
	}
	return 0
}

// AccentY returns the vertical position of the accent (relative to baseline).
func (a *AccentFragment) AccentY() layout.Abs {
	return a.Base.Ascent() + a.AccentGap
}
