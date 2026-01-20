// Package text provides text-related functions for the Typst standard library.
//
// This package includes:
//   - raw: Creates raw text/code elements with optional syntax highlighting
//   - highlight: Syntax highlighting hooks for code blocks
//
// The raw function creates content that displays text verbatim, typically in
// a monospace font. It supports an optional language parameter for syntax
// highlighting.
//
// The highlighting system uses a hook-based design allowing custom syntax
// highlighters to be registered. Built-in highlighters provide basic keyword
// highlighting for common languages like Go, Python, JavaScript, and Rust.
package text
