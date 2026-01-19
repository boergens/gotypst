// Package gotypst provides a Go implementation of the Typst typesetting system.
//
// Typst is a modern typesetting system designed for creating documents
// with a clean syntax and powerful features. This package provides the
// core interfaces and types for embedding Typst compilation in Go applications.
//
// To use this package, implement the World interface to provide access
// to the file system, packages, and other resources needed for compilation.
package gotypst

// World provides access to the external environment during compilation.
//
// Implementations of this interface define how Typst interacts with
// the file system, retrieves packages, and accesses fonts.
type World interface {
	// Library returns the standard library.
	Library() Library

	// MainFile returns the main source file for compilation.
	MainFile() FileID

	// Source returns the source content for a file.
	Source(id FileID) (Source, error)

	// File returns the raw bytes of a file.
	File(id FileID) ([]byte, error)

	// Font returns font data by index.
	Font(index int) (Font, error)

	// Today returns the current date.
	Today(offset *int) Date
}

// FileID uniquely identifies a file in the World.
type FileID struct {
	Package *PackageSpec
	Path    string
}

// PackageSpec identifies a package by namespace, name, and version.
type PackageSpec struct {
	Namespace string
	Name      string
	Version   Version
}

// Version represents a semantic version.
type Version struct {
	Major int
	Minor int
	Patch int
}

// Source represents parsed source content.
type Source struct {
	ID   FileID
	Text string
}

// Library represents the standard library.
type Library interface{}

// Font represents font data.
type Font interface{}

// Date represents a date value.
type Date struct {
	Year  int
	Month int
	Day   int
}

// Document represents a compiled Typst document.
type Document struct {
	Pages []Page
	Title *string
}

// Page represents a single page in a document.
type Page struct {
	Frame  Frame
	Fill   Color
	Number int
}

// Frame represents a laid-out frame of content.
type Frame struct {
	Width  Length
	Height Length
	Items  []FrameItem
}

// FrameItem represents an item within a frame.
type FrameItem interface {
	frameItem()
}

// Length represents a physical length.
type Length float64

// Color represents a color value.
type Color struct {
	R, G, B, A uint8
}
