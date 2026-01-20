package pdf

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/go-text/typesetting/font"
)

// FontCollector tracks fonts used during rendering and manages PDF font embedding.
type FontCollector struct {
	// fonts maps font face pointers to their embedding data.
	fonts map[*font.Face]*EmbeddedFont
	// fontOrder maintains insertion order for deterministic output.
	fontOrder []*font.Face
	// fontPathMap stores file paths for font faces.
	fontPathMap map[*font.Face]string
	// nextID is the next font resource ID (F1, F2, etc.).
	nextID int
}

// EmbeddedFont contains all data needed to embed a font in PDF.
type EmbeddedFont struct {
	// Face is the original font face.
	Face *font.Face
	// ResourceName is the PDF resource name (e.g., "F1").
	ResourceName string
	// UsedGlyphs tracks which glyph IDs have been used.
	UsedGlyphs map[uint16]rune
	// FontRef is the PDF object reference for the Font dictionary.
	FontRef Ref
	// Path is the filesystem path to the font file (for embedding).
	Path string
}

// NewFontCollector creates a new font collector.
func NewFontCollector() *FontCollector {
	return &FontCollector{
		fonts:       make(map[*font.Face]*EmbeddedFont),
		fontPathMap: make(map[*font.Face]string),
		nextID:      1,
	}
}

// RegisterFont registers a font face and returns its PDF resource name.
// If the font is already registered, returns the existing name.
func (fc *FontCollector) RegisterFont(face *font.Face, path string) string {
	if ef, ok := fc.fonts[face]; ok {
		return ef.ResourceName
	}

	name := fmt.Sprintf("F%d", fc.nextID)
	fc.nextID++

	ef := &EmbeddedFont{
		Face:         face,
		ResourceName: name,
		UsedGlyphs:   make(map[uint16]rune),
		Path:         path,
	}
	fc.fonts[face] = ef
	fc.fontOrder = append(fc.fontOrder, face)

	return name
}

// RecordGlyph records that a glyph was used with this font.
func (fc *FontCollector) RecordGlyph(face interface{}, glyphID uint16, char rune) {
	if f, ok := face.(*font.Face); ok {
		if ef, exists := fc.fonts[f]; exists {
			if _, used := ef.UsedGlyphs[glyphID]; !used {
				ef.UsedGlyphs[glyphID] = char
			}
		}
	}
}

// FontName returns the PDF resource name for a font face.
// Auto-registers the font if not already registered.
func (fc *FontCollector) FontName(face interface{}) string {
	if f, ok := face.(*font.Face); ok {
		if ef, exists := fc.fonts[f]; exists {
			return "/" + ef.ResourceName
		}
		// Auto-register the font
		name := fc.RegisterFont(f, fc.fontPathMap[f])
		return "/" + name
	}
	return "/F1" // Fallback for unrecognized face types
}

// SetFontPath associates a font face with its file path.
// Call this before rendering to enable font embedding.
func (fc *FontCollector) SetFontPath(face *font.Face, path string) {
	if fc.fontPathMap == nil {
		fc.fontPathMap = make(map[*font.Face]string)
	}
	fc.fontPathMap[face] = path
}

// Fonts returns all registered fonts in order.
func (fc *FontCollector) Fonts() []*EmbeddedFont {
	result := make([]*EmbeddedFont, 0, len(fc.fontOrder))
	for _, face := range fc.fontOrder {
		result = append(result, fc.fonts[face])
	}
	return result
}

// FontEmitter handles writing font objects to PDF.
type FontEmitter struct {
	writer *Writer
}

// NewFontEmitter creates a new font emitter.
func NewFontEmitter(w *Writer) *FontEmitter {
	return &FontEmitter{writer: w}
}

// EmitFonts writes all font objects and returns a dict of font resources.
func (fe *FontEmitter) EmitFonts(collector *FontCollector) (Dict, error) {
	fontResources := make(Dict)

	for _, ef := range collector.Fonts() {
		fontRef, err := fe.emitFont(ef)
		if err != nil {
			return nil, fmt.Errorf("emit font %s: %w", ef.ResourceName, err)
		}
		ef.FontRef = fontRef
		fontResources[Name(ef.ResourceName)] = fontRef
	}

	return fontResources, nil
}

