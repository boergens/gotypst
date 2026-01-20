// Package math provides layout for mathematical content.
//
// This package handles the spatial arrangement of math elements such as
// fractions, roots, subscripts, and superscripts. It produces frames that
// can be rendered by the PDF renderer.
//
// The math layout process:
//  1. Takes math content elements from eval (MathFracElement, etc.)
//  2. Recursively lays out nested content
//  3. Produces MathFrame structures with positioned items
//  4. Handles baseline alignment and spacing according to math typography rules
package math
