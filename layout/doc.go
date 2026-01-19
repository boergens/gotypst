// Package layout provides the layout engine for Typst.
//
// This package is a Go translation of typst-layout from the original Typst
// compiler. It handles the conversion of abstract document content into
// positioned frames ready for rendering.
//
// The layout engine handles:
// - Page layout and pagination
// - Block-level flow layout
// - Inline text layout with line breaking
// - Mathematical equation layout
// - Grid and table layout
// - Geometric transformations
package layout
