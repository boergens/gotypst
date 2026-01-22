// Spacing elements for horizontal and vertical space.
// Translated from typst-library/src/layout/spacing.rs

package layout

import "github.com/boergens/gotypst/library/foundations"

// HElem represents horizontal spacing.
type HElem struct {
	// Amount is the spacing amount.
	Amount Spacing
	// Weak indicates if this is weak spacing that collapses.
	Weak bool
}

func (*HElem) IsContentElement() {}

// VElem represents vertical spacing.
type VElem struct {
	// Amount is the spacing amount.
	Amount Spacing
	// Weak indicates if this is weak spacing that collapses.
	Weak bool
	// Attach indicates if this spacing attaches to previous element.
	Attach bool
}

func (*VElem) IsContentElement() {}

// Spacing represents a spacing amount (absolute or fractional).
type Spacing struct {
	// Abs is the absolute spacing in points (if not fractional).
	Abs foundations.Length
	// Fr is the fractional spacing (if fractional).
	Fr foundations.Fraction
	// IsFractional indicates if this is fractional spacing.
	IsFractional bool
}

// IsFrac returns true if this is fractional spacing.
func (s Spacing) IsFrac() bool {
	return s.IsFractional
}
