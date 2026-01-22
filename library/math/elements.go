// Math element types for Typst.
// Translated from typst-library/src/math/

package math

import "github.com/boergens/gotypst/library/foundations"

// EquationElem represents a mathematical equation.
// Matches Rust: typst-library/src/math/equation.rs
type EquationElem struct {
	// Body is the equation content.
	Body foundations.Content
	// Block indicates if this is a block (display) equation.
	Block bool
}

func (*EquationElem) IsContentElement() {}

// FracElem represents a fraction.
// Matches Rust: typst-library/src/math/frac.rs
type FracElem struct {
	// Num is the numerator content.
	Num foundations.Content
	// Denom is the denominator content.
	Denom foundations.Content
}

func (*FracElem) IsContentElement() {}

// RootElem represents a root (square root, nth root).
// Matches Rust: typst-library/src/math/root.rs
type RootElem struct {
	// Index is the optional root index (nil for square root).
	Index foundations.Content
	// Radicand is the content under the root sign.
	Radicand foundations.Content
}

func (*RootElem) IsContentElement() {}

// AttachElem represents subscripts and superscripts.
// Matches Rust: typst-library/src/math/attach.rs
type AttachElem struct {
	// Base is the base expression.
	Base foundations.Content
	// T is the top/superscript content.
	T *foundations.Content
	// B is the bottom/subscript content.
	B *foundations.Content
	// TR is the top-right content (for primes).
	TR *foundations.Content
}

func (*AttachElem) IsContentElement() {}

// LrElem represents left-right delimited content.
// Matches Rust: typst-library/src/math/lr.rs
type LrElem struct {
	// Body is the delimited content (including delimiters).
	Body foundations.Content
}

func (*LrElem) IsContentElement() {}

// AlignPointElem represents an alignment point in equations.
// Matches Rust: typst-library/src/math/align.rs
type AlignPointElem struct{}

func (*AlignPointElem) IsContentElement() {}

// Shared instance for align points.
var sharedAlignPoint = &AlignPointElem{}

// SharedAlignPoint returns a shared align point element.
func SharedAlignPoint() *AlignPointElem {
	return sharedAlignPoint
}

// PrimesElem represents prime marks.
// Matches Rust: typst-library/src/math/attach.rs
type PrimesElem struct {
	// Count is the number of prime marks.
	Count int
}

func (*PrimesElem) IsContentElement() {}

// LimitsElem represents an operator with limits above and below.
// Matches Rust: typst-library/src/math/op.rs
type LimitsElem struct {
	// Body is the main operator content.
	Body foundations.Content
	// Inline indicates inline (side) positioning vs display (above/below).
	Inline bool
}

func (*LimitsElem) IsContentElement() {}

// AccentElem represents a math accent (hat, tilde, bar, vec, etc.).
// Matches Rust: typst-library/src/math/accent.rs
type AccentElem struct {
	// Base is the content being accented.
	Base foundations.Content
	// Accent is the accent character.
	Accent rune
}

func (*AccentElem) IsContentElement() {}

// Common accent characters.
const (
	AccentHat    = '\u0302' // COMBINING CIRCUMFLEX ACCENT
	AccentTilde  = '\u0303' // COMBINING TILDE
	AccentBar    = '\u0304' // COMBINING MACRON
	AccentVec    = '\u20D7' // COMBINING RIGHT ARROW ABOVE
	AccentDot    = '\u0307' // COMBINING DOT ABOVE
	AccentDDot   = '\u0308' // COMBINING DIAERESIS
	AccentBreve  = '\u0306' // COMBINING BREVE
	AccentAcute  = '\u0301' // COMBINING ACUTE ACCENT
	AccentGrave  = '\u0300' // COMBINING GRAVE ACCENT
)
