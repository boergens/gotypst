// Package grid provides grid and table layout for GoTypst.
//
// This package implements the grid layout system including:
//   - Grid line rendering (strokes, borders)
//   - Grid layouter with multi-region support
//   - Repeated headers/footers
//   - Rowspan management
//
// The grid layout system is used by both #grid() and #table() elements
// through a unified GridLayouter that supports multi-region pagination,
// rowspans, colspans, and repeating headers/footers.
//
// Pipeline position:
//
//	Source -> Parse -> Evaluate -> Realize -> LAYOUT (grid) -> Render
package grid
