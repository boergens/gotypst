package text

import (
	"github.com/boergens/gotypst/layout/inline"
)

// Size represents a text size value (typically in points).
type Size float64

// SizeFromPt creates a Size from points.
func SizeFromPt(pt float64) Size {
	return Size(pt)
}

// SizeFromEm creates a Size from em units at a given base size.
func SizeFromEm(em float64, base Size) Size {
	return Size(em * float64(base))
}

// Points returns the size in typographic points.
func (s Size) Points() float64 {
	return float64(s)
}

// ToAbs converts the size to layout.Abs.
func (s Size) ToAbs() inline.Abs {
	return inline.Abs(s)
}

// Em represents an em unit (relative to font size).
type Em = inline.Em

// Dir represents text direction.
type Dir int

const (
	// DirLTR is left-to-right direction.
	DirLTR Dir = iota
	// DirRTL is right-to-left direction.
	DirRTL
)

// String returns the direction as a string.
func (d Dir) String() string {
	switch d {
	case DirRTL:
		return "rtl"
	default:
		return "ltr"
	}
}

// ToInline converts to the inline package Dir type.
func (d Dir) ToInline() inline.Dir {
	if d == DirRTL {
		return inline.DirRTL
	}
	return inline.DirLTR
}

// FontWeight represents font weight (100-900).
type FontWeight int

// Predefined font weights.
const (
	FontWeightThin       FontWeight = 100
	FontWeightExtraLight FontWeight = 200
	FontWeightLight      FontWeight = 300
	FontWeightNormal     FontWeight = 400
	FontWeightMedium     FontWeight = 500
	FontWeightSemiBold   FontWeight = 600
	FontWeightBold       FontWeight = 700
	FontWeightExtraBold  FontWeight = 800
	FontWeightBlack      FontWeight = 900
)

// String returns the weight as a string.
func (w FontWeight) String() string {
	switch w {
	case FontWeightThin:
		return "thin"
	case FontWeightExtraLight:
		return "extralight"
	case FontWeightLight:
		return "light"
	case FontWeightNormal:
		return "normal"
	case FontWeightMedium:
		return "medium"
	case FontWeightSemiBold:
		return "semibold"
	case FontWeightBold:
		return "bold"
	case FontWeightExtraBold:
		return "extrabold"
	case FontWeightBlack:
		return "black"
	default:
		return "normal"
	}
}

// ToInline converts to the inline package FontWeight type.
func (w FontWeight) ToInline() inline.FontWeight {
	return inline.FontWeight(w)
}

// FontStyle represents font style.
type FontStyle int

const (
	// FontStyleNormal is upright text.
	FontStyleNormal FontStyle = iota
	// FontStyleItalic is italic text.
	FontStyleItalic
	// FontStyleOblique is oblique (slanted) text.
	FontStyleOblique
)

// String returns the style as a string.
func (s FontStyle) String() string {
	switch s {
	case FontStyleItalic:
		return "italic"
	case FontStyleOblique:
		return "oblique"
	default:
		return "normal"
	}
}

// ToInline converts to the inline package FontStyle type.
func (s FontStyle) ToInline() inline.FontStyle {
	switch s {
	case FontStyleItalic:
		return inline.FontStyleItalic
	case FontStyleOblique:
		return inline.FontStyleOblique
	default:
		return inline.FontStyleNormal
	}
}

// FontStretch represents font stretch/width.
type FontStretch int

const (
	// FontStretchNormal is normal width.
	FontStretchNormal FontStretch = iota
	// FontStretchCondensed is condensed width.
	FontStretchCondensed
	// FontStretchExpanded is expanded width.
	FontStretchExpanded
)

// String returns the stretch as a string.
func (s FontStretch) String() string {
	switch s {
	case FontStretchCondensed:
		return "condensed"
	case FontStretchExpanded:
		return "expanded"
	default:
		return "normal"
	}
}

// ToInline converts to the inline package FontStretch type.
func (s FontStretch) ToInline() inline.FontStretch {
	switch s {
	case FontStretchCondensed:
		return inline.FontStretchCondensed
	case FontStretchExpanded:
		return inline.FontStretchExpanded
	default:
		return inline.FontStretchNormal
	}
}

// NumberType controls number styling.
type NumberType int

const (
	// NumberTypeLining uses lining (uppercase) figures.
	NumberTypeLining NumberType = iota
	// NumberTypeOldstyle uses oldstyle (lowercase) figures.
	NumberTypeOldstyle
)

// String returns the number type as a string.
func (n NumberType) String() string {
	switch n {
	case NumberTypeOldstyle:
		return "oldstyle"
	default:
		return "lining"
	}
}

// NumberWidth controls number width.
type NumberWidth int

const (
	// NumberWidthProportional uses proportional-width figures.
	NumberWidthProportional NumberWidth = iota
	// NumberWidthTabular uses tabular (fixed-width) figures.
	NumberWidthTabular
)

// String returns the number width as a string.
func (n NumberWidth) String() string {
	switch n {
	case NumberWidthTabular:
		return "tabular"
	default:
		return "proportional"
	}
}
