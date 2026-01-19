// Package math provides mathematical equation layout for Typst.
//
// This package is a Go translation of the math module from typst-layout.
// It implements the layout of mathematical expressions following OpenType
// MATH table specifications.
//
// Components:
// - fragment.go: Core math fragment types (Glyph, Frame, Space, etc.)
// - math.go: MathContext and entry points for equation layout
// - run.go: Math run handling and row alignment
// - scripts.go: Superscript/subscript positioning
// - fraction.go: Fraction layout
// - radical.go: Square root and radical symbols
// - accent.go: Accent positioning above/below bases
// - fenced.go: Parentheses and brackets with delimiter stretching
// - table.go: Matrix and table layout
// - text.go: Text in math mode
// - shaping.go: Math-specific text shaping
// - line.go: Overline/underline layout
// - cancel.go: Cancel mark layout
package math
