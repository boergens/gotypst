// Package pages provides document-level page layout for Typst.
//
// This package implements the pages/* module from typst-layout, handling:
// - Document layout into pages
// - Page breaking and parallel layout
// - Page run collection and processing
// - Final page assembly with margins and marginals
//
// The main entry point is LayoutDocument which takes content and styles
// and produces a PagedDocument.
package pages