// emitFont writes a single font and all related objects.
func (fe *FontEmitter) emitFont(ef *EmbeddedFont) (Ref, error) {
	if ef.Face == nil || ef.Face.Font == nil {
		return fe.emitType1Fallback(ef)
	}

	// Determine if this is a TrueType or CFF font
	fontData, err := fe.loadFontData(ef.Path)
	if err != nil {
		// Fall back to Type1 standard font if we can't load
		return fe.emitType1Fallback(ef)
	}

	// Check font type by examining table tags
	isCFF := fe.hasCFFOutlines(fontData)

	if isCFF {
		return fe.emitCFFFont(ef, fontData)
	}
	return fe.emitTrueTypeFont(ef, fontData)
}

// loadFontData loads the raw font file data.
func (fe *FontEmitter) loadFontData(path string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("no font path available")
	}
	return os.ReadFile(path)
}

// hasCFFOutlines checks if font has CFF outlines (OpenType with CFF).
func (fe *FontEmitter) hasCFFOutlines(data []byte) bool {
	// Check for 'CFF ' table tag at typical offset positions
	// SFNT fonts have table directory starting at offset 12
	if len(data) < 16 {
		return false
	}

	// Read number of tables
	numTables := int(data[4])<<8 | int(data[5])
	if len(data) < 12+numTables*16 {
		return false
	}

	// Search for CFF table
	for i := 0; i < numTables; i++ {
		offset := 12 + i*16
		tag := string(data[offset : offset+4])
		if tag == "CFF " {
			return true
		}
	}
	return false
}

// emitTrueTypeFont embeds a TrueType font.
func (fe *FontEmitter) emitTrueTypeFont(ef *EmbeddedFont, fontData []byte) (Ref, error) {
	// Create ToUnicode CMap
	tounicodeRef, err := fe.emitToUnicodeCMap(ef)
	if err != nil {
		return Ref{}, err
	}

	// Create FontDescriptor
	descriptorRef, err := fe.emitFontDescriptor(ef, fontData, false)
	if err != nil {
		return Ref{}, err
	}

	// Create Font dictionary
	fontDict := Dict{
		Name("Type"):           Name("Font"),
		Name("Subtype"):        Name("TrueType"),
		Name("BaseFont"):       Name(fe.getBaseFont(ef)),
		Name("FirstChar"):      Int(0),
		Name("LastChar"):       Int(255),
		Name("Widths"):         fe.getWidthsArray(ef),
		Name("FontDescriptor"): descriptorRef,
		Name("ToUnicode"):      tounicodeRef,
		Name("Encoding"):       Name("WinAnsiEncoding"),
	}

	return fe.writer.addObject(fontDict), nil
}

// emitCFFFont embeds a CFF (OpenType) font.
func (fe *FontEmitter) emitCFFFont(ef *EmbeddedFont, fontData []byte) (Ref, error) {
	// Create ToUnicode CMap
	tounicodeRef, err := fe.emitToUnicodeCMap(ef)
	if err != nil {
		return Ref{}, err
	}

	// Create FontDescriptor with CFF font file
	descriptorRef, err := fe.emitFontDescriptor(ef, fontData, true)
	if err != nil {
		return Ref{}, err
	}

	// Create Font dictionary (Type0 for CFF)
	// For simplicity, we use a simple Type1 wrapper
	fontDict := Dict{
		Name("Type"):           Name("Font"),
		Name("Subtype"):        Name("Type1"),
		Name("BaseFont"):       Name(fe.getBaseFont(ef)),
		Name("FirstChar"):      Int(0),
		Name("LastChar"):       Int(255),
		Name("Widths"):         fe.getWidthsArray(ef),
		Name("FontDescriptor"): descriptorRef,
		Name("ToUnicode"):      tounicodeRef,
	}

	return fe.writer.addObject(fontDict), nil
}

// emitType1Fallback emits a Type1 standard font as fallback.
func (fe *FontEmitter) emitType1Fallback(ef *EmbeddedFont) (Ref, error) {
	fontDict := Dict{
		Name("Type"):     Name("Font"),
		Name("Subtype"):  Name("Type1"),
		Name("BaseFont"): Name("Helvetica"),
		Name("Encoding"): Name("WinAnsiEncoding"),
	}

	return fe.writer.addObject(fontDict), nil
}

