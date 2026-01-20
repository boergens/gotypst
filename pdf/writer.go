package pdf

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// Writer writes PDF content with object management.
type Writer struct {
	buf       bytes.Buffer
	objects   []pdfObject
	xref      []int64 // Byte offsets of objects
	fontMap   *FontMap
	fontRefs  map[int]int // FontRef.ID -> PDF object number
	pageRefs  []int       // Object numbers of pages
	resources *Resources
}

// pdfObject represents a PDF object.
type pdfObject struct {
	num  int
	data []byte
}

// Resources tracks PDF resources for a page or document.
type Resources struct {
	Fonts    map[string]int // Font name -> object number
	XObjects map[string]int
	ExtGState map[string]int
}

// NewWriter creates a new PDF writer.
func NewWriter() *Writer {
	return &Writer{
		fontMap:   NewFontMap(),
		fontRefs:  make(map[int]int),
		resources: &Resources{
			Fonts:    make(map[string]int),
			XObjects: make(map[string]int),
			ExtGState: make(map[string]int),
		},
	}
}

// FontMap returns the writer's font map for registering glyphs.
func (w *Writer) FontMap() *FontMap {
	return w.fontMap
}

// allocObject allocates a new object number.
func (w *Writer) allocObject() int {
	num := len(w.objects) + 1
	w.objects = append(w.objects, pdfObject{num: num})
	return num
}

// writeObject writes a PDF object.
func (w *Writer) writeObject(num int, content string) {
	w.xref = append(w.xref, int64(w.buf.Len()))
	fmt.Fprintf(&w.buf, "%d 0 obj\n%s\nendobj\n\n", num, content)
}

// writeStreamObject writes a PDF stream object.
func (w *Writer) writeStreamObject(num int, dict string, data []byte, compress bool) {
	w.xref = append(w.xref, int64(w.buf.Len()))

	streamData := data
	filter := ""
	if compress && len(data) > 100 {
		var compressed bytes.Buffer
		zw := zlib.NewWriter(&compressed)
		zw.Write(data)
		zw.Close()
		if compressed.Len() < len(data) {
			streamData = compressed.Bytes()
			filter = "/Filter /FlateDecode\n"
		}
	}

	fmt.Fprintf(&w.buf, "%d 0 obj\n<<\n%s%s/Length %d\n>>\nstream\n",
		num, dict, filter, len(streamData))
	w.buf.Write(streamData)
	fmt.Fprintf(&w.buf, "\nendstream\nendobj\n\n")
}

// WriteFonts writes all font resources to the PDF.
func (w *Writer) WriteFonts() error {
	for _, usage := range w.fontMap.AllFonts() {
		if err := w.writeFont(usage); err != nil {
			return err
		}
	}
	return nil
}

// writeFont writes a single font to the PDF.
func (w *Writer) writeFont(usage *FontUsage) error {
	subsetter := NewFontSubsetter(usage)
	embedding, err := subsetter.Subset()
	if err != nil {
		return fmt.Errorf("failed to subset font %d: %w", usage.Ref.ID, err)
	}

	if embedding.IsCIDFont {
		return w.writeCIDFont(embedding)
	}
	return w.writeSimpleFont(embedding)
}

// writeSimpleFont writes a simple (non-CID) TrueType font.
func (w *Writer) writeSimpleFont(emb *FontEmbedding) error {
	// Allocate object numbers
	fontObjNum := w.allocObject()
	descriptorNum := w.allocObject()
	streamNum := w.allocObject()
	toUnicodeNum := w.allocObject()

	// Store font reference
	fontName := fmt.Sprintf("F%d", emb.Ref.ID)
	w.fontRefs[emb.Ref.ID] = fontObjNum
	w.resources.Fonts[fontName] = fontObjNum

	// Build widths array
	widths := buildWidthsArray(emb.Widths)

	// Font dictionary
	fontDict := fmt.Sprintf(`<<
/Type /Font
/Subtype /TrueType
/BaseFont /%s
/FirstChar %d
/LastChar %d
/Widths %s
/FontDescriptor %d 0 R
/ToUnicode %d 0 R
/Encoding /WinAnsiEncoding
>>`, pdfName(emb.PostScriptName), 0, 255, widths, descriptorNum, toUnicodeNum)
	w.writeObject(fontObjNum, fontDict)

	// Font descriptor
	descriptor := fmt.Sprintf(`<<
/Type /FontDescriptor
/FontName /%s
/Flags 32
/FontBBox [-500 -300 1500 1000]
/ItalicAngle 0
/Ascent 800
/Descent -200
/CapHeight 700
/StemV 80
/FontFile2 %d 0 R
>>`, pdfName(emb.PostScriptName), streamNum)
	w.writeObject(descriptorNum, descriptor)

	// Font stream
	w.writeStreamObject(streamNum, "/Subtype /TrueType\n", emb.Subset, true)

	// ToUnicode CMap
	toUnicode := buildToUnicodeCMap(emb.ToUnicode)
	w.writeStreamObject(toUnicodeNum, "", []byte(toUnicode), true)

	return nil
}

