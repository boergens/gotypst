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
	// fontRefs maps font names to their PDF references.
	fontRefs map[string]Ref
	// glyphCollector tracks glyph usage for font subsetting.
	glyphCollector *GlyphCollector
}

// NewWriter creates a new PDF writer.
func NewWriter() *Writer {
	return &Writer{
		nextID:   1,
		images:   make(map[*pages.Image]Ref),
		fontRefs: make(map[string]Ref),
	}
}

// SetGlyphCollector sets the glyph collector for font subsetting.
func (w *Writer) SetGlyphCollector(gc *GlyphCollector) {
	w.glyphCollector = gc
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

	// Embed fonts from glyph collector
	if err := w.embedFonts(); err != nil {
		return fmt.Errorf("embed fonts: %w", err)
	}

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

		// Add font resources
		if len(w.fontRefs) > 0 {
			fonts := make(Dict)
			for name, ref := range w.fontRefs {
				fonts[Name(name)] = ref
			}
			resources[Name("Font")] = fonts
		}

		// Add image resources
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

// embedFonts creates PDF font objects for all collected fonts.
func (w *Writer) embedFonts() error {
	if w.glyphCollector == nil {
		return nil
	}

	fonts := w.glyphCollector.Fonts()
	if len(fonts) == 0 {
		return nil
	}

	for _, gs := range fonts {
		if gs.Font == nil || len(gs.Font.Data) == 0 {
			// Skip fonts without data - they can't be embedded
			continue
		}

		// Create subset font
		fontXObj, err := SubsetFont(gs)
		if err != nil {
			// If subsetting fails, skip this font
			// In production, might want to fall back to full embedding
			continue
		}

		// Create font file stream (the embedded font data)
		fontFileRef := w.allocRef()
		fontFileStream := Stream{
			Dict: Dict{
				Name("Filter"): Name("FlateDecode"),
				Name("Length1"): Int(len(fontXObj.SubsetData)), // Original length hint
			},
			Data: fontXObj.SubsetData,
		}
		w.addObjectWithRef(fontFileRef, fontFileStream)

		// Create CIDSystemInfo dictionary
		cidSystemInfo := Dict{
			Name("Registry"):   String("Adobe"),
			Name("Ordering"):   String("Identity"),
			Name("Supplement"): Int(0),
		}

		// Create font descriptor
		fontDescRef := w.allocRef()
		fontDesc := Dict{
			Name("Type"):        Name("FontDescriptor"),
			Name("FontName"):    Name(fontXObj.BaseFont),
			Name("Flags"):       Int(32), // Symbolic font
			Name("FontBBox"):    Array{Int(0), Int(-200), Int(1000), Int(1000)},
			Name("ItalicAngle"): Int(0),
			Name("Ascent"):      Int(800),
			Name("Descent"):     Int(-200),
			Name("CapHeight"):   Int(700),
			Name("StemV"):       Int(80),
			Name("FontFile2"):   fontFileRef, // TrueType font
		}
		w.addObjectWithRef(fontDescRef, fontDesc)

		// Create CIDFont dictionary
		cidFontRef := w.allocRef()

		// Build widths array for CID font
		widthsArray := buildCIDWidthsArray(fontXObj.Widths)

		cidFont := Dict{
			Name("Type"):           Name("Font"),
			Name("Subtype"):        Name("CIDFontType2"),
			Name("BaseFont"):       Name(fontXObj.BaseFont),
			Name("CIDSystemInfo"):  cidSystemInfo,
			Name("FontDescriptor"): fontDescRef,
			Name("W"):              widthsArray,
			Name("CIDToGIDMap"):    Name("Identity"),
		}
		w.addObjectWithRef(cidFontRef, cidFont)

		// Create ToUnicode CMap for text extraction
		toUnicodeRef := w.allocRef()
		toUnicodeData := buildToUnicodeCMap(gs)
		toUnicodeStream := Stream{
			Dict: make(Dict),
			Data: []byte(toUnicodeData),
		}
		if err := toUnicodeStream.Compress(); err != nil {
			return err
		}
		w.addObjectWithRef(toUnicodeRef, toUnicodeStream)

		// Create Type0 composite font
		fontRef := w.allocRef()
		font := Dict{
			Name("Type"):            Name("Font"),
			Name("Subtype"):         Name("Type0"),
			Name("BaseFont"):        Name(fontXObj.BaseFont),
			Name("Encoding"):        Name("Identity-H"),
			Name("DescendantFonts"): Array{cidFontRef},
			Name("ToUnicode"):       toUnicodeRef,
		}
		w.addObjectWithRef(fontRef, font)

		// Store font reference
		w.fontRefs[gs.Name] = fontRef
	}

	return nil
}

// buildCIDWidthsArray builds the /W array for CID font widths.
// Format: [cid [w1 w2 w3...]] for consecutive glyphs starting at cid
func buildCIDWidthsArray(widths []int) Array {
	if len(widths) == 0 {
		return Array{}
	}

	// Simple approach: list all widths starting from CID 0
	widthArray := make(Array, len(widths))
	for i, w := range widths {
		widthArray[i] = Int(w)
	}

	return Array{Int(0), widthArray}
}

// buildToUnicodeCMap creates a ToUnicode CMap for text extraction.
func buildToUnicodeCMap(gs *GlyphSet) string {
	var b bytes.Buffer

	b.WriteString("/CIDInit /ProcSet findresource begin\n")
	b.WriteString("12 dict begin\n")
	b.WriteString("begincmap\n")
	b.WriteString("/CIDSystemInfo <<\n")
	b.WriteString("/Registry (Adobe)\n")
	b.WriteString("/Ordering (UCS)\n")
	b.WriteString("/Supplement 0\n")
	b.WriteString(">> def\n")
	b.WriteString("/CMapName /Adobe-Identity-UCS def\n")
	b.WriteString("/CMapType 2 def\n")
	b.WriteString("1 begincodespacerange\n")
	b.WriteString("<0000> <FFFF>\n")
	b.WriteString("endcodespacerange\n")

	// Build character mappings
	glyphIDs := gs.SortedGlyphIDs()
	if len(glyphIDs) > 0 {
		// Write mappings in batches of 100 (PDF limit)
		for i := 0; i < len(glyphIDs); i += 100 {
			end := i + 100
			if end > len(glyphIDs) {
				end = len(glyphIDs)
			}
			batch := glyphIDs[i:end]

			fmt.Fprintf(&b, "%d beginbfchar\n", len(batch))
			for _, gid := range batch {
				char := gs.Glyphs[gid]
				fmt.Fprintf(&b, "<%04X> <%04X>\n", gid, char)
			}
			b.WriteString("endbfchar\n")
		}
	}

	b.WriteString("endcmap\n")
	b.WriteString("CMapName currentdict /CMap defineresource pop\n")
	b.WriteString("end\n")
	b.WriteString("end\n")

	return b.String()
}

// Export is a convenience function that exports a PagedDocument to PDF.
func Export(doc *pages.PagedDocument, out io.Writer) error {
	w := NewWriter()
	return w.Write(doc, out)
}
