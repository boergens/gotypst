package math

import (
	"github.com/boergens/gotypst/eval"
)

// LayoutEquation lays out an equation element.
// This is the main entry point for equation layout.
func LayoutEquation(elem *eval.EquationElement, fontSize Abs) *MathFrame {
	// Determine the math style based on whether it's a block equation
	style := TextStyle
	if elem.Block {
		style = DisplayStyle
	}

	ctx := &MathContext{
		FontSize: fontSize,
		Style:    style,
		Cramped:  false,
	}

	constants := DefaultMathConstants()

	return LayoutContent(&elem.Body, ctx, constants)
}

// EquationLayoutResult contains the result of equation layout
// along with rendering information.
type EquationLayoutResult struct {
	// Frame is the laid out math content.
	Frame *MathFrame
	// FontSize is the base font size used for layout.
	FontSize Abs
	// IsBlock indicates if this was a block (display) equation.
	IsBlock bool
}

// LayoutEquationWithResult lays out an equation and returns a result
// with additional metadata for rendering.
func LayoutEquationWithResult(elem *eval.EquationElement, fontSize Abs) *EquationLayoutResult {
	frame := LayoutEquation(elem, fontSize)
	return &EquationLayoutResult{
		Frame:    frame,
		FontSize: fontSize,
		IsBlock:  elem.Block,
	}
}
