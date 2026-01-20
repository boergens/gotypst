package pdf

import (
	"bytes"
	"fmt"
	"io"

	"github.com/boergens/gotypst/layout"
	"github.com/boergens/gotypst/layout/pages"
)

// Writer handles PDF document generation.
type Writer struct {
	// objects stores all indirect objects to be written.
	objects []IndirectObject
	// nextID is the next available object ID.
	nextID int
	// images maps image data pointers to their XObject references.
	images map[*pages.Image]Ref
	// pageRefs stores references to page objects.
	pageRefs []Ref
	// fontCollector tracks fonts for embedding.
	fontCollector *FontCollector
	// fontResources is the dict of font refs after emission.
	fontResources Dict
}

// NewWriter creates a new PDF writer.
func NewWriter() *Writer {
	return &Writer{
		nextID:        1,
		images:        make(map[*pages.Image]Ref),
		fontCollector: NewFontCollector(),
	}
}

// FontCollector returns the font collector for registering fonts.
func (w *Writer) FontCollector() *FontCollector {
	return w.fontCollector
}

// allocRef allocates a new object reference.
func (w *Writer) allocRef() Ref {
	ref := Ref{ID: w.nextID, Gen: 0}
	w.nextID++
	return ref
}

// addObject adds an indirect object and returns its reference.
func (w *Writer) addObject(obj Object) Ref {
	ref := w.allocRef()
	w.objects = append(w.objects, IndirectObject{Ref: ref, Object: obj})
	return ref
}

// addObjectWithRef adds an indirect object with a pre-allocated reference.
func (w *Writer) addObjectWithRef(ref Ref, obj Object) {
	w.objects = append(w.objects, IndirectObject{Ref: ref, Object: obj})
}

// Write generates a PDF from a PagedDocument and writes it to w.
func (w *Writer) Write(doc *pages.PagedDocument, out io.Writer) error {
	// Reserve object IDs for catalog and page tree
	catalogRef := w.allocRef()
	pagesRef := w.allocRef()

	// Emit font objects first
	fontEmitter := NewFontEmitter(w)
	fontResources, err := fontEmitter.EmitFonts(w.fontCollector)
	if err != nil {
		return fmt.Errorf("emit fonts: %w", err)
	}
	w.fontResources = fontResources

	// Process all pages and collect image XObjects
	var pageContentsRefs []Ref
	var pageImageRefs []map[string]Ref // per-page image resources

	for _, page := range doc.Pages {
		contentRef, imageRefs, err := w.processPage(&page, pagesRef)
		if err != nil {
			return err
		}
		pageContentsRefs = append(pageContentsRefs, contentRef)
		pageImageRefs = append(pageImageRefs, imageRefs)
	}

	// Create page objects
	for i, page := range doc.Pages {
		pageRef := w.allocRef()
		w.pageRefs = append(w.pageRefs, pageRef)

		pageDict := Dict{
			Name("Type"):     Name("Page"),
			Name("Parent"):   pagesRef,
			Name("Contents"): pageContentsRefs[i],
			Name("MediaBox"): Array{
				Int(0), Int(0),
				Real(page.Frame.Size.Width),
				Real(page.Frame.Size.Height),
			},
		}

		// Build resources dictionary
		resources := make(Dict)

		// Add fonts
		if len(w.fontResources) > 0 {
			resources[Name("Font")] = w.fontResources
		}

		// Add images
		if len(pageImageRefs[i]) > 0 {
			xobjects := make(Dict)
			for name, ref := range pageImageRefs[i] {
				xobjects[Name(name)] = ref
			}
			resources[Name("XObject")] = xobjects
		}

		if len(resources) > 0 {
			pageDict[Name("Resources")] = resources
		}

		w.addObjectWithRef(pageRef, pageDict)
	}

	// Create page tree
	kids := make(Array, len(w.pageRefs))
	for i, ref := range w.pageRefs {
		kids[i] = ref
	}

	w.addObjectWithRef(pagesRef, Dict{
		Name("Type"):  Name("Pages"),
		Name("Kids"):  kids,
		Name("Count"): Int(len(w.pageRefs)),
	})

	// Create catalog
	w.addObjectWithRef(catalogRef, Dict{
		Name("Type"):  Name("Catalog"),
		Name("Pages"): pagesRef,
	})

	// Add document info if present
	var infoRef *Ref
	if doc.Info.Title != nil || len(doc.Info.Author) > 0 {
		info := make(Dict)
		if doc.Info.Title != nil {
			info[Name("Title")] = String(*doc.Info.Title)
		}
		if len(doc.Info.Author) > 0 {
			info[Name("Author")] = String(doc.Info.Author[0])
		}
		ref := w.addObject(info)
		infoRef = &ref
	}

	// Write PDF
	return w.writePDF(out, catalogRef, infoRef)
}

