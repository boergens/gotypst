package pdf

import "io"

// Document provides a high-level API for building PDF documents.
type Document struct {
	writer   *Writer
	catalog  *Catalog
	info     *Info
	pageTree *PageTree
}

// NewDocument creates a new PDF document.
func NewDocument(version Version) *Document {
	w := NewWriter(version)
	return &Document{
		writer:   w,
		catalog:  NewCatalog(w),
		info:     NewInfo(w),
		pageTree: NewPageTree(w),
	}
}

// Writer returns the underlying PDF writer for advanced operations.
func (d *Document) Writer() *Writer {
	return d.writer
}

// Catalog returns the document catalog for configuration.
func (d *Document) Catalog() *Catalog {
	return d.catalog
}

// Info returns the document info for metadata.
func (d *Document) Info() *Info {
	return d.info
}

// PageTree returns the page tree for adding pages.
func (d *Document) PageTree() *PageTree {
	return d.pageTree
}

// AddPage adds a new page with the specified dimensions (in points).
func (d *Document) AddPage(width, height float64) *PageBuilder {
	return d.pageTree.AddPage(width, height)
}

// AddContentStream creates a content stream and returns its reference.
func (d *Document) AddContentStream(data []byte) Ref {
	ref := d.writer.Alloc()
	stream := NewStream(data)
	d.writer.Write(ref, stream)
	return ref
}

// Finish writes the complete PDF document to the writer.
func (d *Document) Finish(w io.Writer) error {
	// Finalize page tree.
	d.pageTree.Finish()

	// Set catalog pages reference.
	d.catalog.SetPages(d.pageTree.Root())

	// Finalize catalog and info.
	d.catalog.Finish()
	d.info.Finish()

	// Write the PDF.
	return d.writer.Finish(w)
}

// Standard page sizes in points (1 point = 1/72 inch).
const (
	// A4 page dimensions.
	A4Width  = 595.276
	A4Height = 841.89

	// US Letter dimensions.
	LetterWidth  = 612.0
	LetterHeight = 792.0

	// US Legal dimensions.
	LegalWidth  = 612.0
	LegalHeight = 1008.0

	// A3 page dimensions.
	A3Width  = 841.89
	A3Height = 1190.55

	// A5 page dimensions.
	A5Width  = 419.528
	A5Height = 595.276
)
