// Package flow provides block-level flow layout for the Typst layout engine.
//
// This package is a Go translation of typst-layout/src/flow from the original
// Typst compiler. It handles:
//
//   - Page composition with multiple columns
//   - Float placement (top/bottom positioning)
//   - Footnote layout and migration
//   - Line number positioning
//   - Content distribution across regions
//
// The main entry point is the Compose function which takes work items and
// distributes them across regions while handling out-of-flow content like
// floats and footnotes.
package flow
