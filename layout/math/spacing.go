package math

import "github.com/boergens/gotypst/layout"

// MathStyle represents the style context for math layout.
// This affects sizing and spacing of math elements.
type MathStyle int

const (
	// StyleDisplay is display style (large, centered equations).
	StyleDisplay MathStyle = iota
	// StyleText is text style (inline equations).
	StyleText
	// StyleScript is script style (superscripts, subscripts).
	StyleScript
	// StyleScriptScript is script-script style (nested scripts).
	StyleScriptScript
)

// IsCramped returns true if the style is cramped (reduced spacing).
func (s MathStyle) IsCramped() bool {
	return false // For now, all styles are non-cramped
}

// ScriptStyle returns the style to use for subscripts/superscripts.
func (s MathStyle) ScriptStyle() MathStyle {
	switch s {
	case StyleDisplay, StyleText:
		return StyleScript
	case StyleScript:
		return StyleScriptScript
	default:
		return StyleScriptScript
	}
}

// ScaledSize returns the font size multiplier for this style.
func (s MathStyle) ScaledSize() float64 {
	switch s {
	case StyleDisplay, StyleText:
		return 1.0
	case StyleScript:
		return 0.7 // Typical script size ratio
	case StyleScriptScript:
		return 0.5 // Typical script-script size ratio
	default:
		return 1.0
	}
}

// SpaceType represents the type of spacing between math elements.
type SpaceType int

const (
	// SpaceNone is no space.
	SpaceNone SpaceType = iota
	// SpaceThin is thin space (3mu in TeX).
	SpaceThin
	// SpaceMedium is medium space (4mu in TeX).
	SpaceMedium
	// SpaceThick is thick space (5mu in TeX).
	SpaceThick
)

// Amount returns the space amount as a fraction of an em.
// Based on TeX spacing: 1em = 18mu, so thin=3/18, med=4/18, thick=5/18.
func (t SpaceType) Amount() layout.Em {
	switch t {
	case SpaceNone:
		return 0
	case SpaceThin:
		return layout.Em(3.0 / 18.0)
	case SpaceMedium:
		return layout.Em(4.0 / 18.0)
	case SpaceThick:
		return layout.Em(5.0 / 18.0)
	default:
		return 0
	}
}

// spacingTable defines the spacing between math classes.
// This follows the TeX spacing rules from The TeXbook, Chapter 18.
// Row is left class, column is right class.
// Values: 0=none, 1=thin, 2=medium(thin in script), 3=thick(none in script).
var spacingTable = [9][9]SpaceType{
	//          Ord   Op    Bin   Rel   Open  Close Punct Inner None
	/* Ord */   {0, 1, 2, 3, 0, 0, 0, 1, 0},
	/* Op */    {1, 1, 0, 3, 0, 0, 0, 1, 0},
	/* Bin */   {2, 2, 0, 0, 2, 0, 0, 2, 0},
	/* Rel */   {3, 3, 0, 0, 3, 0, 0, 3, 0},
	/* Open */  {0, 0, 0, 0, 0, 0, 0, 0, 0},
	/* Close */ {0, 1, 2, 3, 0, 0, 0, 1, 0},
	/* Punct */ {1, 1, 0, 1, 1, 1, 1, 1, 0},
	/* Inner */ {1, 1, 2, 3, 1, 0, 1, 1, 0},
	/* None */  {0, 0, 0, 0, 0, 0, 0, 0, 0},
}

// GetSpacing returns the spacing type between two math classes.
func GetSpacing(left, right MathClass, style MathStyle) SpaceType {
	if left < 0 || int(left) >= len(spacingTable) {
		return SpaceNone
	}
	if right < 0 || int(right) >= len(spacingTable[0]) {
		return SpaceNone
	}

	space := spacingTable[left][right]

	// In script and scriptscript styles, reduce spacing.
	if style == StyleScript || style == StyleScriptScript {
		switch space {
		case SpaceMedium:
			space = SpaceThin
		case SpaceThick:
			space = SpaceNone
		}
	}

	return space
}

// GetSpacingAbs returns the absolute spacing between two math classes at a given font size.
func GetSpacingAbs(left, right MathClass, style MathStyle, fontSize layout.Abs) layout.Abs {
	space := GetSpacing(left, right, style)
	return space.Amount().At(fontSize)
}

// SpaceFragment helpers.

// NewThinSpace creates a thin space fragment at the given font size.
func NewThinSpace(fontSize layout.Abs) *SpaceFragment {
	return &SpaceFragment{Amount: SpaceThin.Amount().At(fontSize)}
}

// NewMediumSpace creates a medium space fragment at the given font size.
func NewMediumSpace(fontSize layout.Abs) *SpaceFragment {
	return &SpaceFragment{Amount: SpaceMedium.Amount().At(fontSize)}
}

// NewThickSpace creates a thick space fragment at the given font size.
func NewThickSpace(fontSize layout.Abs) *SpaceFragment {
	return &SpaceFragment{Amount: SpaceThick.Amount().At(fontSize)}
}

// NewSpace creates a space fragment with the given absolute amount.
func NewSpace(amount layout.Abs) *SpaceFragment {
	return &SpaceFragment{Amount: amount}
}

// InsertSpacing inserts appropriate spacing between fragments based on their classes.
func InsertSpacing(fragments []MathFragment, style MathStyle, fontSize layout.Abs) []MathFragment {
	if len(fragments) <= 1 {
		return fragments
	}

	result := make([]MathFragment, 0, len(fragments)*2-1)

	for i, frag := range fragments {
		if i > 0 {
			// Get spacing between previous and current fragment
			prevClass := fragments[i-1].Class()
			currClass := frag.Class()
			spacing := GetSpacingAbs(prevClass, currClass, style, fontSize)

			if spacing > 0 {
				result = append(result, NewSpace(spacing))
			}
		}
		result = append(result, frag)
	}

	return result
}
