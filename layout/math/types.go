package math

import (
	"github.com/boergens/gotypst/layout"
)

// Abs is an alias for layout.Abs for convenience.
type Abs = layout.Abs

// Em is an alias for layout.Em for convenience.
type Em = layout.Em

// MathFrame represents a layouted math element.
// It contains positioned items and size/baseline information.
type MathFrame struct {
	// Size is the total size of the frame.
	Size Size
	// Baseline is the distance from the top to the baseline.
	// For fractions, this is the math axis (center of the fraction bar).
	Baseline Abs
	// Items contains the positioned items in the frame.
	Items []FrameEntry
}

// Size represents 2D dimensions.
type Size struct {
	Width, Height Abs
}

// Point represents a 2D coordinate.
type Point struct {
	X, Y Abs
}

// FrameEntry is a positioned item in a frame.
type FrameEntry struct {
	Pos  Point
	Item MathFrameItem
}

// MathFrameItem represents an item that can be placed in a math frame.
// This is distinct from fragment.FrameItem which is used in FrameFragment.
type MathFrameItem interface {
	isMathFrameItem()
}

// TextItem represents text content in a math frame.
type TextItem struct {
	// Text is the text content to display.
	Text string
	// FontSize is the font size for this text.
	FontSize Abs
}

func (TextItem) isMathFrameItem() {}

// LineItem represents a line (e.g., fraction bar) in a math frame.
type LineItem struct {
	// Length is the length of the line.
	Length Abs
	// Thickness is the line thickness.
	Thickness Abs
}

func (LineItem) isMathFrameItem() {}

// ChildFrame represents a nested math frame.
type ChildFrame struct {
	// Frame is the nested frame.
	Frame *MathFrame
}

func (ChildFrame) isMathFrameItem() {}

// Push adds an item to the frame at the given position.
func (f *MathFrame) Push(pos Point, item MathFrameItem) {
	f.Items = append(f.Items, FrameEntry{Pos: pos, Item: item})
}

// PushFrame adds a child frame at the given position.
func (f *MathFrame) PushFrame(pos Point, child *MathFrame) {
	f.Push(pos, ChildFrame{Frame: child})
}

// Width returns the frame's width.
func (f *MathFrame) Width() Abs {
	return f.Size.Width
}

// Height returns the frame's height.
func (f *MathFrame) Height() Abs {
	return f.Size.Height
}

// MathContext provides layout context for math elements.
type MathContext struct {
	// FontSize is the current font size.
	FontSize Abs
	// Style is the current math style (display, text, script, scriptscript).
	Style MathStyle
	// Cramped indicates if the style is cramped (affects superscript positioning).
	Cramped bool
}

// FontSizeForStyle returns the font size for a given math style.
func (ctx *MathContext) FontSizeForStyle(style MathStyle) Abs {
	switch style {
	case StyleDisplay, StyleText:
		return ctx.FontSize
	case StyleScript:
		return ctx.FontSize * 0.7 // 70% of base size
	case StyleScriptScript:
		return ctx.FontSize * 0.5 // 50% of base size
	default:
		return ctx.FontSize
	}
}

// MathConstants contains typographic constants for math layout.
// These values are based on traditional math typography conventions.
type MathConstants struct {
	// AxisHeight is the height of the math axis above the baseline.
	// This is where the fraction bar is drawn.
	AxisHeight Em

	// FractionRuleThickness is the thickness of the fraction bar.
	FractionRuleThickness Em

	// FractionNumeratorGapMin is the minimum gap between numerator and fraction bar.
	FractionNumeratorGapMin Em

	// FractionDenominatorGapMin is the minimum gap between fraction bar and denominator.
	FractionDenominatorGapMin Em

	// FractionNumeratorShiftUp is how much to shift the numerator up in display style.
	FractionNumeratorShiftUp Em

	// FractionDenominatorShiftDown is how much to shift the denominator down in display style.
	FractionDenominatorShiftDown Em

	// StackTopShiftUp is how much to shift the top of a stack up (for inline fractions).
	StackTopShiftUp Em

	// StackBottomShiftDown is how much to shift the bottom of a stack down.
	StackBottomShiftDown Em

	// StackGapMin is the minimum gap between stacked elements.
	StackGapMin Em
}

// DefaultMathConstants returns the default math typography constants.
// These are based on traditional TeX values and common math fonts.
func DefaultMathConstants() MathConstants {
	return MathConstants{
		AxisHeight:                   Em(0.25),  // Quarter of em
		FractionRuleThickness:        Em(0.04),  // 4% of em
		FractionNumeratorGapMin:      Em(0.10),  // 10% of em
		FractionDenominatorGapMin:    Em(0.10),  // 10% of em
		FractionNumeratorShiftUp:     Em(0.68),  // ~2/3 em for display style
		FractionDenominatorShiftDown: Em(0.68),  // ~2/3 em for display style
		StackTopShiftUp:              Em(0.34),  // ~1/3 em for inline style
		StackBottomShiftDown:         Em(0.34),  // ~1/3 em for inline style
		StackGapMin:                  Em(0.20),  // 20% of em
	}
}
