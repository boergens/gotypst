// Package eval layout element functions.
//
// This package contains layout element functions that correspond to Typst's
// layout module. Each element type is defined in its own file:
//
//   - elem_layout_stack.go:     stack() - stacking layout
//   - elem_layout_align.go:     align() - alignment container
//   - elem_layout_columns.go:   columns() - multi-column layout
//   - elem_layout_container.go: box(), block() - container elements
//   - elem_layout_pad.go:       pad() - padding container
//   - elem_layout_grid.go:      grid() - grid-based layout
//   - elem_layout_page.go:      page() - page configuration
//
// Reference: typst-reference/crates/typst-library/src/layout/
package eval