// processPage processes a page frame and returns content stream ref and image refs.
func (w *Writer) processPage(page *pages.Page, pagesRef Ref) (Ref, map[string]Ref, error) {
	var content bytes.Buffer
	imageRefs := make(map[string]Ref)
	imageCounter := 0

	// Process frame items
	err := w.processFrame(&page.Frame, &content, imageRefs, &imageCounter, 0, 0)
	if err != nil {
		return Ref{}, nil, err
	}

	// Create content stream
	stream := Stream{
		Dict: make(Dict),
		Data: content.Bytes(),
	}
	if err := stream.Compress(); err != nil {
		return Ref{}, nil, err
	}

	contentRef := w.addObject(stream)
	return contentRef, imageRefs, nil
}

// processFrame recursively processes frame items and generates PDF operators.
func (w *Writer) processFrame(frame *pages.Frame, content *bytes.Buffer, imageRefs map[string]Ref, imageCounter *int, offsetX, offsetY layout.Abs) error {
	for _, item := range frame.Items {
		x := offsetX + item.Pos.X
		y := offsetY + item.Pos.Y

		switch v := item.Item.(type) {
		case pages.GroupItem:
			// Recurse into nested frame
			if err := w.processFrame(&v.Frame, content, imageRefs, imageCounter, x, y); err != nil {
				return err
			}

		case pages.ImageItem:
			// Get or create image XObject
			imgRef, err := w.getOrCreateImageXObject(&v.Image)
			if err != nil {
				return err
			}

			// Register image for this page's resources
			imgName := fmt.Sprintf("Im%d", *imageCounter)
			*imageCounter++
			imageRefs[imgName] = imgRef

			// Generate image placement operator
			// PDF images are placed with a transformation matrix
			// cm operator: a b c d e f -> [a b c d e f] transformation matrix
			// For images: [width 0 0 height x y] cm
			// Y coordinate needs to be flipped for PDF coordinate system
			width := float64(v.Size.Width)
			height := float64(v.Size.Height)
			xPos := float64(x)
			// PDF origin is bottom-left, Typst origin is top-left
			// We need to account for this in the page rendering
			yPos := float64(y)

			fmt.Fprintf(content, "q\n")                                             // Save graphics state
			fmt.Fprintf(content, "%g 0 0 %g %g %g cm\n", width, height, xPos, yPos) // Transform matrix
			fmt.Fprintf(content, "/%s Do\n", imgName)                               // Draw image
			fmt.Fprintf(content, "Q\n")                                             // Restore graphics state

		case pages.TagItem:
			// Tags don't produce PDF content
		}
	}
	return nil
}

// getOrCreateImageXObject returns a reference to an image XObject,
// creating it if it doesn't exist.
func (w *Writer) getOrCreateImageXObject(img *pages.Image) (Ref, error) {
	// Check cache
	if ref, ok := w.images[img]; ok {
		return ref, nil
	}

	// Allocate ref(s) for image and potential SMask
	imgRef := w.allocRef()

	// Encode image
	xobj, err := EncodeImage(img, imgRef)
	if err != nil {
		return Ref{}, err
	}

	// Add SMask if present
	if xobj.SMask != nil {
		xobj.SMask.Ref = w.allocRef()
		smaskObj := xobj.SMask.ToIndirectObject()
		w.objects = append(w.objects, smaskObj)
	}

	// Add main image
	imgObj := xobj.ToIndirectObject()
	w.objects = append(w.objects, imgObj)

	// Cache the reference
	w.images[img] = imgRef

	return imgRef, nil
}

// writePDF writes the complete PDF document.
func (w *Writer) writePDF(out io.Writer, catalogRef Ref, infoRef *Ref) error {
	var buf bytes.Buffer

	// Header
	buf.WriteString("%PDF-1.7\n")
	// Binary comment to indicate binary content
	buf.WriteString("%\x80\x80\x80\x80\n")

	// Track object byte offsets for xref table
	offsets := make(map[int]int64)

	// Write all objects
	for _, obj := range w.objects {
		offsets[obj.Ref.ID] = int64(buf.Len())
		if err := obj.writeTo(&buf); err != nil {
			return err
		}
		buf.WriteByte('\n')
	}

	// Write xref table
	xrefOffset := buf.Len()
	fmt.Fprintf(&buf, "xref\n")
	fmt.Fprintf(&buf, "0 %d\n", w.nextID)
	fmt.Fprintf(&buf, "0000000000 65535 f \n") // Free entry

	for i := 1; i < w.nextID; i++ {
		if offset, ok := offsets[i]; ok {
			fmt.Fprintf(&buf, "%010d 00000 n \n", offset)
		} else {
			// Object ID was allocated but not used
			fmt.Fprintf(&buf, "0000000000 65535 f \n")
		}
	}

	// Write trailer
	trailer := Dict{
		Name("Size"): Int(w.nextID),
		Name("Root"): catalogRef,
	}
	if infoRef != nil {
		trailer[Name("Info")] = *infoRef
	}

	buf.WriteString("trailer\n")
	if err := trailer.writeTo(&buf); err != nil {
		return err
	}
	fmt.Fprintf(&buf, "\nstartxref\n%d\n%%%%EOF\n", xrefOffset)

	// Write to output
	_, err := out.Write(buf.Bytes())
	return err
}

// Export is a convenience function that exports a PagedDocument to PDF.
func Export(doc *pages.PagedDocument, out io.Writer) error {
	w := NewWriter()
	return w.Write(doc, out)
}
