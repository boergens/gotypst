package layout

import (
	"github.com/boergens/gotypst/library/foundations"
)

// ColumnsElement represents a multi-column layout element.
// It arranges its body content into multiple columns.
//
// Reference: typst-reference/crates/typst-library/src/layout/columns.rs
type ColumnsElement struct {
	// Count is the number of columns.
	// If nil, defaults to 2.
	Count *int
	// Gutter is the gap between columns (in points).
	// If nil, defaults to 4% of page width.
	Gutter *float64
	// Body is the content to arrange in columns.
	Body foundations.Content
}

func (*ColumnsElement) IsContentElement() {}
