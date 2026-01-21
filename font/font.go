// Package font provides font loading, discovery, and management for gotypst.
//
// This package handles:
//   - Loading fonts from TTF/OTF/TTC files
//   - Discovering fonts from system directories
//   - Managing a collection of fonts (FontBook)
//   - Font matching by family, weight, style, and stretch
package font

import (
	"github.com/go-text/typesetting/font"
)

// Font represents a loaded font with metadata.
// It implements the gotypst.Font interface.
type Font struct {
	// face is the underlying font face for text shaping.
	face *font.Face

	// Info contains font metadata (family, style, weight, etc.).
	Info FontInfo

	// Path is the filesystem path where the font was loaded from.
	// Empty for embedded fonts.
	Path string

	// Index is the face index within a font collection (TTC).
	// Zero for single-face fonts (TTF/OTF).
	Index int

	// RawData stores the original font file bytes for subsetting.
	// This is nil for TTC fonts where the data is shared.
	RawData []byte
}

// Family returns the font family name.
// Implements gotypst.Font.
func (f *Font) Family() string {
	return f.Info.Family
}

// Style returns the font style as an integer (0=normal, 1=italic, 2=oblique).
// Implements gotypst.Font.
func (f *Font) Style() Style {
	return f.Info.Style
}

// Weight returns the font weight (100-900).
// Implements gotypst.Font.
func (f *Font) Weight() int {
	return int(f.Info.Weight)
}

// Face returns the underlying font face for text shaping.
// Implements gotypst.Font.
func (f *Font) Face() *font.Face {
	return f.face
}

// FontInfo contains metadata about a font.
type FontInfo struct {
	// Family is the font family name (e.g., "Arial", "Times New Roman").
	Family string

	// PostScriptName is the PostScript name (e.g., "Arial-BoldItalic").
	PostScriptName string

	// FullName is the full font name including style.
	FullName string

	// Style is the font style (normal, italic, oblique).
	Style Style

	// Weight is the font weight (100-900).
	Weight Weight

	// Stretch is the font stretch/width.
	Stretch Stretch
}

// Style represents font style.
type Style uint8

const (
	StyleNormal  Style = iota // Upright
	StyleItalic               // Italic
	StyleOblique              // Oblique (slanted)
)

func (s Style) String() string {
	switch s {
	case StyleNormal:
		return "normal"
	case StyleItalic:
		return "italic"
	case StyleOblique:
		return "oblique"
	default:
		return "unknown"
	}
}

// Weight represents font weight on a scale of 100-900.
type Weight int

const (
	WeightThin       Weight = 100
	WeightExtraLight Weight = 200
	WeightLight      Weight = 300
	WeightNormal     Weight = 400
	WeightMedium     Weight = 500
	WeightSemiBold   Weight = 600
	WeightBold       Weight = 700
	WeightExtraBold  Weight = 800
	WeightBlack      Weight = 900
)

func (w Weight) String() string {
	switch {
	case w <= 100:
		return "thin"
	case w <= 200:
		return "extra-light"
	case w <= 300:
		return "light"
	case w <= 400:
		return "normal"
	case w <= 500:
		return "medium"
	case w <= 600:
		return "semi-bold"
	case w <= 700:
		return "bold"
	case w <= 800:
		return "extra-bold"
	default:
		return "black"
	}
}

// Stretch represents font width/stretch.
type Stretch float32

const (
	StretchUltraCondensed Stretch = 0.5
	StretchExtraCondensed Stretch = 0.625
	StretchCondensed      Stretch = 0.75
	StretchSemiCondensed  Stretch = 0.875
	StretchNormal         Stretch = 1.0
	StretchSemiExpanded   Stretch = 1.125
	StretchExpanded       Stretch = 1.25
	StretchExtraExpanded  Stretch = 1.5
	StretchUltraExpanded  Stretch = 2.0
)

func (s Stretch) String() string {
	switch {
	case s <= 0.5:
		return "ultra-condensed"
	case s <= 0.625:
		return "extra-condensed"
	case s <= 0.75:
		return "condensed"
	case s <= 0.875:
		return "semi-condensed"
	case s <= 1.0:
		return "normal"
	case s <= 1.125:
		return "semi-expanded"
	case s <= 1.25:
		return "expanded"
	case s <= 1.5:
		return "extra-expanded"
	default:
		return "ultra-expanded"
	}
}

// Variant combines style, weight, and stretch for font matching.
type Variant struct {
	Style   Style
	Weight  Weight
	Stretch Stretch
}

// NormalVariant returns the default variant (normal style, weight, stretch).
func NormalVariant() Variant {
	return Variant{
		Style:   StyleNormal,
		Weight:  WeightNormal,
		Stretch: StretchNormal,
	}
}

// BoldVariant returns a bold variant.
func BoldVariant() Variant {
	return Variant{
		Style:   StyleNormal,
		Weight:  WeightBold,
		Stretch: StretchNormal,
	}
}

// ItalicVariant returns an italic variant.
func ItalicVariant() Variant {
	return Variant{
		Style:   StyleItalic,
		Weight:  WeightNormal,
		Stretch: StretchNormal,
	}
}

// BoldItalicVariant returns a bold italic variant.
func BoldItalicVariant() Variant {
	return Variant{
		Style:   StyleItalic,
		Weight:  WeightBold,
		Stretch: StretchNormal,
	}
}

// CanSubset returns true if the font has raw data available for subsetting.
func (f *Font) CanSubset() bool {
	return len(f.RawData) > 0
}

// NewSubsetter creates a subsetter for this font.
// Returns nil if the font doesn't have raw data for subsetting.
func (f *Font) NewSubsetter() *Subsetter {
	if !f.CanSubset() {
		return nil
	}
	return NewSubsetter(f, f.RawData)
}
