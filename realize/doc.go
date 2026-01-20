// Package realize implements the realization subsystem for GoTypst.
//
// Realization is the process that transforms evaluated content into well-known
// elements suitable for layout and rendering. It sits between evaluation and
// layout in the compilation pipeline:
//
//	Source → Parse → Evaluate → Realize → Layout → Render
//
// # Core Function
//
// The main entry point is the Realize function:
//
//	pairs, err := realize.Realize(kind, engine, content, styles)
//
// This takes:
//   - kind: The realization context (LayoutDocument, LayoutFragment, etc.)
//   - engine: The evaluation/layout engine
//   - content: The content tree from evaluation
//   - styles: The cascading style chain
//
// And returns realized pairs (elements with their associated styles).
//
// # RealizationKind
//
// Different kinds affect how content is processed:
//
//   - LayoutDocument: Full document layout preparation
//   - LayoutFragment: Fragment layout with block/inline detection
//   - LayoutPar: Paragraph-specific realization
//   - HtmlDocument: HTML export preparation
//   - HtmlFragment: HTML fragment export
//   - Math: Mathematical content realization
//
// # Show Rules
//
// Realization applies show rules to transform content:
//
//   - User-defined recipes (from Typst code)
//   - Built-in element transformations
//
// # Grouping
//
// Related elements are grouped for unified processing:
//
//   - Inline content → Paragraphs
//   - List items → Lists
//   - Citations → Bibliography handling
//
// # Space Collapsing
//
// The algorithm handles typesetting space rules:
//
//   - Removes spaces at content boundaries
//   - Collapses adjacent spaces
//   - Removes spaces adjacent to destructive elements (breaks, blocks)
package realize