// emitFontDescriptor creates and emits the FontDescriptor dictionary.
func (fe *FontEmitter) emitFontDescriptor(ef *EmbeddedFont, fontData []byte, isCFF bool) (Ref, error) {
	// Get font metrics
	metrics := fe.getFontMetrics(ef)

	// Embed font file
	fontFileRef, err := fe.emitFontFile(fontData, isCFF)
	if err != nil {
		return Ref{}, err
	}

	fontFileKey := Name("FontFile2") // TrueType
	if isCFF {
		fontFileKey = Name("FontFile3")
	}

	descriptorDict := Dict{
		Name("Type"):        Name("FontDescriptor"),
		Name("FontName"):    Name(fe.getBaseFont(ef)),
		Name("Flags"):       Int(metrics.Flags),
		Name("FontBBox"):    metrics.BBox,
		Name("ItalicAngle"): Real(metrics.ItalicAngle),
		Name("Ascent"):      Int(metrics.Ascent),
		Name("Descent"):     Int(metrics.Descent),
		Name("CapHeight"):   Int(metrics.CapHeight),
		Name("StemV"):       Int(metrics.StemV),
		fontFileKey:         fontFileRef,
	}

	return fe.writer.addObject(descriptorDict), nil
}

// emitFontFile embeds the font file data as a stream.
func (fe *FontEmitter) emitFontFile(fontData []byte, isCFF bool) (Ref, error) {
	// Compress the font data
	var compressed bytes.Buffer
	zw := zlib.NewWriter(&compressed)
	if _, err := zw.Write(fontData); err != nil {
		return Ref{}, err
	}
	if err := zw.Close(); err != nil {
		return Ref{}, err
	}

	streamDict := Dict{
		Name("Filter"): Name("FlateDecode"),
		Name("Length1"): Int(len(fontData)), // Original length
	}

	if isCFF {
		streamDict[Name("Subtype")] = Name("CIDFontType0C")
	}

	stream := Stream{
		Dict: streamDict,
		Data: compressed.Bytes(),
	}

	return fe.writer.addObject(stream), nil
}

// emitToUnicodeCMap creates the ToUnicode CMap for text extraction.
func (fe *FontEmitter) emitToUnicodeCMap(ef *EmbeddedFont) (Ref, error) {
	var buf bytes.Buffer
	fe.writeToUnicodeCMap(&buf, ef)

	// Compress the CMap
	var compressed bytes.Buffer
	zw := zlib.NewWriter(&compressed)
	if _, err := zw.Write(buf.Bytes()); err != nil {
		return Ref{}, err
	}
	if err := zw.Close(); err != nil {
		return Ref{}, err
	}

	stream := Stream{
		Dict: Dict{
			Name("Filter"): Name("FlateDecode"),
		},
		Data: compressed.Bytes(),
	}

	return fe.writer.addObject(stream), nil
}

// writeToUnicodeCMap writes the ToUnicode CMap data.
func (fe *FontEmitter) writeToUnicodeCMap(w io.Writer, ef *EmbeddedFont) {
	// CMap header
	fmt.Fprintf(w, "/CIDInit /ProcSet findresource begin\n")
	fmt.Fprintf(w, "12 dict begin\n")
	fmt.Fprintf(w, "begincmap\n")
	fmt.Fprintf(w, "/CIDSystemInfo <<\n")
	fmt.Fprintf(w, "  /Registry (Adobe)\n")
	fmt.Fprintf(w, "  /Ordering (UCS)\n")
	fmt.Fprintf(w, "  /Supplement 0\n")
	fmt.Fprintf(w, ">> def\n")
	fmt.Fprintf(w, "/CMapName /Adobe-Identity-UCS def\n")
	fmt.Fprintf(w, "/CMapType 2 def\n")

	// Codespace range
	fmt.Fprintf(w, "1 begincodespacerange\n")
	fmt.Fprintf(w, "<00> <FF>\n")
	fmt.Fprintf(w, "endcodespacerange\n")

	// Build sorted list of glyph mappings
	type mapping struct {
		code uint16
		char rune
	}
	var mappings []mapping
	for glyphID, char := range ef.UsedGlyphs {
		mappings = append(mappings, mapping{glyphID, char})
	}
	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].code < mappings[j].code
	})

	// Write character mappings in chunks of 100
	for i := 0; i < len(mappings); i += 100 {
		end := i + 100
		if end > len(mappings) {
			end = len(mappings)
		}
		chunk := mappings[i:end]

		fmt.Fprintf(w, "%d beginbfchar\n", len(chunk))
		for _, m := range chunk {
			// Map character code to Unicode value
			if m.char < 0x10000 {
				fmt.Fprintf(w, "<%02X> <%04X>\n", m.code&0xFF, m.char)
			} else {
				// Surrogate pair for characters outside BMP
				hi := 0xD800 + ((uint32(m.char)-0x10000)>>10)&0x3FF
				lo := 0xDC00 + (uint32(m.char) & 0x3FF)
				fmt.Fprintf(w, "<%02X> <%04X%04X>\n", m.code&0xFF, hi, lo)
			}
		}
		fmt.Fprintf(w, "endbfchar\n")
	}

	// CMap footer
	fmt.Fprintf(w, "endcmap\n")
	fmt.Fprintf(w, "CMapName currentdict /CMap defineresource pop\n")
	fmt.Fprintf(w, "end\n")
	fmt.Fprintf(w, "end\n")
}