// writeCIDFont writes a CID font (for large character sets).
func (w *Writer) writeCIDFont(emb *FontEmbedding) error {
	// Allocate object numbers
	fontObjNum := w.allocObject()
	cidFontNum := w.allocObject()
	descriptorNum := w.allocObject()
	streamNum := w.allocObject()
	toUnicodeNum := w.allocObject()
	widthsNum := w.allocObject()

	// Store font reference
	fontName := fmt.Sprintf("F%d", emb.Ref.ID)
	w.fontRefs[emb.Ref.ID] = fontObjNum
	w.resources.Fonts[fontName] = fontObjNum

	// Type 0 font dictionary (top level)
	fontDict := fmt.Sprintf(`<<
/Type /Font
/Subtype /Type0
/BaseFont /%s
/Encoding /Identity-H
/DescendantFonts [%d 0 R]
/ToUnicode %d 0 R
>>`, pdfName(emb.PostScriptName), cidFontNum, toUnicodeNum)
	w.writeObject(fontObjNum, fontDict)

	// CIDFont dictionary
	cidFont := fmt.Sprintf(`<<
/Type /Font
/Subtype /CIDFontType2
/BaseFont /%s
/CIDSystemInfo <<
  /Registry (Adobe)
  /Ordering (Identity)
  /Supplement 0
>>
/FontDescriptor %d 0 R
/W %d 0 R
/CIDToGIDMap /Identity
>>`, pdfName(emb.PostScriptName), descriptorNum, widthsNum)
	w.writeObject(cidFontNum, cidFont)

	// Font descriptor
	descriptor := fmt.Sprintf(`<<
/Type /FontDescriptor
/FontName /%s
/Flags 4
/FontBBox [-500 -300 1500 1000]
/ItalicAngle 0
/Ascent 880
/Descent -120
/CapHeight 700
/StemV 80
/FontFile2 %d 0 R
>>`, pdfName(emb.PostScriptName), streamNum)
	w.writeObject(descriptorNum, descriptor)

	// CID widths array
	cidWidths := buildCIDWidths(emb.Widths)
	w.writeObject(widthsNum, cidWidths)

	// Font stream
	w.writeStreamObject(streamNum,
		fmt.Sprintf("/Subtype /CIDFontType2\n/Length1 %d\n", len(emb.Subset)),
		emb.Subset, true)

	// ToUnicode CMap
	toUnicode := buildCIDToUnicodeCMap(emb.ToUnicode)
	w.writeStreamObject(toUnicodeNum, "", []byte(toUnicode), true)

	return nil
}

// GetFontRef returns the PDF reference string for a font.
func (w *Writer) GetFontRef(fontID int) string {
	if objNum, ok := w.fontRefs[fontID]; ok {
		return fmt.Sprintf("%d 0 R", objNum)
	}
	return ""
}

// GetFontName returns the resource name for a font.
func (w *Writer) GetFontName(fontID int) string {
	return fmt.Sprintf("F%d", fontID)
}

// WriteResources writes the resources dictionary.
func (w *Writer) WriteResources() string {
	var sb strings.Builder
	sb.WriteString("<<\n")

	// Fonts
	if len(w.resources.Fonts) > 0 {
		sb.WriteString("/Font <<\n")
		names := make([]string, 0, len(w.resources.Fonts))
		for name := range w.resources.Fonts {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			fmt.Fprintf(&sb, "/%s %d 0 R\n", name, w.resources.Fonts[name])
		}
		sb.WriteString(">>\n")
	}

	// XObjects
	if len(w.resources.XObjects) > 0 {
		sb.WriteString("/XObject <<\n")
		for name, objNum := range w.resources.XObjects {
			fmt.Fprintf(&sb, "/%s %d 0 R\n", name, objNum)
		}
		sb.WriteString(">>\n")
	}

	// ExtGState
	if len(w.resources.ExtGState) > 0 {
		sb.WriteString("/ExtGState <<\n")
		for name, objNum := range w.resources.ExtGState {
			fmt.Fprintf(&sb, "/%s %d 0 R\n", name, objNum)
		}
		sb.WriteString(">>\n")
	}

	sb.WriteString("/ProcSet [/PDF /Text /ImageB /ImageC /ImageI]\n")
	sb.WriteString(">>")
	return sb.String()
}

