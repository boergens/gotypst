// Package grid implements grid and table layout for Typst documents.
//
// The grid module provides the GridLayouter which handles both #grid() and
// #table() elements through a unified layout algorithm that supports:
//   - Multi-region pagination across page/column breaks
//   - Rowspans (cells spanning multiple rows)
//   - Colspans (cells spanning multiple columns)
//   - Repeating headers and footers
//   - Auto, relative, and fractional column/row sizing
//
// This is a Go translation of typst-layout/src/grid/ from the Typst project.
package grid
