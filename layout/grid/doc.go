// Package grid implements grid and table layout with multi-region pagination support.
//
// This package provides the GridLayouter which handles both #grid() and #table() elements
// through a unified layout algorithm that supports:
//   - Multi-region pagination (content spanning multiple pages)
//   - Rowspans and colspans
//   - Repeating headers and footers across page breaks
//   - Grid line rendering with stroke priorities
//   - RTL (right-to-left) support
//
// The layout algorithm works in phases:
//  1. Column measurement: Resolve column widths (fixed, auto, fractional)
//  2. Row layout: Layout rows with region break detection
//  3. Region finalization: Handle orphan prevention, headers/footers, rowspans
package grid