// Finish completes the PDF and returns the bytes.
func (w *Writer) Finish() ([]byte, error) {
	// Write header
	var final bytes.Buffer
	final.WriteString("%PDF-1.7\n")
	final.WriteString("%\xE2\xE3\xCF\xD3\n\n") // Binary marker

	// Write all buffered content
	final.Write(w.buf.Bytes())

	// Write xref table
	xrefOffset := final.Len()
	fmt.Fprintf(&final, "xref\n0 %d\n", len(w.xref)+1)
	fmt.Fprintf(&final, "0000000000 65535 f \n")
	for _, offset := range w.xref {
		fmt.Fprintf(&final, "%010d 00000 n \n", offset+int64(len("%PDF-1.7\n%\xE2\xE3\xCF\xD3\n\n")))
	}

	// Write trailer
	fmt.Fprintf(&final, "trailer\n<<\n/Size %d\n>>\n", len(w.xref)+1)
	fmt.Fprintf(&final, "startxref\n%d\n%%%%EOF\n", xrefOffset)

	return final.Bytes(), nil
}

// Helper functions for PDF generation.

// pdfName escapes a string for use as a PDF name.
func pdfName(s string) string {
	var sb strings.Builder
	for _, c := range s {
		if c == ' ' || c == '/' || c == '#' || c == '(' || c == ')' ||
			c == '<' || c == '>' || c == '[' || c == ']' || c == '{' || c == '}' ||
			c < 33 || c > 126 {
			fmt.Fprintf(&sb, "#%02X", c)
		} else {
			sb.WriteRune(c)
		}
	}
	return sb.String()
}

// pdfString escapes a string for PDF string literal.
func pdfString(s string) string {
	var sb strings.Builder
	sb.WriteByte('(')
	for _, c := range s {
		switch c {
		case '\\', '(', ')':
			sb.WriteByte('\\')
			sb.WriteRune(c)
		case '\n':
			sb.WriteString("\\n")
		case '\r':
			sb.WriteString("\\r")
		case '\t':
			sb.WriteString("\\t")
		default:
			if c < 32 || c > 126 {
				fmt.Fprintf(&sb, "\\%03o", c)
			} else {
				sb.WriteRune(c)
			}
		}
	}
	sb.WriteByte(')')
	return sb.String()
}

// pdfHexString creates a hex string.
func pdfHexString(s string) string {
	var sb strings.Builder
	sb.WriteByte('<')
	for _, b := range []byte(s) {
		fmt.Fprintf(&sb, "%02X", b)
	}
	sb.WriteByte('>')
	return sb.String()
}

// buildWidthsArray builds a PDF widths array for a simple font.
func buildWidthsArray(widths map[uint16]int) string {
	// For simple fonts, widths array covers FirstChar to LastChar
	arr := make([]int, 256)
	for gid, w := range widths {
		if int(gid) < 256 {
			arr[gid] = w
		}
	}

	var sb strings.Builder
	sb.WriteByte('[')
	for i, w := range arr {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(strconv.Itoa(w))
	}
	sb.WriteByte(']')
	return sb.String()
}

// buildCIDWidths builds a CID widths array.
func buildCIDWidths(widths map[uint16]int) string {
	// Format: [CID [w1 w2 ...]] or [CIDfirst CIDlast w]
	if len(widths) == 0 {
		return "[]"
	}

	// Sort glyph IDs
	gids := make([]uint16, 0, len(widths))
	for gid := range widths {
		gids = append(gids, gid)
	}
	sort.Slice(gids, func(i, j int) bool { return gids[i] < gids[j] })

	var sb strings.Builder
	sb.WriteByte('[')

	i := 0
	for i < len(gids) {
		// Try to find a run of consecutive CIDs with same width
		start := gids[i]
		w := widths[start]

		// Check for consecutive CIDs with same width
		end := start
		for j := i + 1; j < len(gids); j++ {
			if gids[j] == end+1 && widths[gids[j]] == w {
				end = gids[j]
			} else {
				break
			}
		}

		if end > start {
			// Range format: CIDfirst CIDlast width
			fmt.Fprintf(&sb, " %d %d %d", start, end, w)
			i += int(end - start + 1)
		} else {
			// Array format: CID [w1 w2 ...]
			fmt.Fprintf(&sb, " %d [%d", start, w)
			i++
			// Add consecutive CIDs (possibly with different widths)
			for i < len(gids) && gids[i] == gids[i-1]+1 {
				fmt.Fprintf(&sb, " %d", widths[gids[i]])
				i++
			}
			sb.WriteByte(']')
		}
	}

	sb.WriteByte(']')
	return sb.String()
}

