package pdf

import (
	"bytes"
	"fmt"
	"io"
)

// Version represents a PDF version.
type Version struct {
	Major int
	Minor int
}

var (
	// V1_4 is PDF version 1.4.
	V1_4 = Version{1, 4}
	// V1_5 is PDF version 1.5.
	V1_5 = Version{1, 5}
	// V1_6 is PDF version 1.6.
	V1_6 = Version{1, 6}
	// V1_7 is PDF version 1.7.
	V1_7 = Version{1, 7}
	// V2_0 is PDF version 2.0.
	V2_0 = Version{2, 0}
)

// Writer writes PDF documents.
type Writer struct {
	version Version
	buf     *bytes.Buffer
	objects []*IndirectObject
	offsets []int64
	nextNum int
	catalog Ref
	info    Ref
}

// NewWriter creates a new PDF writer.
func NewWriter(version Version) *Writer {
	return &Writer{
		version: version,
		buf:     new(bytes.Buffer),
		nextNum: 1, // Object numbers start at 1.
	}
}

// Alloc allocates a new object number and returns its reference.
func (w *Writer) Alloc() Ref {
	ref := NewRef(w.nextNum, 0)
	w.nextNum++
	return ref
}

// Write adds an object to the document.
// The object must have been allocated with Alloc().
func (w *Writer) Write(ref Ref, obj Object) {
	iobj := NewIndirectObject(ref, obj)
	w.objects = append(w.objects, iobj)
}

// SetCatalog sets the document catalog (root) reference.
func (w *Writer) SetCatalog(ref Ref) {
	w.catalog = ref
}

// SetInfo sets the document info dictionary reference.
func (w *Writer) SetInfo(ref Ref) {
	w.info = ref
}

// Finish writes the complete PDF document to the writer.
func (w *Writer) Finish(out io.Writer) error {
	// Reset buffer.
	w.buf.Reset()
	w.offsets = make([]int64, w.nextNum)

	// Write header.
	if err := w.writeHeader(); err != nil {
		return fmt.Errorf("writing header: %w", err)
	}

	// Write body (all objects).
	if err := w.writeBody(); err != nil {
		return fmt.Errorf("writing body: %w", err)
	}

	// Record xref position.
	xrefOffset := int64(w.buf.Len())

	// Write xref table.
	if err := w.writeXref(); err != nil {
		return fmt.Errorf("writing xref: %w", err)
	}

	// Write trailer.
	if err := w.writeTrailer(xrefOffset); err != nil {
		return fmt.Errorf("writing trailer: %w", err)
	}

	// Copy to output.
	_, err := w.buf.WriteTo(out)
	return err
}

func (w *Writer) writeHeader() error {
	// PDF header.
	_, err := fmt.Fprintf(w.buf, "%%PDF-%d.%d\n", w.version.Major, w.version.Minor)
	if err != nil {
		return err
	}

	// Binary marker (high-bit characters indicate binary content).
	_, err = w.buf.WriteString("%\x80\x81\x82\x83\n")
	return err
}

func (w *Writer) writeBody() error {
	for _, iobj := range w.objects {
		// Record offset for this object.
		w.offsets[iobj.ref.num] = int64(w.buf.Len())

		// Write the object.
		_, err := iobj.writeTo(w.buf)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) writeXref() error {
	_, err := fmt.Fprintf(w.buf, "xref\n0 %d\n", w.nextNum)
	if err != nil {
		return err
	}

	// Entry for object 0 (always free, points to next free).
	_, err = fmt.Fprintf(w.buf, "%010d %05d f \n", 0, 65535)
	if err != nil {
		return err
	}

	// Entries for all allocated objects.
	for i := 1; i < w.nextNum; i++ {
		offset := w.offsets[i]
		if offset > 0 {
			_, err = fmt.Fprintf(w.buf, "%010d %05d n \n", offset, 0)
		} else {
			// Object not written (shouldn't happen in well-formed docs).
			_, err = fmt.Fprintf(w.buf, "%010d %05d f \n", 0, 0)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) writeTrailer(xrefOffset int64) error {
	// Build trailer dictionary.
	trailer := make(Dict)
	trailer[Name("Size")] = Int(w.nextNum)
	if !w.catalog.IsZero() {
		trailer[Name("Root")] = w.catalog
	}
	if !w.info.IsZero() {
		trailer[Name("Info")] = w.info
	}

	// Write trailer.
	_, err := io.WriteString(w.buf, "trailer\n")
	if err != nil {
		return err
	}

	_, err = trailer.writeTo(w.buf)
	if err != nil {
		return err
	}

	// Write startxref and EOF.
	_, err = fmt.Fprintf(w.buf, "\nstartxref\n%d\n%%%%EOF\n", xrefOffset)
	return err
}
