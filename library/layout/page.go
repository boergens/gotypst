package layout

import (
	"github.com/boergens/gotypst/library/foundations"
)

// Paper represents a standard paper size.
//
// Reference: typst-reference/crates/typst-library/src/layout/page.rs
type Paper struct {
	Name   string
	Width  float64 // in points
	Height float64 // in points
}

// Standard paper sizes (width and height in points).
var Papers = map[string]Paper{
	"a0":         {Name: "a0", Width: 2383.94, Height: 3370.39},
	"a1":         {Name: "a1", Width: 1683.78, Height: 2383.94},
	"a2":         {Name: "a2", Width: 1190.55, Height: 1683.78},
	"a3":         {Name: "a3", Width: 841.89, Height: 1190.55},
	"a4":         {Name: "a4", Width: 595.28, Height: 841.89},
	"a5":         {Name: "a5", Width: 419.53, Height: 595.28},
	"a6":         {Name: "a6", Width: 297.64, Height: 419.53},
	"a7":         {Name: "a7", Width: 209.76, Height: 297.64},
	"a8":         {Name: "a8", Width: 147.40, Height: 209.76},
	"us-letter":  {Name: "us-letter", Width: 612, Height: 792},
	"us-legal":   {Name: "us-legal", Width: 612, Height: 1008},
	"us-tabloid": {Name: "us-tabloid", Width: 792, Height: 1224},
}

// PageElement represents a page layout element.
// It configures page properties like size, margins, headers, footers, etc.
//
// Reference: typst-reference/crates/typst-library/src/layout/page.rs
type PageElement struct {
	// Paper is a standard paper size name (e.g., "a4", "us-letter").
	Paper *string
	// Width is the page width in points. If nil, uses paper width.
	Width *float64
	// Height is the page height in points. If nil, uses paper height.
	// Can also be "auto" for infinite height.
	Height *float64
	// HeightAuto indicates height should grow to fit content.
	HeightAuto bool
	// Flipped indicates landscape orientation.
	Flipped bool
	// Margin is the page margins. Can be a single value or per-side.
	Margin foundations.Value
	// MarginLeft, MarginTop, MarginRight, MarginBottom are individual margins.
	MarginLeft   *float64
	MarginTop    *float64
	MarginRight  *float64
	MarginBottom *float64
	// Columns is the number of columns on the page.
	Columns *int
	// Fill is the page background fill.
	Fill foundations.Value
	// Numbering is the page numbering pattern.
	Numbering foundations.Value
	// NumberAlign is the alignment of page numbers.
	NumberAlign foundations.Value
	// Header is the page header content.
	Header foundations.Value
	// HeaderAscent is how much the header is raised into the margin.
	HeaderAscent *float64
	// Footer is the page footer content.
	Footer foundations.Value
	// FooterDescent is how much the footer is lowered into the margin.
	FooterDescent *float64
	// Background is content behind the page body.
	Background foundations.Value
	// Foreground is content in front of the page body.
	Foreground foundations.Value
	// Body is the page content (only when used as constructor).
	Body foundations.Content
}

func (*PageElement) IsContentElement() {}