// buildToUnicodeCMap builds a ToUnicode CMap for simple fonts.
func buildToUnicodeCMap(mapping map[uint16][]rune) string {
	var sb strings.Builder

	sb.WriteString("/CIDInit /ProcSet findresource begin\n")
	sb.WriteString("12 dict begin\n")
	sb.WriteString("begincmap\n")
	sb.WriteString("/CIDSystemInfo <<\n")
	sb.WriteString("  /Registry (Adobe)\n")
	sb.WriteString("  /Ordering (UCS)\n")
	sb.WriteString("  /Supplement 0\n")
	sb.WriteString(">> def\n")
	sb.WriteString("/CMapName /Adobe-Identity-UCS def\n")
	sb.WriteString("/CMapType 2 def\n")
	sb.WriteString("1 begincodespacerange\n")
	sb.WriteString("<00> <FF>\n")
	sb.WriteString("endcodespacerange\n")

	// Sort for deterministic output
	gids := make([]uint16, 0, len(mapping))
	for gid := range mapping {
		if gid < 256 {
			gids = append(gids, gid)
		}
	}
	sort.Slice(gids, func(i, j int) bool { return gids[i] < gids[j] })

	if len(gids) > 0 {
		// Write in chunks of 100 (PDF limit)
		for i := 0; i < len(gids); i += 100 {
			end := i + 100
			if end > len(gids) {
				end = len(gids)
			}
			chunk := gids[i:end]

			fmt.Fprintf(&sb, "%d beginbfchar\n", len(chunk))
			for _, gid := range chunk {
				chars := mapping[gid]
				if len(chars) > 0 {
					fmt.Fprintf(&sb, "<%02X> <%s>\n", gid, runeToUTF16BE(chars[0]))
				}
			}
			sb.WriteString("endbfchar\n")
		}
	}

	sb.WriteString("endcmap\n")
	sb.WriteString("CMapName currentdict /CMap defineresource pop\n")
	sb.WriteString("end\n")
	sb.WriteString("end\n")

	return sb.String()
}

// buildCIDToUnicodeCMap builds a ToUnicode CMap for CID fonts.
func buildCIDToUnicodeCMap(mapping map[uint16][]rune) string {
	var sb strings.Builder

	sb.WriteString("/CIDInit /ProcSet findresource begin\n")
	sb.WriteString("12 dict begin\n")
	sb.WriteString("begincmap\n")
	sb.WriteString("/CIDSystemInfo <<\n")
	sb.WriteString("  /Registry (Adobe)\n")
	sb.WriteString("  /Ordering (UCS)\n")
	sb.WriteString("  /Supplement 0\n")
	sb.WriteString(">> def\n")
	sb.WriteString("/CMapName /Adobe-Identity-UCS def\n")
	sb.WriteString("/CMapType 2 def\n")
	sb.WriteString("1 begincodespacerange\n")
	sb.WriteString("<0000> <FFFF>\n")
	sb.WriteString("endcodespacerange\n")

	// Sort for deterministic output
	gids := make([]uint16, 0, len(mapping))
	for gid := range mapping {
		gids = append(gids, gid)
	}
	sort.Slice(gids, func(i, j int) bool { return gids[i] < gids[j] })

	if len(gids) > 0 {
		// Write in chunks of 100 (PDF limit)
		for i := 0; i < len(gids); i += 100 {
			end := i + 100
			if end > len(gids) {
				end = len(gids)
			}
			chunk := gids[i:end]

			fmt.Fprintf(&sb, "%d beginbfchar\n", len(chunk))
			for _, gid := range chunk {
				chars := mapping[gid]
				if len(chars) > 0 {
					// For CID fonts, use 4-digit hex for glyph ID
					fmt.Fprintf(&sb, "<%04X> <%s>\n", gid, runeToUTF16BE(chars[0]))
				}
			}
			sb.WriteString("endbfchar\n")
		}
	}

	sb.WriteString("endcmap\n")
	sb.WriteString("CMapName currentdict /CMap defineresource pop\n")
	sb.WriteString("end\n")
	sb.WriteString("end\n")

	return sb.String()
}

// runeToUTF16BE converts a rune to UTF-16BE hex string.
func runeToUTF16BE(r rune) string {
	if r <= 0xFFFF {
		return fmt.Sprintf("%04X", r)
	}
	// Surrogate pair for characters outside BMP
	r -= 0x10000
	high := 0xD800 + (r >> 10)
	low := 0xDC00 + (r & 0x3FF)
	return fmt.Sprintf("%04X%04X", high, low)
}

// CompressStream compresses data using zlib.
func CompressStream(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecompressStream decompresses zlib data.
func DecompressStream(data []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}
