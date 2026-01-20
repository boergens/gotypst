// Package text provides the text element and styling for Typst documents.
//
// The text element is the fundamental building block for text content in Typst.
// It controls how text is rendered, including font selection, size, weight,
// style, fill, stroke, and decorations (underline, strikethrough, overline).
//
// This package also includes:
//   - raw: Creates raw text/code elements with optional syntax highlighting
//   - highlight: Syntax highlighting hooks for code blocks
//
// # Basic Usage
//
// Create a text element with styling:
//
//	t := text.New("Hello, World!").
//		WithFont("New Computer Modern").
//		WithSize(text.SizeFromPt(12)).
//		WithWeight(text.FontWeightBold).
//		WithFill(text.Blue)
//
// # Text Decorations
//
// Apply decorations to text:
//
//	t := text.New("Important").
//		WithUnderline(text.NewUnderline().WithEvade(true)).
//		WithFill(text.Red)
//
// # Font Properties
//
// The text element supports full font configuration:
//
//	t := text.New("Styled").
//		WithFont("Libertinus Serif", "New Computer Modern").
//		WithWeight(text.FontWeightSemiBold).
//		WithStyle(text.FontStyleItalic).
//		WithStretch(text.FontStretchNormal)
//
// # Fill and Stroke
//
// Text can have both fill and stroke:
//
//	t := text.New("Outlined").
//		WithFill(text.White).
//		WithStroke(text.NewStroke(text.Black, 0.5))
//
// # Raw Text and Code
//
// The raw function creates content that displays text verbatim, typically in
// a monospace font. It supports an optional language parameter for syntax
// highlighting.
//
// The highlighting system uses a hook-based design allowing custom syntax
// highlighters to be registered. Built-in highlighters provide basic keyword
// highlighting for common languages like Go, Python, JavaScript, and Rust.
//
// # Typst Compatibility
//
// This package implements the text element as defined in Typst's text module.
// The properties map directly to Typst's text function parameters:
//
//   - font: Font families for text display
//   - size: Text size in points
//   - weight: Font weight (100-900)
//   - style: normal, italic, or oblique
//   - stretch: normal, condensed, or expanded
//   - fill: Fill paint (color, gradient, or pattern)
//   - stroke: Optional stroke paint
//   - underline: Underline decoration
//   - strikethrough: Strikethrough decoration
//   - overline: Overline decoration
package text
