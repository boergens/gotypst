package pdf

import (
	"bytes"
	"fmt"
	"io"

	"github.com/boergens/gotypst/layout/inline"
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
	// tagManager handles PDF/UA accessibility tagging.
	tagManager *TagManager
	// tagged indicates whether to generate tagged PDF output.
	tagged bool
	// renderer is used for rendering content and managing fonts.
	renderer *Renderer
	// fontRefs maps font resource names to their references.
	fontRefs map[string]Ref
}

// NewWriter creates a new PDF writer.
func NewWriter() *Writer {
	return &Writer{
		nextID:   1,
		images:   make(map[*pages.Image]Ref),
		renderer: NewRenderer(),
		fontRefs: make(map[string]Ref),
	}
}

// NewTaggedWriter creates a new PDF writer with accessibility tagging enabled.
func NewTaggedWriter() *Writer {
	w := NewWriter()
	w.tagged = true
	w.tagManager = NewTagManager()
	return w
}

// EnableTagging enables PDF/UA accessibility tagging.
func (w *Writer) EnableTagging() {
	w.tagged = true
	if w.tagManager == nil {
		w.tagManager = NewTagManager()
	}
}

// TagManager returns the tag manager for this writer.
func (w *Writer) TagManager() *TagManager {
	return w.tagManager
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

	// Reserve struct tree root ref if tagged
	var structTreeRootRef Ref
	if w.tagged && w.tagManager != nil {
		structTreeRootRef = w.allocRef()
	}

	// Process all pages and collect image XObjects
	var pageContentsRefs []Ref
	var pageImageRefs []map[string]Ref // per-page image resources

	for i, page := range doc.Pages {
		// Set current page in tag manager
		if w.tagManager != nil {
			w.tagManager.SetCurrentPage(i)
		}

		contentRef, imageRefs, err := w.processPage(&page, pagesRef)
		if err != nil {
			return err
		}
		pageContentsRefs = append(pageContentsRefs, contentRef)
		pageImageRefs = append(pageImageRefs, imageRefs)
	}

	// Generate font resources from the font manager
	fontResources := w.renderer.FontManager.GenerateResources(w.allocRef)
	for _, fontRes := range fontResources {
		// Add all font-related objects
		for _, obj := range fontRes.Objects {
			w.objects = append(w.objects, obj)
		}
		// Map font resource name to its reference
		w.fontRefs[fontRes.ResourceName] = fontRes.Ref
	}

	// Create page objects
	for i, page := range doc.Pages {
		pageRef := w.allocRef()
		w.pageRefs = append(w.pageRefs, pageRef)

		// Register page ref with tag manager
		if w.tagManager != nil {
			w.tagManager.SetPageRef(i, pageRef)
		}

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

		// Add font resources - use registered CID fonts if available, else fallback
		if len(w.fontRefs) > 0 {
			fonts := make(Dict)
			for name, ref := range w.fontRefs {
				fonts[Name(name)] = ref
			}
			resources[Name("Font")] = fonts
		} else {
			// Fallback to default Type1 font
			resources[Name("Font")] = Dict{
				Name("F1"): Dict{
					Name("Type"):     Name("Font"),
					Name("Subtype"):  Name("Type1"),
					Name("BaseFont"): Name("Helvetica"),
				},
			}
		}

		// Add image XObjects
		if len(pageImageRefs[i]) > 0 {
			xobjects := make(Dict)
			for name, ref := range pageImageRefs[i] {
				xobjects[Name(name)] = ref
			}
			resources[Name("XObject")] = xobjects
		}

		// Add resources to page
		pageDict[Name("Resources")] = resources

		// Add StructParents if tagged
		if w.tagged {
			pageDict[Name("StructParents")] = Int(i)
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

	// Build structure tree if tagged
	if w.tagged && w.tagManager != nil && w.tagManager.HasTags() {
		w.buildStructTree(structTreeRootRef)
	}

	// Create catalog
	catalogDict := Dict{
		Name("Type"):  Name("Catalog"),
		Name("Pages"): pagesRef,
	}

	// Add StructTreeRoot and MarkInfo if tagged
	if w.tagged && w.tagManager != nil && w.tagManager.HasTags() {
		catalogDict[Name("StructTreeRoot")] = structTreeRootRef
		catalogDict[Name("MarkInfo")] = Dict{
			Name("Marked"): Bool(true),
		}
	}

	w.addObjectWithRef(catalogRef, catalogDict)

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

// buildStructTree builds the PDF structure tree for accessibility.
func (w *Writer) buildStructTree(rootRef Ref) {
	if w.tagManager == nil {
		return
	}

	// Assign refs to all structure elements
	w.tagManager.AssignRefs(w.allocRef)

	// Build parent tree (maps MCID to struct elem)
	parentTree := w.tagManager.BuildParentTree()
	var parentTreeRef Ref
	if len(parentTree) > 0 {
		parentTreeRef = w.buildParentTree(parentTree)
	}

	// Build role map if there are custom roles
	customRoles := w.tagManager.CustomRoles()
	var roleMapRef *Ref
	if len(customRoles) > 0 {
		ref := w.buildRoleMap(customRoles)
		roleMapRef = &ref
	}

	// Create struct elem objects
	for _, elem := range w.tagManager.StructElements() {
		w.buildStructElem(elem)
	}

	// Create root struct tree dict
	rootDict := Dict{
		Name("Type"): Name("StructTreeRoot"),
		Name("K"):    w.tagManager.RootElement().Ref,
	}

	if parentTreeRef.ID != 0 {
		rootDict[Name("ParentTree")] = parentTreeRef
	}

	if roleMapRef != nil {
		rootDict[Name("RoleMap")] = *roleMapRef
	}

	w.addObjectWithRef(rootRef, rootDict)
}

// buildStructElem creates a PDF object for a structure element.
func (w *Writer) buildStructElem(elem *StructElem) {
	dict := Dict{
		Name("Type"): Name("StructElem"),
		Name("S"):    Name(string(elem.Role)),
	}

	// Add parent if not root
	if elem.Parent.ID != 0 {
		dict[Name("P")] = elem.Parent
	}

	// Add kids
	if len(elem.Kids) > 0 {
		kids := make(Array, len(elem.Kids))
		for i, kid := range elem.Kids {
			switch k := kid.(type) {
			case StructKidElem:
				kids[i] = k.Ref
			case StructKidMCID:
				kids[i] = Int(k.MCID)
			}
		}
		if len(kids) == 1 {
			dict[Name("K")] = kids[0]
		} else {
			dict[Name("K")] = kids
		}
	}

	// Add page reference if present
	if elem.PageRef.ID != 0 {
		dict[Name("Pg")] = elem.PageRef
	}

	// Add alt text if present
	if elem.AltText != "" {
		dict[Name("Alt")] = String(elem.AltText)
	}

	// Add actual text if present
	if elem.ActualText != "" {
		dict[Name("ActualText")] = String(elem.ActualText)
	}

	// Add language if present
	if elem.Lang != "" {
		dict[Name("Lang")] = String(elem.Lang)
	}

	w.addObjectWithRef(elem.Ref, dict)
}

// buildParentTree builds the parent tree number tree.
func (w *Writer) buildParentTree(parentTree map[int]Ref) Ref {
	// Build nums array (pairs of MCID and struct elem ref)
	nums := make(Array, 0, len(parentTree)*2)

	// Sort MCIDs for consistent output
	mcids := make([]int, 0, len(parentTree))
	for mcid := range parentTree {
		mcids = append(mcids, mcid)
	}
	// Simple insertion sort for small arrays
	for i := 1; i < len(mcids); i++ {
		for j := i; j > 0 && mcids[j-1] > mcids[j]; j-- {
			mcids[j-1], mcids[j] = mcids[j], mcids[j-1]
		}
	}

	for _, mcid := range mcids {
		nums = append(nums, Int(mcid), parentTree[mcid])
	}

	dict := Dict{
		Name("Nums"): nums,
	}

	return w.addObject(dict)
}

// buildRoleMap builds the role mapping dictionary.
func (w *Writer) buildRoleMap(customRoles map[string]StructRole) Ref {
	dict := make(Dict)
	for custom, standard := range customRoles {
		dict[Name(custom)] = Name(string(standard))
	}
	return w.addObject(dict)
}

// processPage processes a page frame and returns content stream ref and image refs.
func (w *Writer) processPage(page *pages.Page, pagesRef Ref) (Ref, map[string]Ref, error) {
	var content bytes.Buffer
	imageRefs := make(map[string]Ref)
	imageCounter := 0
	pageHeight := float64(page.Frame.Size.Height)

	// Apply coordinate system transform: flip Y axis so origin is top-left
	// Transform matrix: [1 0 0 -1 0 pageHeight] means:
	//   - scale Y by -1 (flip)
	//   - translate by pageHeight (move origin to top)
	// This converts from Typst's top-left origin to PDF's bottom-left origin
	fmt.Fprintf(&content, "q\n")                              // Save initial state
	fmt.Fprintf(&content, "1 0 0 -1 0 %g cm\n", pageHeight)   // Flip Y coordinate system

	// Process frame items using transform-based positioning
	err := w.processFrameWithTransforms(&page.Frame, &content, imageRefs, &imageCounter)
	if err != nil {
		return Ref{}, nil, err
	}

	fmt.Fprintf(&content, "Q\n") // Restore initial state

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

// processFrameWithTransforms recursively processes frame items using PDF transforms.
// This follows the Rust approach: use q/Q for state save/restore and cm for positioning.
// The page-level transform has already flipped Y, so we use Typst coordinates directly.
func (w *Writer) processFrameWithTransforms(frame *pages.Frame, content *bytes.Buffer, imageRefs map[string]Ref, imageCounter *int) error {
	for _, item := range frame.Items {
		x := float64(item.Pos.X)
		y := float64(item.Pos.Y)

		switch v := item.Item.(type) {
		case pages.GroupItem:
			// Save state, translate to item position, recurse, restore
			fmt.Fprintf(content, "q\n")                    // Save graphics state
			fmt.Fprintf(content, "1 0 0 1 %g %g cm\n", x, y) // Translate to position
			if err := w.processFrameWithTransforms(&v.Frame, content, imageRefs, imageCounter); err != nil {
				return err
			}
			fmt.Fprintf(content, "Q\n") // Restore graphics state

		case pages.InlineItem:
			// Render inline text content
			// TODO: Update renderer to use transform-based approach
			fmt.Fprintf(content, "q\n")
			fmt.Fprintf(content, "1 0 0 1 %g %g cm\n", x, y)
			if finalFrame, ok := v.Frame.(*inline.FinalFrame); ok {
				w.renderInlineFrameLocal(content, finalFrame)
			}
			fmt.Fprintf(content, "Q\n")

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

			// Images in PDF are 1x1 unit, scaled by transform
			// With Y already flipped, we translate to position and scale
			width := float64(v.Size.Width)
			height := float64(v.Size.Height)

			fmt.Fprintf(content, "q\n")                                   // Save graphics state
			fmt.Fprintf(content, "1 0 0 1 %g %g cm\n", x, y)              // Translate to position
			fmt.Fprintf(content, "%g 0 0 %g 0 0 cm\n", width, height)     // Scale to size
			fmt.Fprintf(content, "/%s Do\n", imgName)                     // Draw image
			fmt.Fprintf(content, "Q\n")                                   // Restore graphics state

		case pages.TagItem:
			// Tags don't produce PDF content

		case pages.TextItem:
			// Render text at local position
			// Since Y is already flipped at page level, we use coordinates directly
			// But text baseline needs adjustment: text is drawn from baseline up
			fontSize := float64(v.FontSize)

			fmt.Fprintf(content, "BT\n")                            // Begin text
			fmt.Fprintf(content, "/F1 %g Tf\n", fontSize)           // Set font and size
			// Position text: x is direct, y needs baseline offset (text draws upward from baseline)
			// In flipped coordinates, we add fontSize to move baseline down
			fmt.Fprintf(content, "%g %g Td\n", x, y+fontSize)       // Position at baseline
			fmt.Fprintf(content, "(%s) Tj\n", escapeString(v.Text)) // Show text
			fmt.Fprintf(content, "ET\n")                            // End text

		case pages.ShapedTextItem:
			// ShapedTextItem support - render using fallback for now
			if len(v.Glyphs) == 0 {
				continue
			}

			fontSize := float64(v.FontSize)

			fmt.Fprintf(content, "BT\n")
			fmt.Fprintf(content, "/F1 %g Tf\n", fontSize)
			fmt.Fprintf(content, "%g %g Td\n", x, y+fontSize)
			fmt.Fprintf(content, "( ) Tj\n") // Placeholder
			fmt.Fprintf(content, "ET\n")
		}
	}
	return nil
}

// renderInlineFrameLocal renders an inline frame at the current transform position.
// This handles all inline frame item types in the transformed coordinate system
// where Y is already flipped at the page level.
func (w *Writer) renderInlineFrameLocal(content *bytes.Buffer, frame *inline.FinalFrame) {
	baseline := float64(frame.Baseline)

	for _, entry := range frame.Items {
		x := float64(entry.Pos.X)
		y := float64(entry.Pos.Y)

		switch item := entry.Item.(type) {
		case inline.FinalTextItem:
			w.renderShapedTextLocal(content, item.Text, x, y, baseline)

		case inline.FinalMathScriptItem:
			w.renderMathScriptLocal(content, item, x, y, baseline)

		case inline.FinalMathLimitsItem:
			w.renderMathLimitsLocal(content, item, x, y, baseline)
		}
	}
}

// renderShapedTextLocal renders shaped text in transformed coordinates.
func (w *Writer) renderShapedTextLocal(content *bytes.Buffer, text *inline.ShapedText, x, y, baseline float64) {
	if text == nil || text.Glyphs.Len() == 0 {
		return
	}

	glyphs := text.Glyphs.Kept()
	if len(glyphs) == 0 {
		return
	}

	// Get font info from first glyph
	firstGlyph := &glyphs[0]
	fontSize := float64(firstGlyph.Size)

	// Register glyphs with font manager and get font name
	widthUnits := int(float64(firstGlyph.XAdvance) * 1000)
	fontName := w.renderer.FontManager.RegisterGlyph(firstGlyph.Font, firstGlyph.GlyphID, firstGlyph.Char, widthUnits)

	fmt.Fprintf(content, "BT\n")
	fmt.Fprintf(content, "/%s %g Tf\n", fontName, fontSize)
	// In transformed coords, y goes down. Text baseline is at y + baseline offset.
	fmt.Fprintf(content, "%g %g Td\n", x, y+baseline)

	// Build hex string with glyph IDs for CID font
	// For Identity-H encoding, we output glyph IDs as 2-byte big-endian values
	fmt.Fprintf(content, "<")
	for i := range glyphs {
		g := &glyphs[i]
		// Register each glyph
		widthUnits := int(float64(g.XAdvance) * 1000)
		w.renderer.FontManager.RegisterGlyph(g.Font, g.GlyphID, g.Char, widthUnits)
		// Output glyph ID as 2-byte hex
		fmt.Fprintf(content, "%04X", g.GlyphID)
	}
	fmt.Fprintf(content, "> Tj\n")
	fmt.Fprintf(content, "ET\n")
}

// renderMathScriptLocal renders math scripts (superscript/subscript) in transformed coordinates.
func (w *Writer) renderMathScriptLocal(content *bytes.Buffer, item inline.FinalMathScriptItem, x, y, baseline float64) {
	// Render base content
	if item.BaseFrame != nil {
		fmt.Fprintf(content, "q\n")
		fmt.Fprintf(content, "1 0 0 1 %g %g cm\n", x, y)
		w.renderInlineFrameLocal(content, item.BaseFrame)
		fmt.Fprintf(content, "Q\n")
	}

	// Calculate script X position
	scriptX := x + float64(item.ScriptXOffset)

	// Render superscript
	if item.SuperFrame != nil {
		superY := y + float64(item.SuperOffset)
		fmt.Fprintf(content, "q\n")
		fmt.Fprintf(content, "1 0 0 1 %g %g cm\n", scriptX, superY)
		w.renderInlineFrameLocal(content, item.SuperFrame)
		fmt.Fprintf(content, "Q\n")
	}

	// Render subscript
	if item.SubFrame != nil {
		subY := y + float64(item.SubOffset)
		fmt.Fprintf(content, "q\n")
		fmt.Fprintf(content, "1 0 0 1 %g %g cm\n", scriptX, subY)
		w.renderInlineFrameLocal(content, item.SubFrame)
		fmt.Fprintf(content, "Q\n")
	}
}

// renderMathLimitsLocal renders math limits (operator with limits above/below) in transformed coordinates.
func (w *Writer) renderMathLimitsLocal(content *bytes.Buffer, item inline.FinalMathLimitsItem, x, y, baseline float64) {
	centerX := x + float64(item.CenterX)

	// Render upper limit (centered above nucleus)
	if item.UpperFrame != nil {
		upperX := centerX - float64(item.UpperFrame.Size.Width)/2
		upperY := y + float64(item.UpperOffset)
		fmt.Fprintf(content, "q\n")
		fmt.Fprintf(content, "1 0 0 1 %g %g cm\n", upperX, upperY)
		w.renderInlineFrameLocal(content, item.UpperFrame)
		fmt.Fprintf(content, "Q\n")
	}

	// Render nucleus (main operator)
	if item.NucleusFrame != nil {
		nucleusX := centerX - float64(item.NucleusFrame.Size.Width)/2
		fmt.Fprintf(content, "q\n")
		fmt.Fprintf(content, "1 0 0 1 %g %g cm\n", nucleusX, y)
		w.renderInlineFrameLocal(content, item.NucleusFrame)
		fmt.Fprintf(content, "Q\n")
	}

	// Render lower limit (centered below nucleus)
	if item.LowerFrame != nil {
		lowerX := centerX - float64(item.LowerFrame.Size.Width)/2
		lowerY := y + float64(item.LowerOffset)
		fmt.Fprintf(content, "q\n")
		fmt.Fprintf(content, "1 0 0 1 %g %g cm\n", lowerX, lowerY)
		w.renderInlineFrameLocal(content, item.LowerFrame)
		fmt.Fprintf(content, "Q\n")
	}
}

// escapeString escapes special characters for PDF string literals.
func escapeString(s string) string {
	var result bytes.Buffer
	for _, r := range s {
		switch r {
		case '(', ')', '\\':
			result.WriteByte('\\')
			result.WriteRune(r)
		default:
			result.WriteRune(r)
		}
	}
	return result.String()
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

// ExportTagged exports a PagedDocument to PDF with accessibility tagging enabled.
func ExportTagged(doc *pages.PagedDocument, out io.Writer) error {
	w := NewTaggedWriter()
	return w.Write(doc, out)
}