// FontMetrics holds font metric data for the FontDescriptor.
type FontMetrics struct {
	Flags       int
	BBox        Array
	ItalicAngle float64
	Ascent      int
	Descent     int
	CapHeight   int
	StemV       int
}

// getFontMetrics extracts metrics from a font face.
func (fe *FontEmitter) getFontMetrics(ef *EmbeddedFont) FontMetrics {
	// Default metrics if we can't extract from font
	metrics := FontMetrics{
		Flags:       32, // Non-symbolic
		BBox:        Array{Int(-200), Int(-200), Int(1200), Int(900)},
		ItalicAngle: 0,
		Ascent:      800,
		Descent:     -200,
		CapHeight:   700,
		StemV:       80,
	}

	if ef.Face == nil || ef.Face.Font == nil {
		return metrics
	}

	face := ef.Face
	f := face.Font

	// Get units per em for scaling
	upem := float64(f.Upem())
	if upem == 0 {
		upem = 1000
	}

	// Scale factor to normalize to 1000 units
	scale := 1000.0 / upem

	// Font extents (horizontal metrics)
	if extents, ok := face.FontHExtents(); ok {
		metrics.Ascent = int(float64(extents.Ascender) * scale)
		metrics.Descent = int(float64(extents.Descender) * scale)
		// Estimate cap height from ascender
		metrics.CapHeight = int(float64(extents.Ascender) * scale * 0.7)
	}

	// Create bounding box (approximate)
	metrics.BBox = Array{
		Int(0),
		Int(metrics.Descent),
		Int(1000),
		Int(metrics.Ascent),
	}

	// Flags
	flags := 0
	desc := f.Describe()
	if desc.Aspect.Style != font.StyleNormal {
		flags |= 64 // Italic
		metrics.ItalicAngle = -12 // Common italic angle
	}
	if f.IsMonospace() {
		flags |= 1 // FixedPitch
	}
	flags |= 32 // Nonsymbolic (Latin text)
	metrics.Flags = flags

	// Estimate StemV from weight
	weight := int(desc.Aspect.Weight)
	if weight == 0 {
		weight = 400
	}
	metrics.StemV = weight / 5

	return metrics
}

// getBaseFont returns the PostScript name for the font.
func (fe *FontEmitter) getBaseFont(ef *EmbeddedFont) string {
	if ef.Face == nil || ef.Face.Font == nil {
		return "Helvetica"
	}

	desc := ef.Face.Font.Describe()
	if desc.Family != "" {
		// Sanitize name for PDF
		name := sanitizeFontName(desc.Family)
		return name
	}

	return "Helvetica"
}

// getWidthsArray creates the Widths array for the font.
func (fe *FontEmitter) getWidthsArray(ef *EmbeddedFont) Array {
	widths := make(Array, 256)

	// Default width
	defaultWidth := Int(600)

	for i := 0; i < 256; i++ {
		widths[i] = defaultWidth
	}

	if ef.Face == nil || ef.Face.Font == nil {
		return widths
	}

	face := ef.Face
	f := face.Font
	upem := float64(f.Upem())
	if upem == 0 {
		upem = 1000
	}
	scale := 1000.0 / upem

	// Get widths for each character code
	for i := 0; i < 256; i++ {
		r := rune(i)
		glyphID, ok := f.NominalGlyph(r)
		if ok {
			advance := face.HorizontalAdvance(glyphID)
			widths[i] = Int(float64(advance) * scale)
		}
	}

	return widths
}

// sanitizeFontName removes characters not allowed in PDF names.
func sanitizeFontName(name string) string {
	var result bytes.Buffer
	for _, c := range name {
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_' {
			result.WriteRune(c)
		}
	}
	if result.Len() == 0 {
		return "Font"
	}
	return result.String()
}
